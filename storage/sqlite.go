package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ehabterra/workflow"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements the Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(db *sql.DB) *SQLiteStorage {
	return &SQLiteStorage{db: db}
}

// LoadState loads the workflow state from SQLite
func (s *SQLiteStorage) LoadState(id string) ([]workflow.Place, error) {
	// Extract the numeric ID from the workflow ID (e.g., "website_approval_123" -> "123")
	parts := strings.Split(id, "_")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid workflow ID format: %s", id)
	}
	numericID := parts[2]

	var state string
	err := s.db.QueryRow("SELECT state FROM website_workflows WHERE id = ?", numericID).Scan(&state)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to load state: %w", err)
	}
	return []workflow.Place{workflow.Place(state)}, nil
}

// SaveState saves the workflow state to SQLite
func (s *SQLiteStorage) SaveState(id string, places []workflow.Place) error {
	if len(places) == 0 {
		return fmt.Errorf("no places to save")
	}

	// Extract the numeric ID from the workflow ID
	parts := strings.Split(id, "_")
	if len(parts) < 3 {
		return fmt.Errorf("invalid workflow ID format: %s", id)
	}
	numericID := parts[2]

	state := string(places[0])
	_, err := s.db.Exec("UPDATE website_workflows SET state = ? WHERE id = ?", state, numericID)
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

// DeleteState removes the workflow state from SQLite
func (s *SQLiteStorage) DeleteState(id string) error {
	// Extract the numeric ID from the workflow ID
	parts := strings.Split(id, "_")
	if len(parts) < 3 {
		return fmt.Errorf("invalid workflow ID format: %s", id)
	}
	numericID := parts[2]

	_, err := s.db.Exec("DELETE FROM website_workflows WHERE id = ?", numericID)
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}
	return nil
}
