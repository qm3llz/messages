package repository

import (
	"context"
	"errors"
	"messages/internal/apperr"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) CreateSession(ctx context.Context, token string, userID int, expiresAt time.Time) error {
	sql := `INSERT INTO sessions(hashed_token, user_id, expires_at)
			VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(ctx, sql, token, userID, expiresAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepository) GetUserIDByToken(ctx context.Context, token string) (int, error) {
	sql := `SELECT user_id FROM sessions
			WHERE hashed_token = $1 AND expires_at >= NOW()`

	var sessionUserID int

	err := r.pool.QueryRow(ctx, sql, token).Scan(&sessionUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, apperr.ErrSessionNotFound
	}
	if err != nil {
		return 0, err
	}

	return sessionUserID, nil
}

func (r *SessionRepository) DeleteSessionByToken(ctx context.Context, token string) error {
	sql := `DELETE FROM sessions
			WHERE hashed_token = $1`

	_, err := r.pool.Exec(ctx, sql, token)
	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepository) DeleteSessionByUser(ctx context.Context, userID int) error {
	sql := `DELETE FROM sessions
			WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, sql, userID)
	if err != nil {
		return err
	}

	return nil
}
