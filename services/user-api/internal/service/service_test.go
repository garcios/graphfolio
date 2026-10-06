package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"user-api/internal/domain"
	"user-api/internal/repository"
	mocks "user-api/internal/repository/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestUserService_GetUserPreferences(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()
	mockUser := &domain.User{
		ID:              validUUID,
		Email:           "oscar@example.com",
		DisplayName:     "Oscar Garcia",
		DisplayCurrency: "USD",
		Theme:           "DARK",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().
			GetUserByID(ctx, validUUID).
			Return(mockUser, nil)

		svc := NewUserService(mockRepo)
		user, err := svc.GetUserPreferences(ctx, validUUID.String())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if user == nil || user.ID != validUUID {
			t.Errorf("expected user ID %s, got: %v", validUUID, user)
		}
		if user.DisplayName != "Oscar Garcia" {
			t.Errorf("expected display name Oscar Garcia, got: %s", user.DisplayName)
		}
	})

	t.Run("success with demo user alias '1'", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		demoUUID := uuid.MustParse(DefaultDemoUserID)
		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().
			GetUserByID(ctx, demoUUID).
			Return(mockUser, nil)

		svc := NewUserService(mockRepo)
		user, err := svc.GetUserPreferences(ctx, "1")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if user == nil {
			t.Fatalf("expected non-nil user")
		}
	})

	t.Run("invalid user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewUserService(mockRepo)

		user, err := svc.GetUserPreferences(ctx, "invalid-uuid")
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("expected ErrInvalidUserID, got: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got: %v", user)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().
			GetUserByID(ctx, validUUID).
			Return(nil, repository.ErrUserNotFound)

		svc := NewUserService(mockRepo)
		user, err := svc.GetUserPreferences(ctx, validUUID.String())
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got: %v", user)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dbErr := errors.New("db connection failure")
		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().
			GetUserByID(ctx, validUUID).
			Return(nil, dbErr)

		svc := NewUserService(mockRepo)
		user, err := svc.GetUserPreferences(ctx, validUUID.String())
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected dbErr, got: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got: %v", user)
		}
	})
}

