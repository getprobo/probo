package coredata

import (
	"database/sql/driver"
	"fmt"
)

type (
	ExportJobType string
)

const (
	ExportJobTypeFramework ExportJobType = "FRAMEWORK"
	ExportJobTypeDocument  ExportJobType = "DOCUMENT"
)

func (ejt ExportJobType) String() string {
	return string(ejt)
}

func (ejt *ExportJobType) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ExportJobType: %T", value)
	}

	switch s {
	case ExportJobTypeFramework.String():
		*ejt = ExportJobTypeFramework
	case ExportJobTypeDocument.String():
		*ejt = ExportJobTypeDocument
	default:
		return fmt.Errorf("invalid ExportJobType value: %q", s)
	}
	return nil
}

func (ejt ExportJobType) Value() (driver.Value, error) {
	return ejt.String(), nil
}
