package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ehabterra/workflow"
	"github.com/ehabterra/workflow/history"
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

// Define a custom type for context keys to avoid collisions
type contextKey string

const notesKey contextKey = "notes"

var (
	db           *sql.DB
	workflowDef  *workflow.Definition
	templates    *template.Template
	workflowMgr  *workflow.Manager
	sqlStore     *storage.SQLiteStorage
	historyStore *history.SQLiteHistory
)

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./website_workflow.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// 1. Configure the generic SQL storage
	sqlStore, err = storage.NewSQLiteStorage(db,
		storage.WithTable("workflows"),
		storage.WithCustomFields(map[string]string{
			"title":   "title TEXT",
			"content": "content TEXT",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create sql store: %v", err)
	}

	// 2. Generate and execute the schema
	schema := sqlStore.GenerateSchema()
	log.Printf("Generated Schema:\n%s\n", schema)
	if err := storage.Initialize(db, schema); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// 2b. Initialize history store
	historyStore = history.NewSQLiteHistory(db,
		history.WithTable("transition_history"),
	)
	if err := historyStore.Initialize(); err != nil {
		log.Fatalf("Failed to initialize transition_history table: %v", err)
	}

	// 3. Define the workflow structure
	workflowReg := workflow.NewRegistry()
	workflowDef, err = workflow.NewDefinition(
		[]workflow.Place{"draft", "review", "approved", "published"},
		[]workflow.Transition{
			*workflow.MustNewTransition("submit_for_review", []workflow.Place{"draft"}, []workflow.Place{"review"}),
			*workflow.MustNewTransition("request_changes", []workflow.Place{"review"}, []workflow.Place{"draft"}),
			*workflow.MustNewTransition("approve", []workflow.Place{"review"}, []workflow.Place{"approved"}),
			*workflow.MustNewTransition("publish", []workflow.Place{"approved"}, []workflow.Place{"published"}),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create workflow definition: %v", err)
	}

	// 4. Create the manager with the generic store
	workflowMgr = workflow.NewManager(workflowReg, sqlStore)

	// Add a listener for logging/history (optional)
	workflowMgr.AddEventListener(workflow.EventAfterTransition, func(e workflow.Event) error {
		notesVal := e.Context().Value(notesKey)
		notesStr, _ := notesVal.(string)
		return historyStore.SaveTransition(&history.TransitionRecord{
			WorkflowID: e.Workflow().Name(),
			FromState:  fmt.Sprintf("%v", e.From()),
			ToState:    fmt.Sprintf("%v", e.To()),
			Transition: e.Transition().Name(),
			Notes:      notesStr,
			Actor:      "", // fill in if you have user info
			CreatedAt:  time.Now(),
		})
	})

	// Load templates
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	os.MkdirAll("templates", 0755)

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/workflow/new", handleNewWorkflowForm)
	http.HandleFunc("/workflow/create", handleCreateWorkflow)
	http.HandleFunc("/workflow/", handleWorkflowPage)
	http.HandleFunc("/diagram", handleDiagram)

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Simplified representation for the list view
type WorkflowSummary struct {
	ID    string
	Title string
	State string
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// For the home page, we query the table directly for a summary list.
	rows, err := db.Query("SELECT id, title, state FROM workflows ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var summaries []WorkflowSummary
	for rows.Next() {
		var summary WorkflowSummary
		var stateJSON string
		if err := rows.Scan(&summary.ID, &summary.Title, &stateJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var places []string
		if err := json.Unmarshal([]byte(stateJSON), &places); err == nil && len(places) > 0 {
			summary.State = places[0]
		} else {
			summary.State = "?"
		}
		summaries = append(summaries, summary)
	}

	templates.ExecuteTemplate(w, "home.html", summaries)
}

func handleNewWorkflowForm(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "workflow-form.html", nil)
}

func handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("content_%d", len(listWorkflowsHelper())+1)

	wf, err := workflowMgr.CreateWorkflow(id, workflowDef, "draft")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wf.SetContext("title", title)
	wf.SetContext("content", content)

	if err := workflowMgr.SaveWorkflow(id, wf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type WorkflowPageData struct {
	ID          string
	Workflow    *workflow.Workflow
	Title       interface{}
	Content     interface{}
	Transitions []workflow.Transition
	History     []history.TransitionRecord
}

func handleWorkflowPage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/workflow/")
	parts := strings.Split(path, "/")
	id := parts[0]

	if r.Method == "POST" {
		action := r.FormValue("action")
		notes := r.FormValue("notes")
		wf, err := workflowMgr.GetWorkflow(id, workflowDef)
		if err != nil {
			http.Error(w, "Workflow not found", http.StatusNotFound)
			return
		}

		transitions, _ := wf.EnabledTransitions()
		var targetTransition *workflow.Transition
		for i := range transitions {
			if transitions[i].Name() == action {
				targetTransition = &transitions[i]
				break
			}
		}

		if targetTransition == nil {
			http.Error(w, "Transition not allowed or does not exist", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(context.Background(), notesKey, notes)
		if err := wf.ApplyWithContext(ctx, targetTransition.To()); err != nil {
			http.Error(w, fmt.Sprintf("Failed to apply transition: %v", err), http.StatusInternalServerError)
			return
		}
		if err := workflowMgr.SaveWorkflow(id, wf); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save workflow: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}

	wf, err := workflowMgr.GetWorkflow(id, workflowDef)
	if err != nil {
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	title, _ := wf.Context("title")
	content, _ := wf.Context("content")
	enabledTransitions, _ := wf.EnabledTransitions()
	history, _ := historyStore.ListHistory(id, history.QueryOptions{Limit: 50, Offset: 0})

	data := WorkflowPageData{
		ID:          id,
		Workflow:    wf,
		Title:       title,
		Content:     content,
		Transitions: enabledTransitions,
		History:     history,
	}

	templates.ExecuteTemplate(w, "workflow.html", data)
}

func handleDiagram(w http.ResponseWriter, r *http.Request) {
	// Create a temporary workflow to generate the diagram from the definition
	tempWf, err := workflow.NewWorkflow("diagram-generator", workflowDef, "draft")
	if err != nil {
		http.Error(w, "Failed to create diagram generator", http.StatusInternalServerError)
		return
	}
	diagram := tempWf.Diagram()
	templates.ExecuteTemplate(w, "diagram.html", diagram)
}

func listWorkflowsHelper() []string {
	rows, err := db.Query("SELECT id FROM workflows")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
