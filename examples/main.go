package main

import (
	"fmt"
	"log"

	"github.com/ehabterra/workflow"
)

func main() {
	// Define order places
	places := []workflow.Place{
		"pending",
		"payment_processing",
		"payment_approved",
		"payment_failed",
		"inventory_check",
		"shipping",
		"delivered",
		"cancelled",
	}

	// Define transitions
	transitions := []workflow.Transition{
		createTransition("process_payment", []workflow.Place{"pending"}, []workflow.Place{"payment_processing"}),
		createTransition("payment_success", []workflow.Place{"payment_processing"}, []workflow.Place{"payment_approved"}),
		createTransition("payment_failure", []workflow.Place{"payment_processing"}, []workflow.Place{"payment_failed"}),
		createTransition("check_inventory", []workflow.Place{"payment_approved"}, []workflow.Place{"inventory_check"}),
		createTransition("start_shipping", []workflow.Place{"inventory_check"}, []workflow.Place{"shipping"}),
		createTransition("mark_delivered", []workflow.Place{"shipping"}, []workflow.Place{"delivered"}),
		createTransition("cancel_pending", []workflow.Place{"pending"}, []workflow.Place{"cancelled"}),
		createTransition("cancel_payment_processing", []workflow.Place{"payment_processing"}, []workflow.Place{"cancelled"}),
		createTransition("cancel_payment_approved", []workflow.Place{"payment_approved"}, []workflow.Place{"cancelled"}),
	}

	// Create workflow definition
	definition, err := workflow.NewDefinition(places, transitions)
	if err != nil {
		log.Fatal(err)
	}

	// Create workflow instance
	wf, err := workflow.NewWorkflow("order-123", definition, "pending")
	if err != nil {
		log.Fatal(err)
	}

	// Add order context
	wf.SetContext("order_amount", 150.0)
	wf.SetContext("customer_id", "cust-456")

	// Add event listeners for business logic
	wf.AddEventListener(workflow.EventBeforeTransition, func(event workflow.Event) error {
		log.Printf("Processing transition: %s", event.Transition())
		return nil
	})

	wf.AddEventListener(workflow.EventAfterTransition, func(event workflow.Event) error {
		log.Printf("Completed transition: %s", event.Transition())
		return nil
	})

	// Process the order
	if err := wf.Apply([]workflow.Place{"payment_processing"}); err != nil {
		log.Fatal(err)
	}

	if err := wf.Apply([]workflow.Place{"payment_approved"}); err != nil {
		log.Fatal(err)
	}

	// Get current places
	currentPlaces := wf.CurrentPlaces()
	fmt.Printf("Current places: %v\n", currentPlaces)

	// Generate workflow diagram
	diagram := wf.GenerateMermaidDiagram()
	fmt.Println(diagram)
}

func createTransition(name string, from, to []workflow.Place) workflow.Transition {
	tr, err := workflow.NewTransition(name, from, to)
	if err != nil {
		log.Fatal(err)
	}
	return *tr
}
