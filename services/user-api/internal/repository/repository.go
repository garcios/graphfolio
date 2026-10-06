package repository

import (
	"context"
	"errors"

	"user-api/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// Repository defines the persistence operations for user identity and preferences.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_repository.go -package=mocks user-api/internal/repository Repository
type Repository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUserPreferences(ctx context.Context, id uuid.UUID, displayName *string, currency *string, theme *string) (*domain.User, error)
}
