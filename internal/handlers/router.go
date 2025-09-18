package handlers

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(uh *UserHandler) *chi.Mux {
	r := chi.NewRouter()
	initRoutes(r, uh)
	return r
}

func initRoutes(r *chi.Mux, uh *UserHandler) {
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", uh.Register)
		r.Post("/login", uh.Login)
	})
}
