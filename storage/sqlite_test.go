package storage

import (
	"database/sql"
	"testing"

	"github.com/euphoria-laxis/workflow"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}

func TestSQLiteStorage_Basic(t *testing.T) {
	db := setupTestDB(t)
	s, err := NewSQLiteStorage(db, WithCustomFields(map[string]string{
		"foo": "foo TEXT",
	}))
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if err := Initialize(db, s.GenerateSchema()); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	places := []workflow.Place{"draft"}
	context := map[string]interface{}{"foo": "bar"}
	if err := s.SaveState("wf1", places, context); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	loadedPlaces, loadedContext, err := s.LoadState("wf1")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if len(loadedPlaces) != 1 || loadedPlaces[0] != "draft" {
		t.Errorf("unexpected places: %+v", loadedPlaces)
	}
	if loadedContext["foo"] != "bar" {
		t.Errorf("unexpected context: %+v", loadedContext)
	}
}

func TestSQLiteStorage_CustomFields(t *testing.T) {
	db := setupTestDB(t)
	s, err := NewSQLiteStorage(db, WithCustomFields(map[string]string{
		"ip_address": "ip_address TEXT",
		"user_agent": "user_agent TEXT",
	}))
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if err := Initialize(db, s.GenerateSchema()); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	places := []workflow.Place{"review"}
	context := map[string]interface{}{
		"ip_address": "127.0.0.1",
		"user_agent": "test-agent",
	}
	if err := s.SaveState("wf2", places, context); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	_, loadedContext, err := s.LoadState("wf2")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if loadedContext["ip_address"] != "127.0.0.1" || loadedContext["user_agent"] != "test-agent" {
		t.Errorf("unexpected custom fields: %+v", loadedContext)
	}
}

func TestSQLiteStorage_DeleteState(t *testing.T) {
	db := setupTestDB(t)
	s, err := NewSQLiteStorage(db)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	if err := Initialize(db, s.GenerateSchema()); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}
	places := []workflow.Place{"draft"}
	context := map[string]interface{}{}
	if err := s.SaveState("wf3", places, context); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}
	if err := s.DeleteState("wf3"); err != nil {
		t.Fatalf("failed to delete state: %v", err)
	}
	_, _, err = s.LoadState("wf3")
	if err == nil {
		t.Errorf("expected error when loading deleted state")
	}
}
