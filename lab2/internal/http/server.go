package http

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"transport-auth/internal/config"
)

type Server struct {
	httpServer *http.Server
	db         *sql.DB
	cfg        config.Config
}

type contextKey string

const userCtxKey contextKey = "user"

const (
	adminLogin    = "admin"
	adminPassword = "admin123"
)

type userClaims struct {
	UserID   int64  `json:"user_id"`
	Login    string `json:"login"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func (s *Server) userExists(id int64) (bool, error) {
	var one int
	err := s.db.QueryRow("SELECT 1 FROM users WHERE id = ?", id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
