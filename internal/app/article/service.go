package article

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gopress/internal/app/ports"
	"gopress/internal/domain/article"
)

var (
	ErrNotFound    = errors.New("article not found")
	ErrInvalidData = errors.New("invalid data")
)

type Service struct {
	repo ports.ArticleRepo
}

func NewService(repo ports.ArticleRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, title, content string) error {
	if title == "" || content == "" {
		return ErrInvalidData
	}

	a := &article.Article{
		Title:    title,
		Content:  content,
		AuthorID: userID,
	}

	return s.repo.Create(ctx, a)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*article.Article, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*article.Article, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *Service) Update(ctx context.Context, userID uuid.UUID, id int64, title, content string) error {
	if title == "" || content == "" {
		return ErrInvalidData
	}

	ok, err := s.repo.UpdateOwned(ctx, id, userID, title, content)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID uuid.UUID, id int64) error {
	ok, err := s.repo.DeleteOwned(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}
