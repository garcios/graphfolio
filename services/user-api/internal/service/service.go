package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"user-api/internal/auth"
	"user-api/internal/crypto"
	"user-api/internal/domain"
	"user-api/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID            = errors.New("invalid user ID")
	ErrUserNotFound             = repository.ErrUserNotFound
	ErrEmailAlreadyRegistered   = repository.ErrEmailAlreadyExists
	ErrInvalidCurrency          = errors.New("unsupported currency code")
	ErrInvalidDisplayName       = errors.New("display name must be between 1 and 100 characters")
	ErrInvalidTheme             = errors.New("theme must be DARK, LIGHT, or SYSTEM")
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrAccountInactive          = errors.New("account is inactive")
	ErrInvalidEmail             = errors.New("invalid email address")
	ErrInvalidPassword          = errors.New("password must be between 8 and 128 characters and include uppercase, lowercase, digit, and symbol")
	ErrInvalidSession           = errors.New("invalid or expired session")
	ErrInvalidCurrentPassword   = errors.New("invalid current password")
	ErrSamePassword             = errors.New("new password cannot be the same as current password")
	ErrPasswordComplexityFailed = errors.New("new password does not meet complexity requirements")
	ErrInvalidFinancialYearEnd  = errors.New("invalid financial year end date")
)

const (
	AccessTokenDuration     = 15 * time.Minute
	RefreshTokenDuration    = 7 * 24 * time.Hour
	RememberMeTokenDuration = 30 * 24 * time.Hour
	DefaultDevJWTSecret     = "graphfolio-default-dev-secret-key-256bit-ok!"
)

// UpdatePreferencesInput represents parameters for updating user preferences.
type UpdatePreferencesInput struct {
	UserID                string
	DisplayName           *string
	DisplayCurrency       *string
	Theme                 *string
	FinancialYearEndMonth *int
	FinancialYearEndDay   *int
}

// AuthResult represents the payload returned upon successful registration, login, or token refresh.
type AuthResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // In seconds
}

// UserService defines the business logic and validations for user identity, authentication, and preferences.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_service.go -package=mocks user-api/internal/service UserService
type UserService interface {
	Register(ctx context.Context, email, password, displayName, currency string) (*AuthResult, error)
	Login(ctx context.Context, email, password string, rememberMe bool) (*AuthResult, error)
	RefreshToken(ctx context.Context, rawRefreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	GetMe(ctx context.Context, userID string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string, revokeOtherSessions bool, currentRefreshToken string) error

	GetUserPreferences(ctx context.Context, userID string) (*domain.User, error)
	UpdateUserPreferences(ctx context.Context, input UpdatePreferencesInput) (*domain.User, error)
	ListSupportedCurrencies(ctx context.Context) ([]domain.CurrencyInfo, error)
}

type userService struct {
	repo      repository.Repository
	jwtSecret []byte
}

// NewUserService constructs a new UserService with the provided repository and secret.
func NewUserService(repo repository.Repository, jwtSecret ...[]byte) UserService {
	var secret []byte
	if len(jwtSecret) > 0 && len(jwtSecret[0]) > 0 {
		secret = jwtSecret[0]
	} else if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		secret = []byte(envSecret)
	} else {
		secret = []byte(DefaultDevJWTSecret)
	}

	return &userService{
		repo:      repo,
		jwtSecret: secret,
	}
}

// DefaultDemoUserID is the fixed UUID for single-user/demo sessions.
const DefaultDemoUserID = "018f0000-0000-7000-8000-000000000001"

func parseUserID(raw string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "1" {
		return uuid.MustParse(DefaultDemoUserID), nil
	}
	uid, err := uuid.Parse(trimmed)
	if err != nil {
		return uuid.Nil, ErrInvalidUserID
	}
	return uid, nil
}

func validateEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if len(trimmed) < 3 || !strings.Contains(trimmed, "@") || !strings.Contains(trimmed, ".") {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return ErrInvalidPassword
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~", ch):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrInvalidPassword
	}
	return nil
}

