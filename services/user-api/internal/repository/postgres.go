package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"user-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const getUserByIDSQL = `
SELECT id, email, password_hash, is_active, display_name, base_currency, theme, role, must_change_password, financial_year_end_month, financial_year_end_day, email_verified_at, last_login_at, created_at, updated_at
FROM users.users
WHERE id = $1;
`

const getUserByEmailSQL = `
SELECT id, email, password_hash, is_active, display_name, base_currency, theme, role, must_change_password, financial_year_end_month, financial_year_end_day, email_verified_at, last_login_at, created_at, updated_at
FROM users.users
WHERE lower(email) = lower($1);
`

const createUserSQL = `
INSERT INTO users.users (email, password_hash, display_name, base_currency, theme, role, must_change_password, is_active)
VALUES (lower($1), $2, $3, $4, 'DARK', 'USER', false, true)
RETURNING id, email, password_hash, is_active, display_name, base_currency, theme, role, must_change_password, financial_year_end_month, financial_year_end_day, email_verified_at, last_login_at, created_at, updated_at;
`

const updateUserPreferencesSQL = `
UPDATE users.users
SET display_name = COALESCE($2::text, display_name),
    base_currency = COALESCE($3::char(3), base_currency),
    theme = COALESCE($4::text, theme),
    financial_year_end_month = COALESCE($5::smallint, financial_year_end_month),
    financial_year_end_day = COALESCE($6::smallint, financial_year_end_day),
    updated_at = now()
WHERE id = $1
RETURNING id, email, password_hash, is_active, display_name, base_currency, theme, role, must_change_password, financial_year_end_month, financial_year_end_day, email_verified_at, last_login_at, created_at, updated_at;
`

const updateLastLoginSQL = `
UPDATE users.users
SET last_login_at = now()
WHERE id = $1;
`

const createSessionSQL = `
INSERT INTO users.sessions (id, user_id, refresh_token_hash, family_id, is_revoked, user_agent, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now());
`

const getSessionByHashSQL = `
SELECT id, user_id, refresh_token_hash, family_id, is_revoked, COALESCE(user_agent, ''), expires_at, created_at, updated_at
FROM users.sessions
WHERE refresh_token_hash = $1;
`

const revokeSessionSQL = `
UPDATE users.sessions
SET is_revoked = true, updated_at = now()
WHERE id = $1;
`

const revokeSessionFamilySQL = `
UPDATE users.sessions
SET is_revoked = true, updated_at = now()
WHERE family_id = $1;
`

const updatePasswordSQL = `
UPDATE users.users
SET password_hash = $1,
    must_change_password = false,
    updated_at = now()
WHERE id = $2 AND is_active = true;
`

const revokeOtherUserSessionsSQL = `
UPDATE users.sessions
SET is_revoked = true, updated_at = now()
WHERE user_id = $1
  AND is_revoked = false
  AND ($2 = '' OR refresh_token_hash != $2);
`

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.IsActive,
		&u.DisplayName,
		&u.DisplayCurrency,
		&u.Theme,
		&u.Role,
		&u.MustChangePassword,
		&u.FinancialYearEndMonth,
		&u.FinancialYearEndDay,
		&u.EmailVerifiedAt,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, getUserByIDSQL, id)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to get user by id %s: %w", id, err)
	}
	return u, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, getUserByEmailSQL, strings.TrimSpace(email))
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to get user by email: %w", err)
	}
	return u, nil
}

func (r *PostgresRepository) CreateUser(
	ctx context.Context,
	email, passwordHash, displayName, baseCurrency string,
) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, createUserSQL, strings.TrimSpace(email), passwordHash, displayName, baseCurrency)
	u, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("repository: failed to create user: %w", err)
	}
	return u, nil
}

func (r *PostgresRepository) UpdateUserPreferences(
	ctx context.Context,
	id uuid.UUID,
	displayName *string,
	currency *string,
	theme *string,
	fyMonth *int,
	fyDay *int,
) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, updateUserPreferencesSQL, id, displayName, currency, theme, fyMonth, fyDay)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to update user preferences for id %s: %w", id, err)
	}
	return u, nil
}

func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, updateLastLoginSQL, id)
	if err != nil {
		return fmt.Errorf("repository: failed to update last login for %s: %w", id, err)
	}
	return nil
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	if session.FamilyID == uuid.Nil {
		session.FamilyID = uuid.New()
	}

	_, err := r.pool.Exec(
		ctx,
		createSessionSQL,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.FamilyID,
		session.IsRevoked,
		session.UserAgent,
		session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("repository: failed to create session: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetSessionByHash(ctx context.Context, hash string) (*domain.Session, error) {
	row := r.pool.QueryRow(ctx, getSessionByHashSQL, hash)
	var s domain.Session
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.FamilyID,
		&s.IsRevoked,
		&s.UserAgent,
		&s.ExpiresAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("repository: failed to get session by hash: %w", err)
	}
	return &s, nil
}

func (r *PostgresRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, revokeSessionSQL, id)
	if err != nil {
		return fmt.Errorf("repository: failed to revoke session: %w", err)
	}
	return nil
}

func (r *PostgresRepository) RevokeSessionFamily(ctx context.Context, familyID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, revokeSessionFamilySQL, familyID)
	if err != nil {
		return fmt.Errorf("repository: failed to revoke session family: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	res, err := r.pool.Exec(ctx, updatePasswordSQL, passwordHash, id)
	if err != nil {
		return fmt.Errorf("repository: failed to update password: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *PostgresRepository) RevokeOtherUserSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) error {
	_, err := r.pool.Exec(ctx, revokeOtherUserSessionsSQL, userID, currentTokenHash)
	if err != nil {
		return fmt.Errorf("repository: failed to revoke other user sessions: %w", err)
	}
	return nil
}

