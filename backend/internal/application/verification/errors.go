package verification

import "errors"

// ErrInvalidState is returned when a task is not in a state that allows the
// requested verification action.
var ErrInvalidState = errors.New("verification: invalid task state for this action")
