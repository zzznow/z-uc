package common

import (
	"fmt"
	"runtime"
	"strings"
)

type AppError struct {
	Err     string `json:"error"`
	Source  string `json:"source,omitempty"`
	Chain   string `json:"chain,omitempty"`
	TraceID string `json:"traceId,omitempty"`
}

func (e *AppError) Error() string { return e.Err }

func NewError(msg string) *AppError {
	_, file, line, _ := runtime.Caller(1)
	return &AppError{
		Err:    msg,
		Source: fmt.Sprintf("%s:%d", trimPath(file), line),
	}
}

func NewErrorf(format string, args ...interface{}) *AppError {
	return NewError(fmt.Sprintf(format, args...))
}

func WrapError(err *AppError, svc string) *AppError {
	if err == nil {
		return nil
	}
	if err.Chain == "" {
		err.Chain = svc
	} else {
		err.Chain = svc + " <- " + err.Chain
	}
	return err
}

func trimPath(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		path = path[i+1:]
	}
	return path
}
