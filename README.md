# Go Workflow

A flexible and extensible workflow engine for Go applications. This package provides a robust implementation of a Petri net-based workflow system that supports complex state transitions, event handling, and constraints. Inspired by [Symfony Workflow Component](https://symfony.com/doc/current/workflow.html).

## Features

- Petri net-based workflow engine
- Support for multiple states and transitions
- Event system for workflow lifecycle hooks
- Constraint system for transition validation
- Thread-safe workflow registry
- Comprehensive test coverage
- Mermaid diagram generation for visualization
- Flexible storage interface for workflow persistence
- Workflow manager for lifecycle management
- Support for parallel transitions and branching

## Feature Checklist

### Current Features ✅
- [x] Basic workflow definition and execution
- [x] Multiple states and transitions
- [x] Event system for workflow hooks
- [x] Constraint system for transitions
- [x] Thread-safe workflow registry
- [x] Mermaid diagram visualization
- [x] Workflow manager for lifecycle management
- [x] Storage interface for persistence
- [x] SQLite storage implementation
- [x] Support for parallel transitions and branching
- [x] Workflow history and audit trail (in examples)
- [x] Web interface for workflow management (in examples)
- [x] REST API endpoints (in examples)

### Planned Features 🚀

#### High Priority
- [ ] YAML/JSON configuration support
- [ ] Standalone web interface for workflow management
- [ ] Enhanced REST API endpoints
- [ ] Workflow validation system
- [ ] Dynamic workflow definition loading

#### Medium Priority
- [ ] Custom scripting for transition conditions
- [ ] Workflow versioning
- [ ] Workflow templates
- [ ] Role-based access control
- [ ] Workflow timeout and scheduling

#### Low Priority
- [ ] Workflow statistics and analytics
- [ ] Export/Import workflow definitions

## Installation

```bash
go get github.com/ehabterra/workflow
```

## Quick Start

Here's a simple example of how to use the workflow package:

```go
package main

import (
    "fmt"
    "github.com/ehabterra/workflow/workflow"
)

func main() {
    // Create a workflow definition
    definition, err := workflow.NewDefinition(
        []workflow.State{"start", "middle", "end"},
		[]workflow.Transition{
			func() workflow.Transition {
				tr, _ := workflow.NewTransition("to-middle", []workflow.State{"start"}, []workflow.State{"middle"})
				return *tr
			}(),
			func() workflow.Transition {
				tr, _ := workflow.NewTransition("to-end", []workflow.State{"middle"}, []workflow.State{"end"})
				return *tr
			}(),
        },
    )
    if err != nil {
        panic(err)
    }

    // Create a new workflow
    wf, err := workflow.NewWorkflow("my-workflow", definition, "start")
    if err != nil {
        panic(err)
    }

    // Add event listeners
    wf.AddEventListener(workflow.EventBeforeTransition, func(event workflow.Event) error {
        fmt.Printf("Before transition: %s\n", event.Transition.Name)
        return nil
    })

    // Apply transitions
    err = wf.Apply([]workflow.State{"middle"})
    if err != nil {
        panic(err)
    }

    // Get current states
    states := wf.CurrentStates()
    fmt.Printf("Current states: %v\n", states)

    // Generate and print the workflow diagram
    diagram := wf.GenerateMermaidDiagram()
    fmt.Println(diagram)
}
```

## Advanced Usage

### Using the Workflow Manager

The workflow manager provides a high-level interface for managing workflow lifecycles and persistence:

```go
// Create a registry and storage
registry := workflow.NewRegistry()
storage := workflow.NewSQLiteStorage("workflows.db")

// Create a workflow manager
manager := workflow.NewManager(registry, storage)

// Create a new workflow
wf, err := manager.CreateWorkflow("my-workflow", definition, "start")
if err != nil {
    panic(err)
}

// Get a workflow (loads from storage if not in registry)
wf, err = manager.GetWorkflow("my-workflow", definition)
if err != nil {
    panic(err)
}

// Save workflow state
err = manager.SaveWorkflow("my-workflow", wf)
if err != nil {
    panic(err)
}

// Delete a workflow
err = manager.DeleteWorkflow("my-workflow")
if err != nil {
    panic(err)
}
```

### Storage Interface

The package provides a flexible storage interface for persisting workflow states:

```go
type Storage interface {
    LoadState(id string) ([]Place, error)
    SaveState(id string, places []Place) error
    DeleteState(id string) error
}
```

You can implement your own storage backend by implementing this interface. The package includes a SQLite implementation:

```go
// Create a SQLite storage
storage := workflow.NewSQLiteStorage("workflows.db")

// Use it with the workflow manager
manager := workflow.NewManager(registry, storage)
```

### Adding Constraints

You can add constraints to transitions to control when they can be applied:

```go
type MyConstraint struct{}

func (c *MyConstraint) Validate(event workflow.Event) error {
    // Add your validation logic here
    return nil
}

// Add the constraint to a transition
tr.AddConstraint(&MyConstraint{})
```

### Using the Registry

The registry allows you to manage multiple workflows:

```go
registry := workflow.NewRegistry()

// Add a workflow
err := registry.AddWorkflow(wf)

// Get a workflow
wf, err := registry.Workflow("my-workflow")

// List all workflows
names := registry.ListWorkflows()

// Get context value
value, ok := wf.Context("key")
if ok {
    fmt.Printf("Context value: %v\n", value)
}
```

### Event Types

The workflow engine supports several event types:

- `EventBeforeTransition`: Fired before a transition is applied
- `EventAfterTransition`: Fired after a transition is applied
- `EventGuard`: Fired to check if a transition is allowed

### Context

You can attach context data to workflows:

```go
wf.SetContext("key", "value")
value, ok := wf.GetContext("key")
```

### Workflow Visualization

The package includes a Mermaid diagram generator for visualizing workflows. The generated diagrams can be rendered in any Mermaid-compatible viewer (like GitHub, GitLab, or the Mermaid Live Editor).

```go
// Generate a Mermaid diagram
diagram := wf.GenerateMermaidDiagram()
fmt.Println(diagram)
```

Example output:
```mermaid
stateDiagram-v2
    start
    middle
    end
    start --> middle : to-middle
    middle --> end : to-end

    %% Current states
    state start as start
```

## Benchmarks

The package includes benchmarks for common operations. Run them with:

```bash
go test -bench=. ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 