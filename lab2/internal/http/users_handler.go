package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !u.IsAdmin {
		row := s.db.QueryRow("SELECT id, login, is_admin FROM users WHERE id=?", u.UserID)
		var id int64
		var login string
		var isAdmin bool
		if err := row.Scan(&id, &login, &isAdmin); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, []map[string]any{{"id": id, "login": login, "is_admin": isAdmin}})
		return
	}

	rows, err := s.db.Query("SELECT id, login, is_admin FROM users ORDER BY id")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var login string
		var isAdmin bool
		if err := rows.Scan(&id, &login, &isAdmin); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		data = append(data, map[string]any{"id": id, "login": login, "is_admin": isAdmin})
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	res, err := s.db.Exec("INSERT INTO users(login, password_hash, is_admin) VALUES(?,?,?)", req.Login, string(hash), req.IsAdmin)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !u.IsAdmin && u.UserID != userID {
		writeError(w, http.StatusForbidden, "cannot edit other users")
		return
	}
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		IsAdmin  *bool  `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	hash := ""
	if req.Password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		hash = string(b)
	} else {
		row := s.db.QueryRow("SELECT password_hash FROM users WHERE id=?", userID)
		if err := row.Scan(&hash); err != nil {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
	}

	isAdminValue := false
	if u.IsAdmin && req.IsAdmin != nil {
		isAdminValue = *req.IsAdmin
	} else {
		row := s.db.QueryRow("SELECT is_admin FROM users WHERE id=?", userID)
		if err := row.Scan(&isAdminValue); err != nil {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
	}

	_, err = s.db.Exec("UPDATE users SET login=?, password_hash=?, is_admin=? WHERE id=?", req.Login, hash, isAdminValue, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": true})
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !u.IsAdmin && u.UserID != userID {
		writeError(w, http.StatusForbidden, "cannot delete other users")
		return
	}
	_, err = s.db.Exec("DELETE FROM users WHERE id=?", userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
