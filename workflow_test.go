package workflow_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/euphoria-laxis/workflow"
)

func TestNewWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		definition    func() (*workflow.Definition, error)
		initialPlace  workflow.Place
		wantErr       bool
		errorContains string
	}{
		{
			name: "valid workflow",
			definition: func() (*workflow.Definition, error) {
				return workflow.NewDefinition(
					[]workflow.Place{"start", "middle", "end"},
					[]workflow.Transition{},
				)
			},
			initialPlace: "start",
			wantErr:      false,
		},
		{
			name: "",
			definition: func() (*workflow.Definition, error) {
				return workflow.NewDefinition([]workflow.Place{"start"}, []workflow.Transition{})
			},
			initialPlace:  "start",
			wantErr:       true,
			errorContains: "workflow name cannot be empty",
		},
		{
			name:          "nil definition",
			definition:    func() (*workflow.Definition, error) { return nil, nil },
			initialPlace:  "start",
			wantErr:       true,
			errorContains: "workflow definition cannot be nil",
		},
		{
			name: "invalid initial place",
			definition: func() (*workflow.Definition, error) {
				return workflow.NewDefinition([]workflow.Place{"valid"}, []workflow.Transition{})
			},
			initialPlace:  "invalid",
			wantErr:       true,
			errorContains: "initial place 'invalid' is not defined in workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := tt.definition()
			if err != nil {
				t.Fatalf("failed to create definition: %v", err)
			}
			wf, err := workflow.NewWorkflow(tt.name, def, tt.initialPlace)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				name := wf.Name()
				if name != tt.name {
					t.Errorf("NewWorkflow() name = %v, want %v", name, tt.name)
				}
				if len(wf.CurrentPlaces()) != 1 || wf.CurrentPlaces()[0] != tt.initialPlace {
					t.Errorf("NewWorkflow() current place = %v, want %v", wf.CurrentPlaces(), []workflow.Place{tt.initialPlace})
				}
			} else if err != nil && tt.errorContains != "" && err.Error() != tt.errorContains && !tt.wantErr {
				t.Errorf("NewWorkflow() error = %v, want error containing %v", err, tt.errorContains)
			}
		})
	}
}

