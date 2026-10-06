package service

import (
	"context"
	"errors"
	"strings"

	"user-api/internal/domain"
	"user-api/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrUserNotFound       = repository.ErrUserNotFound
	ErrInvalidCurrency    = errors.New("unsupported currency code")
	ErrInvalidDisplayName = errors.New("display name must be between 1 and 100 characters")
	ErrInvalidTheme       = errors.New("theme must be DARK, LIGHT, or SYSTEM")
)

// UpdatePreferencesInput represents parameters for updating user preferences.
type UpdatePreferencesInput struct {
	UserID          string
	DisplayName     *string
	DisplayCurrency *string
	Theme           *string
}

// UserService defines the business logic and validations for user identity and preferences.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_service.go -package=mocks user-api/internal/service UserService
type UserService interface {
	GetUserPreferences(ctx context.Context, userID string) (*domain.User, error)
	UpdateUserPreferences(ctx context.Context, input UpdatePreferencesInput) (*domain.User, error)
	ListSupportedCurrencies(ctx context.Context) ([]domain.CurrencyInfo, error)
}

type userService struct {
	repo repository.Repository
}

// NewUserService constructs a new UserService with the provided repository.
func NewUserService(repo repository.Repository) UserService {
	return &userService{repo: repo}
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

	return s.repo.UpdateUserPreferences(ctx, uid, displayName, currency, theme)
}

func (s *userService) ListSupportedCurrencies(ctx context.Context) ([]domain.CurrencyInfo, error) {
	return domain.SupportedCurrencies, nil
}
