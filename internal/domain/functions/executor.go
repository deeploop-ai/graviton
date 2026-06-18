package functions

import "context"

// Execution describes a single function invocation.
type Execution struct {
	FunctionID string
	Runtime    string // e.g. node-18.0, python-3.9
	SourcePath string // path or archive location of function source
	Entrypoint string // e.g. "index.main"
	Timeout    int64  // seconds
	Env        map[string]string
	Data       string // JSON payload
}

// ExecutionResult is the output of a function invocation.
type ExecutionResult struct {
	StatusCode int
	Stdout     string
	Stderr     string
	Response   string
	DurationMS int64
}

// Executor is the function runtime port.
type Executor interface {
	Execute(ctx context.Context, exec Execution) (*ExecutionResult, error)
}
