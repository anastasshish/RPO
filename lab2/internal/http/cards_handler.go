package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) listCards(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	query := "SELECT id, card_number, balance, blocked, owner_name, key_id, user_id FROM cards ORDER BY id"
	args := []any{}
	if !u.IsAdmin {
		query = "SELECT id, card_number, balance, blocked, owner_name, key_id, user_id FROM cards WHERE user_id=? ORDER BY id"
		args = append(args, u.UserID)
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id, keyID, userID int64
		var cardNumber, ownerName string
		var balance float64
		var blocked bool
		if err := rows.Scan(&id, &cardNumber, &balance, &blocked, &ownerName, &keyID, &userID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		data = append(data, map[string]any{
			"id": id, "card_number": cardNumber, "balance": balance, "blocked": blocked, "owner_name": ownerName, "key_id": keyID, "user_id": userID,
		})
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) createCard(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		CardNumber string  `json:"card_number"`
		Balance    float64 `json:"balance"`
		Blocked    bool    `json:"blocked"`
		OwnerName  string  `json:"owner_name"`
		KeyID      int64   `json:"key_id"`
		UserID     *int64  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.UserID != nil && !u.IsAdmin && *req.UserID != u.UserID {
		writeError(w, http.StatusForbidden, "cannot assign card to another user")
		return
	}
	ownerUserID := u.UserID
	if u.IsAdmin && req.UserID != nil {
		ownerUserID = *req.UserID
	}
	ok, err := s.userExists(ownerUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusBadRequest, "user not found")
		return
	}
	res, err := s.db.Exec("INSERT INTO cards(card_number, balance, blocked, owner_name, key_id, user_id) VALUES(?,?,?,?,?,?)", req.CardNumber, req.Balance, req.Blocked, req.OwnerName, req.KeyID, ownerUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) updateCard(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		CardNumber string  `json:"card_number"`
		Balance    float64 `json:"balance"`
		Blocked    bool    `json:"blocked"`
		OwnerName  string  `json:"owner_name"`
		KeyID      int64   `json:"key_id"`
		UserID     *int64  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var currentOwnerID int64
	if err := s.db.QueryRow("SELECT user_id FROM cards WHERE id=?", id).Scan(&currentOwnerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !u.IsAdmin && currentOwnerID != u.UserID {
		writeError(w, http.StatusForbidden, "cannot edit other users cards")
		return
	}
	if !u.IsAdmin && req.UserID != nil && *req.UserID != currentOwnerID {
		writeError(w, http.StatusForbidden, "cannot change card owner")
		return
	}
	newOwnerID := currentOwnerID
	if u.IsAdmin && req.UserID != nil {
		newOwnerID = *req.UserID
	}
	ok, err := s.userExists(newOwnerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusBadRequest, "user not found")
		return
	}
	_, err = s.db.Exec("UPDATE cards SET card_number=?, balance=?, blocked=?, owner_name=?, key_id=?, user_id=? WHERE id=?", req.CardNumber, req.Balance, req.Blocked, req.OwnerName, req.KeyID, newOwnerID, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": true})
}

func (s *Server) deleteCard(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if !u.IsAdmin {
		var ownerID int64
		if err := s.db.QueryRow("SELECT user_id FROM cards WHERE id=?", id).Scan(&ownerID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "card not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if ownerID != u.UserID {
			writeError(w, http.StatusForbidden, "cannot delete other users cards")
			return
		}
	}
	_, err := s.db.Exec("DELETE FROM cards WHERE id=?", id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
