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
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// GCPIAMPolicyReader is the narrow seam CloudGCPDriver depends
	// on for IAM-policy reads. The driver discriminates between
	// project-scope (single Project IAM policy) and org-scope
	// (cloudasset transitive enumeration) at construction time.
	// Exported so the pkg/probo factory can wire the seam.
	GCPIAMPolicyReader interface {
		// GetProjectIAMPolicy reads the IAM policy at the project
		// resource. Used by GCP_PROJECT scope.
		GetProjectIAMPolicy(ctx context.Context, projectID string) (GCPIAMPolicy, error)

		// SearchAllIAMPolicies enumerates IAM bindings across an
		// organization tree via cloudasset. Used by GCP_ORGANIZATION
		// scope.
		SearchAllIAMPolicies(ctx context.Context, orgID string) ([]GCPIAMPolicy, error)
	}

	// GCPIAMPolicy is the seam-friendly subset of the GCP IAM
	// policy shape the driver consumes. Keeps the driver decoupled
	// from the cloudresourcemanager / cloudasset SDK types so unit
	// tests can fabricate fixtures cleanly.
	GCPIAMPolicy struct {
		Bindings []GCPIAMBinding
	}

	// GCPIAMBinding is the seam-friendly subset of one IAM binding.
	GCPIAMBinding struct {
		Role    string
		Members []string
	}

	// CloudGCPDriver fetches IAM bindings from a customer's GCP
	// project or organization. The fetch path is keyed on
	// scope_kind: project scope reads only the project policy;
	// org scope additionally enumerates child projects via
	// cloudasset.
	CloudGCPDriver struct {
		policy    GCPIAMPolicyReader
		scopeKind coredata.CloudAccountScopeKind
		scopeID   string
	}
)

// NewCloudGCPDriver wires the driver to the supplied IAM-policy
// seam. The scopeKind drives the dual fetch path: GCP_PROJECT uses
// only cloudresourcemanager.Projects.GetIamPolicy; GCP_ORGANIZATION
// adds cloudasset.SearchAllIamPolicies for transitive enumeration
// across child projects.
func NewCloudGCPDriver(policy GCPIAMPolicyReader, scopeKind coredata.CloudAccountScopeKind, scopeID string) *CloudGCPDriver {
	return &CloudGCPDriver{
		policy:    policy,
		scopeKind: scopeKind,
		scopeID:   scopeID,
	}
}

// ListAccounts emits one AccountRecord per distinct principal in
// the IAM bindings. Membership type ("user:" / "serviceAccount:" /
// "group:") is preserved in AccountType so reviewers can
// distinguish humans from machine identities.
func (d *CloudGCPDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var policies []GCPIAMPolicy

	switch d.scopeKind {
	case coredata.CloudAccountScopeKindGCPProject:
		p, err := d.policy.GetProjectIAMPolicy(ctx, d.scopeID)
		if err != nil {
			return nil, fmt.Errorf("cannot read gcp project iam policy: %w", err)
		}
		policies = []GCPIAMPolicy{p}

	case coredata.CloudAccountScopeKindGCPOrganization:
		ps, err := d.policy.SearchAllIAMPolicies(ctx, d.scopeID)
		if err != nil {
			return nil, fmt.Errorf("cannot search gcp iam policies: %w", err)
		}
		policies = ps

	default:
		return nil, fmt.Errorf("cannot list gcp accounts: unsupported scope kind %q", d.scopeKind)
	}

	type aggregate struct {
		record AccountRecord
		roles  []string
	}
	byEmail := make(map[string]*aggregate)

	for _, policy := range policies {
		for _, binding := range policy.Bindings {
			for _, member := range binding.Members {
				kind, identifier := splitGCPMember(member)
				if identifier == "" {
					continue
				}

				agg, ok := byEmail[identifier]
				if !ok {
					agg = &aggregate{
						record: AccountRecord{
							Email:       identifier,
							ExternalID:  identifier,
							MFAStatus:   coredata.MFAStatusUnknown,
							AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
							AccountType: gcpAccountTypeForKind(kind),
						},
					}
					byEmail[identifier] = agg
				}
				agg.roles = append(agg.roles, binding.Role)
			}
		}
	}

	records := make([]AccountRecord, 0, len(byEmail))
	for _, agg := range byEmail {
		agg.record.Role = strings.Join(agg.roles, ", ")
		records = append(records, agg.record)
	}

	return records, nil
}

// splitGCPMember splits a GCP IAM member string like
// "user:foo@bar.com" into ("user", "foo@bar.com"). Membership
// types Probo doesn't model are returned with kind="" and
// preserved as-is.
func splitGCPMember(member string) (string, string) {
	idx := strings.Index(member, ":")
	if idx == -1 {
		return "", ""
	}
	return member[:idx], member[idx+1:]
}

func gcpAccountTypeForKind(kind string) coredata.AccessEntryAccountType {
	switch kind {
	case "serviceAccount":
		return coredata.AccessEntryAccountTypeServiceAccount
	case "user", "group", "domain":
		return coredata.AccessEntryAccountTypeUser
	default:
		return coredata.AccessEntryAccountTypeUser
	}
}
