package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"user-api/internal/domain"
	"user-api/internal/service"
	mocks "user-api/internal/service/mocks"

	userpb "graphfolio/proto/user/v1"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserServer_GetUserPreferences(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()
	now := time.Now().UTC()
	mockUser := &domain.User{
		ID:              validUUID,
		Email:           "oscar@example.com",
		DisplayName:     "Oscar Garcia",
		DisplayCurrency: "USD",
		Theme:           "DARK",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			GetUserPreferences(ctx, validUUID.String()).
			Return(mockUser, nil)

		server := NewUserServer(mockSvc)
		resp, err := server.GetUserPreferences(ctx, &userpb.GetUserPreferencesRequest{
			UserId: validUUID.String(),
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resp == nil || resp.Preferences == nil {
			t.Fatalf("expected preferences in response")
		}
		if resp.Preferences.UserId != validUUID.String() {
			t.Errorf("expected user id %s, got: %s", validUUID.String(), resp.Preferences.UserId)
		}
		if resp.Preferences.Email != mockUser.Email {
			t.Errorf("expected email %s, got: %s", mockUser.Email, resp.Preferences.Email)
		}
		if resp.Preferences.DisplayName != mockUser.DisplayName {
			t.Errorf("expected display name %s, got: %s", mockUser.DisplayName, resp.Preferences.DisplayName)
		}
		if resp.Preferences.DisplayCurrency != mockUser.DisplayCurrency {
			t.Errorf("expected currency %s, got: %s", mockUser.DisplayCurrency, resp.Preferences.DisplayCurrency)
		}
		if resp.Preferences.Theme != mockUser.Theme {
			t.Errorf("expected theme %s, got: %s", mockUser.Theme, resp.Preferences.Theme)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		server := NewUserServer(nil)
		_, err := server.GetUserPreferences(ctx, nil)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Fatalf("expected Unavailable with nil svc, got: %v", err)
		}
	})

	t.Run("nil request with active svc", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		server := NewUserServer(mockSvc)

		_, err := server.GetUserPreferences(ctx, nil)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("empty user id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		server := NewUserServer(mockSvc)

		_, err := server.GetUserPreferences(ctx, &userpb.GetUserPreferencesRequest{UserId: ""})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("invalid user ID format", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			GetUserPreferences(ctx, "invalid-uuid").
			Return(nil, service.ErrInvalidUserID)

		server := NewUserServer(mockSvc)
		_, err := server.GetUserPreferences(ctx, &userpb.GetUserPreferencesRequest{UserId: "invalid-uuid"})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			GetUserPreferences(ctx, validUUID.String()).
			Return(nil, service.ErrUserNotFound)

		server := NewUserServer(mockSvc)
		_, err := server.GetUserPreferences(ctx, &userpb.GetUserPreferencesRequest{UserId: validUUID.String()})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			t.Fatalf("expected NotFound, got: %v", err)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			GetUserPreferences(ctx, validUUID.String()).
			Return(nil, errors.New("db timeout"))

		server := NewUserServer(mockSvc)
		_, err := server.GetUserPreferences(ctx, &userpb.GetUserPreferencesRequest{UserId: validUUID.String()})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Fatalf("expected Internal, got: %v", err)
		}
	})
}

func TestUserServer_UpdateUserPreferences(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()
	newName := "Oscar G."
	newCurrency := "EUR"
	newTheme := "LIGHT"
	now := time.Now().UTC()

	mockUpdatedUser := &domain.User{
		ID:              validUUID,
		Email:           "oscar@example.com",
		DisplayName:     newName,
		DisplayCurrency: newCurrency,
		Theme:           newTheme,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			UpdateUserPreferences(ctx, service.UpdatePreferencesInput{
				UserID:          validUUID.String(),
				DisplayName:     &newName,
				DisplayCurrency: &newCurrency,
				Theme:           &newTheme,
			}).
			Return(mockUpdatedUser, nil)

		server := NewUserServer(mockSvc)
		resp, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{
			UserId:          validUUID.String(),
			DisplayName:     &newName,
			DisplayCurrency: &newCurrency,
			Theme:           &newTheme,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resp == nil || resp.Preferences == nil {
			t.Fatalf("expected preferences in response")
		}
		if resp.Preferences.DisplayName != newName {
			t.Errorf("expected %s, got: %s", newName, resp.Preferences.DisplayName)
		}
		if resp.Preferences.DisplayCurrency != newCurrency {
			t.Errorf("expected %s, got: %s", newCurrency, resp.Preferences.DisplayCurrency)
		}
		if resp.Preferences.Theme != newTheme {
			t.Errorf("expected %s, got: %s", newTheme, resp.Preferences.Theme)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		server := NewUserServer(mockSvc)

		_, err := server.UpdateUserPreferences(ctx, nil)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("empty user id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		server := NewUserServer(mockSvc)

		_, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{UserId: ""})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("invalid display name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			UpdateUserPreferences(ctx, gomock.Any()).
			Return(nil, service.ErrInvalidDisplayName)

		server := NewUserServer(mockSvc)
		emptyName := ""
		_, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{
			UserId:      validUUID.String(),
			DisplayName: &emptyName,
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			UpdateUserPreferences(ctx, gomock.Any()).
			Return(nil, service.ErrInvalidCurrency)

		server := NewUserServer(mockSvc)
		badCurr := "XYZ"
		_, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{
			UserId:          validUUID.String(),
			DisplayCurrency: &badCurr,
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("invalid theme", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			UpdateUserPreferences(ctx, gomock.Any()).
			Return(nil, service.ErrInvalidTheme)

		server := NewUserServer(mockSvc)
		badTheme := "RETROWAVE"
		_, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{
			UserId: validUUID.String(),
			Theme:  &badTheme,
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			UpdateUserPreferences(ctx, gomock.Any()).
			Return(nil, service.ErrUserNotFound)

		server := NewUserServer(mockSvc)
		_, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{
			UserId: validUUID.String(),
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			t.Fatalf("expected NotFound, got: %v", err)
		}
	})
}

func TestUserServer_ListSupportedCurrencies(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			ListSupportedCurrencies(ctx).
			Return(domain.SupportedCurrencies, nil)

		server := NewUserServer(mockSvc)
		resp, err := server.ListSupportedCurrencies(ctx, &userpb.ListSupportedCurrenciesRequest{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resp == nil || len(resp.Currencies) != len(domain.SupportedCurrencies) {
			t.Fatalf("expected %d currencies, got %v", len(domain.SupportedCurrencies), resp)
		}
		if resp.Currencies[0].Code != "USD" || resp.Currencies[0].Symbol != "$" {
			t.Errorf("unexpected first currency: %+v", resp.Currencies[0])
		}
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			ListSupportedCurrencies(ctx).
			Return(nil, errors.New("failed to load"))

		server := NewUserServer(mockSvc)
		_, err := server.ListSupportedCurrencies(ctx, &userpb.ListSupportedCurrenciesRequest{})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Fatalf("expected Internal, got: %v", err)
		}
	})

	t.Run("nil service", func(t *testing.T) {
		server := NewUserServer(nil)
		_, err := server.ListSupportedCurrencies(ctx, &userpb.ListSupportedCurrenciesRequest{})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unavailable {
			t.Fatalf("expected Unavailable, got: %v", err)
		}
	})
}
