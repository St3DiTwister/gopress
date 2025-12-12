package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopress/internal/app/ports"
	"gopress/internal/domain/user"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) ports.UserRepo {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, u *user.User) error {
	const query = `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	row := r.pool.QueryRow(ctx, query, u.Email, u.Username, u.Password)
	if err := row.Scan(&u.ID, &u.CreatedAt); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	const query = `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var u user.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}
func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	const query = `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u user.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}
func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM users WHERE id = $1`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
