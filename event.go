package workflow

// EventType represents the type of workflow event
type EventType string

const (
	// EventBeforeTransition is fired before a transition is applied
	EventBeforeTransition EventType = "before_transition"
	// EventAfterTransition is fired after a transition is applied
	EventAfterTransition EventType = "after_transition"
	// EventGuard is fired to check if a transition is allowed
	EventGuard EventType = "guard"
)

// Event defines the common interface for all event types
type Event interface {
	Type() EventType
	Transition() *Transition
	From() []Place
	To() []Place
	Workflow() *Workflow
	Context(key string) (interface{}, bool)
}

// BaseEvent represents a workflow event
type BaseEvent struct {
	eventType  EventType
	transition *Transition
	from       []Place
	to         []Place
	workflow   *Workflow
	context    map[string]interface{}
}

// NewEvent creates a new BaseEvent instance
func NewEvent(eventType EventType, transition *Transition, from []Place, to []Place, workflow *Workflow, context map[string]interface{}) *BaseEvent {
	return &BaseEvent{
		eventType:  eventType,
		transition: transition,
		from:       from,
		to:         to,
		workflow:   workflow,
		context:    context,
	}
}

// Type returns the event type
func (e *BaseEvent) Type() EventType {
	return e.eventType
}

// Transition returns the transition associated with the event
func (e *BaseEvent) Transition() *Transition {
	return e.transition
}

// From returns the source places of the transition
func (e *BaseEvent) From() []Place {
	return e.from
}

// To returns the target places of the transition
func (e *BaseEvent) To() []Place {
	return e.to
}

// Workflow returns the workflow instance
func (e *BaseEvent) Workflow() *Workflow {
	return e.workflow
}

// Context returns the value for the given key from the event context
func (e *BaseEvent) Context(key string) (interface{}, bool) {
	value, ok := e.context[key]
	return value, ok
}

// GuardEvent represents a guard event in the workflow
type GuardEvent struct {
	BaseEvent
	isBlocking bool
}

// NewGuardEvent creates a new Guard Event instance
func NewGuardEvent(transition *Transition, from []Place, to []Place, workflow *Workflow, context map[string]interface{}) *GuardEvent {
	return &GuardEvent{
		BaseEvent: BaseEvent{
			eventType:  EventGuard,
			transition: transition,
			from:       from,
			to:         to,
			workflow:   workflow,
			context:    context,
		},
		isBlocking: false,
	}
}

// IsBlocking returns whether the event is blocking
func (e *GuardEvent) IsBlocking() bool {
	return e.isBlocking
}

// SetBlocking sets whether the event is blocking
func (e *GuardEvent) SetBlocking(blocking bool) {
	e.isBlocking = blocking
}

// EventListener is a function that handles workflow events
type EventListener func(Event) error

// GuardEventListener is a function that handles guard events
type GuardEventListener func(*GuardEvent) error

// Listener interface for handling events
type Listener interface {
	HandleEvent(Event) error
}

// GuardListener interface for handling guard events
type GuardListener interface {
	HandleGuardEvent(*GuardEvent) error
}
