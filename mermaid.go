package workflow

import (
	"fmt"
	"strings"
)

// GenerateMermaidDiagram generates a Mermaid state diagram for the workflow
func (w *Workflow) GenerateMermaidDiagram() string {
	var diagram strings.Builder
	diagram.WriteString("stateDiagram-v2\n")
	diagram.WriteString("    classDef currentPlace font-weight:bold,stroke-width:4px\n")

	// Add all places
	for _, place := range w.definition.Places {
		diagram.WriteString(fmt.Sprintf("    %s\n", place))
	}

	// Add all transitions
	for _, trans := range w.definition.Transitions {
		// Handle multiple to places
		if len(trans.To) > 1 {
			// This is a fork
			forkState := fmt.Sprintf("%s_fork", trans.Name)
			diagram.WriteString(fmt.Sprintf("    state %s <<fork>>\n", forkState))
			if len(trans.From) > 1 {
				// This is a join
				joinState := fmt.Sprintf("%s_join", trans.Name)
				diagram.WriteString(fmt.Sprintf("    state %s <<join>>\n", joinState))
				for _, from := range trans.From {
					diagram.WriteString(fmt.Sprintf("    %s --> %s : %s\n", from, joinState, trans.Name))
				}
				diagram.WriteString(fmt.Sprintf("    %s --> %s\n", joinState, forkState))
			} else {
				diagram.WriteString(fmt.Sprintf("    %s --> %s : %s\n", trans.From[0], forkState, trans.Name))
			}
			for _, to := range trans.To {
				diagram.WriteString(fmt.Sprintf("    %s --> %s\n", forkState, to))
			}
		} else {
			if len(trans.From) > 1 {
				// This is a join
				joinState := fmt.Sprintf("%s_join", trans.Name)
				diagram.WriteString(fmt.Sprintf("    state %s <<join>>\n", joinState))
				for _, from := range trans.From {
					diagram.WriteString(fmt.Sprintf("    %s --> %s : %s\n", from, joinState, trans.Name))
				}
				diagram.WriteString(fmt.Sprintf("    %s --> %s\n", joinState, trans.To[0]))
			} else {
				// Regular transition
				diagram.WriteString(fmt.Sprintf("    %s --> %s : %s\n", trans.From[0], trans.To[0], trans.Name))
			}
		}
	}

	// Add current place highlighting
	currentPlaces := w.marking.Places()
	if len(currentPlaces) > 0 {
		diagram.WriteString("\n    %% Current places\n")
		for _, place := range currentPlaces {
			diagram.WriteString(fmt.Sprintf("    class %s currentPlace\n", place))
		}
	}

	diagram.WriteString("\n    %% Initial place\n")
	diagram.WriteString(fmt.Sprintf("    [*] --> %s\n", w.InitialPlace()))

	return diagram.String()
}
