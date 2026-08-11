package repository

import (
	"context"
	"errors"
	"messages/internal/apperr"
	"messages/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, username, passwordHash string) (int, error) { // TODO: Сделать возвращаемым не id, а models.User
	sql := `INSERT INTO users(username, password_hash) 
			VALUES ($1, $2)
			RETURNING id;`

	var id int
	err := r.pool.QueryRow(ctx, sql, username, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, apperr.ErrUsernameTaken
		}
		return 0, err
	}
	return id, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, username string) error {
	sql := `DELETE FROM users WHERE username = $1;`

	_, err := r.pool.Exec(ctx, sql, username)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, ID int) (*models.User, error) {
	sql := `SELECT id, username, password_hash FROM users WHERE id = $1;`

	var user models.User

	err := r.pool.QueryRow(ctx, sql, ID).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	sql := `SELECT id, username, password_hash 
			FROM users 
			WHERE username = $1;`

	var user models.User

	err := r.pool.QueryRow(ctx, sql, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, passwordHash, username string) error { // TODO: Возможно, тут тоже стоит возвращать models.User
	sql := `UPDATE users 
			SET password_hash = $1 
			WHERE username = $2;`

	_, err := r.pool.Exec(ctx, sql, passwordHash, username)
	if err != nil {
		return err
	}

	return nil
}
