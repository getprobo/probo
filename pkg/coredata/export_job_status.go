package coredata

import (
	"database/sql/driver"
	"fmt"
)

type (
	ExportJobStatus string
)

const (
	ExportJobStatusPending    ExportJobStatus = "PENDING"
	ExportJobStatusProcessing ExportJobStatus = "PROCESSING"
	ExportJobStatusCompleted  ExportJobStatus = "COMPLETED"
	ExportJobStatusFailed     ExportJobStatus = "FAILED"
)

func (ejs ExportJobStatus) String() string {
	return string(ejs)
}

func (ejs *ExportJobStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ExportJobStatus: %T", value)
	}

	switch s {
	case ExportJobStatusPending.String():
		*ejs = ExportJobStatusPending
	case ExportJobStatusProcessing.String():
		*ejs = ExportJobStatusProcessing
	case ExportJobStatusCompleted.String():
		*ejs = ExportJobStatusCompleted
	case ExportJobStatusFailed.String():
		*ejs = ExportJobStatusFailed
	default:
		return fmt.Errorf("invalid ExportJobStatus value: %q", s)
	}
	return nil
}

func (ejs ExportJobStatus) Value() (driver.Value, error) {
	return ejs.String(), nil
}
