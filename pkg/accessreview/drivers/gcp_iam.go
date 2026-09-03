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
	"net/url"
	"slices"
	"strings"

	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	iam "google.golang.org/api/iam/v1"
)

const (
	gcpPrincipalUserPrefix           = "user:"
	gcpPrincipalGroupPrefix          = "group:"
	gcpPrincipalServiceAccountPrefix = "serviceAccount:"
	gcpUserManagedKeyType            = "USER_MANAGED"
	gcpServiceAccountsPageSize       = 100
)

type (
	gcpPrincipalKind string

	// gcpIdentity is one IAM principal of a single GCP project, joined from
	// the project policy and, when listing is allowed, the project's service
	// accounts and their keys.
	gcpIdentity struct {
		Principal         string
		Email             string
		Kind              gcpPrincipalKind
		Roles             []string
		UniqueID          string
		DisplayName       string
		Disabled          *bool
		HasUserManagedKey bool
	}

	gcpServiceAccount struct {
		Email       string
		UniqueID    string
		DisplayName string
		Disabled    bool
	}
)

const (
	gcpPrincipalUser           gcpPrincipalKind = "user"
	gcpPrincipalGroup          gcpPrincipalKind = "group"
	gcpPrincipalServiceAccount gcpPrincipalKind = "serviceAccount"
)

func listProjectIAMPrincipals(ctx context.Context, session *cloudgcp.Session) ([]gcpIdentity, error) {
	svc, err := cloudresourcemanager.NewService(ctx, session.ServiceOptions()...)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp resource manager client: %w", err)
	}

	policy, err := svc.Projects.GetIamPolicy(
		session.AccountID(),
		&cloudresourcemanager.GetIamPolicyRequest{},
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot get gcp project iam policy: %w", err)
	}

	return principalsFromPolicy(policy), nil
}

func principalsFromPolicy(policy *cloudresourcemanager.Policy) []gcpIdentity {
	if policy == nil {
		return nil
	}

	byPrincipal := make(map[string]gcpIdentity)

	for _, binding := range policy.Bindings {
		if binding == nil || binding.Role == "" {
			continue
		}

		for _, member := range binding.Members {
			identity, ok := parseGCPPrincipal(member)
			if !ok {
				continue
			}

			existing, ok := byPrincipal[identity.Principal]
			if !ok {
				existing = identity
			}

			if !slices.Contains(existing.Roles, binding.Role) {
				existing.Roles = append(existing.Roles, binding.Role)
			}

			byPrincipal[identity.Principal] = existing
		}
	}

	principals := make([]gcpIdentity, 0, len(byPrincipal))
	for _, identity := range byPrincipal {
		principals = append(principals, identity)
	}

	return principals
}

func parseGCPPrincipal(member string) (gcpIdentity, bool) {
	member = strings.TrimSpace(member)
	if skipGCPPrincipal(member) {
		return gcpIdentity{}, false
	}

	switch {
	case strings.HasPrefix(member, gcpPrincipalUserPrefix):
		email := strings.TrimPrefix(member, gcpPrincipalUserPrefix)
		if email == "" {
			return gcpIdentity{}, false
		}

		return gcpIdentity{
			Principal: member,
			Email:     email,
			Kind:      gcpPrincipalUser,
		}, true
	case strings.HasPrefix(member, gcpPrincipalGroupPrefix):
		email := strings.TrimPrefix(member, gcpPrincipalGroupPrefix)
		if email == "" {
			return gcpIdentity{}, false
		}

		return gcpIdentity{
			Principal: member,
			Email:     email,
			Kind:      gcpPrincipalGroup,
		}, true
	case strings.HasPrefix(member, gcpPrincipalServiceAccountPrefix):
		email := strings.TrimPrefix(member, gcpPrincipalServiceAccountPrefix)
		if email == "" {
			return gcpIdentity{}, false
		}

		return gcpIdentity{
			Principal: member,
			Email:     email,
			Kind:      gcpPrincipalServiceAccount,
		}, true
	default:
		return gcpIdentity{}, false
	}
}

