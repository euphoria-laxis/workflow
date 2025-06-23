package history

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}

func TestSQLiteHistory_Basic(t *testing.T) {
	db := setupTestDB(t)
	h := NewSQLiteHistory(db)
	if err := h.Initialize(); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	rec := &TransitionRecord{
		WorkflowID: "wf1",
		FromState:  "draft",
		ToState:    "review",
		Transition: "submit_for_review",
		Notes:      "test note",
		Actor:      "user1",
		CreatedAt:  time.Now(),
	}
	if err := h.SaveTransition(rec); err != nil {
		t.Fatalf("failed to save transition: %v", err)
	}

	history, err := h.ListHistory("wf1", QueryOptions{})
	if err != nil {
		t.Fatalf("failed to list history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 record, got %d", len(history))
	}
	if history[0].FromState != "draft" || history[0].ToState != "review" {
		t.Errorf("unexpected states: %+v", history[0])
	}
}

func TestSQLiteHistory_CustomFields(t *testing.T) {
	db := setupTestDB(t)
	h := NewSQLiteHistory(db, WithCustomFields(map[string]string{
		"ip_address": "ip_address TEXT",
		"user_agent": "user_agent TEXT",
	}))
	if err := h.Initialize(); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	rec := &TransitionRecord{
		WorkflowID: "wf2",
		FromState:  "review",
		ToState:    "approved",
		Transition: "approve",
		Notes:      "approved by admin",
		Actor:      "admin",
		CreatedAt:  time.Now(),
		CustomFields: map[string]interface{}{
			"ip_address": "127.0.0.1",
			"user_agent": "test-agent",
		},
	}
	if err := h.SaveTransition(rec); err != nil {
		t.Fatalf("failed to save transition with custom fields: %v", err)
	}

	history, err := h.ListHistory("wf2", QueryOptions{})
	if err != nil {
		t.Fatalf("failed to list history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 record, got %d", len(history))
	}
	cf := history[0].CustomFields
	if cf["ip_address"] != "127.0.0.1" || cf["user_agent"] != "test-agent" {
		t.Errorf("unexpected custom fields: %+v", cf)
	}
}

func TestSQLiteHistory_PaginationAndFiltering(t *testing.T) {
	db := setupTestDB(t)
	h := NewSQLiteHistory(db)
	if err := h.Initialize(); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}
	now := time.Now()
	for i := 0; i < 10; i++ {
		rec := &TransitionRecord{
			WorkflowID: "wf3",
			FromState:  "s1",
			ToState:    "s2",
			Transition: "t",
			Notes:      "n",
			Actor:      "actor",
			CreatedAt:  now.Add(time.Duration(i) * time.Minute),
		}
		if err := h.SaveTransition(rec); err != nil {
			t.Fatalf("failed to save transition: %v", err)
		}
	}
	// Test limit
	hist, err := h.ListHistory("wf3", QueryOptions{Limit: 3})
	if err != nil {
		t.Fatalf("failed to list history: %v", err)
	}
	if len(hist) != 3 {
		t.Errorf("expected 3 records, got %d", len(hist))
	}
	// Test offset
	hist2, err := h.ListHistory("wf3", QueryOptions{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("failed to list history: %v", err)
	}
	if len(hist2) != 2 {
		t.Errorf("expected 2 records, got %d", len(hist2))
	}
	// Test filtering by actor
	hist3, err := h.ListHistory("wf3", QueryOptions{Actor: "actor"})
	if err != nil {
		t.Fatalf("failed to list history: %v", err)
	}
	if len(hist3) != 10 {
		t.Errorf("expected 10 records, got %d", len(hist3))
	}
}
