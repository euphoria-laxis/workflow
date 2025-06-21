# Order Processing Workflow Example

This example demonstrates a comprehensive order processing workflow using the Petri net-based workflow engine. It showcases realistic business logic including payment processing, inventory management, and order fulfillment.

## Workflow Overview

The order processing workflow consists of the following states and transitions:

### States (Places)
- **pending**: Initial state when order is created
- **payment_processing**: Payment is being processed
- **payment_approved**: Payment was successful
- **payment_failed**: Payment failed
- **inventory_check**: Checking if items are in stock
- **inventory_insufficient**: Not enough stock available
- **shipping**: Order is being shipped
- **delivered**: Order has been delivered
- **cancelled**: Order was cancelled
- **refunded**: Order was refunded

### Key Transitions
- `process_payment`: pending → payment_processing
- `payment_success`: payment_processing → payment_approved
- `payment_failure`: payment_processing → payment_failed
- `retry_payment`: payment_failed → payment_processing
- `check_inventory`: payment_approved → inventory_check
- `inventory_available`: inventory_check → shipping
- `inventory_insufficient`: inventory_check → inventory_insufficient
- `restock_and_ship`: inventory_insufficient → shipping
- `mark_delivered`: shipping → delivered
- `refund_order`: delivered → refunded

## Business Logic Components

### Order Management
- **Order**: Represents an order with items, customer info, and status
- **OrderItem**: Individual items in an order with quantity and pricing

### Payment Processing
- **PaymentProcessor**: Simulates payment processing with configurable success rate
- Handles payment retries and failure scenarios
- Simulates processing delays for realism

### Inventory Management
- **Inventory**: Manages product stock levels
- Checks availability before shipping
- Handles restocking scenarios
- Prevents overselling

## Features Demonstrated

1. **Event-Driven Architecture**: Uses workflow events to trigger business logic
2. **Error Handling**: Comprehensive error handling for each transition
3. **Context Management**: Stores order data and business state in workflow context
4. **Realistic Simulation**: Simulates real-world scenarios like payment failures and inventory shortages
5. **State Persistence**: Maintains order status throughout the workflow
6. **Visualization**: Generates Mermaid diagrams for workflow visualization

## Running the Example

```bash
cd examples/order_processing
go run main.go
```

## Example Output

The example will demonstrate:

1. **Order Creation**: Creates a sample order with multiple items
2. **Payment Processing**: Simulates payment with success/failure scenarios
3. **Inventory Check**: Verifies stock availability
4. **Shipping Process**: Handles the shipping workflow
5. **Delivery Confirmation**: Marks order as delivered
6. **Status Reporting**: Shows final order status and workflow diagram

## Key Learning Points

- **Workflow Design**: How to design complex business workflows with multiple paths
- **Event Handling**: Using events to trigger business logic at specific workflow stages
- **Error Recovery**: Handling failure scenarios and retry logic
- **State Management**: Managing complex state across multiple workflow stages
- **Business Integration**: Integrating external systems (payment, inventory) with workflows

## Extending the Example

You can extend this example by:

1. **Adding Database Storage**: Integrate with the WorkflowManager for persistent storage
2. **Web Interface**: Add HTTP endpoints for order management
3. **Notification System**: Add email/SMS notifications at key stages
4. **Analytics**: Track workflow performance and bottlenecks
5. **Parallel Processing**: Add concurrent inventory checks for multiple warehouses

## Integration with Workflow Manager

This example can be enhanced with the WorkflowManager for production use:

```go
// Create manager with SQLite storage
manager := workflow.NewManager(storage.NewSQLiteStorage("orders.db"))

// Save workflow instance
err := manager.SaveWorkflow(wf)

// Load workflow instance
wf, err := manager.LoadWorkflow("order-123")
```

This provides persistent storage, workflow instance management, and better scalability for production environments. 