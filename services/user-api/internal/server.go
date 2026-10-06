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

	input := service.UpdatePreferencesInput{
		UserID:          req.GetUserId(),
		DisplayName:     req.DisplayName,
		DisplayCurrency: req.DisplayCurrency,
		Theme:           req.Theme,
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
		errors.Is(err, service.ErrInvalidTheme):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func domainUserToProto(u *domain.User) *userpb.UserPreferences {
	if u == nil {
		return nil
	}
	return &userpb.UserPreferences{
		UserId:          u.ID.String(),
		Email:           u.Email,
		DisplayName:     u.DisplayName,
		DisplayCurrency: u.DisplayCurrency,
		Theme:           u.Theme,
		CreatedAt:       u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       u.UpdatedAt.Format(time.RFC3339),
	}
}
