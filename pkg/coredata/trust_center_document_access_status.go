package coredata

import (
	"database/sql/driver"
	"fmt"
)

type TrustCenterDocumentAccessStatus string

const (
	TrustCenterDocumentAccessStatusRequested TrustCenterDocumentAccessStatus = "REQUESTED"
	TrustCenterDocumentAccessStatusGranted   TrustCenterDocumentAccessStatus = "GRANTED"
	TrustCenterDocumentAccessStatusRejected  TrustCenterDocumentAccessStatus = "REJECTED"
	TrustCenterDocumentAccessStatusRevoked   TrustCenterDocumentAccessStatus = "REVOKED"
)

func (tcdas TrustCenterDocumentAccessStatus) String() string {
	return string(tcdas)
}

func (tcdas *TrustCenterDocumentAccessStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for TrustCenterDocumentAccessStatus: %T", value)
	}

	switch s {
	case "REQUESTED":
		*tcdas = TrustCenterDocumentAccessStatusRequested
	case "GRANTED":
		*tcdas = TrustCenterDocumentAccessStatusGranted
	case "REJECTED":
		*tcdas = TrustCenterDocumentAccessStatusRejected
	case "REVOKED":
		*tcdas = TrustCenterDocumentAccessStatusRevoked
	default:
		return fmt.Errorf("invalid TrustCenterDocumentAccessStatus value: %q", s)
	}
	return nil
}

func (tcdas TrustCenterDocumentAccessStatus) Value() (driver.Value, error) {
	return tcdas.String(), nil
}