func (s *userService) Register(ctx context.Context, email, password, displayName, currency string) (*AuthResult, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	trimmedName := strings.TrimSpace(displayName)
	if len(trimmedName) < 1 || len(trimmedName) > 100 {
		return nil, ErrInvalidDisplayName
	}

	curr := strings.ToUpper(strings.TrimSpace(currency))
	if !domain.IsSupportedCurrency(curr) {
		return nil, ErrInvalidCurrency
	}

	hashedPassword, err := crypto.HashPassword(password, crypto.DefaultParams)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, email, hashedPassword, trimmedName, curr)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, RefreshTokenDuration)
}

func (s *userService) Login(ctx context.Context, email, password string, rememberMe bool) (*AuthResult, error) {
	if err := validateEmail(email); err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Timing-attack mitigation: compute dummy verification
			_, _ = crypto.VerifyPassword(password, crypto.DummyHash)
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	match, err := crypto.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	duration := RefreshTokenDuration
	if rememberMe {
		duration = RememberMeTokenDuration
	}

	return s.issueTokens(ctx, user, duration)
}

func (s *userService) RefreshToken(ctx context.Context, rawRefreshToken string) (*AuthResult, error) {
	rawToken := strings.TrimSpace(rawRefreshToken)
	if rawToken == "" {
		return nil, ErrInvalidSession
	}

	hashHex := auth.HashRefreshToken(rawToken)
	session, err := s.repo.GetSessionByHash(ctx, hashHex)
	if err != nil {
		return nil, ErrInvalidSession
	}

	// Reuse detection: If a revoked token is presented, revoke entire token family
	if session.IsRevoked {
		_ = s.repo.RevokeSessionFamily(ctx, session.FamilyID)
		return nil, ErrInvalidSession
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.repo.RevokeSession(ctx, session.ID)
		return nil, ErrInvalidSession
	}

	// Invalidate current session (token rotation)
	_ = s.repo.RevokeSession(ctx, session.ID)

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil || !user.IsActive {
		return nil, ErrInvalidSession
	}

	// Issue new token pair under the same family
	newRawToken, newHashHex, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newSession := &domain.Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		RefreshTokenHash: newHashHex,
		FamilyID:         session.FamilyID,
		IsRevoked:        false,
		ExpiresAt:        time.Now().Add(RefreshTokenDuration),
	}
	if err := s.repo.CreateSession(ctx, newSession); err != nil {
		return nil, err
	}

	now := time.Now()
	accessToken, err := auth.GenerateAccessToken(s.jwtSecret, auth.Claims{
		Subject:      user.ID.String(),
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		BaseCurrency: user.DisplayCurrency,
		Role:         user.Role,
		IssuedAt:     now.Unix(),
		ExpiresAt:    now.Add(AccessTokenDuration).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRawToken,
		ExpiresIn:    int64(AccessTokenDuration / time.Second),
	}, nil
}

func (s *userService) Logout(ctx context.Context, rawRefreshToken string) error {
	rawToken := strings.TrimSpace(rawRefreshToken)
	if rawToken == "" {
		return nil
	}

	hashHex := auth.HashRefreshToken(rawToken)
	session, err := s.repo.GetSessionByHash(ctx, hashHex)
	if err != nil {
		return nil // Idempotent
	}

	return s.repo.RevokeSession(ctx, session.ID)
}

func (s *userService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	return s.GetUserPreferences(ctx, userID)
}

