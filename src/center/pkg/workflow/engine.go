package workflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// WorkflowState represents workflow execution state
type WorkflowState string

const (
	WorkflowStatePending   WorkflowState = "pending"
	WorkflowStateRunning   WorkflowState = "running"
	WorkflowStatePaused    WorkflowState = "paused"
	WorkflowStateCompleted WorkflowState = "completed"
	WorkflowStateFailed    WorkflowState = "failed"
	WorkflowStateCancelled WorkflowState = "cancelled"
)

// StepState represents step execution state
type StepState string

const (
	StepStatePending   StepState = "pending"
	StepStateRunning   StepState = "running"
	StepStateCompleted StepState = "completed"
	StepStateFailed    StepState = "failed"
	StepStateSkipped   StepState = "skipped"
)

// WorkflowDefinition defines a workflow
type WorkflowDefinition struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Steps       []StepDef     `json:"steps"`
	Triggers    []Trigger     `json:"triggers"`
	Timeout     time.Duration `json:"timeout"`
	RetryPolicy *RetryPolicy  `json:"retry_policy"`
	Metadata    map[string]string `json:"metadata"`
}

// StepDef defines a workflow step
type StepDef struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        StepType      `json:"type"`     // task, condition, parallel, wait, subworkflow
	SkillID     string        `json:"skill_id"`
	Operation   string        `json:"operation"`
	Params      map[string]interface{} `json:"params"`
	Depends     []string      `json:"depends"`  // Step IDs this depends on
	Condition   *Condition    `json:"condition"`
	Timeout     time.Duration `json:"timeout"`
	RetryPolicy *RetryPolicy  `json:"retry_policy"`
	OnError     ErrorAction   `json:"on_error"` // fail, skip, retry
}

// StepType defines step types
type StepType string

const (
	StepTypeTask       StepType = "task"
	StepTypeCondition  StepType = "condition"
	StepTypeParallel   StepType = "parallel"
	StepTypeWait       StepType = "wait"
	StepTypeSubworkflow StepType = "subworkflow"
)

// ErrorAction defines error handling actions
type ErrorAction string

const (
	ErrorActionFail  ErrorAction = "fail"
	ErrorActionSkip  ErrorAction = "skip"
	ErrorActionRetry ErrorAction = "retry"
)

// Condition defines a conditional expression
type Condition struct {
	Type      ConditionType `json:"type"` // equals, not_equals, greater, less, contains, exists
	Variable  string        `json:"variable"`
	Value     interface{}   `json:"value"`
	ThenStep  string        `json:"then_step"`  // Step to execute if true
	ElseStep  string        `json:"else_step"`  // Step to execute if false
}

// ConditionType defines condition types
type ConditionType string

const (
	ConditionEquals     ConditionType = "equals"
	ConditionNotEquals  ConditionType = "not_equals"
	ConditionGreater    ConditionType = "greater"
	ConditionLess       ConditionType = "less"
	ConditionContains   ConditionType = "contains"
	ConditionExists     ConditionType = "exists"
)

// Trigger defines workflow trigger
type Trigger struct {
	Type      TriggerType `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Enabled   bool        `json:"enabled"`
}

// TriggerType defines trigger types
type TriggerType string

const (
	TriggerTypeManual   TriggerType = "manual"
	TriggerTypeSchedule TriggerType = "schedule"
	TriggerTypeEvent    TriggerType = "event"
	TriggerTypeMessage  TriggerType = "message"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries  int           `json:"max_retries"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay    time.Duration `json:"max_delay"`
	Multiplier  float64       `json:"multiplier"`
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflow_id"`
	WorkflowVer  string                 `json:"workflow_version"`
	State        WorkflowState          `json:"state"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  time.Time              `json:"completed_at"`
	CurrentStep  string                 `json:"current_step"`
	StepStates   map[string]StepState   `json:"step_states"`
	StepResults  map[string]interface{} `json:"step_results"`
	Variables    map[string]interface{} `json:"variables"`
	Error        string                 `json:"error"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	TriggeredBy  string                 `json:"triggered_by"`
}

