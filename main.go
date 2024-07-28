package main

import (
	"log"
	"os"
)

func main() {
	var logger = log.New(os.Stdout, "", log.LstdFlags)

	workflowEngine, err := NewSimpleWorkflowEngine(logger, 10)
	if err != nil {
		log.Fatalf("Failed to create workflow engine: %v", err)
	}

	// Run the workflows
	payloads := []WorkflowToPayload{
		{
			WorkflowID: "phishing-email",
			Payload: WorkflowPayload{
				"alert":    "Phishing",
				"email-id": "e2"},
		},
	}
	results, err := workflowEngine.Run(payloads)
	if err != nil {
		log.Fatalf("Failed to run workflows: %v", err)
	}

	// Print the results
	for i, result := range results {
		logger.Printf("Workflow %d result: %v\n", i, result)
	}

}
