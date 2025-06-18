package workflow

import (
	"fmt"
)

// Transition represents a transition between places in the workflow
type Transition struct {
	Name        string
	From        []Place
	To          []Place
	Metadata    map[string]interface{}
	Constraints []Constraint
}

// Constraint represents a validation constraint for a transition
type Constraint interface {
	Validate(Event) error
}

// NewTransition creates a new transition
func NewTransition(name string, from []Place, to []Place) (*Transition, error) {
	if name == "" {
		return nil, fmt.Errorf("transition name cannot be empty")
	}

	if len(from) == 0 {
		return nil, fmt.Errorf("transition must have at least one 'from' place")
	}

	if len(to) == 0 {
		return nil, fmt.Errorf("transition must have at least one 'to' place")
	}

	// Check for duplicate places in from
	fromSet := make(map[Place]bool)
	for _, place := range from {
		if fromSet[place] {
			return nil, fmt.Errorf("duplicate 'from' place: %s", place)
		}
		fromSet[place] = true
	}

	// Check for duplicate places in to
	toSet := make(map[Place]bool)
	for _, place := range to {
		if toSet[place] {
			return nil, fmt.Errorf("duplicate 'to' place: %s", place)
		}
		toSet[place] = true
	}

	return &Transition{
		Name:        name,
		From:        from,
		To:          to,
		Metadata:    make(map[string]interface{}),
		Constraints: make([]Constraint, 0),
	}, nil
}

// AddConstraint adds a constraint to the transition
func (t *Transition) AddConstraint(constraint Constraint) {
	t.Constraints = append(t.Constraints, constraint)
}

// SetMetadata sets metadata for the transition
func (t *Transition) SetMetadata(key string, value interface{}) {
	t.Metadata[key] = value
}

// GetMetadataValue returns the value for the given key from the transition metadata
func (t *Transition) GetMetadataValue(key string) (interface{}, bool) {
	value, ok := t.Metadata[key]
	return value, ok
}

// Validate validates the transition against all constraints
func (t *Transition) Validate(event Event) error {
	for _, constraint := range t.Constraints {
		if err := constraint.Validate(event); err != nil {
			return err
		}
	}
	return nil
}
