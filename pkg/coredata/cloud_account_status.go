// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"database/sql/driver"
	"fmt"
)

type CloudAccountStatus string

const (
	CloudAccountStatusPendingVerification CloudAccountStatus = "PENDING_VERIFICATION"
	CloudAccountStatusVerified            CloudAccountStatus = "VERIFIED"
	CloudAccountStatusErrored             CloudAccountStatus = "ERRORED"
	CloudAccountStatusDisconnected        CloudAccountStatus = "DISCONNECTED"
)

func CloudAccountStatuses() []CloudAccountStatus {
	return []CloudAccountStatus{
		CloudAccountStatusPendingVerification,
		CloudAccountStatusVerified,
		CloudAccountStatusErrored,
		CloudAccountStatusDisconnected,
	}
}

func (s CloudAccountStatus) String() string {
	return string(s)
}

func (s *CloudAccountStatus) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan CloudAccountStatus: unsupported type %T", value)
	}

	switch str {
	case "PENDING_VERIFICATION":
		*s = CloudAccountStatusPendingVerification
	case "VERIFIED":
		*s = CloudAccountStatusVerified
	case "ERRORED":
		*s = CloudAccountStatusErrored
	case "DISCONNECTED":
		*s = CloudAccountStatusDisconnected
	default:
		return fmt.Errorf("cannot parse CloudAccountStatus: invalid value %q", str)
	}

	return nil
}

func (s CloudAccountStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
