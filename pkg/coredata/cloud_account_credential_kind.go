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

type CloudAccountCredentialKind string

const (
	CloudAccountCredentialKindAWSAssumeRole                 CloudAccountCredentialKind = "AWS_ASSUME_ROLE"
	CloudAccountCredentialKindGCPServiceAccountKey          CloudAccountCredentialKind = "GCP_SERVICE_ACCOUNT_KEY"
	CloudAccountCredentialKindAzureClientSecret             CloudAccountCredentialKind = "AZURE_CLIENT_SECRET"
	CloudAccountCredentialKindGCPWorkloadIdentityFederation CloudAccountCredentialKind = "GCP_WORKLOAD_IDENTITY_FEDERATION" // v2 placeholder
	CloudAccountCredentialKindAzureFederatedCredential      CloudAccountCredentialKind = "AZURE_FEDERATED_CREDENTIAL"       // v2 placeholder
)

func CloudAccountCredentialKinds() []CloudAccountCredentialKind {
	return []CloudAccountCredentialKind{
		CloudAccountCredentialKindAWSAssumeRole,
		CloudAccountCredentialKindGCPServiceAccountKey,
		CloudAccountCredentialKindAzureClientSecret,
		CloudAccountCredentialKindGCPWorkloadIdentityFederation,
		CloudAccountCredentialKindAzureFederatedCredential,
	}
}

func (k CloudAccountCredentialKind) String() string {
	return string(k)
}

func (k *CloudAccountCredentialKind) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("cannot scan CloudAccountCredentialKind: unsupported type %T", value)
	}

	switch s {
	case "AWS_ASSUME_ROLE":
		*k = CloudAccountCredentialKindAWSAssumeRole
	case "GCP_SERVICE_ACCOUNT_KEY":
		*k = CloudAccountCredentialKindGCPServiceAccountKey
	case "AZURE_CLIENT_SECRET":
		*k = CloudAccountCredentialKindAzureClientSecret
	case "GCP_WORKLOAD_IDENTITY_FEDERATION":
		*k = CloudAccountCredentialKindGCPWorkloadIdentityFederation
	case "AZURE_FEDERATED_CREDENTIAL":
		*k = CloudAccountCredentialKindAzureFederatedCredential
	default:
		return fmt.Errorf("cannot parse CloudAccountCredentialKind: invalid value %q", s)
	}

	return nil
}

func (k CloudAccountCredentialKind) Value() (driver.Value, error) {
	return k.String(), nil
}
