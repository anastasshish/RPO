package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) listTerminals(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id, serial_number, address, name FROM terminals ORDER BY id")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var data []map[string]any
	for rows.Next() {
		var id int64
		var serial, address, name string
		if err := rows.Scan(&id, &serial, &address, &name); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		data = append(data, map[string]any{"id": id, "serial_number": serial, "address": address, "name": name})
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) createTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var req struct {
		SerialNumber string `json:"serial_number"`
		Address      string `json:"address"`
		Name         string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := s.db.Exec("INSERT INTO terminals(serial_number, address, name) VALUES(?,?,?)", req.SerialNumber, req.Address, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) updateTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		SerialNumber string `json:"serial_number"`
		Address      string `json:"address"`
		Name         string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	_, err := s.db.Exec("UPDATE terminals SET serial_number=?, address=?, name=? WHERE id=?", req.SerialNumber, req.Address, req.Name, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": true})
}

func (s *Server) deleteTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	_, err := s.db.Exec("DELETE FROM terminals WHERE id=?", id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
