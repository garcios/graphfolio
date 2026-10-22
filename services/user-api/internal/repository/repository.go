package repository

import (
	"context"
	"errors"

	"user-api/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already registered")
	ErrSessionNotFound       = errors.New("session not found")
)

// Repository defines the persistence operations for user identity, authentication, and preferences.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_repository.go -package=mocks user-api/internal/repository Repository
type Repository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, email, passwordHash, displayName, baseCurrency string) (*domain.User, error)
	UpdateUserPreferences(ctx context.Context, id uuid.UUID, displayName *string, currency *string, theme *string, fyMonth *int, fyDay *int) (*domain.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	CreateSession(ctx context.Context, session *domain.Session) error
	GetSessionByHash(ctx context.Context, hash string) (*domain.Session, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeSessionFamily(ctx context.Context, familyID uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	RevokeOtherUserSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) error
}
