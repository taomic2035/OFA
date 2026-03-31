package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorCode defines standard error codes
type ErrorCode int

const (
	// Success
	Success ErrorCode = 0

	// Client errors (1xx)
	ErrInvalidRequest   ErrorCode = 100
	ErrInvalidParameter ErrorCode = 101
	ErrUnauthorized     ErrorCode = 102
	ErrForbidden        ErrorCode = 103
	ErrNotFound         ErrorCode = 104
	ErrAlreadyExists    ErrorCode = 105
	ErrRateLimited      ErrorCode = 106

	// Agent errors (2xx)
	ErrAgentNotFound     ErrorCode = 200
	ErrAgentOffline      ErrorCode = 201
	ErrAgentBusy         ErrorCode = 202
	ErrAgentDisconnected ErrorCode = 203

	// Task errors (3xx)
	ErrTaskNotFound  ErrorCode = 300
	ErrTaskFailed    ErrorCode = 301
	ErrTaskTimeout   ErrorCode = 302
	ErrTaskCancelled ErrorCode = 303
	ErrTaskRunning   ErrorCode = 304

	// Skill errors (4xx)
	ErrSkillNotFound    ErrorCode = 400
	ErrSkillError       ErrorCode = 401
	ErrInvalidSkillInput ErrorCode = 402

	// Message errors (5xx)
	ErrMessageNotFound ErrorCode = 500
	ErrMessageFailed   ErrorCode = 501
	ErrMessageTimeout  ErrorCode = 502

	// Server errors (9xx)
	ErrInternal      ErrorCode = 900
	ErrDatabase      ErrorCode = 901
	ErrCache         ErrorCode = 902
	ErrNetwork       ErrorCode = 903
	ErrConfiguration ErrorCode = 904
)

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    ErrorCode       `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty"`
	Cause   error           `json:"-"`
}

// Error implements the error interface
func (e *ErrorInfo) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *ErrorInfo) Unwrap() error {
	return e.Cause
}

// ToJSON returns JSON representation
func (e *ErrorInfo) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// NewError creates a new ErrorInfo
func NewError(code ErrorCode, message string) *ErrorInfo {
	return &ErrorInfo{
		Code:    code,
		Message: message,
	}
}

// WithCause adds a cause to the error
func (e *ErrorInfo) WithCause(cause error) *ErrorInfo {
	e.Cause = cause
	return e
}

// WithDetails adds details to the error
func (e *ErrorInfo) WithDetails(details interface{}) *ErrorInfo {
	data, _ := json.Marshal(details)
	e.Details = data
	return e
}

// ===== Predefined Errors =====

var (
	// Client errors
	ErrInvalidRequestMsg   = NewError(ErrInvalidRequest, "Invalid request")
	ErrInvalidParameterMsg = NewError(ErrInvalidParameter, "Invalid parameter")
	ErrUnauthorizedMsg     = NewError(ErrUnauthorized, "Unauthorized")
	ErrForbiddenMsg        = NewError(ErrForbidden, "Forbidden")
	ErrNotFoundMsg         = NewError(ErrNotFound, "Resource not found")
	ErrAlreadyExistsMsg    = NewError(ErrAlreadyExists, "Resource already exists")
	ErrRateLimitedMsg      = NewError(ErrRateLimited, "Rate limited")

	// Agent errors
	ErrAgentNotFoundMsg     = NewError(ErrAgentNotFound, "Agent not found")
	ErrAgentOfflineMsg      = NewError(ErrAgentOffline, "Agent is offline")
	ErrAgentBusyMsg         = NewError(ErrAgentBusy, "Agent is busy")
	ErrAgentDisconnectedMsg = NewError(ErrAgentDisconnected, "Agent disconnected")

	// Task errors
	ErrTaskNotFoundMsg  = NewError(ErrTaskNotFound, "Task not found")
	ErrTaskFailedMsg    = NewError(ErrTaskFailed, "Task failed")
	ErrTaskTimeoutMsg   = NewError(ErrTaskTimeout, "Task timeout")
	ErrTaskCancelledMsg = NewError(ErrTaskCancelled, "Task cancelled")

	// Skill errors
	ErrSkillNotFoundMsg    = NewError(ErrSkillNotFound, "Skill not found")
	ErrSkillErrorMsg       = NewError(ErrSkillError, "Skill execution error")
	ErrInvalidSkillInputMsg = NewError(ErrInvalidSkillInput, "Invalid skill input")

	// Server errors
	ErrInternalMsg      = NewError(ErrInternal, "Internal server error")
	ErrDatabaseMsg      = NewError(ErrDatabase, "Database error")
	ErrCacheMsg         = NewError(ErrCache, "Cache error")
	ErrNetworkMsg       = NewError(ErrNetwork, "Network error")
	ErrConfigurationMsg = NewError(ErrConfiguration, "Configuration error")
)

// ===== Helper Functions =====

// Is checks if error has specific code
func Is(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*ErrorInfo); ok {
		return e.Code == code
	}
	return false
}

// GetCode extracts error code from error
func GetCode(err error) ErrorCode {
	if err == nil {
		return Success
	}
	if e, ok := err.(*ErrorInfo); ok {
		return e.Code
	}
	return ErrInternal
}

// Wrap wraps an error with code and message
func Wrap(err error, code ErrorCode, message string) *ErrorInfo {
	return &ErrorInfo{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// WrapWithCode wraps an error with just a code
func WrapWithCode(err error, code ErrorCode) *ErrorInfo {
	return &ErrorInfo{
		Code:    code,
		Message: err.Error(),
		Cause:   err,
	}
}