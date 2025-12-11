package http

import (
	"gopress/internal/middleware"
	jwtpkg "gopress/pkg/jwt"
	"net/http"
)

type Handlers struct {
	Auth    *AuthHandler
	Article *ArticleHandler
}

type Router struct {
	mux *http.ServeMux
}

func NewRouter(h Handlers, jwtManager *jwtpkg.Manager) *Router {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", h.Auth.Login)
	mux.HandleFunc("/register", h.Auth.Register)
	mux.Handle("/me", middleware.RequireAuth(jwtManager, http.HandlerFunc(h.Auth.GetMe)))

	mux.Handle("/articles", middleware.RequireAuth(jwtManager, http.HandlerFunc(h.Article.Articles)))
	mux.Handle("/articles/", middleware.RequireAuth(jwtManager, http.HandlerFunc(h.Article.ArticlesByID)))

	return &Router{mux: mux}
}

func (r *Router) Handler() http.Handler {
	return r.mux
}
