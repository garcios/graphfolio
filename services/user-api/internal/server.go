package internal

import (
	"context"
	"errors"
	"time"

	"user-api/internal/domain"
	"user-api/internal/service"

	userpb "graphfolio/proto/user/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServer implements the gRPC UserServiceServer contract.
type UserServer struct {
	userpb.UnimplementedUserServiceServer
	svc service.UserService
}

// NewUserServer creates a new instance of UserServer.
func NewUserServer(svc service.UserService) *UserServer {
	return &UserServer{svc: svc}
}

// Register registers a new user with credentials and issues initial session tokens.
func (s *UserServer) Register(
	ctx context.Context,
	req *userpb.RegisterRequest,
) (*userpb.RegisterResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	result, err := s.svc.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetDisplayName(), req.GetDisplayCurrency())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.RegisterResponse{
		User:         domainUserToProto(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

// Login authenticates an investor with credentials and issues session tokens.
func (s *UserServer) Login(
	ctx context.Context,
	req *userpb.LoginRequest,
) (*userpb.LoginResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	result, err := s.svc.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetRememberMe())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.LoginResponse{
		User:         domainUserToProto(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

// RefreshToken renews an access token using a valid refresh token.
func (s *UserServer) RefreshToken(
	ctx context.Context,
	req *userpb.RefreshTokenRequest,
) (*userpb.RefreshTokenResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	result, err := s.svc.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User:         domainUserToProto(result.User),
	}, nil
}

// Logout revokes an active session refresh token.
func (s *UserServer) Logout(
	ctx context.Context,
	req *userpb.LogoutRequest,
) (*userpb.LogoutResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	err := s.svc.Logout(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.LogoutResponse{
		Success: true,
	}, nil
}

// GetMe retrieves profile and preferences for the authenticated user.
func (s *UserServer) GetMe(
	ctx context.Context,
	req *userpb.GetMeRequest,
) (*userpb.GetMeResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.svc.GetMe(ctx, req.GetUserId())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.GetMeResponse{
		User: domainUserToProto(user),
	}, nil
}

// ChangePassword updates an authenticated user's password and optionally revokes other sessions.
func (s *UserServer) ChangePassword(
	ctx context.Context,
	req *userpb.ChangePasswordRequest,
) (*userpb.ChangePasswordResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	err := s.svc.ChangePassword(
		ctx,
		req.GetUserId(),
		req.GetCurrentPassword(),
		req.GetNewPassword(),
		req.GetRevokeOtherSessions(),
		req.GetCurrentRefreshToken(),
	)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.ChangePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

// GetUserPreferences retrieves preferences for a specified user.
func (s *UserServer) GetUserPreferences(
	ctx context.Context,
	req *userpb.GetUserPreferencesRequest,
) (*userpb.GetUserPreferencesResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.svc.GetUserPreferences(ctx, req.GetUserId())
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.GetUserPreferencesResponse{
		Preferences: domainUserToProto(user),
	}, nil
}

// UpdateUserPreferences updates optional preferences for a specified user.
func (s *UserServer) UpdateUserPreferences(
	ctx context.Context,
	req *userpb.UpdateUserPreferencesRequest,
) (*userpb.UpdateUserPreferencesResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var fyMonth *int
	if req.FinancialYearEndMonth != nil {
		m := int(*req.FinancialYearEndMonth)
		fyMonth = &m
	}
	var fyDay *int
	if req.FinancialYearEndDay != nil {
		d := int(*req.FinancialYearEndDay)
		fyDay = &d
	}

	input := service.UpdatePreferencesInput{
		UserID:                req.GetUserId(),
		DisplayName:           req.DisplayName,
		DisplayCurrency:       req.DisplayCurrency,
		Theme:                 req.Theme,
		FinancialYearEndMonth: fyMonth,
		FinancialYearEndDay:   fyDay,
	}

	user, err := s.svc.UpdateUserPreferences(ctx, input)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &userpb.UpdateUserPreferencesResponse{
		Preferences: domainUserToProto(user),
	}, nil
}

// ListSupportedCurrencies returns the list of supported display currencies.
func (s *UserServer) ListSupportedCurrencies(
	ctx context.Context,
	_ *userpb.ListSupportedCurrenciesRequest,
) (*userpb.ListSupportedCurrenciesResponse, error) {
	if s.svc == nil {
		return nil, status.Error(codes.Unavailable, "user service unavailable")
	}
	currencies, err := s.svc.ListSupportedCurrencies(ctx)
	if err != nil {
		return nil, mapServiceError(err)
	}

	protoCurrencies := make([]*userpb.CurrencyInfo, len(currencies))
	for i, c := range currencies {
		protoCurrencies[i] = &userpb.CurrencyInfo{
			Code:   c.Code,
			Name:   c.Name,
			Symbol: c.Symbol,
		}
	}

	return &userpb.ListSupportedCurrenciesResponse{
		Currencies: protoCurrencies,
	}, nil
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidDisplayName),
		errors.Is(err, service.ErrInvalidCurrency),
		errors.Is(err, service.ErrInvalidTheme),
		errors.Is(err, service.ErrInvalidEmail),
		errors.Is(err, service.ErrInvalidPassword),
		errors.Is(err, service.ErrSamePassword),
		errors.Is(err, service.ErrPasswordComplexityFailed),
		errors.Is(err, service.ErrInvalidFinancialYearEnd):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrEmailAlreadyRegistered):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrInvalidCurrentPassword),
		errors.Is(err, service.ErrInvalidSession):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, service.ErrAccountInactive):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProtoRole(role string) userpb.UserRole {
	if role == domain.RoleAdmin {
		return userpb.UserRole_USER_ROLE_ADMIN
	}
	return userpb.UserRole_USER_ROLE_USER
}

func domainUserToProto(u *domain.User) *userpb.UserPreferences {
	if u == nil {
		return nil
	}
	return &userpb.UserPreferences{
		UserId:             u.ID.String(),
		Email:              u.Email,
		DisplayName:        u.DisplayName,
		DisplayCurrency:    u.DisplayCurrency,
		Theme:              u.Theme,
		Role:                  toProtoRole(u.Role),
		MustChangePassword:    u.MustChangePassword,
		FinancialYearEndMonth: int32(u.FinancialYearEndMonth),
		FinancialYearEndDay:   int32(u.FinancialYearEndDay),
		CreatedAt:             u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             u.UpdatedAt.Format(time.RFC3339),
	}
}
