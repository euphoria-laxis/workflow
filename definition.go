package workflow

import (
	"fmt"
)

// Definition represents a workflow definition with places and transitions
type Definition struct {
	Places      []Place
	Transitions []Transition
}

// NewDefinition creates a new workflow definition
func NewDefinition(places []Place, transitions []Transition) (*Definition, error) {
	// Create a map of valid places for quick lookup
	validPlaces := make(map[Place]bool)
	for _, place := range places {
		validPlaces[place] = true
	}

	// Validate that all places in transitions exist in places
	for _, trans := range transitions {
		// Validate 'from' places
		for _, place := range trans.From {
			if !validPlaces[place] {
				return nil, fmt.Errorf("place '%s' in transition '%s' is not defined in workflow places", place, trans.Name)
			}
		}
		// Validate 'to' places
		for _, place := range trans.To {
			if !validPlaces[place] {
				return nil, fmt.Errorf("place '%s' in transition '%s' is not defined in workflow places", place, trans.Name)
			}
		}
	}

	return &Definition{
		Places:      places,
		Transitions: transitions,
	}, nil
}

// AllPlaces returns all places (places) in the definition
func (d *Definition) AllPlaces() []Place {
	places := make([]Place, len(d.Places))
	copy(places, d.Places)
	return places
}

// AllTransitions returns all transitions in the definition
func (d *Definition) AllTransitions() []Transition {
	transitions := make([]Transition, len(d.Transitions))
	copy(transitions, d.Transitions)
	return transitions
}

// Transition returns a transition by name
func (d *Definition) Transition(name string) *Transition {
	for _, t := range d.Transitions {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// Place checks if a place exists in the definition
func (d *Definition) Place(place Place) bool {
	for _, p := range d.Places {
		if p == place {
			return true
		}
	}
	return false
}
