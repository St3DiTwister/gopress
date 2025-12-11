package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopress/internal/models"
)

type ArticleRepo interface {
	Create(ctx context.Context, a *models.Article) error
	GetByID(ctx context.Context, id int64) (*models.Article, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*models.Article, error)
	List(ctx context.Context, limit int, offset int) ([]*models.Article, error)
}

type articleRepo struct {
	pool *pgxpool.Pool
}

func NewArticleRepo(pool *pgxpool.Pool) ArticleRepo {
	return &articleRepo{pool: pool}
}

func (r *articleRepo) Create(ctx context.Context, a *models.Article) error {
	const query = `
        INSERT INTO articles (author_id, title, content)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `

	row := r.pool.QueryRow(ctx, query, a.AuthorID, a.Title, a.Content)
	if err := row.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return fmt.Errorf("insert article: %w", err)
	}
	return nil
}

func (r *articleRepo) GetByID(ctx context.Context, id int64) (*models.Article, error) {
	const query = `
        SELECT a.id, title, content, author_id, a.created_at, a.updated_at, u.username
        FROM articles a
        LEFT JOIN users u on u.id = a.author_id
        WHERE a.id = $1
    `

	var a models.Article
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID,
		&a.Title,
		&a.Content,
		&a.AuthorID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.AuthorUsername,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get article by id: %w", err)
	}

	return &a, nil
}

func (r *articleRepo) ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*models.Article, error) {
	const query = `
		SELECT a.id, a.title, a.content, a.author_id, a.created_at, a.updated_at, u.username
		FROM articles a
		LEFT JOIN users u on u.id = a.author_id
		WHERE author_id = $1
		ORDER BY a.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("get articles by author: %w", err)
	}
	defer rows.Close()

	var res []*models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(
			&a.ID,
			&a.Title,
			&a.Content,
			&a.AuthorID,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.AuthorUsername,
		); err != nil {
			return nil, fmt.Errorf("scan articles by author: %w", err)
		}
		res = append(res, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get articles by author: %w", err)
	}

	return res, nil
}

func (r *articleRepo) List(ctx context.Context, limit int, offset int) ([]*models.Article, error) {
	if limit < 1 || limit > 20 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT a.id, author_id, title, content, a.created_at, a.updated_at, u.username
		FROM articles a
		LEFT JOIN users u on u.id = a.author_id
		ORDER BY a.created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get articles: %w", err)
	}
	defer rows.Close()

	var res []*models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(
			&a.ID,
			&a.AuthorID,
			&a.Title,
			&a.Content,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.AuthorUsername,
		); err != nil {
			return nil, fmt.Errorf("scan articles: %w", err)
		}
		res = append(res, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get articles: %w", err)
	}
	return res, nil
}
