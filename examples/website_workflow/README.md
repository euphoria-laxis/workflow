# Website Workflow Example

This example demonstrates a website content approval workflow using the workflow package. It includes a web interface for managing website content through various states: draft, review, approved, and published.

## Features

- Web interface for workflow management
- SQLite storage for workflow persistence
- Workflow manager for lifecycle management
- Mermaid diagram visualization
- Transition history tracking
- Form validation and error handling
- Clean and responsive UI with Tailwind CSS

## Workflow States

The workflow consists of the following states:

1. **Draft**: Initial state for new content
2. **Review**: Content is under review
3. **Approved**: Content has been approved
4. **Published**: Content is live on the website

## Transitions

- `to_review`: Draft → Review
- `to_approved`: Review → Approved
- `to_published`: Approved → Published
- `to_draft`: Review → Draft (rejection)

## Prerequisites

- Go 1.16 or later
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ehabterra/workflow.git
cd workflow/examples/website_workflow
```

2. Install dependencies:
```bash
go mod download
```

## Running the Example

1. Start the server:
```bash
go run main.go
```

2. Open your browser and navigate to:
```
http://localhost:8080
```

## Usage

1. **Creating Content**:
   - Click "Create New Content"
   - Fill in the content details
   - Submit to create a new workflow instance

2. **Managing Content**:
   - View all content items on the main page
   - Click on a content item to view details
   - Use the available transitions to move content through the workflow
   - View the workflow diagram to understand the process

3. **Viewing History**:
   - Each content item shows its transition history
   - History includes timestamps and transition details

## Implementation Details

### Workflow Manager

The example uses the workflow manager to handle workflow lifecycle:

```go
// Initialize the workflow manager with SQLite storage
registry := workflow.NewRegistry()
storage := workflow.NewSQLiteStorage("website_workflow.db")
manager := workflow.NewManager(registry, storage)

// Create a new workflow instance
wf, err := manager.CreateWorkflow("website_approval_1", definition, "draft")
if err != nil {
    // Handle error
}

// Get a workflow instance
wf, err = manager.GetWorkflow("website_approval_1", definition)
if err != nil {
    // Handle error
}

// Save workflow state
err = manager.SaveWorkflow("website_approval_1", wf)
if err != nil {
    // Handle error
}
```

### Storage

The example uses SQLite for persistent storage:

```go
// Initialize SQLite storage
storage := workflow.NewSQLiteStorage("website_workflow.db")

// The storage interface provides these methods:
// - LoadState(id string) ([]Place, error)
// - SaveState(id string, places []Place) error
// - DeleteState(id string) error
```

### Web Interface

The web interface is built using:
- Standard Go `html/template` for templating
- Tailwind CSS for styling
- Mermaid.js for workflow visualization

## Project Structure

```
website_workflow/
├── main.go           # Main application code
├── templates/        # HTML templates
│   ├── index.html    # Main page template
│   ├── workflow.html # Workflow page template
│   └── diagram.html  # Diagram page template
├── website_workflow.db # SQLite database
└── README.md         # This file
```

## Testing

1. Create a new content item
2. Try different transitions
3. Verify the state is persisted
4. Check the transition history
5. View the workflow diagram

## Contributing

Feel free to submit issues and enhancement requests! 