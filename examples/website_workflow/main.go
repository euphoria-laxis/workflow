package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ehabterra/workflow"
	"github.com/ehabterra/workflow/storage"
	_ "github.com/mattn/go-sqlite3"
)

// WebsiteWorkflow represents a website approval workflow
type WebsiteWorkflow struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

// TransitionHistory represents a workflow transition record
type TransitionHistory struct {
	ID         int64     `json:"id"`
	WorkflowID int64     `json:"workflow_id"`
	FromState  string    `json:"from_state"`
	ToState    string    `json:"to_state"`
	Transition string    `json:"transition"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	db          *sql.DB
	workflowReg *workflow.Registry
	workflowDef *workflow.Definition
	templates   *template.Template
	workflowMgr *workflow.Manager
)

func init() {
	// Initialize SQLite database
	var err error
	db, err = sql.Open("sqlite3", "./website_workflow.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	_, err = db.Exec(`
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
	if err != nil {
		log.Fatal(err)
	}

	// Initialize workflow registry and definition
	workflowReg = workflow.NewRegistry()
	workflowDef, err = workflow.NewDefinition([]workflow.Place{"draft", "review", "approved", "published"}, []workflow.Transition{})
	if err != nil {
		log.Fatal(err)
	}

	// Define transitions
	tr1, _ := workflow.NewTransition("submit_for_review", []workflow.Place{"draft"}, []workflow.Place{"review"})
	tr2, _ := workflow.NewTransition("request_changes", []workflow.Place{"review"}, []workflow.Place{"draft"})
	tr3, _ := workflow.NewTransition("approve", []workflow.Place{"review"}, []workflow.Place{"approved"})
	tr4, _ := workflow.NewTransition("publish", []workflow.Place{"approved"}, []workflow.Place{"published"})
	workflowDef.Transitions = append(workflowDef.Transitions, *tr1, *tr2, *tr3, *tr4)

	// Initialize workflow manager with SQLite storage
	sqliteStorage := storage.NewSQLiteStorage(db)
	workflowMgr = workflow.NewManager(workflowReg, sqliteStorage)

	// Load templates
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	// Create templates directory if it doesn't exist
	os.MkdirAll("templates", 0755)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// API routes
	http.HandleFunc("/api/workflows", handleWorkflows)
	http.HandleFunc("/api/workflows/", handleWorkflowAndHistory)

	// Web routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/workflow/", handleWorkflowPage)
	http.HandleFunc("/diagram", handleDiagram)

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var workflows []WebsiteWorkflow
	rows, err := db.Query("SELECT id, title, description, state, notes FROM website_workflows")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var wf WebsiteWorkflow
		if err := rows.Scan(&wf.ID, &wf.Title, &wf.Description, &wf.State, &wf.Notes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		workflows = append(workflows, wf)
	}

	templates.ExecuteTemplate(w, "home.html", workflows)
}

