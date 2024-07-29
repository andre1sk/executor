package main

import (
	"log"
	"sync"
)

type Executor interface {
	Enrich(alerts []Alert) ([]Alert, error)
	EnrichParallel(alerts []Alert) ([]Alert, error)
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
	se.logger.Printf("mapped payloads: %v", wfPayloads)

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

func (se *SimpleExecutor) EnrichParallel(alerts []Alert) ([]Alert, error) {
	var wfPayloads []PayloadWorkflows
	for _, alert := range alerts {
		workflows := se.ruleMatcher.Match(alert)
		wfPayloads = append(wfPayloads, PayloadWorkflows{
			WorkflowIDs: workflows,
			Payload:     WorkflowPayload(alert),
		})
	}
	se.logger.Printf("mapped payloads: %v", wfPayloads)

	var wg sync.WaitGroup
	resultsChan := make(chan WorkflowPayload, len(wfPayloads))
	errChan := make(chan error, len(wfPayloads))
	// Semaphore to limit concurrency
	semaphore := make(chan struct{}, 5)

	for _, wfPayload := range wfPayloads {
		wg.Add(1)
		go func(wfPayload PayloadWorkflows) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			res, err := se.workflowEngine.Run([]PayloadWorkflows{wfPayload})
			if err != nil {
				errChan <- err
				return
			}
			for _, r := range res {
				resultsChan <- r
			}
		}(wfPayload)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
		close(errChan)
	}()

	var results []WorkflowPayload
	var err error
	for r := range resultsChan {
		results = append(results, r)
	}

	if len(errChan) > 0 {
		for e := range errChan {
			se.logger.Printf("failed to run workflows: %v", e)
			err = e
		}
		return nil, err
	}

	se.logger.Printf("results: %v", results)

	var enrichedAlerts []Alert
	for _, result := range results {
		enrichedAlerts = append(enrichedAlerts, Alert(result))
	}
	return enrichedAlerts, nil
}
