package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type WorkflowStep struct {
	Step   string `json:"step"`
	Action string `json:"action"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Workflow struct {
	ID     string
	Steps  []WorkflowStep
	logger *log.Logger
}

func NewWorkflow(id string, steps []WorkflowStep, logger *log.Logger) *Workflow {
	return &Workflow{
		ID:     id,
		Steps:  steps,
		logger: logger,
	}
}

func (wd *Workflow) Validate() (errors []error) {
	for _, step := range wd.Steps {
		if !ActionExists(step.Action) {
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
		output, err := RunAction(step.Action, cpayload[step.Input])
		if err != nil {
			return cpayload, err
		}
		cpayload[step.Output] = output
	}
	return cpayload, nil
}

type WorkflowPayload map[string]string

type WorkflowToPayload struct {
	WorkflowID string
	Payload    WorkflowPayload
}

type WorkflowEngine interface {
	Run([]WorkflowToPayload) ([]WorkflowPayload, error)
}

type SimpleWorkflowEngine struct {
	logger      *log.Logger
	workflows   map[string]*Workflow
	MaxParallel int
}

func NewSimpleWorkflowEngine(logger *log.Logger, maxParallel int) (*SimpleWorkflowEngine, error) {
	workflowDefs, err := LoadWorkflowDefs()
	if err != nil {
		return nil, err
	}
	workflows := make(map[string]*Workflow)
	for id, steps := range workflowDefs {
		workflows[id] = NewWorkflow(id, steps, logger)
		errors := workflows[id].Validate()
		if len(errors) > 0 {
			return nil, fmt.Errorf("workflow %s has errors: %v", id, errors)
		}
	}

	return &SimpleWorkflowEngine{
		logger:      logger,
		workflows:   workflows,
		MaxParallel: maxParallel,
	}, nil
}

func (we *SimpleWorkflowEngine) Run(workflows []WorkflowToPayload) ([]WorkflowPayload, error) {
	var err error
	results := make([]WorkflowPayload, len(workflows))
	for i, wf := range workflows {
		w, ok := we.workflows[wf.WorkflowID]
		if !ok {
			msg := fmt.Sprintf("workflow %s not found", wf.WorkflowID)
			we.logger.Printf(msg)
			return results, fmt.Errorf(msg)
		}
		results[i], err = w.Run(wf.Payload)
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

func LoadWorkflowDefs() (map[string][]WorkflowStep, error) {
	workflowDefs := map[string][]WorkflowStep{}
	jsonFile, err := os.Open("config/workflows.json")
	if err != nil {
		return workflowDefs, fmt.Errorf("failed to open JSON file: %v", err)
	}
	defer jsonFile.Close()

	// Read the JSON file
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return workflowDefs, fmt.Errorf("failed to read JSON file: %v", err)
	}

	err = json.Unmarshal(byteValue, &workflowDefs)
	if err != nil {
		return workflowDefs, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return workflowDefs, nil
}
