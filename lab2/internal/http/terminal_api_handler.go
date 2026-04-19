package http

import (
	"encoding/json"
	"net/http"
)

func (s *Server) authorizePayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CardNumber     string  `json:"card_number"`
		Amount         float64 `json:"amount"`
		TerminalSerial string  `json:"terminal_serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	var terminalID int64
	if err := tx.QueryRow("SELECT id FROM terminals WHERE serial_number=?", req.TerminalSerial).Scan(&terminalID); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"approved": false, "reason": "unknown terminal"})
		return
	}

	var balance float64
	var blocked bool
	var ownerUserID int64
	err = tx.QueryRow("SELECT balance, blocked, user_id FROM cards WHERE card_number=?", req.CardNumber).Scan(&balance, &blocked, &ownerUserID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"approved": false, "reason": "card not found"})
		return
	}
	u := s.currentUser(r)
	if u != nil && !u.IsAdmin && ownerUserID != u.UserID {
		writeJSON(w, http.StatusOK, map[string]any{"approved": false, "reason": "card not found"})
		return
	}
	if blocked {
		_, _ = tx.Exec("INSERT INTO transactions(amount, card_number, terminal_id, approved) VALUES(?,?,?,?)", req.Amount, req.CardNumber, terminalID, false)
		_ = tx.Commit()
		writeJSON(w, http.StatusOK, map[string]any{"approved": false, "reason": "card blocked"})
		return
	}
	if balance < req.Amount {
		_, _ = tx.Exec("INSERT INTO transactions(amount, card_number, terminal_id, approved) VALUES(?,?,?,?)", req.Amount, req.CardNumber, terminalID, false)
		_ = tx.Commit()
		writeJSON(w, http.StatusOK, map[string]any{"approved": false, "reason": "insufficient funds"})
		return
	}
	_, err = tx.Exec("UPDATE cards SET balance = balance - ? WHERE card_number=?", req.Amount, req.CardNumber)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_, err = tx.Exec("INSERT INTO transactions(amount, card_number, terminal_id, approved) VALUES(?,?,?,?)", req.Amount, req.CardNumber, terminalID, true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"approved": true})
}

func (s *Server) uploadKeys(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id, key_name, key_value FROM card_keys ORDER BY id")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var keys []map[string]any
	for rows.Next() {
		var id int64
		var name, value string
		if err := rows.Scan(&id, &name, &value); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		keys = append(keys, map[string]any{"id": id, "key_name": name, "key_value": value})
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": keys})
}
