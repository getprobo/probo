package coredata

import (
	"database/sql/driver"
	"fmt"
)

type DocumentClassification string

const (
	DocumentClassificationPublic       DocumentClassification = "PUBLIC"
	DocumentClassificationInternal     DocumentClassification = "INTERNAL"
	DocumentClassificationConfidential DocumentClassification = "CONFIDENTIAL"
	DocumentClassificationSecret       DocumentClassification = "SECRET"
)

func DocumentClassifications() []DocumentClassification {
	return []DocumentClassification{
		DocumentClassificationPublic,
		DocumentClassificationInternal,
		DocumentClassificationConfidential,
		DocumentClassificationSecret,
	}
}

func (dc DocumentClassification) String() string {
	switch dc {
	case DocumentClassificationPublic:
		return "PUBLIC"
	case DocumentClassificationInternal:
		return "INTERNAL"
	case DocumentClassificationConfidential:
		return "CONFIDENTIAL"
	case DocumentClassificationSecret:
		return "SECRET"
	}
	panic(fmt.Errorf("invalid DocumentClassification value: %s", string(dc)))
}

// Scan implements the sql.Scanner interface for database deserialization.
func (dc *DocumentClassification) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var sv string
	switch v := value.(type) {
	case string:
		sv = v
	case []byte:
		sv = string(v)
	default:
		return fmt.Errorf("cannot scan DocumentClassification: expected string or []byte, got %T", value)
	}

	*dc = DocumentClassification(sv)
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (dc DocumentClassification) Value() (driver.Value, error) {
	return string(dc), nil
}
