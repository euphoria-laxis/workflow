package workflow_test

import (
	"testing"

	"github.com/ehabterra/workflow"
)

func TestNewTransition(t *testing.T) {
	tests := []struct {
		name    string
		trName  string
		from    []workflow.Place
		to      []workflow.Place
		wantErr bool
	}{
		{
			name:    "valid transition",
			trName:  "valid transition",
			from:    []workflow.Place{"start"},
			to:      []workflow.Place{"end"},
			wantErr: false,
		},
		{
			name:    "empty name",
			trName:  "",
			from:    []workflow.Place{"start"},
			to:      []workflow.Place{"end"},
			wantErr: true,
		},
		{
			name:    "empty from places",
			trName:  "empty from places",
			from:    []workflow.Place{},
			to:      []workflow.Place{"end"},
			wantErr: true,
		},
		{
			name:    "empty to places",
			trName:  "empty to places",
			from:    []workflow.Place{"start"},
			to:      []workflow.Place{},
			wantErr: true,
		},
		{
			name:    "duplicate from places",
			trName:  "duplicate from places",
			from:    []workflow.Place{"start", "start"},
			to:      []workflow.Place{"end"},
			wantErr: true,
		},
		{
			name:    "duplicate to places",
			trName:  "duplicate to places",
			from:    []workflow.Place{"start"},
			to:      []workflow.Place{"end", "end"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := workflow.NewTransition(tt.trName, tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransition_Metadata(t *testing.T) {
	tr, err := workflow.NewTransition("test", []workflow.Place{"start"}, []workflow.Place{"end"})
	if err != nil {
		t.Fatalf("failed to create transition: %v", err)
	}

	// Test setting metadata
	tr.SetMetadata("key", "value")
	value, ok := tr.GetMetadataValue("key")
	if !ok {
		t.Error("metadata value not found")
	}
	if value != "value" {
		t.Errorf("metadata value = %v, want %v", value, "value")
	}

	// Test getting non-existent metadata
	_, ok = tr.GetMetadataValue("non-existent")
	if ok {
		t.Error("non-existent metadata value found")
	}
}

func TestTransition_Constraints(t *testing.T) {
	tr, err := workflow.NewTransition("test", []workflow.Place{"start"}, []workflow.Place{"end"})
	if err != nil {
		t.Fatalf("failed to create transition: %v", err)
	}

	// Create a test constraint
	testConstraint := &testConstraint{shouldFail: false}
	tr.AddConstraint(testConstraint)

	// Test validation with passing constraint
	event := workflow.NewGuardEvent(tr, []workflow.Place{"start"}, []workflow.Place{"end"}, nil, nil)
	err = tr.Validate(event)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}

	// Test validation with failing constraint
	testConstraint.shouldFail = true
	err = tr.Validate(event)
	if err == nil {
		t.Error("Validate() error = nil, want error")
	}
}

// testConstraint is a simple constraint for testing
type testConstraint struct {
	shouldFail bool
}

func (c *testConstraint) Validate(event workflow.Event) error {
	if c.shouldFail {
		return workflow.ErrTransitionNotAllowed
	}
	return nil
}
