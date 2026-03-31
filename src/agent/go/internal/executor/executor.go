package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pb "github.com/ofa/agent/proto"
)

// Skill defines the interface for a skill
type Skill interface {
	ID() string
	Name() string
	Version() string
	Category() string
	Execute(ctx context.Context, input []byte) ([]byte, error)
}

// Executor manages task execution
type Executor struct {
	skills sync.Map // map[string]Skill
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	return &Executor{}
}

// RegisterSkill registers a skill
func (e *Executor) RegisterSkill(skill Skill) {
	e.skills.Store(skill.ID(), skill)
}

// UnregisterSkill unregisters a skill
func (e *Executor) UnregisterSkill(skillID string) {
	e.skills.Delete(skillID)
}

// GetCapabilities returns all registered capabilities
func (e *Executor) GetCapabilities() []*pb.Capability {
	var caps []*pb.Capability
	e.skills.Range(func(key, value interface{}) bool {
		skill := value.(Skill)
		caps = append(caps, &pb.Capability{
			Id:       skill.ID(),
			Name:     skill.Name(),
			Version:  skill.Version(),
			Category: skill.Category(),
		})
		return true
	})
	return caps
}

// Execute executes a task
func (e *Executor) Execute(ctx context.Context, task *pb.TaskAssignment) (*pb.TaskResult, error) {
	skill, ok := e.skills.Load(task.SkillId)
	if !ok {
		return &pb.TaskResult{
			TaskId: task.TaskId,
			Status: pb.TaskStatus_TASK_STATUS_FAILED,
			Error:  fmt.Sprintf("Skill not found: %s", task.SkillId),
		}, nil
	}

	s := skill.(Skill)
	output, err := s.Execute(ctx, task.Input)
	if err != nil {
		return &pb.TaskResult{
			TaskId: task.TaskId,
			Status: pb.TaskStatus_TASK_STATUS_FAILED,
			Error:  err.Error(),
		}, nil
	}

	return &pb.TaskResult{
		TaskId: task.TaskId,
		Status: pb.TaskStatus_TASK_STATUS_COMPLETED,
		Output: output,
	}, nil
}

// ===== Built-in Skills =====

// TextProcessSkill processes text
type TextProcessSkill struct{}

func (s *TextProcessSkill) ID() string { return "text.process" }
func (s *TextProcessSkill) Name() string { return "Text Process" }
func (s *TextProcessSkill) Version() string { return "1.0.0" }
func (s *TextProcessSkill) Category() string { return "text" }

func (s *TextProcessSkill) Execute(ctx context.Context, input []byte) ([]byte, error) {
	var req struct {
		Text      string `json:"text"`
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, err
	}

	var result string
	switch req.Operation {
	case "uppercase":
		result = toUpper(req.Text)
	case "lowercase":
		result = toLower(req.Text)
	case "reverse":
		result = reverse(req.Text)
	case "length":
		result = fmt.Sprintf("%d", len(req.Text))
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}

	return json.Marshal(map[string]string{"result": result})
}

func toUpper(s string) string {
	// Simple uppercase implementation
	result := make([]byte, len(s))
	for i, c := range s {
		if c >= 'a' && c <= 'z' {
			result[i] = byte(c - 32)
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			result[i] = byte(c + 32)
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

func reverse(s string) string {
	result := make([]byte, len(s))
	for i := len(s) - 1; i >= 0; i-- {
		result[len(s)-1-i] = s[i]
	}
	return string(result)
}

// JSONProcessSkill processes JSON
type JSONProcessSkill struct{}

func (s *JSONProcessSkill) ID() string { return "json.process" }
func (s *JSONProcessSkill) Name() string { return "JSON Process" }
func (s *JSONProcessSkill) Version() string { return "1.0.0" }
func (s *JSONProcessSkill) Category() string { return "data" }

func (s *JSONProcessSkill) Execute(ctx context.Context, input []byte) ([]byte, error) {
	var req struct {
		Data     interface{} `json:"data"`
		Operation string      `json:"operation"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, err
	}

	// Operations: get_keys, get_values, pretty
	switch req.Operation {
	case "get_keys":
		if m, ok := req.Data.(map[string]interface{}); ok {
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			return json.Marshal(map[string]interface{}{"keys": keys})
		}
		return nil, fmt.Errorf("data is not an object")
	case "get_values":
		if m, ok := req.Data.(map[string]interface{}); ok {
			var values []interface{}
			for _, v := range m {
				values = append(values, v)
			}
			return json.Marshal(map[string]interface{}{"values": values})
		}
		return nil, fmt.Errorf("data is not an object")
	case "pretty":
		return json.MarshalIndent(req.Data, "", "  ")
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}
}

// CalculatorSkill performs mathematical calculations
type CalculatorSkill struct{}

func (s *CalculatorSkill) ID() string { return "calculator" }
func (s *CalculatorSkill) Name() string { return "Calculator" }
func (s *CalculatorSkill) Version() string { return "1.0.0" }
func (s *CalculatorSkill) Category() string { return "math" }

func (s *CalculatorSkill) Execute(ctx context.Context, input []byte) ([]byte, error) {
	var req struct {
		Operation string  `json:"operation"`
		A         float64 `json:"a"`
		B         float64 `json:"b"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, err
	}

	var result float64
	switch req.Operation {
	case "add":
		result = req.A + req.B
	case "sub":
		result = req.A - req.B
	case "mul":
		result = req.A * req.B
	case "div":
		if req.B == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = req.A / req.B
	case "pow":
		result = pow(req.A, req.B)
	case "sqrt":
		if req.A < 0 {
			return nil, fmt.Errorf("cannot sqrt negative number")
		}
		result = sqrt(req.A)
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}

	return json.Marshal(map[string]interface{}{
		"result":     result,
		"operation":  req.Operation,
		"operands":   []float64{req.A, req.B},
	})
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

func sqrt(x float64) float64 {
	// Newton's method for square root
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// EchoSkill echoes back the input
type EchoSkill struct{}

func (s *EchoSkill) ID() string { return "echo" }
func (s *EchoSkill) Name() string { return "Echo" }
func (s *EchoSkill) Version() string { return "1.0.0" }
func (s *EchoSkill) Category() string { return "utility" }

func (s *EchoSkill) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"echo":   string(input),
		"length": len(input),
	})
}