package workflow

import (
	"fmt"
)

// Storage defines the interface for workflow state persistence
type Storage interface {
	// LoadState loads the workflow state for a given ID
	LoadState(id string) ([]Place, error)
	// SaveState saves the workflow state for a given ID
	SaveState(id string, places []Place) error
	// DeleteState removes the workflow state for a given ID
	DeleteState(id string) error
}

// Manager handles workflow instances and their persistence
type Manager struct {
	registry *Registry
	storage  Storage
}

// NewManager creates a new workflow manager
func NewManager(registry *Registry, storage Storage) *Manager {
	return &Manager{
		registry: registry,
		storage:  storage,
	}
}

// LoadWorkflow loads a workflow instance from storage
func (m *Manager) LoadWorkflow(id string, definition *Definition) (*Workflow, error) {
	// Try to get from registry first
	wf, err := m.registry.Workflow(id)
	if err == nil {
		return wf, nil
	}

	// Load state from storage
	places, err := m.storage.LoadState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow state: %w", err)
	}

	// Create new workflow instance
	wf, err = NewWorkflow(id, definition, places[0])
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Set the current marking
	wf.Marking().SetPlaces(places)

	// Add to registry
	m.registry.AddWorkflow(wf)
	return wf, nil
}

// SaveWorkflow saves a workflow instance state to storage
func (m *Manager) SaveWorkflow(id string, wf *Workflow) error {
	return m.storage.SaveState(id, wf.Marking().Places())
}

// GetWorkflow gets a workflow instance from the registry or loads it from storage
func (m *Manager) GetWorkflow(id string, definition *Definition) (*Workflow, error) {
	// Try to get from registry first
	wf, err := m.registry.Workflow(id)
	if err == nil {
		return wf, nil
	}

	// If not in registry, load from storage
	return m.LoadWorkflow(id, definition)
}

// CreateWorkflow creates a new workflow instance and saves it to storage
func (m *Manager) CreateWorkflow(id string, definition *Definition, initialPlace Place) (*Workflow, error) {
	wf, err := NewWorkflow(id, definition, initialPlace)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Save initial state
	if err := m.storage.SaveState(id, []Place{initialPlace}); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %w", err)
	}

	// Add to registry
	m.registry.AddWorkflow(wf)
	return wf, nil
}

// DeleteWorkflow removes a workflow instance and its state
func (m *Manager) DeleteWorkflow(id string) error {
	// Remove from registry
	m.registry.RemoveWorkflow(id)

	// Remove from storage
	return m.storage.DeleteState(id)
}