// StepExecution represents a step execution instance
type StepExecution struct {
	StepID       string        `json:"step_id"`
	ExecutionID  string        `json:"execution_id"`
	State        StepState     `json:"state"`
	StartedAt    time.Time     `json:"started_at"`
	CompletedAt  time.Time     `json:"completed_at"`
	Result       interface{}   `json:"result"`
	Error        string        `json:"error"`
	RetryCount   int           `json:"retry_count"`
	TaskID       string        `json:"task_id"`    // Associated task ID if step is a task
}

// WorkflowEngine manages workflow execution
type WorkflowEngine struct {
	definitions sync.Map // map[string]*WorkflowDefinition
	executions  sync.Map // map[string]*WorkflowExecution
	steps       sync.Map // map[string]*StepExecution

	taskSubmitter TaskSubmitter

	// Scheduler for scheduled workflows
	scheduler *WorkflowScheduler

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// TaskSubmitter submits tasks to the task system
type TaskSubmitter interface {
	SubmitTask(skillID, operation string, params map[string]interface{}) (string, error)
	GetTaskResult(taskID string) (interface{}, error)
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(submitter TaskSubmitter) (*WorkflowEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &WorkflowEngine{
		taskSubmitter: submitter,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Create scheduler
	engine.scheduler = NewWorkflowScheduler(engine)

	return engine, nil
}

// Start begins workflow execution
func (e *WorkflowEngine) Start() {
	e.scheduler.Start()
}

// Stop stops the workflow engine
func (e *WorkflowEngine) Stop() {
	e.scheduler.Stop()
	e.cancel()
}

// RegisterWorkflow registers a workflow definition
func (e *WorkflowEngine) RegisterWorkflow(def *WorkflowDefinition) error {
	if def.ID == "" {
		return errors.New("workflow ID required")
	}

	if len(def.Steps) == 0 {
		return errors.New("workflow must have at least one step")
	}

	// Validate step dependencies
	for _, step := range def.Steps {
		for _, dep := range step.Depends {
			found := false
			for _, s := range def.Steps {
				if s.ID == dep {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("step %s depends on non-existent step %s", step.ID, dep)
			}
		}
	}

	e.definitions.Store(def.ID, def)

	// Register triggers
	for _, trigger := range def.Triggers {
		if trigger.Enabled {
			e.scheduler.RegisterTrigger(def.ID, trigger)
		}
	}

	return nil
}

// UnregisterWorkflow removes a workflow definition
func (e *WorkflowEngine) UnregisterWorkflow(workflowID string) {
	e.definitions.Delete(workflowID)
	e.scheduler.UnregisterTriggers(workflowID)
}

// GetWorkflow returns a workflow definition
func (e *WorkflowEngine) GetWorkflow(workflowID string) (*WorkflowDefinition, bool) {
	if v, ok := e.definitions.Load(workflowID); ok {
		return v.(*WorkflowDefinition), true
	}
	return nil, false
}

// ExecuteWorkflow starts a workflow execution
func (e *WorkflowEngine) ExecuteWorkflow(workflowID string, input map[string]interface{}) (string, error) {
	def, ok := e.GetWorkflow(workflowID)
	if !ok {
		return "", fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Create execution
	execID := generateExecutionID()
	exec := &WorkflowExecution{
		ID:          execID,
		WorkflowID:  workflowID,
		WorkflowVer: def.Version,
		State:       WorkflowStatePending,
		StartedAt:   time.Now(),
		StepStates:  make(map[string]StepState),
		StepResults: make(map[string]interface{}),
		Variables:   make(map[string]interface{}),
		Input:       input,
		TriggeredBy: "manual",
	}

	// Initialize step states
	for _, step := range def.Steps {
		exec.StepStates[step.ID] = StepStatePending
	}

	e.executions.Store(execID, exec)

	// Start execution
	go e.runWorkflow(execID)

	return execID, nil
}

// runWorkflow executes a workflow
func (e *WorkflowEngine) runWorkflow(execID string) {
	exec, ok := e.getExecution(execID)
	if !ok {
		return
	}

	def, ok := e.GetWorkflow(exec.WorkflowID)
	if !ok {
		exec.State = WorkflowStateFailed
		exec.Error = "workflow definition not found"
		exec.CompletedAt = time.Now()
		return
	}

	exec.State = WorkflowStateRunning

	// Copy input to variables
	for k, v := range exec.Input {
		exec.Variables[k] = v
	}

	// Execute steps in order (respecting dependencies)
	for {
		// Find next executable step
		nextStep := e.findNextStep(def, exec)
		if nextStep == nil {
			// Check if all steps are completed or failed
			allDone := true
			for _, state := range exec.StepStates {
				if state == StepStatePending || state == StepStateRunning {
					allDone = false
					break
				}
			}

			if allDone {
				exec.State = WorkflowStateCompleted
				exec.CompletedAt = time.Now()
				e.collectOutput(def, exec)
				return
			}

			// Wait for running steps
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Execute step
		e.executeStep(execID, nextStep)
	}
}

// findNextStep finds the next step that can be executed
func (e *WorkflowEngine) findNextStep(def *WorkflowDefinition, exec *WorkflowExecution) *StepDef {
	for _, step := range def.Steps {
		if exec.StepStates[step.ID] != StepStatePending {
			continue
		}

		// Check dependencies
		allDepsComplete := true
		for _, dep := range step.Depends {
			if exec.StepStates[dep] != StepStateCompleted {
				allDepsComplete = false
				break
			}
		}

		if allDepsComplete {
			return &step
		}
	}

	return nil
}

// executeStep executes a single step
func (e *WorkflowEngine) executeStep(execID string, step *StepDef) {
	exec, _ := e.getExecution(execID)

	exec.StepStates[step.ID] = StepStateRunning
	exec.CurrentStep = step.ID

	// Resolve parameters from variables
	params := e.resolveParams(step.Params, exec.Variables)

	var result interface{}
	var err error

	switch step.Type {
	case StepTypeTask:
		result, err = e.executeTaskStep(step, params)
	case StepTypeCondition:
		result, err = e.executeConditionStep(step, exec)
	case StepTypeWait:
		result, err = e.executeWaitStep(step)
	case StepTypeParallel:
		result, err = e.executeParallelStep(execID, step)
	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	if err != nil {
		e.handleStepError(execID, step, err)
		return
	}

	exec.StepStates[step.ID] = StepStateCompleted
	exec.StepResults[step.ID] = result

	// Update variables
	if result != nil {
		exec.Variables[step.ID+"_result"] = result
	}
}

// executeTaskStep executes a task step
func (e *WorkflowEngine) executeTaskStep(step *StepDef, params map[string]interface{}) (interface{}, error) {
	taskID, err := e.taskSubmitter.SubmitTask(step.SkillID, step.Operation, params)
	if err != nil {
		return nil, err
	}

	// Wait for task completion
	timeout := step.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(e.ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("task timeout")
		case <-time.After(100 * time.Millisecond):
			result, err := e.taskSubmitter.GetTaskResult(taskID)
			if err == nil {
				return result, nil
			}
		}
	}
}

// executeConditionStep executes a condition step
func (e *WorkflowEngine) executeConditionStep(step *StepDef, exec *WorkflowExecution) (interface{}, error) {
	if step.Condition == nil {
		return nil, errors.New("condition not defined")
	}

	cond := step.Condition
	value := exec.Variables[cond.Variable]

	matched := e.evaluateCondition(cond.Type, value, cond.Value)

	if matched {
		if cond.ThenStep != "" {
			exec.StepStates[cond.ThenStep] = StepStatePending
		}
	} else {
		if cond.ElseStep != "" {
			exec.StepStates[cond.ElseStep] = StepStatePending
		}
	}

	return matched, nil
}

// evaluateCondition evaluates a condition
func (e *WorkflowEngine) evaluateCondition(condType ConditionType, value, expected interface{}) bool {
	switch condType {
	case ConditionEquals:
		return value == expected
	case ConditionNotEquals:
		return value != expected
	case ConditionExists:
		return value != nil
	default:
		return false
	}
}

// executeWaitStep executes a wait step
func (e *WorkflowEngine) executeWaitStep(step *StepDef) (interface{}, error) {
	if step.Timeout > 0 {
		time.Sleep(step.Timeout)
	}
	return nil, nil
}

// executeParallelStep executes parallel steps
func (e *WorkflowEngine) executeParallelStep(execID string, step *StepDef) (interface{}, error) {
	// Parallel execution would spawn multiple goroutines
	// For now, simplified implementation
	return nil, nil
}

// handleStepError handles step execution error
func (e *WorkflowEngine) handleStepError(execID string, step *StepDef, err error) {
	exec, _ := e.getExecution(execID)

	switch step.OnError {
	case ErrorActionFail:
		exec.StepStates[step.ID] = StepStateFailed
		exec.State = WorkflowStateFailed
		exec.Error = err.Error()
		exec.CompletedAt = time.Now()
	case ErrorActionSkip:
		exec.StepStates[step.ID] = StepStateSkipped
		log.Printf("Step %s skipped due to error: %v", step.ID, err)
	case ErrorActionRetry:
		// Retry logic would be implemented here
		exec.StepStates[step.ID] = StepStateFailed
		exec.State = WorkflowStateFailed
		exec.Error = err.Error()
	}
}

// resolveParams resolves parameters from variables
func (e *WorkflowEngine) resolveParams(params map[string]interface{}, variables map[string]interface{}) map[string]interface{} {
	resolved := make(map[string]interface{})

	for k, v := range params {
		// Check if value is a variable reference
		if str, ok := v.(string); ok && len(str) > 2 && str[0] == '$' {
			varName := str[1:]
			if varValue, ok := variables[varName]; ok {
				resolved[k] = varValue
			} else {
				resolved[k] = v
			}
		} else {
			resolved[k] = v
		}
	}

	return resolved
}

// collectOutput collects workflow output
func (e *WorkflowEngine) collectOutput(def *WorkflowDefinition, exec *WorkflowExecution) {
	exec.Output = make(map[string]interface{})

	// Collect all step results as output
	for k, v := range exec.StepResults {
		exec.Output[k] = v
	}
}

// getExecution gets an execution by ID
func (e *WorkflowEngine) getExecution(execID string) (*WorkflowExecution, bool) {
	if v, ok := e.executions.Load(execID); ok {
		return v.(*WorkflowExecution), true
	}
	return nil, false
}

// GetExecution returns an execution by ID
func (e *WorkflowEngine) GetExecution(execID string) (*WorkflowExecution, bool) {
	return e.getExecution(execID)
}

// CancelExecution cancels a workflow execution
func (e *WorkflowEngine) CancelExecution(execID string) error {
	exec, ok := e.getExecution(execID)
	if !ok {
		return fmt.Errorf("execution not found: %s", execID)
	}

	exec.State = WorkflowStateCancelled
	exec.CompletedAt = time.Now()

	return nil
}

// PauseExecution pauses a workflow execution
func (e *WorkflowEngine) PauseExecution(execID string) error {
	exec, ok := e.getExecution(execID)
	if !ok {
		return fmt.Errorf("execution not found: %s", execID)
	}

	if exec.State == WorkflowStateRunning {
		exec.State = WorkflowStatePaused
		return nil
	}

	return errors.New("can only pause running workflow")
}

// ResumeExecution resumes a paused workflow execution
func (e *WorkflowEngine) ResumeExecution(execID string) error {
	exec, ok := e.getExecution(execID)
	if !ok {
		return fmt.Errorf("execution not found: %s", execID)
	}

	if exec.State == WorkflowStatePaused {
		exec.State = WorkflowStateRunning
		go e.runWorkflow(execID)
		return nil
	}

	return errors.New("can only resume paused workflow")
}

// ListExecutions returns all executions
func (e *WorkflowEngine) ListExecutions() []*WorkflowExecution {
	var executions []*WorkflowExecution
	e.executions.Range(func(key, value interface{}) bool {
		executions = append(executions, value.(*WorkflowExecution))
		return true
	})
	return executions
}

// ListWorkflows returns all workflow definitions
func (e *WorkflowEngine) ListWorkflows() []*WorkflowDefinition {
	var workflows []*WorkflowDefinition
	e.definitions.Range(func(key, value interface{}) bool {
		workflows = append(workflows, value.(*WorkflowDefinition))
		return true
	})
	return workflows
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return "exec-" + time.Now().Format("20060102-150405") + "-" + randomSuffix()
}

// randomSuffix generates a random suffix
func randomSuffix() string {
	return fmt.Sprintf("%04d", time.Now().Nanosecond()%10000)
}