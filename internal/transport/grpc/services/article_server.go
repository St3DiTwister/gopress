package services

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopress/internal/app/ports"
	"gopress/internal/domain/article"
	"gopress/internal/transport/grpc/interceptor"

	articlepb "gopress/api/proto/article"
)

type ArticleServer struct {
	articlepb.UnimplementedArticleServiceServer
	repo ports.ArticleRepo
}

func NewArticleServer(repo ports.ArticleRepo) *ArticleServer {
	return &ArticleServer{repo: repo}
}

func (s *ArticleServer) List(ctx context.Context, req *articlepb.ListArticlesRequest) (*articlepb.ListArticlesResponse, error) {
	limit := int(req.Limit)
	offset := int(req.Offset)

	items, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list articles")
	}

	res := &articlepb.ListArticlesResponse{Articles: make([]*articlepb.Article, 0, len(items))}
	for _, a := range items {
		res.Articles = append(res.Articles, mapArticle(a))
	}
	return res, nil
}

func (s *ArticleServer) Get(ctx context.Context, req *articlepb.GetArticleRequest) (*articlepb.GetArticleResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	a, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get article")
	}
	if a == nil {
		return nil, status.Error(codes.NotFound, "article not found")
	}

	return &articlepb.GetArticleResponse{Article: mapArticle(a)}, nil
}

func (s *ArticleServer) Create(ctx context.Context, req *articlepb.CreateArticleRequest) (*articlepb.CreateArticleResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth")
	}

	title := req.Title
	content := req.Content
	if title == "" || content == "" {
		return nil, status.Error(codes.InvalidArgument, "title and content required")
	}

	a := &article.Article{
		AuthorID: userID,
		Title:    title,
		Content:  content,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, status.Error(codes.Internal, "failed to create article")
	}

	return &articlepb.CreateArticleResponse{
		Status: "ok",
		Id:     a.ID,
	}, nil
}

func (s *ArticleServer) Update(ctx context.Context, req *articlepb.UpdateArticleRequest) (*articlepb.UpdateArticleResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth")
	}

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if req.Title == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "title and content required")
	}

	okOwned, err := s.repo.UpdateOwned(ctx, req.Id, userID, req.Title, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update article")
	}
	if !okOwned {
		// либо нет статьи, либо не владелец
		return nil, status.Error(codes.NotFound, "article not found")
	}

	return &articlepb.UpdateArticleResponse{Status: "ok"}, nil
}

func (s *ArticleServer) Delete(ctx context.Context, req *articlepb.DeleteArticleRequest) (*articlepb.DeleteArticleResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth")
	}

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	okOwned, err := s.repo.DeleteOwned(ctx, req.Id, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete article")
	}
	if !okOwned {
		return nil, status.Error(codes.NotFound, "article not found")
	}

	return &articlepb.DeleteArticleResponse{Status: "ok"}, nil
}

func mapArticle(a *article.Article) *articlepb.Article {
	var createdUnix int64
	var updatedUnix int64
	if !a.CreatedAt.IsZero() {
		createdUnix = a.CreatedAt.Unix()
	}
	if !a.UpdatedAt.IsZero() {
		updatedUnix = a.UpdatedAt.Unix()
	}

	return &articlepb.Article{
		Id:             a.ID,
		Title:          a.Title,
		Content:        a.Content,
		AuthorId:       a.AuthorID.String(),
		AuthorUsername: a.AuthorUsername,
		CreatedAtUnix:  createdUnix,
		UpdatedAtUnix:  updatedUnix,
	}
}
