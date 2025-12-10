package http

import "net/http"

type Router struct {
	mux *http.ServeMux
}

func NewRouter(handler *AuthHandler) *Router {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/register", handler.Register)

	return &Router{mux: mux}
}

func (r *Router) Handler() http.Handler {
	return r.mux
}
