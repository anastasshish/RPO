package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) listTransactions(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	query := "SELECT t.id, t.amount, t.card_number, t.terminal_id, t.created_at, t.approved FROM transactions t ORDER BY t.id DESC"
	args := []any{}
	if !u.IsAdmin {
		query = "SELECT t.id, t.amount, t.card_number, t.terminal_id, t.created_at, t.approved FROM transactions t JOIN cards c ON c.card_number=t.card_number WHERE c.user_id=? ORDER BY t.id DESC"
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
		var id, terminalID int64
		var cardNumber string
		var amount float64
		var createdAt string
		var approved bool
		if err := rows.Scan(&id, &amount, &cardNumber, &terminalID, &createdAt, &approved); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		data = append(data, map[string]any{
			"id": id, "amount": amount, "card_number": cardNumber, "terminal_id": terminalID, "created_at": createdAt, "approved": approved,
		})
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Amount     float64 `json:"amount"`
		CardNumber string  `json:"card_number"`
		TerminalID int64   `json:"terminal_id"`
		Approved   bool    `json:"approved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.CardNumber == "" {
		writeError(w, http.StatusBadRequest, "card_number required")
		return
	}
	if req.Amount < 0 {
		writeError(w, http.StatusBadRequest, "amount must be non-negative")
		return
	}

	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	var ownerID int64
	var balance float64
	err = tx.QueryRow("SELECT user_id, balance FROM cards WHERE card_number=?", req.CardNumber).Scan(&ownerID, &balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !u.IsAdmin && ownerID != u.UserID {
		writeError(w, http.StatusForbidden, "cannot create transaction for other users cards")
		return
	}
	if req.Approved {
		if balance < req.Amount {
			writeError(w, http.StatusBadRequest, "insufficient funds")
			return
		}
		if _, err := tx.Exec("UPDATE cards SET balance = balance - ? WHERE card_number=?", req.Amount, req.CardNumber); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	res, err := tx.Exec("INSERT INTO transactions(amount, card_number, terminal_id, approved) VALUES(?,?,?,?)", req.Amount, req.CardNumber, req.TerminalID, req.Approved)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) updateTransaction(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		Amount     float64 `json:"amount"`
		CardNumber string  `json:"card_number"`
		TerminalID int64   `json:"terminal_id"`
		Approved   bool    `json:"approved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.CardNumber == "" {
		writeError(w, http.StatusBadRequest, "card_number required")
		return
	}
	if req.Amount < 0 {
		writeError(w, http.StatusBadRequest, "amount must be non-negative")
		return
	}

	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	var oldAmount float64
	var oldApproved bool
	var oldCardNumber string
	err = tx.QueryRow("SELECT amount, approved, card_number FROM transactions WHERE id=?", id).Scan(&oldAmount, &oldApproved, &oldCardNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !u.IsAdmin {
		var txOwnerID int64
		err := tx.QueryRow("SELECT c.user_id FROM transactions t JOIN cards c ON c.card_number=t.card_number WHERE t.id=?", id).Scan(&txOwnerID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if txOwnerID != u.UserID {
			writeError(w, http.StatusForbidden, "cannot edit other users transactions")
			return
		}
		var newCardOwnerID int64
		if err := tx.QueryRow("SELECT user_id FROM cards WHERE card_number=?", req.CardNumber).Scan(&newCardOwnerID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusBadRequest, "card not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if newCardOwnerID != u.UserID {
			writeError(w, http.StatusForbidden, "cannot move transaction to other users card")
			return
		}
	}

	var newCardOwner int64
	if err := tx.QueryRow("SELECT user_id FROM cards WHERE card_number=?", req.CardNumber).Scan(&newCardOwner); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "card not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if oldApproved {
		if _, err := tx.Exec("UPDATE cards SET balance = balance + ? WHERE card_number=?", oldAmount, oldCardNumber); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if req.Approved {
		var bal float64
		if err := tx.QueryRow("SELECT balance FROM cards WHERE card_number=?", req.CardNumber).Scan(&bal); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusBadRequest, "card not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if bal < req.Amount {
			writeError(w, http.StatusBadRequest, "insufficient funds")
			return
		}
		if _, err := tx.Exec("UPDATE cards SET balance = balance - ? WHERE card_number=?", req.Amount, req.CardNumber); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	if _, err := tx.Exec("UPDATE transactions SET amount=?, card_number=?, terminal_id=?, approved=? WHERE id=?", req.Amount, req.CardNumber, req.TerminalID, req.Approved, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": true})
}

func (s *Server) deleteTransaction(w http.ResponseWriter, r *http.Request) {
	u := s.currentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if !u.IsAdmin {
		var ownerID int64
		err := s.db.QueryRow("SELECT c.user_id FROM transactions t JOIN cards c ON c.card_number=t.card_number WHERE t.id=?", id).Scan(&ownerID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "transaction not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if ownerID != u.UserID {
			writeError(w, http.StatusForbidden, "cannot delete other users transactions")
			return
		}
	}
	_, err := s.db.Exec("DELETE FROM transactions WHERE id=?", id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
