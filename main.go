package main

import (
	"log"
	"os"
)

func main() {
	var logger = log.New(os.Stdout, "", log.LstdFlags)

	actionRunner, err := NewSimpleActionRunner(logger)
	if err != nil {
		log.Fatalf("Failed to create action runner: %v", err)
	}

	workflowDefs, err := LoadWorkflowDefs()
	if err != nil {
		log.Fatalf("Failed to load workflow definitions: %v", err)
	}

	workflowEngine, err := NewSimpleWorkflowEngine(logger, actionRunner, workflowDefs)
	if err != nil {
		log.Fatalf("Failed to create workflow engine: %v", err)
	}

	rules, err := LoadRules()
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}
	ruleMatcher := NewBasicRuleMatcher(rules)

	executor := NewSimpleExecutor(ruleMatcher, workflowEngine, logger)

	apiServer := NewAPIServer(executor, logger)
	if err := apiServer.Start(); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}

}
