package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS numbers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			value INTEGER NOT NULL
		)
	`)
	if err != nil {
		panic(err)
	}

	return db
}

func TestHandleNumbers_GET(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	db.Exec("INSERT INTO numbers (value) VALUES (10), (5), (15)")

	handler := NewHandler(db)

	req := httptest.NewRequest("GET", "/numbers", nil)
	rr := httptest.NewRecorder()

	handler.HandleNumbers(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string][]int
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not parse response: %v", err)
	}

	numbers := response["numbers"]
	if len(numbers) != 3 {
		t.Errorf("Expected 3 numbers, got %d", len(numbers))
	}

	expected := []int{5, 10, 15}
	for i, num := range numbers {
		if num != expected[i] {
			t.Errorf("Expected %d at position %d, got %d", expected[i], i, num)
		}
	}
}

func TestHandleNumbers_POST(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := NewHandler(db)

	jsonData := `{"number": 42}`
	req := httptest.NewRequest("POST", "/numbers", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.HandleNumbers(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM numbers WHERE value = 42").Scan(&count)
	if err != nil {
		t.Fatalf("Could not query database: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 number in database, got %d", count)
	}
}

func TestHandleNumbers_InvalidJSON(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := NewHandler(db)

	jsonData := `{"number": "not_a_number"}`
	req := httptest.NewRequest("POST", "/numbers", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.HandleNumbers(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleClean(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	db.Exec("INSERT INTO numbers (value) VALUES (1), (2), (3)")

	handler := NewHandler(db)

	jsonData := `{"clean": "clean"}`
	req := httptest.NewRequest("POST", "/clean", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.HandleClean(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM numbers").Scan(&count)
	if err != nil {
		t.Fatalf("Could not query database: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 numbers in database after clean, got %d", count)
	}
}

func TestHandleClean_InvalidJSON(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := NewHandler(db)

	jsonData := `{"clean": "wrong_value"}`
	req := httptest.NewRequest("POST", "/clean", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.HandleClean(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestHandleClean_WrongMethod(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := NewHandler(db)

	req := httptest.NewRequest("GET", "/clean", nil)
	rr := httptest.NewRecorder()

	handler.HandleClean(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestHandleNumbers_WrongMethod(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := NewHandler(db)

	req := httptest.NewRequest("PUT", "/numbers", nil)
	rr := httptest.NewRecorder()

	handler.HandleNumbers(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestHandleNumbers_EmptyDatabase(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := NewHandler(db)

	req := httptest.NewRequest("GET", "/numbers", nil)
	rr := httptest.NewRecorder()

	handler.HandleNumbers(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string][]int
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not parse response: %v", err)
	}

	numbers := response["numbers"]
	if len(numbers) != 0 {
		t.Errorf("Expected 0 numbers in empty database, got %d", len(numbers))
	}
}

func TestHandleClean_EmptyJSON(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := NewHandler(db)

	jsonData := `{}`
	req := httptest.NewRequest("POST", "/clean", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.HandleClean(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