func skipGCPPrincipal(member string) bool {
	switch member {
	case "allUsers", "allAuthenticatedUsers":
		return true
	}

	for _, prefix := range []string{
		"domain:",
		"deleted:",
		"principal:",
		"principalSet:",
	} {
		if strings.HasPrefix(member, prefix) {
			return true
		}
	}

	return false
}

func listProjectServiceAccounts(ctx context.Context, session *cloudgcp.Session) ([]gcpServiceAccount, error) {
	svc, err := iam.NewService(ctx, session.ServiceOptions()...)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp iam client: %w", err)
	}

	parent, err := url.JoinPath("projects", url.PathEscape(session.AccountID()))
	if err != nil {
		return nil, fmt.Errorf("cannot build gcp service account parent: %w", err)
	}

	var (
		accounts  []gcpServiceAccount
		pageToken string
	)

	for range maxPaginationPages {
		call := svc.Projects.ServiceAccounts.List(parent).
			PageSize(gcpServiceAccountsPageSize).
			Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("cannot list gcp service accounts: %w", err)
		}

		for _, account := range resp.Accounts {
			if account == nil || account.Email == "" {
				continue
			}

			accounts = append(
				accounts,
				gcpServiceAccount{
					Email:       account.Email,
					UniqueID:    account.UniqueId,
					DisplayName: account.DisplayName,
					Disabled:    account.Disabled,
				},
			)
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			return accounts, nil
		}
	}

	return nil, fmt.Errorf("cannot list gcp service accounts: %w", ErrPaginationLimitReached)
}

func unionGCPIdentities(principals []gcpIdentity, accounts []gcpServiceAccount) []gcpIdentity {
	byEmail := make(map[string]int, len(principals))
	identities := make([]gcpIdentity, 0, len(principals)+len(accounts))

	for _, principal := range principals {
		if principal.Kind == gcpPrincipalServiceAccount {
			byEmail[principal.Email] = len(identities)
		}

		identities = append(identities, principal)
	}

	for _, account := range accounts {
		if i, ok := byEmail[account.Email]; ok {
			identities[i].UniqueID = account.UniqueID
			identities[i].DisplayName = account.DisplayName
			identities[i].Disabled = new(account.Disabled)

			continue
		}

		identities = append(
			identities,
			gcpIdentity{
				Principal:   gcpPrincipalServiceAccountPrefix + account.Email,
				Email:       account.Email,
				Kind:        gcpPrincipalServiceAccount,
				UniqueID:    account.UniqueID,
				DisplayName: account.DisplayName,
				Disabled:    new(account.Disabled),
			},
		)
	}

	slices.SortFunc(
		identities,
		func(a, b gcpIdentity) int {
			return strings.Compare(a.Principal, b.Principal)
		},
	)

	return identities
}

func attachUserManagedKeys(
	ctx context.Context,
	session *cloudgcp.Session,
	identities []gcpIdentity,
) error {
	svc, err := iam.NewService(ctx, session.ServiceOptions()...)
	if err != nil {
		return fmt.Errorf("cannot create gcp iam client: %w", err)
	}

	for i := range identities {
		if identities[i].Kind != gcpPrincipalServiceAccount || identities[i].UniqueID == "" {
			continue
		}

		hasKey, err := serviceAccountHasUserManagedKey(ctx, svc, identities[i].Email)
		if err != nil {
			return fmt.Errorf("cannot list gcp service account keys: %w", err)
		}

		identities[i].HasUserManagedKey = hasKey
	}

	return nil
}

func serviceAccountHasUserManagedKey(ctx context.Context, svc *iam.Service, email string) (bool, error) {
	name, err := url.JoinPath("projects/-/serviceAccounts", url.PathEscape(email))
	if err != nil {
		return false, fmt.Errorf("cannot build gcp service account resource name: %w", err)
	}

	resp, err := svc.Projects.ServiceAccounts.Keys.List(name).
		KeyTypes(gcpUserManagedKeyType).
		Context(ctx).
		Do()
	if err != nil {
		return false, fmt.Errorf("cannot list keys of a gcp service account: %w", err)
	}

	for _, key := range resp.Keys {
		if key != nil && key.KeyType == gcpUserManagedKeyType {
			return true, nil
		}
	}

	return false, nil
}
