package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Login == adminLogin && req.Password == adminPassword {
		s.writeAuthSuccess(w, 1, adminLogin, true)
		return
	}

	var (
		id           int64
		passwordHash string
		isAdmin      bool
	)
	err := s.db.QueryRow("SELECT id, password_hash, is_admin FROM users WHERE login = ?", req.Login).Scan(&id, &passwordHash, &isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "wrong login or password")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "wrong login or password")
		return
	}

	s.writeAuthSuccess(w, id, req.Login, isAdmin)
}

func (s *Server) writeAuthSuccess(w http.ResponseWriter, id int64, login string, isAdmin bool) {
	claims := userClaims{
		UserID:  id,
		Login:   login,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": tokenString,
		"user": map[string]any{
			"id":       id,
			"login":    login,
			"is_admin": isAdmin,
		},
	})
}
