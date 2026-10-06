package repository

import (
	"context"
	"errors"
	"fmt"

	"user-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const getUserByIDSQL = `
SELECT id, email, display_name, base_currency, theme, created_at, updated_at
FROM users.users
WHERE id = $1;
`

const updateUserPreferencesSQL = `
UPDATE users.users
SET display_name = COALESCE($2::text, display_name),
    base_currency = COALESCE($3::char(3), base_currency),
    theme = COALESCE($4::text, theme),
    updated_at = now()
WHERE id = $1
RETURNING id, email, display_name, base_currency, theme, created_at, updated_at;
`

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, getUserByIDSQL, id)

	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.DisplayCurrency,
		&u.Theme,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to get user by id %s: %w", id, err)
	}

	return &u, nil
}

func (r *PostgresRepository) UpdateUserPreferences(
	ctx context.Context,
	id uuid.UUID,
	displayName *string,
	currency *string,
	theme *string,
) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, updateUserPreferencesSQL, id, displayName, currency, theme)

	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.DisplayCurrency,
		&u.Theme,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to update user preferences for id %s: %w", id, err)
	}

	return &u, nil
}
