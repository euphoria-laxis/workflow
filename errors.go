package workflow

import "fmt"

// Common errors
var (
	ErrTransitionNotAllowed = fmt.Errorf("transition not allowed")
	ErrInvalidPlace         = fmt.Errorf("invalid place")
	ErrInvalidTransition    = fmt.Errorf("invalid transition")
)
