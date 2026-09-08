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
	"path"
	"slices"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
)

const azureRoleAssignmentsAtScopeFilter = "atScope()"

func listAzureSubscriptionIdentities(
	ctx context.Context,
	session *cloudazure.Session,
) ([]azureIdentity, error) {
	assignments, err := listAzureRoleAssignments(ctx, session)
	if err != nil {
		return nil, err
	}

	roleNames, err := resolveAzureRoleDefinitions(ctx, session, assignments)
	if err != nil {
		return nil, err
	}

	return foldAzureAssignments(assignments, roleNames), nil
}

func listAzureRoleAssignments(
	ctx context.Context,
	session *cloudazure.Session,
) ([]armauthorization.RoleAssignment, error) {
	client, err := armauthorization.NewRoleAssignmentsClient(
		session.AccountID(),
		session.TokenCredential(),
		session.ARMClientOptions(),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create azure role assignments client: %w", err)
	}

	pager := client.NewListForSubscriptionPager(
		&armauthorization.RoleAssignmentsClientListForSubscriptionOptions{
			Filter: new(azureRoleAssignmentsAtScopeFilter),
		},
	)

	var assignments []armauthorization.RoleAssignment

	for range maxPaginationPages {
		if !pager.More() {
			return assignments, nil
		}

		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot list azure role assignments: %w", err)
		}

		for _, assignment := range page.Value {
			if assignment != nil {
				assignments = append(assignments, *assignment)
			}
		}
	}

	return nil, fmt.Errorf("cannot list all azure role assignments: %w", ErrPaginationLimitReached)
}

func resolveAzureRoleDefinitions(
	ctx context.Context,
	session *cloudazure.Session,
	assignments []armauthorization.RoleAssignment,
) (map[string]string, error) {
	client, err := armauthorization.NewRoleDefinitionsClient(
		session.TokenCredential(),
		session.ARMClientOptions(),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create azure role definitions client: %w", err)
	}

	names := make(map[string]string)

	for _, assignment := range assignments {
		roleDefinitionID := azureAssignmentRoleDefinitionID(assignment)
		if roleDefinitionID == "" {
			continue
		}

		guid := azureRoleDefinitionGUID(roleDefinitionID)
		if guid == "" {
			continue
		}

		if _, ok := names[roleDefinitionID]; ok {
			continue
		}

		definition, err := client.Get(
			ctx,
			"subscriptions/"+session.AccountID(),
			guid,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot get azure role definition: %w", err)
		}

		names[roleDefinitionID] = azureRoleDefinitionName(definition.RoleDefinition)
	}

	return names, nil
}

func foldAzureAssignments(
	assignments []armauthorization.RoleAssignment,
	roleNames map[string]string,
) []azureIdentity {
	byPrincipal := make(map[string]azureIdentity)

	for _, assignment := range assignments {
		principalID := azureAssignmentPrincipalID(assignment)
		if principalID == "" {
			continue
		}

		identity, ok := byPrincipal[principalID]
		if !ok {
			identity = azureIdentity{
				PrincipalID:   principalID,
				PrincipalType: azureAssignmentPrincipalType(assignment),
			}
		}

		roleDefinitionID := azureAssignmentRoleDefinitionID(assignment)
		if guid := azureRoleDefinitionGUID(roleDefinitionID); guid != "" &&
			!slices.Contains(identity.RoleDefinitionIDs, guid) {
			identity.RoleDefinitionIDs = append(identity.RoleDefinitionIDs, guid)
		}

		if name := roleNames[roleDefinitionID]; name != "" && !slices.Contains(identity.Roles, name) {
			identity.Roles = append(identity.Roles, name)
		}

		byPrincipal[principalID] = identity
	}

	identities := make([]azureIdentity, 0, len(byPrincipal))
	for _, identity := range byPrincipal {
		slices.Sort(identity.Roles)
		slices.Sort(identity.RoleDefinitionIDs)
		identities = append(identities, identity)
	}

	slices.SortFunc(
		identities,
		func(a, b azureIdentity) int {
			return strings.Compare(a.PrincipalID, b.PrincipalID)
		},
	)

	return identities
}

func azureAssignmentPrincipalID(assignment armauthorization.RoleAssignment) string {
	if assignment.Properties == nil || assignment.Properties.PrincipalID == nil {
		return ""
	}

	return *assignment.Properties.PrincipalID
}

func azureAssignmentPrincipalType(assignment armauthorization.RoleAssignment) string {
	if assignment.Properties == nil || assignment.Properties.PrincipalType == nil {
		return ""
	}

	return string(*assignment.Properties.PrincipalType)
}

func azureAssignmentRoleDefinitionID(assignment armauthorization.RoleAssignment) string {
	if assignment.Properties == nil || assignment.Properties.RoleDefinitionID == nil {
		return ""
	}

	return *assignment.Properties.RoleDefinitionID
}

func azureRoleDefinitionName(definition armauthorization.RoleDefinition) string {
	if definition.Properties == nil || definition.Properties.RoleName == nil {
		return ""
	}

	return *definition.Properties.RoleName
}

func azureRoleDefinitionGUID(roleDefinitionID string) string {
	if roleDefinitionID == "" {
		return ""
	}

	return strings.ToLower(path.Base(roleDefinitionID))
}
