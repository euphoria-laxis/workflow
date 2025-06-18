package workflow

import (
	"fmt"
	"testing"
)

// MockStorage implements the Storage interface for testing
type MockStorage struct {
	states map[string][]Place
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		states: make(map[string][]Place),
	}
}

func (m *MockStorage) LoadState(id string) ([]Place, error) {
	if states, ok := m.states[id]; ok {
		return states, nil
	}
	return nil, fmt.Errorf("workflow not found")
}

func (m *MockStorage) SaveState(id string, places []Place) error {
	m.states[id] = places
	return nil
}

func (m *MockStorage) DeleteState(id string) error {
	delete(m.states, id)
	return nil
}

func TestNewManager(t *testing.T) {
	registry := NewRegistry()
	storage := NewMockStorage()
	manager := NewManager(registry, storage)

	if manager.registry != registry {
		t.Errorf("Expected registry to be %v, got %v", registry, manager.registry)
	}
	if manager.storage != storage {
		t.Errorf("Expected storage to be %v, got %v", storage, manager.storage)
	}
}

func TestManager_CreateWorkflow(t *testing.T) {
	registry := NewRegistry()
	storage := NewMockStorage()
	manager := NewManager(registry, storage)

	// Create a simple workflow definition
	places := []Place{"draft", "review", "published"}
	definition, err := NewDefinition(places, []Transition{})
	if err != nil {
		t.Fatalf("Failed to create workflow definition: %v", err)
	}

	// Test creating a new workflow
	id := "test_workflow"
	initialPlace := Place("draft")
	wf, err := manager.CreateWorkflow(id, definition, initialPlace)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Verify workflow was created correctly
	if wf.Name() != id {
		t.Errorf("Expected workflow name to be %s, got %s", id, wf.Name())
	}

	// Verify workflow was added to registry
	registryWf, err := registry.Workflow(id)
	if err != nil {
		t.Errorf("Workflow not found in registry: %v", err)
	}
	if registryWf != wf {
		t.Errorf("Expected workflow in registry to be %v, got %v", wf, registryWf)
	}

	// Verify initial state was saved
	states, err := storage.LoadState(id)
	if err != nil {
		t.Errorf("Failed to load workflow state: %v", err)
	}
	if len(states) != 1 || states[0] != initialPlace {
		t.Errorf("Expected initial state to be %v, got %v", initialPlace, states)
	}
}

func TestManager_GetWorkflow(t *testing.T) {
	registry := NewRegistry()
	storage := NewMockStorage()
	manager := NewManager(registry, storage)

	// Create a simple workflow definition
	places := []Place{"draft", "review", "published"}
	definition, err := NewDefinition(places, []Transition{})
	if err != nil {
		t.Fatalf("Failed to create workflow definition: %v", err)
	}

	// Test getting a non-existent workflow
	_, err = manager.GetWorkflow("non_existent", definition)
	if err == nil {
		t.Error("Expected error when getting non-existent workflow")
	}

	// Create a workflow
	id := "test_workflow"
	initialPlace := Place("draft")
	wf, err := manager.CreateWorkflow(id, definition, initialPlace)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Test getting the workflow
	retrievedWf, err := manager.GetWorkflow(id, definition)
	if err != nil {
		t.Errorf("Failed to get workflow: %v", err)
	}
	if retrievedWf != wf {
		t.Errorf("Expected workflow to be %v, got %v", wf, retrievedWf)
	}
}

func TestManager_SaveWorkflow(t *testing.T) {
	registry := NewRegistry()
	storage := NewMockStorage()
	manager := NewManager(registry, storage)

	// Create a simple workflow definition
	places := []Place{"draft", "review", "published"}
	definition, err := NewDefinition(places, []Transition{})
	if err != nil {
		t.Fatalf("Failed to create workflow definition: %v", err)
	}

	// Create a workflow
	id := "test_workflow"
	initialPlace := Place("draft")
	wf, err := manager.CreateWorkflow(id, definition, initialPlace)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Change the workflow state
	newPlace := Place("review")
	wf.Marking().SetPlaces([]Place{newPlace})

	// Save the workflow
	err = manager.SaveWorkflow(id, wf)
	if err != nil {
		t.Errorf("Failed to save workflow: %v", err)
	}

	// Verify the state was saved
	states, err := storage.LoadState(id)
	if err != nil {
		t.Errorf("Failed to load workflow state: %v", err)
	}
	if len(states) != 1 || states[0] != newPlace {
		t.Errorf("Expected state to be %v, got %v", newPlace, states)
	}
}

func TestManager_DeleteWorkflow(t *testing.T) {
	registry := NewRegistry()
	storage := NewMockStorage()
	manager := NewManager(registry, storage)

	// Create a simple workflow definition
	places := []Place{"draft", "review", "published"}
	definition, err := NewDefinition(places, []Transition{})
	if err != nil {
		t.Fatalf("Failed to create workflow definition: %v", err)
	}

	// Create a workflow
	id := "test_workflow"
	initialPlace := Place("draft")
	_, err = manager.CreateWorkflow(id, definition, initialPlace)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// Delete the workflow
	err = manager.DeleteWorkflow(id)
	if err != nil {
		t.Errorf("Failed to delete workflow: %v", err)
	}

	// Verify workflow was removed from registry
	_, err = registry.Workflow(id)
	if err == nil {
		t.Error("Expected error when getting deleted workflow from registry")
	}

	// Verify workflow state was removed from storage
	_, err = storage.LoadState(id)
	if err == nil {
		t.Error("Expected error when getting deleted workflow from storage")
	}
}

func TestManager_LoadWorkflow(t *testing.T) {
	registry := NewRegistry()
	storage := NewMockStorage()
	manager := NewManager(registry, storage)

	// Create a simple workflow definition
	places := []Place{"draft", "review", "published"}
	definition, err := NewDefinition(places, []Transition{})
	if err != nil {
		t.Fatalf("Failed to create workflow definition: %v", err)
	}

	// Test loading a non-existent workflow
	_, err = manager.LoadWorkflow("non_existent", definition)
	if err == nil {
		t.Error("Expected error when loading non-existent workflow")
	}

	// Create a workflow and save its state
	id := "test_workflow"
	initialPlace := Place("draft")
	err = storage.SaveState(id, []Place{initialPlace})
	if err != nil {
		t.Fatalf("Failed to save workflow state: %v", err)
	}

	// Load the workflow
	wf, err := manager.LoadWorkflow(id, definition)
	if err != nil {
		t.Errorf("Failed to load workflow: %v", err)
	}

	// Verify workflow was loaded correctly
	if wf.Name() != id {
		t.Errorf("Expected workflow name to be %s, got %s", id, wf.Name())
	}

	// Verify workflow was added to registry
	registryWf, err := registry.Workflow(id)
	if err != nil {
		t.Errorf("Workflow not found in registry: %v", err)
	}
	if registryWf != wf {
		t.Errorf("Expected workflow in registry to be %v, got %v", wf, registryWf)
	}

	// Verify workflow state was loaded correctly
	places = wf.Marking().Places()
	if len(places) != 1 || places[0] != initialPlace {
		t.Errorf("Expected workflow state to be %v, got %v", initialPlace, places)
	}
}
