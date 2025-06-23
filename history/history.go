package history

import "time"

// TransitionRecord is the base struct for a transition event.
type TransitionRecord struct {
	WorkflowID   string
	FromState    string
	ToState      string
	Transition   string
	Notes        string
	Actor        string
	CreatedAt    time.Time
	CustomFields map[string]interface{} // For custom columns, if any
}

// QueryOptions allows for pagination and filtering.
type QueryOptions struct {
	Limit      int
	Offset     int
	FromDate   *time.Time
	ToDate     *time.Time
	Actor      string
	Transition string
}

// HistoryStore is the interface for saving and querying transition history.
type HistoryStore interface {
	SaveTransition(record *TransitionRecord) error
	ListHistory(workflowID string, opts QueryOptions) ([]TransitionRecord, error)
	GenerateSchema() string
	Initialize() error
}
