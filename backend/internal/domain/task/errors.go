package task

import "errors"

// ErrTaskNotFound is returned when a task id is unknown.
var ErrTaskNotFound = errors.New("task: not found")
