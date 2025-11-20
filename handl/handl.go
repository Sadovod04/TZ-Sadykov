package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) HandleNumbers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "POST":
		h.addNumber(w, r)
	case "GET":
		h.getNumbers(w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) addNumber(w http.ResponseWriter, r *http.Request) {
	var data map[string]int
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	number := data["number"]

	_, err := h.db.Exec("INSERT INTO numbers (value) VALUES ($1)", number)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.getNumbers(w)
}
func (h *Handler) HandleClean(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := h.db.Exec("DELETE FROM numbers")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.db.Exec("DELETE FROM sqlite_sequence WHERE name='numbers'")

	json.NewEncoder(w).Encode(map[string]string{"message": "почистил"})
}

func (h *Handler) getNumbers(w http.ResponseWriter) {
	rows, err := h.db.Query("SELECT value FROM numbers")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var numbers []int
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		numbers = append(numbers, n)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	sort.Ints(numbers)

	if err := json.NewEncoder(w).Encode(map[string][]int{"numbers": numbers}); err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
	}
}
