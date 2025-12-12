package ports

import (
	"context"
	"github.com/google/uuid"
	"gopress/internal/domain/article"
)

type ArticleRepo interface {
	Create(ctx context.Context, a *article.Article) error
	GetByID(ctx context.Context, id int64) (*article.Article, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*article.Article, error)
	List(ctx context.Context, limit int, offset int) ([]*article.Article, error)
	UpdateOwned(ctx context.Context, id int64, authorID uuid.UUID, title, content string) (bool, error)
	DeleteOwned(ctx context.Context, id int64, authorID uuid.UUID) (bool, error)
}
