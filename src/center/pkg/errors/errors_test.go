package errors

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError(ErrNotFound, "Resource not found")

	if err.Code != ErrNotFound {
		t.Errorf("Expected code %d, got %d", ErrNotFound, err.Code)
	}

	if err.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got '%s'", err.Message)
	}
}

func TestWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewError(ErrInternal, "Internal error").WithCause(cause)

	if err.Cause != cause {
		t.Error("Cause not set correctly")
	}

	expected := "[900] Internal error: underlying error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestWithDetails(t *testing.T) {
	details := map[string]string{"field": "name"}
	err := NewError(ErrInvalidParameter, "Invalid parameter").WithDetails(details)

	if err.Details == nil {
		t.Error("Details not set")
	}
}

func TestIs(t *testing.T) {
	err := NewError(ErrNotFound, "Not found")

	if !Is(err, ErrNotFound) {
		t.Error("Is should return true for matching code")
	}

	if Is(err, ErrInternal) {
		t.Error("Is should return false for non-matching code")
	}

	if Is(nil, ErrNotFound) {
		t.Error("Is should return false for nil error")
	}
}

func TestGetCode(t *testing.T) {
	err := NewError(ErrAgentOffline, "Agent offline")
	code := GetCode(err)

	if code != ErrAgentOffline {
		t.Errorf("Expected code %d, got %d", ErrAgentOffline, code)
	}

	// Test standard error
	stdErr := errors.New("standard error")
	code = GetCode(stdErr)

	if code != ErrInternal {
		t.Errorf("Expected ErrInternal for standard error, got %d", code)
	}

	// Test nil
	code = GetCode(nil)
	if code != Success {
		t.Errorf("Expected Success for nil, got %d", code)
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("database connection failed")
	err := Wrap(cause, ErrDatabase, "Failed to connect")

	if err.Code != ErrDatabase {
		t.Errorf("Expected code %d, got %d", ErrDatabase, err.Code)
	}

	if err.Cause != cause {
		t.Error("Cause not preserved")
	}
}

func TestToJSON(t *testing.T) {
	err := NewError(ErrNotFound, "Not found")
	json := err.ToJSON()

	if len(json) == 0 {
		t.Error("ToJSON returned empty")
	}

	expected := `{"code":104,"message":"Not found"}`
	if string(json) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(json))
	}
}

func TestPredefinedErrors(t *testing.T) {
	if ErrAgentNotFoundMsg.Code != ErrAgentNotFound {
		t.Errorf("Predefined error has wrong code")
	}

	if ErrTaskTimeoutMsg.Message != "Task timeout" {
		t.Errorf("Predefined error has wrong message")
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("cause")
	err := NewError(ErrInternal, "error").WithCause(cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap should return cause")
	}
}