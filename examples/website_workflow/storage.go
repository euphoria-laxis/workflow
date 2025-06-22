package main

import "database/sql"

// WebsiteStorage encapsulates all DB operations for workflows and history
type WebsiteStorage struct {
	db *sql.DB
}

func NewWebsiteStorage(db *sql.DB) *WebsiteStorage {
	return &WebsiteStorage{db: db}
}

func (s *WebsiteStorage) Initialize() error {
	// Create tables
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS website_workflows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			state TEXT NOT NULL,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS transition_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workflow_id INTEGER NOT NULL,
			from_state TEXT NOT NULL,
			to_state TEXT NOT NULL,
			transition TEXT NOT NULL,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workflow_id) REFERENCES website_workflows(id)
		);
	`)

	return err
}

func (s *WebsiteStorage) ListWorkflows() ([]WebsiteWorkflow, error) {
	var workflows []WebsiteWorkflow
	rows, err := s.db.Query("SELECT id, title, description, state, notes FROM website_workflows")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var wf WebsiteWorkflow
		if err := rows.Scan(&wf.ID, &wf.Title, &wf.Description, &wf.State, &wf.Notes); err != nil {
			return nil, err
		}
		workflows = append(workflows, wf)
	}
	return workflows, nil
}

func (s *WebsiteStorage) GetWorkflow(id int64) (WebsiteWorkflow, error) {
	var wf WebsiteWorkflow
	err := s.db.QueryRow("SELECT id, title, description, state, notes FROM website_workflows WHERE id = ?", id).
		Scan(&wf.ID, &wf.Title, &wf.Description, &wf.State, &wf.Notes)
	return wf, err
}

func (s *WebsiteStorage) CreateWorkflow(wf *WebsiteWorkflow) error {
	stmt, err := s.db.Prepare(`
		INSERT INTO website_workflows (title, description, state, notes) VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(wf.Title, wf.Description, wf.State, wf.Notes)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	wf.ID = id
	return nil
}

func (s *WebsiteStorage) UpdateWorkflow(wf *WebsiteWorkflow) error {
	stmt, err := s.db.Prepare("UPDATE website_workflows SET title = ?, description = ?, notes = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(wf.Title, wf.Description, wf.Notes, wf.ID)
	return err
}

func (s *WebsiteStorage) AddTransitionHistory(h *TransitionHistory) error {
	stmt, err := s.db.Prepare(`
		INSERT INTO transition_history (workflow_id, from_state, to_state, transition, notes) VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(h.WorkflowID, h.FromState, h.ToState, h.Transition, h.Notes)
	return err
}

func (s *WebsiteStorage) ListTransitionHistory(workflowID int64) ([]TransitionHistory, error) {
	var history []TransitionHistory
	rows, err := s.db.Query(`
		SELECT id, workflow_id, from_state, to_state, transition, notes, created_at 
		FROM transition_history 
		WHERE workflow_id = ? 
		ORDER BY created_at DESC
	`, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var h TransitionHistory
		if err := rows.Scan(&h.ID, &h.WorkflowID, &h.FromState, &h.ToState, &h.Transition, &h.Notes, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}
