package main

import "log"

type Executor interface {
	Enrich(alerts []Alert) ([]Alert, error)
}

type SimpleExecutor struct {
	ruleMatcher    RuleMatcher
	workflowEngine WorkflowEngine
	logger         *log.Logger
}

// verify interface compliance
var _ Executor = (*SimpleExecutor)(nil)

func NewSimpleExecutor(ruleMatcher RuleMatcher, workflowEngine WorkflowEngine, logger *log.Logger) *SimpleExecutor {
	return &SimpleExecutor{
		ruleMatcher:    ruleMatcher,
		workflowEngine: workflowEngine,
		logger:         logger,
	}
}

func (se *SimpleExecutor) Enrich(alerts []Alert) ([]Alert, error) {

	var wfPayloads []PayloadWorkflows
	for _, alert := range alerts {
		workflows := se.ruleMatcher.Match(alert)
		wfPayloads = append(wfPayloads, PayloadWorkflows{
			WorkflowIDs: workflows,
			Payload:     WorkflowPayload(alert),
		})
	}
	se.logger.Printf("wfPayloads: %v", wfPayloads)

	results, err := se.workflowEngine.Run(wfPayloads)
	if err != nil {
		se.logger.Printf("failed to run workflows: %v", err)
		return nil, err
	}
	se.logger.Printf("results: %v", results)

	// Convert results to Alerts
	var enrichedAlerts []Alert
	for _, result := range results {
		enrichedAlerts = append(enrichedAlerts, Alert(result))
	}
	return enrichedAlerts, nil
}
