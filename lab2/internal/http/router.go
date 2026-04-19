package http

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"transport-auth/internal/config"
)

func NewServer(db *sql.DB, cfg config.Config) *http.Server {
	s := &Server{db: db, cfg: cfg}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	r.Get("/api/v1/swagger/*", httpSwagger.Handler(httpSwagger.URL("/api/v1/swagger/openapi.yaml")))
	r.Get("/api/v1/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yaml")
	})

	r.Post("/api/v1/auth/login", s.login)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(s.authMiddleware)

			r.Get("/terminals", s.listTerminals)
			r.Post("/terminals", s.createTerminal)
			r.Put("/terminals/{id}", s.updateTerminal)
			r.Delete("/terminals/{id}", s.deleteTerminal)

			r.Get("/cards", s.listCards)
			r.Post("/cards", s.createCard)
			r.Put("/cards/{id}", s.updateCard)
			r.Delete("/cards/{id}", s.deleteCard)

			r.Get("/transactions", s.listTransactions)
			r.Post("/transactions", s.createTransaction)
			r.Put("/transactions/{id}", s.updateTransaction)
			r.Delete("/transactions/{id}", s.deleteTransaction)

			r.Get("/users", s.listUsers)
			r.Post("/users", s.createUser)
			r.Put("/users/{id}", s.updateUser)
			r.Delete("/users/{id}", s.deleteUser)

			r.Get("/keys", s.listKeys)
			r.Post("/keys", s.createKey)
			r.Put("/keys/{id}", s.updateKey)
			r.Delete("/keys/{id}", s.deleteKey)

			r.Post("/terminal/authorize", s.authorizePayment)
			r.Get("/terminal/keys", s.uploadKeys)
		})
	})

	return &http.Server{
		Addr:    cfg.AppAddr,
		Handler: r,
	}
}
