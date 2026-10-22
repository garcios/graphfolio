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
		ID:                    validUUID,
		Email:                 "oscar@example.com",
		DisplayName:           "Oscar Garcia",
		DisplayCurrency:       "USD",
		Theme:                 "DARK",
		FinancialYearEndMonth: 6,
		FinancialYearEndDay:   30,
		CreatedAt:             now,
		UpdatedAt:             now,
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
		if resp.Preferences.FinancialYearEndMonth != 6 || resp.Preferences.FinancialYearEndDay != 30 {
			t.Errorf("expected EOFY 6/30, got: %d/%d", resp.Preferences.FinancialYearEndMonth, resp.Preferences.FinancialYearEndDay)
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

	t.Run("success with EOFY fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		fyM := int32(4)
		fyD := int32(5)
		expectedM := int(fyM)
		expectedD := int(fyD)

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			UpdateUserPreferences(ctx, service.UpdatePreferencesInput{
				UserID:                validUUID.String(),
				FinancialYearEndMonth: &expectedM,
				FinancialYearEndDay:   &expectedD,
			}).
			Return(&domain.User{
				ID:                    validUUID,
				Email:                 "oscar@example.com",
				DisplayName:           "Oscar Garcia",
				DisplayCurrency:       "GBP",
				Theme:                 "DARK",
				FinancialYearEndMonth: 4,
				FinancialYearEndDay:   5,
				CreatedAt:             now,
				UpdatedAt:             now,
			}, nil)

		server := NewUserServer(mockSvc)
		resp, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{
			UserId:                validUUID.String(),
			FinancialYearEndMonth: &fyM,
			FinancialYearEndDay:   &fyD,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resp.Preferences.FinancialYearEndMonth != 4 || resp.Preferences.FinancialYearEndDay != 5 {
			t.Errorf("expected EOFY 4/5, got: %d/%d", resp.Preferences.FinancialYearEndMonth, resp.Preferences.FinancialYearEndDay)
		}
	})

	t.Run("invalid EOFY returns InvalidArgument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			UpdateUserPreferences(ctx, gomock.Any()).
			Return(nil, service.ErrInvalidFinancialYearEnd)

		server := NewUserServer(mockSvc)
		badM := int32(2)
		badD := int32(31)
		_, err := server.UpdateUserPreferences(ctx, &userpb.UpdateUserPreferencesRequest{
			UserId:                validUUID.String(),
			FinancialYearEndMonth: &badM,
			FinancialYearEndDay:   &badD,
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument for invalid EOFY, got: %v", err)
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

func TestUserServer_ChangePassword(t *testing.T) {
	ctx := context.Background()
	userID := "018f0000-0000-7000-8000-000000000001"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			ChangePassword(ctx, userID, "OldPassword#123", "NewPassword#456", true, "refresh-token").
			Return(nil)

		server := NewUserServer(mockSvc)
		resp, err := server.ChangePassword(ctx, &userpb.ChangePasswordRequest{
			UserId:              userID,
			CurrentPassword:     "OldPassword#123",
			NewPassword:         "NewPassword#456",
			RevokeOtherSessions: true,
			CurrentRefreshToken: "refresh-token",
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success true")
		}
	})

	t.Run("invalid current password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			ChangePassword(ctx, userID, "WrongPass#123", "NewPassword#456", false, "").
			Return(service.ErrInvalidCurrentPassword)

		server := NewUserServer(mockSvc)
		_, err := server.ChangePassword(ctx, &userpb.ChangePasswordRequest{
			UserId:          userID,
			CurrentPassword: "WrongPass#123",
			NewPassword:     "NewPassword#456",
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Unauthenticated {
			t.Fatalf("expected Unauthenticated, got: %v", err)
		}
	})

	t.Run("same password invalid argument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			ChangePassword(ctx, userID, "OldPassword#123", "OldPassword#123", false, "").
			Return(service.ErrSamePassword)

		server := NewUserServer(mockSvc)
		_, err := server.ChangePassword(ctx, &userpb.ChangePasswordRequest{
			UserId:          userID,
			CurrentPassword: "OldPassword#123",
			NewPassword:     "OldPassword#123",
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		server := NewUserServer(mockSvc)
		_, err := server.ChangePassword(ctx, nil)
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.InvalidArgument {
			t.Fatalf("expected InvalidArgument, got: %v", err)
		}
	})
}

func TestUserServer_RoleAndMustChangePasswordMapping(t *testing.T) {
	ctx := context.Background()
	adminUUID := uuid.New()
	now := time.Now().UTC()
	adminUser := &domain.User{
		ID:                 adminUUID,
		Email:              "admin@graphfolio.app",
		DisplayName:        "System Administrator",
		DisplayCurrency:    "USD",
		Theme:              "DARK",
		Role:               domain.RoleAdmin,
		MustChangePassword: true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	t.Run("login returns admin role and must_change_password true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			Login(ctx, "admin@graphfolio.app", "Password123!", false).
			Return(&service.AuthResult{
				User:         adminUser,
				AccessToken:  "mock-access-token",
				RefreshToken: "mock-refresh-token",
				ExpiresIn:    900,
			}, nil)

		server := NewUserServer(mockSvc)
		resp, err := server.Login(ctx, &userpb.LoginRequest{
			Email:    "admin@graphfolio.app",
			Password: "Password123!",
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resp == nil || resp.User == nil {
			t.Fatalf("expected user preferences in response")
		}
		if resp.User.Role != userpb.UserRole_USER_ROLE_ADMIN {
			t.Errorf("expected role USER_ROLE_ADMIN, got: %v", resp.User.Role)
		}
		if !resp.User.MustChangePassword {
			t.Errorf("expected must_change_password true, got false")
		}
	})

	t.Run("get user preferences returns standard user role", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userUUID := uuid.New()
		standardUser := &domain.User{
			ID:                 userUUID,
			Email:              "investor@example.com",
			DisplayName:        "Investor",
			DisplayCurrency:    "USD",
			Theme:              "LIGHT",
			Role:               domain.RoleUser,
			MustChangePassword: false,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		mockSvc := mocks.NewMockUserService(ctrl)
		mockSvc.EXPECT().
			GetUserPreferences(ctx, userUUID.String()).
			Return(standardUser, nil)

		server := NewUserServer(mockSvc)
		resp, err := server.GetUserPreferences(ctx, &userpb.GetUserPreferencesRequest{
			UserId: userUUID.String(),
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resp.Preferences.Role != userpb.UserRole_USER_ROLE_USER {
			t.Errorf("expected role USER_ROLE_USER, got: %v", resp.Preferences.Role)
		}
		if resp.Preferences.MustChangePassword {
			t.Errorf("expected must_change_password false, got true")
		}
	})
}