func handleWorkflowPage(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	var wf WebsiteWorkflow
	err := db.QueryRow("SELECT id, title, description, state, notes FROM website_workflows WHERE id = ?", id).
		Scan(&wf.ID, &wf.Title, &wf.Description, &wf.State, &wf.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ExecuteTemplate(w, "workflow.html", wf)
}

func handleDiagram(w http.ResponseWriter, r *http.Request) {
	// Create a template workflow for the diagram
	templateWorkflow, err := workflow.NewWorkflow("template", workflowDef, workflow.Place("draft"))
	if err != nil {
		http.Error(w, "Failed to create template workflow", http.StatusInternalServerError)
		return
	}
	diagram := templateWorkflow.Diagram()
	templates.ExecuteTemplate(w, "diagram.html", diagram)
}

func handleWorkflows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var workflows []WebsiteWorkflow
		rows, err := db.Query("SELECT id, title, description, state, notes FROM website_workflows")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var wf WebsiteWorkflow
			if err := rows.Scan(&wf.ID, &wf.Title, &wf.Description, &wf.State, &wf.Notes); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			workflows = append(workflows, wf)
		}

		json.NewEncoder(w).Encode(workflows)

	case "POST":
		var wf WebsiteWorkflow
		if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if wf.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		// Use prepared statement
		stmt, err := db.Prepare(`
			INSERT INTO website_workflows 
			(title, description, state, notes) 
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(wf.Title, wf.Description, "draft", wf.Notes)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		wf.ID = id
		wf.State = "draft"

		// Create new workflow instance
		_, err = workflowMgr.CreateWorkflow(fmt.Sprintf("website_approval_%d", id), workflowDef, workflow.Place("draft"))
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(wf)
	}
}

func handleWorkflowAndHistory(w http.ResponseWriter, r *http.Request) {
	// Split the path to get the ID and check if it's a history request
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	// Check if this is a history request
	if len(parts) > 3 && parts[3] == "history" {
		handleTransitionHistory(w, r)
		return
	}

	// Otherwise handle as a regular workflow request
	handleWorkflow(w, r)
}

func handleWorkflow(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	idStr := parts[2] // /api/workflows/{id}

	// Convert string ID to int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
		return
	}

	// Get the workflow instance
	workflowInstance, err := workflowMgr.GetWorkflow(fmt.Sprintf("website_approval_%d", id), workflowDef)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		var wf WebsiteWorkflow
		err := db.QueryRow("SELECT id, title, description, state, notes FROM website_workflows WHERE id = ?", id).
			Scan(&wf.ID, &wf.Title, &wf.Description, &wf.State, &wf.Notes)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Workflow not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		json.NewEncoder(w).Encode(wf)

	case "PUT":
		var wf WebsiteWorkflow
		if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if wf.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		// Use prepared statement to prevent SQL injection
		stmt, err := db.Prepare("UPDATE website_workflows SET title = ?, description = ?, notes = ? WHERE id = ?")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(wf.Title, wf.Description, wf.Notes, id)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(wf)

	case "POST":
		var transition struct {
			Transition string `json:"transition"`
			Notes      string `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&transition); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate transition input
		if transition.Transition == "" {
			http.Error(w, "Transition is required", http.StatusBadRequest)
			return
		}

		var wf WebsiteWorkflow
		err := db.QueryRow("SELECT id, title, description, state, notes FROM website_workflows WHERE id = ?", id).
			Scan(&wf.ID, &wf.Title, &wf.Description, &wf.State, &wf.Notes)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Workflow not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Store the current state before transition
		fromState := wf.State
		targetState := transition.Transition

		// Apply the transition using the workflow instance
		if err := workflowInstance.Apply([]workflow.Place{workflow.Place(targetState)}); err != nil {
			http.Error(w, "Transition failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Save the new state
		if err := workflowMgr.SaveWorkflow(fmt.Sprintf("website_approval_%d", id), workflowInstance); err != nil {
			http.Error(w, "Failed to save workflow state", http.StatusInternalServerError)
			return
		}

		// Record transition in history
		historyStmt, err := db.Prepare(`
			INSERT INTO transition_history 
			(workflow_id, from_state, to_state, transition, notes) 
			VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer historyStmt.Close()

		_, err = historyStmt.Exec(id, fromState, targetState, targetState, transition.Notes)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		wf.State = targetState
		json.NewEncoder(w).Encode(wf)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTransitionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[3] != "history" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	id := parts[2] // /api/workflows/{id}/history

	var history []TransitionHistory
	rows, err := db.Query(`
		SELECT id, workflow_id, from_state, to_state, transition, notes, created_at 
		FROM transition_history 
		WHERE workflow_id = ? 
		ORDER BY created_at DESC
	`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var h TransitionHistory
		if err := rows.Scan(&h.ID, &h.WorkflowID, &h.FromState, &h.ToState, &h.Transition, &h.Notes, &h.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		history = append(history, h)
	}

	json.NewEncoder(w).Encode(history)
}
