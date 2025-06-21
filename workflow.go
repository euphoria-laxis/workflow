package workflow

import (
	"fmt"
)

// Workflow represents a workflow instance
type Workflow struct {
	name         string
	definition   *Definition
	initialPlace Place
	marking      Marking
	listeners    map[EventType][]interface{}
	context      map[string]interface{}
}

// NewWorkflow constructor
func NewWorkflow(name string, definition *Definition, initialPlace Place) (*Workflow, error) {
	if name == "" {
		return nil, fmt.Errorf("workflow name cannot be empty")
	}

	if definition == nil {
		return nil, fmt.Errorf("workflow definition cannot be nil")
	}

	if !definition.Place(initialPlace) {
		return nil, fmt.Errorf("initial place %s is not defined in the workflow", initialPlace)
	}

	marking := NewMarking([]Place{initialPlace})

	return &Workflow{
		name:         name,
		definition:   definition,
		initialPlace: initialPlace,
		marking:      marking,
		listeners:    make(map[EventType][]interface{}),
		context:      make(map[string]interface{}),
	}, nil
}

// Name returns the workflow name
func (w *Workflow) Name() string {
	return w.name
}

// AddEventListener adds an event listener for a specific event type
func (w *Workflow) AddEventListener(eventType EventType, listener EventListener) {
	w.listeners[eventType] = append(w.listeners[eventType], listener)
}

// AddGuardEventListener adds a guard event listener
func (w *Workflow) AddGuardEventListener(listener GuardEventListener) {
	eventType := EventGuard
	w.listeners[eventType] = append(w.listeners[eventType], listener)
}

// RemoveEventListener removes an event listener
func (w *Workflow) RemoveEventListener(eventType EventType, listener interface{}) {
	listeners := w.listeners[eventType]
	for i, l := range listeners {
		if &l == &listener {
			w.listeners[eventType] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

// SetContext sets a value in the workflow context
func (w *Workflow) SetContext(key string, value interface{}) {
	w.context[key] = value
}

// Context returns the value for the given key from the workflow context
func (w *Workflow) Context(key string) (interface{}, bool) {
	value, ok := w.context[key]
	return value, ok
}

// Can check if transition to target places is possible
func (w *Workflow) Can(to []Place) error {
	// Check if transition is valid
	if len(to) == 0 {
		return ErrInvalidTransition
	}

	// Validate that all target places exist in workflow places
	for _, place := range to {
		if !w.definition.Place(place) {
			return ErrInvalidPlace
		}
	}

	// Get enabled transitions
	enabled, err := w.EnabledTransitions()
	if err != nil {
		return err
	}

	// Check if any enabled transition leads to the target places
	for _, t := range enabled {
		if len(t.To()) == len(to) {
			matches := true
			for i := range t.To() {
				if t.To()[i] != to[i] {
					matches = false
					break
				}
			}
			if matches {
				// Create guard event for validation
				event := NewGuardEvent(&t, w.marking.Places(), to, w, w.context)

				// First, validate transition constraints
				if err = t.validate(event); err != nil {
					return err
				}

				// Then, fire guard event listeners
				for _, listener := range w.listeners[EventGuard] {
					if err = listener.(GuardEventListener)(event); err != nil {
						return err
					}
					if event.IsBlocking() {
						return ErrTransitionNotAllowed
					}
				}
				return nil
			}
		}
	}

	return ErrTransitionNotAllowed
}

// Apply applies a transition to the workflow
func (w *Workflow) Apply(targetPlaces []Place) error {
	// Validate target places
	for _, place := range targetPlaces {
		if !w.definition.Place(place) {
			return ErrInvalidPlace
		}
	}

	// Check if the transition is allowed
	if err := w.Can(targetPlaces); err != nil {
		return err
	}

	// Find the transition that leads to these places
	var from []Place
	var transition *Transition
	currentPlaces := w.marking.Places()

	// Check each transition
	for _, t := range w.definition.Transitions {
		// Check if all 'from' places are in current places
		allFromPlacesPresent := true
		for _, fromPlace := range t.From() {
			found := false
			for _, place := range currentPlaces {
				if place == fromPlace {
					found = true
					break
				}
			}
			if !found {
				allFromPlacesPresent = false
				break
			}
		}

		// Check if all 'to' places match
		if allFromPlacesPresent && len(t.To()) == len(targetPlaces) {
			matches := true
			for i := range t.To() {
				if t.To()[i] != targetPlaces[i] {
					matches = false
					break
				}
			}
			if matches {
				from = t.From()
				transition = &t
				break
			}
		}
	}

	if transition == nil {
		return ErrInvalidTransition
	}

	// Fire before transition event
	event := NewEvent(EventBeforeTransition, transition, from, targetPlaces, w, w.context)
	for _, listener := range w.listeners[EventBeforeTransition] {
		if err := listener.(EventListener)(event); err != nil {
			return err
		}
	}

	// Remove the 'from' places from marking
	newPlaces := make([]Place, 0, len(currentPlaces))
	for _, place := range currentPlaces {
		found := false
		for _, fromPlace := range from {
			if place == fromPlace {
				found = true
				break
			}
		}
		if !found {
			newPlaces = append(newPlaces, place)
		}
	}

	// Add the target places to marking
	newPlaces = append(newPlaces, targetPlaces...)
	w.marking.SetPlaces(newPlaces)

	// Fire after transition event
	event = NewEvent(EventAfterTransition, transition, from, targetPlaces, w, w.context)
	for _, listener := range w.listeners[EventAfterTransition] {
		if err := listener.(EventListener)(event); err != nil {
			return err
		}
	}

	return nil
}

// EnabledTransitions returns all transitions that can be applied in the current place
func (w *Workflow) EnabledTransitions() ([]Transition, error) {
	var enabled []Transition
	currentPlaces := w.marking.Places()

	// Check each transition
	for _, trans := range w.definition.Transitions {
		// Check if all 'from' places are in current places
		allFromPlacesPresent := true
		for _, fromPlace := range trans.From() {
			found := false
			for _, place := range currentPlaces {
				if place == fromPlace {
					found = true
					break
				}
			}
			if !found {
				allFromPlacesPresent = false
				break
			}
		}

		if allFromPlacesPresent {
			enabled = append(enabled, trans)
		}
	}
	return enabled, nil
}

// CurrentPlaces returns the current places of the workflow
func (w *Workflow) CurrentPlaces() []Place {
	return w.marking.Places()
}

// Definition returns the workflow definition
func (w *Workflow) Definition() *Definition {
	return w.definition
}

// Marking returns the current marking of the workflow
func (w *Workflow) Marking() Marking {
	return w.marking
}

// SetMarking sets the workflow marking
func (w *Workflow) SetMarking(marking Marking) error {
	if marking == nil {
		return fmt.Errorf("marking cannot be nil")
	}
	w.marking = marking
	return nil
}

// InitialPlace returns the initial place of the workflow
func (w *Workflow) InitialPlace() Place {
	return w.initialPlace
}
