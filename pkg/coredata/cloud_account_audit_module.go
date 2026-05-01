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

// CloudAccountAuditModule names a discrete audit pipeline (e.g.
// access-review enrichment, S3-public-bucket scan) a cloud account
// can opt into. Owned by coredata so the GraphQL @goModel for
// CloudAccountAuditModule resolves to a coredata type, matching
// every other cloud-account enum. The behaviour side -- per-(provider,
// scope, module) action lists -- lives in pkg/cloudaccount/permissions.go
// and consumes this enum.
type CloudAccountAuditModule string

const (
	CloudAccountAuditModuleAccessReview CloudAccountAuditModule = "ACCESS_REVIEW"
	CloudAccountAuditModuleCSPMS3Public CloudAccountAuditModule = "CSPM_S3_PUBLIC" // v2 placeholder
	CloudAccountAuditModuleCSPMIAMAudit CloudAccountAuditModule = "CSPM_IAM_AUDIT" // v2 placeholder
)

func CloudAccountAuditModules() []CloudAccountAuditModule {
	return []CloudAccountAuditModule{
		CloudAccountAuditModuleAccessReview,
		CloudAccountAuditModuleCSPMS3Public,
		CloudAccountAuditModuleCSPMIAMAudit,
	}
}

func (m CloudAccountAuditModule) String() string {
	return string(m)
}

func (m *CloudAccountAuditModule) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan CloudAccountAuditModule: unsupported type %T", value)
	}

	switch s {
	case "ACCESS_REVIEW":
		*m = CloudAccountAuditModuleAccessReview
	case "CSPM_S3_PUBLIC":
		*m = CloudAccountAuditModuleCSPMS3Public
	case "CSPM_IAM_AUDIT":
		*m = CloudAccountAuditModuleCSPMIAMAudit
	default:
		return fmt.Errorf("cannot parse CloudAccountAuditModule: invalid value %q", s)
	}

	return nil
}

func (m CloudAccountAuditModule) Value() (driver.Value, error) {
	return m.String(), nil
}