func TestUserService_UpdateUserPreferences(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()
	newName := "Oscar G."
	newCurrency := "EUR"
	newTheme := "LIGHT"

	mockUpdatedUser := &domain.User{
		ID:              validUUID,
		Email:           "oscar@example.com",
		DisplayName:     newName,
		DisplayCurrency: newCurrency,
		Theme:           newTheme,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	t.Run("success with all fields and case normalization", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rawCurrency := "eur"
		rawTheme := "light"
		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().
			UpdateUserPreferences(ctx, validUUID, gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, id uuid.UUID, name *string, curr *string, th *string) (*domain.User, error) {
				if id != validUUID {
					t.Errorf("expected id %s, got %s", validUUID, id)
				}
				if name == nil || *name != newName {
					t.Errorf("expected name %s, got %v", newName, name)
				}
				if curr == nil || *curr != "EUR" {
					t.Errorf("expected uppercase currency EUR, got %v", curr)
				}
				if th == nil || *th != "LIGHT" {
					t.Errorf("expected uppercase theme LIGHT, got %v", th)
				}
				return mockUpdatedUser, nil
			})

		svc := NewUserService(mockRepo)
		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID:          validUUID.String(),
			DisplayName:     &newName,
			DisplayCurrency: &rawCurrency,
			Theme:           &rawTheme,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.DisplayName != newName || res.DisplayCurrency != newCurrency || res.Theme != newTheme {
			t.Errorf("unexpected user returned: %+v", res)
		}
	})

	t.Run("success with partial fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().
			UpdateUserPreferences(ctx, validUUID, nil, gomock.Any(), nil).
			DoAndReturn(func(_ context.Context, id uuid.UUID, name *string, curr *string, th *string) (*domain.User, error) {
				if name != nil {
					t.Errorf("expected nil name, got %v", name)
				}
				if curr == nil || *curr != "AUD" {
					t.Errorf("expected currency AUD, got %v", curr)
				}
				if th != nil {
					t.Errorf("expected nil theme, got %v", th)
				}
				return &domain.User{
					ID:              validUUID,
					DisplayCurrency: "AUD",
				}, nil
			})

		svc := NewUserService(mockRepo)
		curr := "AUD"
		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID:          validUUID.String(),
			DisplayCurrency: &curr,
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.DisplayCurrency != "AUD" {
			t.Errorf("expected currency AUD, got: %s", res.DisplayCurrency)
		}
	})

	t.Run("invalid user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewUserService(mockRepo)

		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID: "not-a-uuid",
		})
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("expected ErrInvalidUserID, got: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil user, got: %v", res)
		}
	})

	t.Run("invalid display name - empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewUserService(mockRepo)

		emptyName := "   "
		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID:      validUUID.String(),
			DisplayName: &emptyName,
		})
		if !errors.Is(err, ErrInvalidDisplayName) {
			t.Fatalf("expected ErrInvalidDisplayName, got: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil user, got: %v", res)
		}
	})

	t.Run("invalid display name - exceeds 100 characters", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewUserService(mockRepo)

		longName := strings.Repeat("A", 101)
		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID:      validUUID.String(),
			DisplayName: &longName,
		})
		if !errors.Is(err, ErrInvalidDisplayName) {
			t.Fatalf("expected ErrInvalidDisplayName, got: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil user, got: %v", res)
		}
	})

	t.Run("invalid currency - unsupported code", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewUserService(mockRepo)

		badCurr := "XYZ"
		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID:          validUUID.String(),
			DisplayCurrency: &badCurr,
		})
		if !errors.Is(err, ErrInvalidCurrency) {
			t.Fatalf("expected ErrInvalidCurrency, got: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil user, got: %v", res)
		}
	})

	t.Run("invalid theme - unknown theme option", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewUserService(mockRepo)

		badTheme := "RETROWAVE"
		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID: validUUID.String(),
			Theme:  &badTheme,
		})
		if !errors.Is(err, ErrInvalidTheme) {
			t.Fatalf("expected ErrInvalidTheme, got: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil user, got: %v", res)
		}
	})

	t.Run("valid themes accepted", func(t *testing.T) {
		for _, validTh := range []string{"DARK", "LIGHT", "SYSTEM", "dark", "system"} {
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().
				UpdateUserPreferences(ctx, validUUID, nil, nil, gomock.Any()).
				Return(mockUpdatedUser, nil)

			svc := NewUserService(mockRepo)
			th := validTh
			_, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
				UserID: validUUID.String(),
				Theme:  &th,
			})
			if err != nil {
				t.Errorf("expected theme %q to be valid, got: %v", validTh, err)
			}
			ctrl.Finish()
		}
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().
			UpdateUserPreferences(ctx, validUUID, gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, repository.ErrUserNotFound)

		svc := NewUserService(mockRepo)
		name := "New Name"
		res, err := svc.UpdateUserPreferences(ctx, UpdatePreferencesInput{
			UserID:      validUUID.String(),
			DisplayName: &name,
		})
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil user, got: %v", res)
		}
	})
}

func TestUserService_ListSupportedCurrencies(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewUserService(mockRepo)

	currencies, err := svc.ListSupportedCurrencies(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(currencies) != len(domain.SupportedCurrencies) {
		t.Errorf("expected %d currencies, got %d", len(domain.SupportedCurrencies), len(currencies))
	}

	foundUSD := false
	foundEUR := false
	for _, c := range currencies {
		if c.Code == "USD" && c.Symbol == "$" {
			foundUSD = true
		}
		if c.Code == "EUR" && c.Symbol == "€" {
			foundEUR = true
		}
	}
	if !foundUSD {
		t.Errorf("expected USD in supported currencies")
	}
	if !foundEUR {
		t.Errorf("expected EUR in supported currencies")
	}
}
