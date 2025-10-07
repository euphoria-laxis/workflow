package workflow_test

import (
	"testing"

	"github.com/euphoria-laxis/workflow"
)

func TestNewDefinition(t *testing.T) {
	tests := []struct {
		name        string
		places      []workflow.Place
		transitions []workflow.Transition
		wantErr     bool
		errContains string
	}{
		{
			name:   "valid definition",
			places: []workflow.Place{"start", "end"},
			transitions: []workflow.Transition{
				func() workflow.Transition {
					t, _ := workflow.NewTransition("to-end", []workflow.Place{"start"}, []workflow.Place{"end"})
					return *t
				}(),
			},
			wantErr: false,
		},
		{
			name:   "invalid transition - missing from place",
			places: []workflow.Place{"end"},
			transitions: []workflow.Transition{
				func() workflow.Transition {
					t, _ := workflow.NewTransition("to-end", []workflow.Place{"start"}, []workflow.Place{"end"})
					return *t
				}(),
			},
			wantErr:     true,
			errContains: "place 'start' in transition 'to-end' is not defined in workflow places",
		},
		{
			name:   "invalid transition - missing to place",
			places: []workflow.Place{"start"},
			transitions: []workflow.Transition{
				func() workflow.Transition {
					t, _ := workflow.NewTransition("to-end", []workflow.Place{"start"}, []workflow.Place{"end"})
					return *t
				}(),
			},
			wantErr:     true,
			errContains: "place 'end' in transition 'to-end' is not defined in workflow places",
		},
		{
			name:   "invalid fork - missing place",
			places: []workflow.Place{"start", "branch1", "end"},
			transitions: []workflow.Transition{
				func() workflow.Transition {
					t, _ := workflow.NewTransition("fork", []workflow.Place{"start"}, []workflow.Place{"branch1", "non-existent"})
					return *t
				}(),
			},
			wantErr:     true,
			errContains: "place 'non-existent' in transition 'fork' is not defined in workflow places",
		},
		{
			name:   "invalid merge - missing place",
			places: []workflow.Place{"start", "branch1", "branch2", "end"},
			transitions: []workflow.Transition{
				func() workflow.Transition {
					t1, _ := workflow.NewTransition("fork", []workflow.Place{"start"}, []workflow.Place{"branch1", "branch2"})
					return *t1
				}(),
				func() workflow.Transition {
					t2, _ := workflow.NewTransition("merge", []workflow.Place{"branch1", "non-existent"}, []workflow.Place{"end"})
					return *t2
				}(),
			},
			wantErr:     true,
			errContains: "place 'non-existent' in transition 'merge' is not defined in workflow places",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := workflow.NewDefinition(tt.places, tt.transitions)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err.Error() != tt.errContains {
				t.Errorf("NewDefinition() error = %v, want error containing %v", err, tt.errContains)
			}
		})
	}
}