func (s *userService) ChangePassword(
	ctx context.Context,
	userID string,
	currentPassword string,
	newPassword string,
	revokeOtherSessions bool,
	currentRefreshToken string,
) error {
	uid, err := parseUserID(userID)
	if err != nil {
		return ErrInvalidUserID
	}

	user, err := s.repo.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Timing-attack mitigation: compute dummy verification
			_, _ = crypto.VerifyPassword(currentPassword, crypto.DummyHash)
			return ErrInvalidCredentials
		}
		return err
	}

	if !user.IsActive {
		return ErrAccountInactive
	}

	match, err := crypto.VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil || !match {
		return ErrInvalidCurrentPassword
	}

	if currentPassword == newPassword {
		return ErrSamePassword
	}

	if err := validatePassword(newPassword); err != nil {
		return ErrPasswordComplexityFailed
	}

	newHashedPassword, err := crypto.HashPassword(newPassword, crypto.DefaultParams)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, uid, newHashedPassword); err != nil {
		return err
	}

	if revokeOtherSessions {
		var currentHash string
		trimmedToken := strings.TrimSpace(currentRefreshToken)
		if trimmedToken != "" {
			currentHash = auth.HashRefreshToken(trimmedToken)
		}
		_ = s.repo.RevokeOtherUserSessions(ctx, uid, currentHash)
	}

	return nil
}


func (s *userService) issueTokens(ctx context.Context, user *domain.User, refreshDuration time.Duration) (*AuthResult, error) {
	rawRefreshToken, hashHex, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	familyID := uuid.New()
	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		RefreshTokenHash: hashHex,
		FamilyID:         familyID,
		IsRevoked:        false,
		ExpiresAt:        time.Now().Add(refreshDuration),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	now := time.Now()
	accessToken, err := auth.GenerateAccessToken(s.jwtSecret, auth.Claims{
		Subject:      user.ID.String(),
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		BaseCurrency: user.DisplayCurrency,
		Role:         user.Role,
		IssuedAt:     now.Unix(),
		ExpiresAt:    now.Add(AccessTokenDuration).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    int64(AccessTokenDuration / time.Second),
	}, nil
}

func (s *userService) GetUserPreferences(ctx context.Context, userID string) (*domain.User, error) {
	uid, err := parseUserID(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	return s.repo.GetUserByID(ctx, uid)
}

func (s *userService) UpdateUserPreferences(ctx context.Context, input UpdatePreferencesInput) (*domain.User, error) {
	uid, err := parseUserID(input.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var displayName *string
	if input.DisplayName != nil {
		name := strings.TrimSpace(*input.DisplayName)
		if len(name) < 1 || len(name) > 100 {
			return nil, ErrInvalidDisplayName
		}
		displayName = &name
	}

	var currency *string
	if input.DisplayCurrency != nil {
		curr := strings.ToUpper(strings.TrimSpace(*input.DisplayCurrency))
		if !domain.IsSupportedCurrency(curr) {
			return nil, ErrInvalidCurrency
		}
		currency = &curr
	}

	var theme *string
	if input.Theme != nil {
		th := strings.ToUpper(strings.TrimSpace(*input.Theme))
		if !domain.IsValidTheme(th) {
			return nil, ErrInvalidTheme
		}
		theme = &th
	}

	var fyMonth *int
	var fyDay *int
	if input.FinancialYearEndMonth != nil || input.FinancialYearEndDay != nil {
		m := 0
		d := 0
		if input.FinancialYearEndMonth != nil && input.FinancialYearEndDay != nil {
			m = *input.FinancialYearEndMonth
			d = *input.FinancialYearEndDay
		} else {
			current, err := s.repo.GetUserByID(ctx, uid)
			if err != nil {
				return nil, err
			}
			m = current.FinancialYearEndMonth
			d = current.FinancialYearEndDay
			if input.FinancialYearEndMonth != nil {
				m = *input.FinancialYearEndMonth
			}
			if input.FinancialYearEndDay != nil {
				d = *input.FinancialYearEndDay
			}
		}
		if !domain.IsValidFinancialYearEnd(m, d) {
			return nil, ErrInvalidFinancialYearEnd
		}
		fyMonth = input.FinancialYearEndMonth
		fyDay = input.FinancialYearEndDay
	}

	return s.repo.UpdateUserPreferences(ctx, uid, displayName, currency, theme, fyMonth, fyDay)
}

func (s *userService) ListSupportedCurrencies(ctx context.Context) ([]domain.CurrencyInfo, error) {
	return domain.SupportedCurrencies, nil
}
