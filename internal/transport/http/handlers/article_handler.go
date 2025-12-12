package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	articleSvc "gopress/internal/app/article"
	"gopress/internal/domain/article"
	"gopress/internal/transport/http/middleware"
	"gopress/pkg/httpx"
)

type ArticleHandler struct {
	service *articleSvc.Service
}

func NewArticleHandler(service *articleSvc.Service) *ArticleHandler {
	return &ArticleHandler{service: service}
}

type newArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *ArticleHandler) Articles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.create(w, r)
	case http.MethodGet:
		h.list(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ArticleHandler) create(w http.ResponseWriter, r *http.Request) {
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

	err := h.service.Create(ctx, userID, req.Title, req.Content)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ArticleHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	limit := httpx.QueryInt(q, "limit", 20)
	offset := httpx.QueryInt(q, "offset", 0)

	articles, err := h.service.List(ctx, limit, offset)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if articles == nil {
		articles = make([]*article.Article, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(articles)
}

func (h *ArticleHandler) ArticlesByID(w http.ResponseWriter, r *http.Request) {
	const prefix = "/articles/"
	idStr := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), prefix)

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ArticleHandler) get(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := r.Context()

	a, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, articleSvc.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}

func (h *ArticleHandler) update(w http.ResponseWriter, r *http.Request, id int64) {
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

	err := h.service.Update(ctx, userID, id, req.Title, req.Content)
	if err != nil {
		http.Error(w, "not found or forbidden", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ArticleHandler) delete(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.service.Delete(ctx, userID, id)
	if err != nil {
		http.Error(w, "not found or forbidden", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
