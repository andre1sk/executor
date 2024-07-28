package main

import (
	"fmt"
	"log"
)

type WorkflowStep struct {
	Step   string `json:"step"`
	Action string `json:"action"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Workflow struct {
	ID           string
	Steps        []WorkflowStep
	logger       *log.Logger
	ActionRunner ActionRunner
}

func NewWorkflow(id string, steps []WorkflowStep, logger *log.Logger, actionRunner ActionRunner) *Workflow {
	return &Workflow{
		ID:           id,
		Steps:        steps,
		logger:       logger,
		ActionRunner: actionRunner,
	}
}

func (wd *Workflow) Validate() (errors []error) {
	for _, step := range wd.Steps {
		if !wd.ActionRunner.ActionExists(step.Action) {
			errors = append(errors, fmt.Errorf("%s's Action %s not found", step.Step, step.Action))
		}
	}

	return
}

func (w *Workflow) Run(payload WorkflowPayload) (WorkflowPayload, error) {
	// Copy the payload to avoid modifying the original
	cpayload := make(WorkflowPayload)
	for k, v := range payload {
		cpayload[k] = v
	}
	for _, step := range w.Steps {
		if _, ok := cpayload[step.Input]; !ok {
			return cpayload, fmt.Errorf("input %s not found", step.Input)
		}
		if _, ok := cpayload[step.Output]; ok && step.Output == "" {
			w.logger.Printf("%s input %s is empty, done early", step.Step, step.Output)
			return cpayload, nil
		}
		output, err := w.ActionRunner.RunAction(step.Action, cpayload[step.Input])
		if err != nil {
			return cpayload, err
		}
		cpayload[step.Output] = output
	}
	return cpayload, nil
}

type WorkflowPayload map[string]string

type PayloadWorkflows struct {
	WorkflowIDs []string
	Payload     WorkflowPayload
}

type WorkflowEngine interface {
	Run([]PayloadWorkflows) ([]WorkflowPayload, error)
}

// verify interface compliance
var _ WorkflowEngine = (*SimpleWorkflowEngine)(nil)

type SimpleWorkflowEngine struct {
	logger    *log.Logger
	workflows map[string]*Workflow
}

func NewSimpleWorkflowEngine(logger *log.Logger, actionRunner ActionRunner, workflowDefs map[string][]WorkflowStep) (*SimpleWorkflowEngine, error) {
	workflows := make(map[string]*Workflow)
	for id, steps := range workflowDefs {
		workflows[id] = NewWorkflow(id, steps, logger, actionRunner)
		errors := workflows[id].Validate()
		if len(errors) > 0 {
			return nil, fmt.Errorf("workflow %s has errors: %v", id, errors)
		}
	}

	return &SimpleWorkflowEngine{
		logger:    logger,
		workflows: workflows,
	}, nil
}

func (we *SimpleWorkflowEngine) Run(workflows []PayloadWorkflows) ([]WorkflowPayload, error) {
	var results []WorkflowPayload
	for _, wfPayload := range workflows {
		currentPayload := wfPayload.Payload
		for _, wfID := range wfPayload.WorkflowIDs {
			wf, ok := we.workflows[wfID]
			if !ok {
				msg := fmt.Sprintf("workflow %s not found", wfID)
				we.logger.Printf(msg)
				return results, fmt.Errorf(msg)
			}
			var err error
			currentPayload, err = wf.Run(currentPayload)
			if err != nil {
				return results, err
			}
		}
		results = append(results, currentPayload)
	}
	return results, nil
}

func LoadWorkflowDefs() (map[string][]WorkflowStep, error) {
	workflowDefs := map[string][]WorkflowStep{}
	err := ParseJsonFileInto("config/workflows.json", &workflowDefs)
	return workflowDefs, err
}
