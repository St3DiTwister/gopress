package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopress/internal/app/ports"
	"gopress/internal/domain/article"
)

type articleRepo struct {
	pool *pgxpool.Pool
}

func NewArticleRepo(pool *pgxpool.Pool) ports.ArticleRepo {
	return &articleRepo{pool: pool}
}

func (r *articleRepo) Create(ctx context.Context, a *article.Article) error {
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

func (r *articleRepo) GetByID(ctx context.Context, id int64) (*article.Article, error) {
	const query = `
        SELECT a.id, title, content, author_id, a.created_at, a.updated_at, u.username
        FROM articles a
        LEFT JOIN users u on u.id = a.author_id
        WHERE a.id = $1
    `

	var a article.Article
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

func (r *articleRepo) ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*article.Article, error) {
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

	var res []*article.Article
	for rows.Next() {
		var a article.Article
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

func (r *articleRepo) List(ctx context.Context, limit int, offset int) ([]*article.Article, error) {
	if limit < 1 || limit > 20 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	const query = `
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

	var res []*article.Article
	for rows.Next() {
		var a article.Article
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

func (r *articleRepo) UpdateOwned(ctx context.Context, id int64, authorID uuid.UUID, title, content string) (bool, error) {
	const query = `
		UPDATE articles
		SET title = $1, content = $2, updated_at = NOW()
		WHERE id = $3
			AND author_id = $4
	`

	res, err := r.pool.Exec(ctx, query, title, content, id, authorID)
	if err != nil {
		return false, fmt.Errorf("update owned articles: %w", err)
	}

	if res.RowsAffected() == 0 {
		// article isn't found or execute not owner
		return false, nil
	}

	return true, nil
}

func (r *articleRepo) DeleteOwned(ctx context.Context, id int64, authorID uuid.UUID) (bool, error) {
	const query = `
		DELETE FROM articles
		WHERE id = $1
			AND author_id = $2
	`

	res, err := r.pool.Exec(ctx, query, id, authorID)
	if err != nil {
		return false, fmt.Errorf("delete owned articles: %w", err)
	}

	if res.RowsAffected() == 0 {
		// article isn't found or execute not owner
		return false, nil
	}
	return true, nil
}
