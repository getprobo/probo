// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package drivers

import (
	"context"
	"fmt"
	"slices"

	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// gcpRoleOwner is the only IAM role this driver treats as an explicit
	// admin signal. A custom role can grant the same permissions under any
	// name, so anything else reports no signal rather than false.
	gcpRoleOwner = "roles/owner"
)

type (
	// GCPDriver lists the identities of the one GCP project the connector
	// names: IAM users, groups, and service accounts from the project policy,
	// plus every service account that lives in the project. One connector
	// produces one source.
	GCPDriver struct {
		session *cloudgcp.Session
	}
)

var _ Driver = (*GCPDriver)(nil)

// NewGCPDriver builds the driver over a session already impersonated on the
// connected project.
func NewGCPDriver(session *cloudgcp.Session) *GCPDriver {
	return &GCPDriver{session: session}
}

func (d *GCPDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	principals, err := listProjectIAMPrincipals(ctx, d.session)
	if err != nil {
		return nil, fmt.Errorf("cannot list iam identities of the gcp project: %w", err)
	}

	accounts, err := listProjectServiceAccounts(ctx, d.session)
	if err != nil {
		return nil, fmt.Errorf("cannot list service accounts of the gcp project: %w", err)
	}

	identities := unionGCPIdentities(principals, accounts)

	if err := attachUserManagedKeys(ctx, d.session, identities); err != nil {
		return nil, err
	}

	records := make([]AccountRecord, 0, len(identities))
	for _, identity := range identities {
		records = append(records, gcpIdentityRecord(identity))
	}

	return records, nil
}

func gcpIdentityRecord(identity gcpIdentity) AccountRecord {
	return AccountRecord{
		Email:       identity.Email,
		FullName:    identity.DisplayName,
		Roles:       slices.Clone(identity.Roles),
		IsAdmin:     gcpIsAdmin(identity.Roles),
		Active:      gcpActive(identity),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  gcpAuthMethod(identity),
		AccountType: gcpAccountType(identity.Kind),
		ExternalID:  gcpExternalID(identity),
	}
}

// gcpIsAdmin reports admin only when a binding names roles/owner. It returns
// nil rather than false otherwise, because a custom role can grant the same
// privileges under any name.
func gcpIsAdmin(roles []string) *bool {
	if slices.Contains(roles, gcpRoleOwner) {
		return new(true)
	}

	return nil
}

// gcpActive reports the service-account disabled flag when the identity was
// listed. Users and groups have no enabled flag on the project policy, so
// they stay unknown.
func gcpActive(identity gcpIdentity) *bool {
	if identity.Kind != gcpPrincipalServiceAccount || identity.Disabled == nil {
		return nil
	}

	return new(!*identity.Disabled)
}

func gcpAuthMethod(identity gcpIdentity) coredata.AccessReviewEntryAuthMethod {
	switch identity.Kind {
	case gcpPrincipalUser:
		return coredata.AccessReviewEntryAuthMethodSSO
	case gcpPrincipalServiceAccount:
		if identity.HasUserManagedKey {
			return coredata.AccessReviewEntryAuthMethodAPIKey
		}

		return coredata.AccessReviewEntryAuthMethodUnknown
	default:
		return coredata.AccessReviewEntryAuthMethodUnknown
	}
}

func gcpAccountType(kind gcpPrincipalKind) coredata.AccessReviewEntryAccountType {
	if kind == gcpPrincipalServiceAccount {
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	}

	return coredata.AccessReviewEntryAccountTypeUser
}

func gcpExternalID(identity gcpIdentity) string {
	if identity.UniqueID != "" {
		return identity.UniqueID
	}

	return identity.Principal
}
