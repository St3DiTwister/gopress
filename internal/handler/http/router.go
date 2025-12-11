package http

import (
	"gopress/internal/middleware"
	jwtpkg "gopress/pkg/jwt"
	"net/http"
)

type Handlers struct {
	Auth *AuthHandler
}

type Router struct {
	mux *http.ServeMux
}

func NewRouter(h Handlers, jwtManager *jwtpkg.Manager) *Router {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", h.Auth.Login)
	mux.HandleFunc("/register", h.Auth.Register)
	mux.Handle("/me", middleware.RequireAuth(jwtManager, http.HandlerFunc(h.Auth.GetMe)))

	return &Router{mux: mux}
}

func (r *Router) Handler() http.Handler {
	return r.mux
}
