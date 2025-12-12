package ports

import (
	"context"
	"github.com/google/uuid"
	"gopress/internal/domain/user"
)

type UserRepo interface {
	Create(ctx context.Context, u *user.User) error
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