func TestWorkflow_Can(t *testing.T) {
	tests := []struct {
		name         string
		initialPlace workflow.Place
		to           []workflow.Place
		defPlaces    []workflow.Place
		defTrans     []struct {
			name string
			from []workflow.Place
			to   []workflow.Place
		}
		wantErr bool
	}{
		{
			name:         "valid single place transition",
			initialPlace: "a",
			to:           []workflow.Place{"b"},
			defPlaces:    []workflow.Place{"a", "b"},
			defTrans: []struct {
				name string
				from []workflow.Place
				to   []workflow.Place
			}{
				{"to-b", []workflow.Place{"a"}, []workflow.Place{"b"}},
			},
			wantErr: false,
		},
		{
			name:         "valid multiple places transition",
			initialPlace: "a",
			to:           []workflow.Place{"b", "c"},
			defPlaces:    []workflow.Place{"a", "b", "c"},
			defTrans: []struct {
				name string
				from []workflow.Place
				to   []workflow.Place
			}{
				{"to-bc", []workflow.Place{"a"}, []workflow.Place{"b", "c"}},
			},
			wantErr: false,
		},
		{
			name:         "invalid transition - no path",
			initialPlace: "a",
			to:           []workflow.Place{"c"},
			defPlaces:    []workflow.Place{"a", "b", "c"},
			defTrans: []struct {
				name string
				from []workflow.Place
				to   []workflow.Place
			}{
				{"to-b", []workflow.Place{"a"}, []workflow.Place{"b"}},
			},
			wantErr: true,
		},
		{
			name:         "invalid transition - empty target",
			initialPlace: "a",
			to:           []workflow.Place{},
			defPlaces:    []workflow.Place{"a", "b"},
			defTrans: []struct {
				name string
				from []workflow.Place
				to   []workflow.Place
			}{
				{"to-b", []workflow.Place{"a"}, []workflow.Place{"b"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var defTransObjs []workflow.Transition
			for _, tr := range tt.defTrans {
				trObj, err := workflow.NewTransition(tr.name, tr.from, tr.to)
				if err != nil {
					t.Fatalf("failed to create transition: %v", err)
				}
				defTransObjs = append(defTransObjs, *trObj)
			}
			definition, err := workflow.NewDefinition(tt.defPlaces, defTransObjs)
			if err != nil {
				t.Fatalf("failed to create definition: %v", err)
			}
			wf, _ := workflow.NewWorkflow("test", definition, tt.initialPlace)
			err = wf.CanWithContext(context.Background(), tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Workflow.Can() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkflow_Apply(t *testing.T) {
	tests := []struct {
		name         string
		definition   func() (*workflow.Definition, error)
		initialPlace workflow.Place
		targetPlaces []workflow.Place
		wantErr      bool
		check        func(*workflow.Workflow) error
	}{
		{
			name: "valid single place transition",
			definition: func() (*workflow.Definition, error) {
				t, _ := workflow.NewTransition("to-middle", []workflow.Place{"start"}, []workflow.Place{"middle"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "middle"},
					[]workflow.Transition{*t},
				)
			},
			initialPlace: "start",
			targetPlaces: []workflow.Place{"middle"},
			wantErr:      false,
			check: func(w *workflow.Workflow) error {
				if len(w.CurrentPlaces()) != 1 || w.CurrentPlaces()[0] != "middle" {
					return fmt.Errorf("expected current place to be 'middle', got %v", w.CurrentPlaces())
				}
				return nil
			},
		},
		{
			name: "valid multiple places transition",
			definition: func() (*workflow.Definition, error) {
				t, _ := workflow.NewTransition("to-multiple", []workflow.Place{"start"}, []workflow.Place{"place1", "place2"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "place1", "place2"},
					[]workflow.Transition{*t},
				)
			},
			initialPlace: "start",
			targetPlaces: []workflow.Place{"place1", "place2"},
			wantErr:      false,
			check: func(w *workflow.Workflow) error {
				if len(w.CurrentPlaces()) != 2 {
					return fmt.Errorf("expected 2 current places, got %d", len(w.CurrentPlaces()))
				}
				places := map[workflow.Place]bool{"place1": true, "place2": true}
				for _, place := range w.CurrentPlaces() {
					if !places[place] {
						return fmt.Errorf("unexpected place %v", place)
					}
				}
				return nil
			},
		},
		{
			name: "self-transition",
			definition: func() (*workflow.Definition, error) {
				t, _ := workflow.NewTransition("self-transition", []workflow.Place{"start"}, []workflow.Place{"start"})
				return workflow.NewDefinition(
					[]workflow.Place{"start"},
					[]workflow.Transition{*t},
				)
			},
			initialPlace: "start",
			targetPlaces: []workflow.Place{"start"},
			wantErr:      false,
			check: func(w *workflow.Workflow) error {
				if len(w.CurrentPlaces()) != 1 || w.CurrentPlaces()[0] != "start" {
					return fmt.Errorf("expected current place to be 'start', got %v", w.CurrentPlaces())
				}
				return nil
			},
		},
		{
			name: "transition with overlapping places",
			definition: func() (*workflow.Definition, error) {
				t, _ := workflow.NewTransition("overlapping", []workflow.Place{"start", "middle"}, []workflow.Place{"middle", "end"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "middle", "end"},
					[]workflow.Transition{*t},
				)
			},
			initialPlace: "start",
			targetPlaces: []workflow.Place{"middle", "end"},
			wantErr:      true,
		},
		{
			name: "invalid transition - empty target",
			definition: func() (*workflow.Definition, error) {
				return workflow.NewDefinition(
					[]workflow.Place{"start"},
					[]workflow.Transition{},
				)
			},
			initialPlace: "start",
			targetPlaces: []workflow.Place{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := tt.definition()
			if err != nil {
				t.Fatalf("failed to create definition: %v", err)
			}
			wf, _ := workflow.NewWorkflow("test", def, tt.initialPlace)

			err = wf.ApplyWithContext(context.Background(), tt.targetPlaces)
			if (err != nil) != tt.wantErr {
				t.Errorf("Workflow.Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				if err := tt.check(wf); err != nil {
					t.Errorf("Workflow.Apply() place check failed: %v", err)
				}
			}
		})
	}
}

func TestWorkflow_GetEnabledTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialPlace workflow.Place
		transitions  []struct {
			from []workflow.Place
			to   []workflow.Place
		}
		defPlaces []workflow.Place
		defTrans  []struct {
			name string
			from []workflow.Place
			to   []workflow.Place
		}
		wantTransitions []string
		wantErr         bool
	}{
		{
			name:         "single enabled transition",
			initialPlace: "a",
			defPlaces:    []workflow.Place{"a", "b", "c"},
			defTrans: []struct {
				name string
				from []workflow.Place
				to   []workflow.Place
			}{
				{"to-b", []workflow.Place{"a"}, []workflow.Place{"b"}},
				{"to-c", []workflow.Place{"b"}, []workflow.Place{"c"}},
			},
			transitions: []struct {
				from []workflow.Place
				to   []workflow.Place
			}{
				{from: []workflow.Place{"a"}, to: []workflow.Place{"b"}},
			},
			wantTransitions: []string{"to-c"},
			wantErr:         false,
		},
		{
			name:         "multiple enabled transitions",
			initialPlace: "a",
			defPlaces:    []workflow.Place{"a", "b", "c", "d"},
			defTrans: []struct {
				name string
				from []workflow.Place
				to   []workflow.Place
			}{
				{"to-b", []workflow.Place{"a"}, []workflow.Place{"b"}},
				{"to-c", []workflow.Place{"a"}, []workflow.Place{"c"}},
				{"to-d", []workflow.Place{"b"}, []workflow.Place{"d"}},
				{"b-to-c", []workflow.Place{"b"}, []workflow.Place{"c"}},
			},
			transitions: []struct {
				from []workflow.Place
				to   []workflow.Place
			}{
				{from: []workflow.Place{"a"}, to: []workflow.Place{"b"}},
			},
			wantTransitions: []string{"to-d", "b-to-c"},
			wantErr:         false,
		},
		{
			name:         "no enabled transitions",
			initialPlace: "c",
			defPlaces:    []workflow.Place{"a", "b", "c"},
			defTrans: []struct {
				name string
				from []workflow.Place
				to   []workflow.Place
			}{
				{"to-b", []workflow.Place{"a"}, []workflow.Place{"b"}},
				{"to-c", []workflow.Place{"b"}, []workflow.Place{"c"}},
			},
			transitions: []struct {
				from []workflow.Place
				to   []workflow.Place
			}{
				{from: []workflow.Place{"a"}, to: []workflow.Place{"b"}},
			},
			wantTransitions: []string{},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var defTransObjs []workflow.Transition
			for _, tr := range tt.defTrans {
				trObj, err := workflow.NewTransition(tr.name, tr.from, tr.to)
				if err != nil {
					t.Fatalf("failed to create transition: %v", err)
				}
				defTransObjs = append(defTransObjs, *trObj)
			}
			definition, err := workflow.NewDefinition(tt.defPlaces, defTransObjs)
			if err != nil {
				t.Fatalf("failed to create definition: %v", err)
			}
			wf, _ := workflow.NewWorkflow("test", definition, tt.initialPlace)

			for _, trans := range tt.transitions {
				err := wf.ApplyWithContext(context.Background(), trans.to)
				if (err != nil) != tt.wantErr {
					t.Errorf("Workflow.Apply() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr {
					return
				}
			}

			got, err := wf.EnabledTransitions()
			if err != nil {
				t.Errorf("Workflow.EnabledTransitions() error = %v", err)
				return
			}
			if len(got) != len(tt.wantTransitions) {
				t.Errorf("Workflow.EnabledTransitions() = %v, want %v", got, tt.wantTransitions)
				return
			}

			// Check that all expected transitions are present
			for _, want := range tt.wantTransitions {
				found := false
				for _, trans := range got {
					if trans.Name() == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Workflow.EnabledTransitions() missing transition %v", want)
				}
			}
		})
	}
}

func TestWorkflow_Events(t *testing.T) {
	tr, _ := workflow.NewTransition("to-end", []workflow.Place{"start"}, []workflow.Place{"end"})
	definition, err := workflow.NewDefinition(
		[]workflow.Place{"start", "end"},
		[]workflow.Transition{*tr},
	)
	if err != nil {
		t.Fatalf("failed to create definition: %v", err)
	}
	wf, _ := workflow.NewWorkflow("test", definition, "start")

	var beforeTransitionCalled bool
	var afterTransitionCalled bool
	var guardCalled bool

	wf.AddEventListener(workflow.EventBeforeTransition, func(event workflow.Event) error {
		beforeTransitionCalled = true
		return nil
	})

	wf.AddEventListener(workflow.EventAfterTransition, func(event workflow.Event) error {
		afterTransitionCalled = true
		return nil
	})

	wf.AddGuardEventListener(func(event *workflow.GuardEvent) error {
		guardCalled = true
		return nil
	})

	err = wf.ApplyWithContext(context.Background(), []workflow.Place{"end"})
	if err != nil {
		t.Errorf("Workflow.Apply() error = %v", err)
	}

	if !beforeTransitionCalled {
		t.Error("before transition event was not called")
	}
	if !afterTransitionCalled {
		t.Error("after transition event was not called")
	}
	if !guardCalled {
		t.Error("guard event was not called")
	}
}

func TestWorkflow_Context(t *testing.T) {
	tr, _ := workflow.NewTransition("to-end", []workflow.Place{"start"}, []workflow.Place{"end"})
	definition, err := workflow.NewDefinition(
		[]workflow.Place{"start", "end"},
		[]workflow.Transition{*tr},
	)
	if err != nil {
		t.Fatalf("failed to create definition: %v", err)
	}
	wf, _ := workflow.NewWorkflow("test", definition, "start")

	// Test setting and getting context
	wf.SetContext("key", "value")
	value, ok := wf.Context("key")
	if !ok || value != "value" {
		t.Errorf("Context() = %v, %v, want %v, %v", value, ok, "value", true)
	}

	// Test getting non-existent context
	_, ok = wf.Context("non-existent")
	if ok {
		t.Error("Context() = true, want false")
	}
}

func TestWorkflow_ForkAndMerge(t *testing.T) {
	tests := []struct {
		name         string
		definition   func() (*workflow.Definition, error)
		initialPlace workflow.Place
		transitions  []struct {
			from []workflow.Place
			to   []workflow.Place
		}
		wantErr bool
		check   func(*workflow.Workflow) error
	}{
		{
			name: "simple fork and merge",
			definition: func() (*workflow.Definition, error) {
				t1, _ := workflow.NewTransition("fork", []workflow.Place{"start"}, []workflow.Place{"branch1", "branch2"})
				t2, _ := workflow.NewTransition("merge", []workflow.Place{"branch1", "branch2"}, []workflow.Place{"end"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "branch1", "branch2", "end"},
					[]workflow.Transition{*t1, *t2},
				)
			},
			initialPlace: "start",
			transitions: []struct {
				from []workflow.Place
				to   []workflow.Place
			}{
				{from: []workflow.Place{"start"}, to: []workflow.Place{"branch1", "branch2"}},
				{from: []workflow.Place{"branch1", "branch2"}, to: []workflow.Place{"end"}},
			},
			wantErr: false,
			check: func(w *workflow.Workflow) error {
				places := w.CurrentPlaces()
				if len(places) != 1 || places[0] != "end" {
					return fmt.Errorf("expected final place to be 'end', got %v", places)
				}
				return nil
			},
		},
		{
			name: "complex fork and merge",
			definition: func() (*workflow.Definition, error) {
				t1, _ := workflow.NewTransition("fork1", []workflow.Place{"start"}, []workflow.Place{"branch1", "branch2"})
				t2, _ := workflow.NewTransition("fork2", []workflow.Place{"branch1"}, []workflow.Place{"branch3"})
				t3, _ := workflow.NewTransition("merge1", []workflow.Place{"branch2", "branch3"}, []workflow.Place{"merge1", "branch1"})
				t4, _ := workflow.NewTransition("merge2", []workflow.Place{"merge1", "branch1"}, []workflow.Place{"merge2"})
				t5, _ := workflow.NewTransition("end", []workflow.Place{"merge2"}, []workflow.Place{"end"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "branch1", "branch2", "branch3", "merge1", "merge2", "end"},
					[]workflow.Transition{*t1, *t2, *t3, *t4, *t5},
				)
			},
			initialPlace: "start",
			transitions: []struct {
				from []workflow.Place
				to   []workflow.Place
			}{
				{from: []workflow.Place{"start"}, to: []workflow.Place{"branch1", "branch2"}},
				{from: []workflow.Place{"branch1"}, to: []workflow.Place{"branch3"}},
				{from: []workflow.Place{"branch2", "branch3"}, to: []workflow.Place{"merge1", "branch1"}},
				{from: []workflow.Place{"merge1", "branch1"}, to: []workflow.Place{"merge2"}},
				{from: []workflow.Place{"merge2"}, to: []workflow.Place{"end"}},
			},
			wantErr: false,
			check: func(w *workflow.Workflow) error {
				places := w.CurrentPlaces()
				if len(places) != 1 || places[0] != "end" {
					return fmt.Errorf("expected final place to be 'end', got %v", places)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := tt.definition()
			if err != nil {
				t.Fatalf("failed to create definition: %v", err)
			}
			wf, _ := workflow.NewWorkflow("test", def, tt.initialPlace)

			for _, trans := range tt.transitions {
				err := wf.ApplyWithContext(context.Background(), trans.to)
				if (err != nil) != tt.wantErr {
					t.Errorf("Workflow.Apply() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr {
					return
				}
			}

			if tt.check != nil {
				if err := tt.check(wf); err != nil {
					t.Errorf("Workflow.Apply() place check failed: %v", err)
				}
			}
		})
	}
}

func TestWorkflow_Diagram(t *testing.T) {
	tests := []struct {
		name         string
		definition   func() (*workflow.Definition, error)
		initialPlace workflow.Place
		want         string
	}{
		{
			name: "simple workflow",
			definition: func() (*workflow.Definition, error) {
				t, _ := workflow.NewTransition("to-end", []workflow.Place{"start"}, []workflow.Place{"end"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "end"},
					[]workflow.Transition{*t},
				)
			},
			initialPlace: "start",
			want: `stateDiagram-v2
    classDef currentPlace font-weight:bold,stroke-width:4px
    start
    end
    start --> end : to-end

    %% Current places
    class start currentPlace

    %% Initial place
    [*] --> start
`,
		},
		{
			name: "complex workflow",
			definition: func() (*workflow.Definition, error) {
				t1, _ := workflow.NewTransition("to-middle", []workflow.Place{"start"}, []workflow.Place{"middle"})
				t2, _ := workflow.NewTransition("to-end", []workflow.Place{"middle"}, []workflow.Place{"end"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "middle", "end"},
					[]workflow.Transition{*t1, *t2},
				)
			},
			initialPlace: "start",
			want: `stateDiagram-v2
    classDef currentPlace font-weight:bold,stroke-width:4px
    start
    middle
    end
    start --> middle : to-middle
    middle --> end : to-end

    %% Current places
    class start currentPlace

    %% Initial place
    [*] --> start
`,
		},
		{
			name: "fork and merge workflow",
			definition: func() (*workflow.Definition, error) {
				t1, _ := workflow.NewTransition("fork", []workflow.Place{"start"}, []workflow.Place{"branch1", "branch2"})
				t2, _ := workflow.NewTransition("merge", []workflow.Place{"branch1", "branch2"}, []workflow.Place{"end"})
				return workflow.NewDefinition(
					[]workflow.Place{"start", "branch1", "branch2", "end"},
					[]workflow.Transition{*t1, *t2},
				)
			},
			initialPlace: "start",
			want: `stateDiagram-v2
    classDef currentPlace font-weight:bold,stroke-width:4px
    start
    branch1
    branch2
    end
    state fork_fork <<fork>>
    start --> fork_fork : fork
    fork_fork --> branch1
    fork_fork --> branch2
    state merge_join <<join>>
    branch1 --> merge_join : merge
    branch2 --> merge_join : merge
    merge_join --> end

    %% Current places
    class start currentPlace

    %% Initial place
    [*] --> start
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := tt.definition()
			if err != nil {
				t.Fatalf("failed to create definition: %v", err)
			}
			wf, err := workflow.NewWorkflow("test", def, tt.initialPlace)
			if err != nil {
				t.Fatalf("failed to create workflow: %v", err)
			}

			got := wf.Diagram()
			if got != tt.want {
				t.Errorf("Diagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkflow_InitialPlace(t *testing.T) {
	tests := []struct {
		name         string
		definition   func() (*workflow.Definition, error)
		initialPlace workflow.Place
		wantPlace    workflow.Place
	}{
		{
			name: "simple workflow",
			definition: func() (*workflow.Definition, error) {
				return workflow.NewDefinition(
					[]workflow.Place{"start", "middle", "end"},
					[]workflow.Transition{},
				)
			},
			initialPlace: "start",
			wantPlace:    "start",
		},
		{
			name: "complex workflow",
			definition: func() (*workflow.Definition, error) {
				return workflow.NewDefinition(
					[]workflow.Place{"draft", "review", "approved"},
					[]workflow.Transition{},
				)
			},
			initialPlace: "draft",
			wantPlace:    "draft",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := tt.definition()
			if err != nil {
				t.Fatalf("failed to create definition: %v", err)
			}
			wf, err := workflow.NewWorkflow("test", def, tt.initialPlace)
			if err != nil {
				t.Fatalf("failed to create workflow: %v", err)
			}

			got := wf.InitialPlace()
			if got != tt.wantPlace {
				t.Errorf("GetInitialPlace() = %v, want %v", got, tt.wantPlace)
			}
		})
	}
}
