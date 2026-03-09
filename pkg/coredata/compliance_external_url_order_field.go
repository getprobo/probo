package coredata

type (
	ComplianceExternalURLOrderField string
)

const (
	ComplianceExternalURLOrderFieldCreatedAt ComplianceExternalURLOrderField = "CREATED_AT"
	ComplianceExternalURLOrderFieldRank      ComplianceExternalURLOrderField = "RANK"
)

func (p ComplianceExternalURLOrderField) Column() string {
	switch p {
	case ComplianceExternalURLOrderFieldCreatedAt:
		return "created_at"
	case ComplianceExternalURLOrderFieldRank:
		return "rank"
	default:
		return string(p)
	}
}

func (p ComplianceExternalURLOrderField) String() string {
	return string(p)
}

func (p ComplianceExternalURLOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *ComplianceExternalURLOrderField) UnmarshalText(text []byte) error {
	*p = ComplianceExternalURLOrderField(text)
	return nil
}
