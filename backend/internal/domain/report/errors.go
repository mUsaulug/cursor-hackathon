package report

import "errors"

// ErrReportNotFound is returned when a report id is unknown.
var ErrReportNotFound = errors.New("report: not found")
