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

type CloudAccountScopeKind string

const (
	CloudAccountScopeKindAWSAccount           CloudAccountScopeKind = "AWS_ACCOUNT"
	CloudAccountScopeKindAWSOrganization      CloudAccountScopeKind = "AWS_ORGANIZATION"
	CloudAccountScopeKindGCPProject           CloudAccountScopeKind = "GCP_PROJECT"
	CloudAccountScopeKindGCPOrganization      CloudAccountScopeKind = "GCP_ORGANIZATION"
	CloudAccountScopeKindAzureSubscription    CloudAccountScopeKind = "AZURE_SUBSCRIPTION"
	CloudAccountScopeKindAzureManagementGroup CloudAccountScopeKind = "AZURE_MANAGEMENT_GROUP"
	CloudAccountScopeKindAzureTenant          CloudAccountScopeKind = "AZURE_TENANT"
)

func CloudAccountScopeKinds() []CloudAccountScopeKind {
	return []CloudAccountScopeKind{
		CloudAccountScopeKindAWSAccount,
		CloudAccountScopeKindAWSOrganization,
		CloudAccountScopeKindGCPProject,
		CloudAccountScopeKindGCPOrganization,
		CloudAccountScopeKindAzureSubscription,
		CloudAccountScopeKindAzureManagementGroup,
		CloudAccountScopeKindAzureTenant,
	}
}

func (s CloudAccountScopeKind) String() string {
	return string(s)
}

func (s *CloudAccountScopeKind) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan CloudAccountScopeKind: unsupported type %T", value)
	}

	switch str {
	case "AWS_ACCOUNT":
		*s = CloudAccountScopeKindAWSAccount
	case "AWS_ORGANIZATION":
		*s = CloudAccountScopeKindAWSOrganization
	case "GCP_PROJECT":
		*s = CloudAccountScopeKindGCPProject
	case "GCP_ORGANIZATION":
		*s = CloudAccountScopeKindGCPOrganization
	case "AZURE_SUBSCRIPTION":
		*s = CloudAccountScopeKindAzureSubscription
	case "AZURE_MANAGEMENT_GROUP":
		*s = CloudAccountScopeKindAzureManagementGroup
	case "AZURE_TENANT":
		*s = CloudAccountScopeKindAzureTenant
	default:
		return fmt.Errorf("cannot parse CloudAccountScopeKind: invalid value %q", str)
	}

	return nil
}

func (s CloudAccountScopeKind) Value() (driver.Value, error) {
	return s.String(), nil
}
