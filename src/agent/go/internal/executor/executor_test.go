package executor

import (
	"context"
	"encoding/json"
	"testing"
)

func TestTextProcessSkill_Uppercase(t *testing.T) {
	skill := &TextProcessSkill{}

	input, _ := json.Marshal(map[string]string{
		"text":      "hello",
		"operation": "uppercase",
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]string
	json.Unmarshal(output, &result)

	if result["result"] != "HELLO" {
		t.Errorf("Expected HELLO, got %s", result["result"])
	}
}

func TestTextProcessSkill_Lowercase(t *testing.T) {
	skill := &TextProcessSkill{}

	input, _ := json.Marshal(map[string]string{
		"text":      "HELLO",
		"operation": "lowercase",
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]string
	json.Unmarshal(output, &result)

	if result["result"] != "hello" {
		t.Errorf("Expected hello, got %s", result["result"])
	}
}

func TestTextProcessSkill_Reverse(t *testing.T) {
	skill := &TextProcessSkill{}

	input, _ := json.Marshal(map[string]string{
		"text":      "hello",
		"operation": "reverse",
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]string
	json.Unmarshal(output, &result)

	if result["result"] != "olleh" {
		t.Errorf("Expected olleh, got %s", result["result"])
	}
}

func TestTextProcessSkill_Length(t *testing.T) {
	skill := &TextProcessSkill{}

	input, _ := json.Marshal(map[string]string{
		"text":      "hello",
		"operation": "length",
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]string
	json.Unmarshal(output, &result)

	if result["result"] != "5" {
		t.Errorf("Expected 5, got %s", result["result"])
	}
}

func TestTextProcessSkill_InvalidOperation(t *testing.T) {
	skill := &TextProcessSkill{}

	input, _ := json.Marshal(map[string]string{
		"text":      "hello",
		"operation": "invalid",
	})

	_, err := skill.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for invalid operation")
	}
}

func TestJSONProcessSkill_GetKeys(t *testing.T) {
	skill := &JSONProcessSkill{}

	input, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"a": 1,
			"b": 2,
			"c": 3,
		},
		"operation": "get_keys",
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(output, &result)

	keys := result["keys"]
	if keys == nil {
		t.Error("Expected keys array")
	}
}

func TestExecutor_RegisterSkill(t *testing.T) {
	exec := NewExecutor()
	skill := &TextProcessSkill{}

	exec.RegisterSkill(skill)

	// Check that skill is registered
	s, ok := exec.skills.Load(skill.ID())
	if !ok {
		t.Error("Skill not registered")
	}
	if s != skill {
		t.Error("Wrong skill registered")
	}
}

func TestExecutor_GetCapabilities(t *testing.T) {
	exec := NewExecutor()
	exec.RegisterSkill(&TextProcessSkill{})
	exec.RegisterSkill(&JSONProcessSkill{})

	caps := exec.GetCapabilities()

	if len(caps) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(caps))
	}
}

func TestCalculatorSkill_Add(t *testing.T) {
	skill := &CalculatorSkill{}

	input, _ := json.Marshal(map[string]interface{}{
		"operation": "add",
		"a":         10,
		"b":         5,
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(output, &result)

	if result["result"].(float64) != 15 {
		t.Errorf("Expected 15, got %v", result["result"])
	}
}

func TestCalculatorSkill_Sub(t *testing.T) {
	skill := &CalculatorSkill{}

	input, _ := json.Marshal(map[string]interface{}{
		"operation": "sub",
		"a":         10,
		"b":         3,
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(output, &result)

	if result["result"].(float64) != 7 {
		t.Errorf("Expected 7, got %v", result["result"])
	}
}

func TestCalculatorSkill_Mul(t *testing.T) {
	skill := &CalculatorSkill{}

	input, _ := json.Marshal(map[string]interface{}{
		"operation": "mul",
		"a":         6,
		"b":         7,
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(output, &result)

	if result["result"].(float64) != 42 {
		t.Errorf("Expected 42, got %v", result["result"])
	}
}

func TestCalculatorSkill_Div(t *testing.T) {
	skill := &CalculatorSkill{}

	input, _ := json.Marshal(map[string]interface{}{
		"operation": "div",
		"a":         20,
		"b":         4,
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(output, &result)

	if result["result"].(float64) != 5 {
		t.Errorf("Expected 5, got %v", result["result"])
	}
}

func TestCalculatorSkill_DivByZero(t *testing.T) {
	skill := &CalculatorSkill{}

	input, _ := json.Marshal(map[string]interface{}{
		"operation": "div",
		"a":         10,
		"b":         0,
	})

	_, err := skill.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for division by zero")
	}
}

func TestEchoSkill(t *testing.T) {
	skill := &EchoSkill{}

	input, _ := json.Marshal(map[string]string{
		"message": "hello world",
	})

	output, err := skill.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(output, &result)

	// The echo returns the JSON string which is: {"message":"hello world"}
	// Length is 25 characters
	if result["length"].(float64) != 25 {
		t.Errorf("Expected length 25, got %v", result["length"])
	}
}