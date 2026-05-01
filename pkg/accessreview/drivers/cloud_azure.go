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

package drivers

import (
	"context"
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// AzureRoleAssignmentLister is the narrow seam CloudAzureDriver
	// depends on for Microsoft.Authorization/roleAssignments reads.
	// Real implementation lives behind *cloudaccount.AzureProvider;
	// tests inject a stub. Exported so the pkg/probo factory can
	// wire the seam.
	AzureRoleAssignmentLister interface {
		// ListRoleAssignments returns the role assignments at scope
		// plus the resolved principal display names. The seam
		// returns a flat slice of records; the driver does no
		// further reshaping.
		ListRoleAssignments(ctx context.Context) ([]AzureRoleAssignmentRecord, error)
	}

	// AzureRoleAssignmentRecord is the typed seam payload one role
	// assignment maps to. Drivers consume this directly without
	// touching the Azure SDK types.
	AzureRoleAssignmentRecord struct {
		PrincipalID    string
		PrincipalType  string // User | ServicePrincipal | Group
		PrincipalEmail string
		PrincipalName  string
		RoleName       string
	}

	// CloudAzureDriver fetches role assignments from a customer's
	// Azure scope (Subscription / Management Group / Tenant). The
	// seam resolves principal IDs to display names + emails via
	// Microsoft Graph behind the scenes; the driver consumes the
	// flattened records.
	CloudAzureDriver struct {
		assignments AzureRoleAssignmentLister
	}
)

// NewCloudAzureDriver wires the driver to the supplied role-
// assignment seam.
func NewCloudAzureDriver(assignments AzureRoleAssignmentLister) *CloudAzureDriver {
	return &CloudAzureDriver{assignments: assignments}
}

// ListAccounts emits one AccountRecord per resolved principal in
// the role-assignment list. Service principals come back as
// AccountTypeServiceAccount; users and groups come back as
// AccountTypeUser (Probo doesn't model Group separately yet).
func (d *CloudAzureDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	assignments, err := d.assignments.ListRoleAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot list azure role assignments: %w", err)
	}

	type aggregate struct {
		record AccountRecord
		roles  []string
	}
	byPrincipal := make(map[string]*aggregate)

	for _, assignment := range assignments {
		key := assignment.PrincipalID
		if key == "" {
			continue
		}

		agg, ok := byPrincipal[key]
		if !ok {
			email := assignment.PrincipalEmail
			if email == "" {
				email = assignment.PrincipalID
			}

			agg = &aggregate{
				record: AccountRecord{
					Email:       email,
					FullName:    assignment.PrincipalName,
					ExternalID:  assignment.PrincipalID,
					MFAStatus:   coredata.MFAStatusUnknown,
					AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
					AccountType: azureAccountTypeForPrincipal(assignment.PrincipalType),
				},
			}
			byPrincipal[key] = agg
		}
		if assignment.RoleName != "" {
			agg.roles = append(agg.roles, assignment.RoleName)
		}
	}

	records := make([]AccountRecord, 0, len(byPrincipal))
	for _, agg := range byPrincipal {
		if len(agg.roles) > 0 {
			agg.record.Role = agg.roles[0]
		}
		records = append(records, agg.record)
	}

	return records, nil
}

func azureAccountTypeForPrincipal(principalType string) coredata.AccessEntryAccountType {
	switch principalType {
	case "ServicePrincipal":
		return coredata.AccessEntryAccountTypeServiceAccount
	default:
		return coredata.AccessEntryAccountTypeUser
	}
}
