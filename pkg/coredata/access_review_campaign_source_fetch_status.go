package coredata

import (
	"database/sql/driver"
	"fmt"
)

type AccessReviewCampaignSourceFetchStatus string

const (
	AccessReviewCampaignSourceFetchStatusQueued   AccessReviewCampaignSourceFetchStatus = "QUEUED"
	AccessReviewCampaignSourceFetchStatusFetching AccessReviewCampaignSourceFetchStatus = "FETCHING"
	AccessReviewCampaignSourceFetchStatusSuccess  AccessReviewCampaignSourceFetchStatus = "SUCCESS"
	AccessReviewCampaignSourceFetchStatusFailed   AccessReviewCampaignSourceFetchStatus = "FAILED"
)

func AccessReviewCampaignSourceFetchStatuses() []AccessReviewCampaignSourceFetchStatus {
	return []AccessReviewCampaignSourceFetchStatus{
		AccessReviewCampaignSourceFetchStatusQueued,
		AccessReviewCampaignSourceFetchStatusFetching,
		AccessReviewCampaignSourceFetchStatusSuccess,
		AccessReviewCampaignSourceFetchStatusFailed,
	}
}

func (s AccessReviewCampaignSourceFetchStatus) IsTerminal() bool {
	return s == AccessReviewCampaignSourceFetchStatusSuccess || s == AccessReviewCampaignSourceFetchStatusFailed
}

func (s AccessReviewCampaignSourceFetchStatus) String() string {
	return string(s)
}

func (s *AccessReviewCampaignSourceFetchStatus) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported type for AccessReviewCampaignSourceFetchStatus: %T", value)
	}

	switch str {
	case "QUEUED":
		*s = AccessReviewCampaignSourceFetchStatusQueued
	case "FETCHING":
		*s = AccessReviewCampaignSourceFetchStatusFetching
	case "SUCCESS":
		*s = AccessReviewCampaignSourceFetchStatusSuccess
	case "FAILED":
		*s = AccessReviewCampaignSourceFetchStatusFailed
	default:
		return fmt.Errorf("invalid AccessReviewCampaignSourceFetchStatus value: %q", str)
	}

	return nil
}

func (s AccessReviewCampaignSourceFetchStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
