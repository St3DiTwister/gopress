package http

import (
	"encoding/json"
	"fmt"
	"gopress/internal/middleware"
	"gopress/internal/models"
	"gopress/internal/repository"
	"gopress/pkg/httpx"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ArticleHandler struct {
	articleRepo repository.ArticleRepo
}

func NewArticleHandler(articleRepo repository.ArticleRepo) *ArticleHandler {
	return &ArticleHandler{articleRepo}
}

type newArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *ArticleHandler) Articles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.createArticle(w, r)
	} else if r.Method == http.MethodGet {
		h.listArticles(w, r)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *ArticleHandler) createArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req newArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	article := &models.Article{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
	}

	err := h.articleRepo.Create(ctx, article)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ArticleHandler) listArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := r.URL.Query()

	limit := httpx.QueryInt(q, "limit", 20)
	offset := httpx.QueryInt(q, "offset", 0)

	articles, err := h.articleRepo.List(ctx, limit, offset)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if articles == nil {
		articles = make([]*models.Article, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(articles)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *ArticleHandler) ArticlesByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	const prefix = "/articles/"
	if !strings.HasPrefix(path, prefix) {
		http.NotFound(w, r)
		return
	}

	idPart := strings.TrimPrefix(path, prefix)
	if idPart == "" {
		http.Error(w, "missing article id", http.StatusBadRequest)
		return
	}

	idPart = strings.TrimSuffix(idPart, "/")

	articleID, err := strconv.Atoi(idPart)
	if err != nil {
		http.Error(w, "invalid article id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getArticleByID(w, r, int64(articleID))
	case http.MethodPut:
		h.updateArticle(w, r, int64(articleID))
	case http.MethodDelete:
		h.deleteArticle(w, r, int64(articleID))
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type getArticleByIDResponse struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
}

func (h *ArticleHandler) getArticleByID(w http.ResponseWriter, r *http.Request, articleID int64) {
	ctx := r.Context()
	article, err := h.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if article == nil {
		http.Error(w, "article not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(getArticleByIDResponse{
		Title:     article.Title,
		Content:   article.Content,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
		Username:  article.AuthorUsername,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

type updateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *ArticleHandler) updateArticle(w http.ResponseWriter, r *http.Request, articleID int64) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req updateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	updated, err := h.articleRepo.UpdateOwned(ctx, articleID, userID, req.Title, req.Content)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !updated {
		http.Error(w, "article not found or forbidden", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ArticleHandler) deleteArticle(w http.ResponseWriter, r *http.Request, articleID int64) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deleted, err := h.articleRepo.DeleteOwned(ctx, articleID, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "article not found or forbidden", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
