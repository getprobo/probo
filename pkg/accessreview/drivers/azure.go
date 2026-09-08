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

	"go.gearno.de/kit/log"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	azurePrincipalUser                  = "User"
	azurePrincipalAgentUser             = "AgentUser"
	azurePrincipalGroup                 = "Group"
	azurePrincipalForeignGroup          = "ForeignGroup"
	azurePrincipalServicePrincipal      = "ServicePrincipal"
	azurePrincipalAgentServicePrincipal = "AgentServicePrincipal"
	azurePrincipalDevice                = "Device"
	azureRoleDefinitionOwner            = "8e3af657-a8ff-443c-a75c-2fe8c4bcb635"
	azureRoleDefinitionUserAccessAdmin  = "18d7d88d-d35e-4fb5-a5c3-7773c20a72d9"
	azureRoleDefinitionRBACAdmin        = "f58310d9-a9f6-439a-9e8d-f62e7b41a168"
)

type (
	// AzureDriver lists the principals that hold Azure RBAC on the one
	// subscription the connector names. One connector produces one source.
	AzureDriver struct {
		session *cloudazure.Session
		logger  *log.Logger
	}

	azureIdentity struct {
		PrincipalID       string
		PrincipalType     string
		Roles             []string
		RoleDefinitionIDs []string
		DisplayName       string
		Email             string
		AccountEnabled    *bool
	}
)

var (
	_ Driver = (*AzureDriver)(nil)

	azureAdminRoleDefinitionIDs = map[string]struct{}{
		azureRoleDefinitionOwner:           {},
		azureRoleDefinitionUserAccessAdmin: {},
		azureRoleDefinitionRBACAdmin:       {},
	}
)

// NewAzureDriver builds the driver over a session already federated on the
// connected subscription.
func NewAzureDriver(session *cloudazure.Session, logger *log.Logger) *AzureDriver {
	return &AzureDriver{session: session, logger: logger}
}

func (d *AzureDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	identities, err := listAzureSubscriptionIdentities(ctx, d.session)
	if err != nil {
		return nil, fmt.Errorf("cannot list azure role assignments: %w", err)
	}

	if err := resolveAzurePrincipals(ctx, d.session, identities); err != nil {
		if ctx.Err() != nil {
			return nil, err
		}

		d.logger.WarnCtx(
			ctx,
			"cannot resolve azure principals, reporting identifiers only",
			cloudazure.SafeLogFields(err)...,
		)
	}

	records := make([]AccountRecord, 0, len(identities))
	for _, identity := range identities {
		records = append(records, azureIdentityRecord(identity))
	}

	return records, nil
}

func azureIdentityRecord(identity azureIdentity) AccountRecord {
	return AccountRecord{
		Email:       identity.Email,
		FullName:    identity.DisplayName,
		Roles:       slices.Clone(identity.Roles),
		IsAdmin:     azureIsAdmin(identity.RoleDefinitionIDs),
		Active:      azureActive(identity),
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  azureAuthMethod(identity.PrincipalType),
		AccountType: azureAccountType(identity.PrincipalType),
		ExternalID:  identity.PrincipalID,
	}
}

// azureIsAdmin reports admin only when a bound role definition is Owner,
// User Access Administrator, or Role Based Access Control Administrator.
// It matches the definition GUID because display names are localizable, and
// returns nil rather than false otherwise: a custom role can grant the same
// privileges under any name.
func azureIsAdmin(roleDefinitionIDs []string) *bool {
	for _, id := range roleDefinitionIDs {
		if _, ok := azureAdminRoleDefinitionIDs[id]; ok {
			return new(true)
		}
	}

	return nil
}

func azureActive(identity azureIdentity) *bool {
	switch identity.PrincipalType {
	case azurePrincipalGroup, azurePrincipalForeignGroup:
		return nil
	case azurePrincipalUser, azurePrincipalAgentUser,
		azurePrincipalServicePrincipal, azurePrincipalAgentServicePrincipal,
		azurePrincipalDevice:
		if identity.AccountEnabled == nil {
			return nil
		}

		return new(*identity.AccountEnabled)
	default:
		return nil
	}
}

func azureAuthMethod(principalType string) coredata.AccessReviewEntryAuthMethod {
	switch principalType {
	case azurePrincipalUser, azurePrincipalAgentUser:
		return coredata.AccessReviewEntryAuthMethodSSO
	default:
		return coredata.AccessReviewEntryAuthMethodUnknown
	}
}

func azureAccountType(principalType string) coredata.AccessReviewEntryAccountType {
	switch principalType {
	case azurePrincipalServicePrincipal, azurePrincipalAgentServicePrincipal, azurePrincipalDevice:
		return coredata.AccessReviewEntryAccountTypeServiceAccount
	default:
		return coredata.AccessReviewEntryAccountTypeUser
	}
}

func azureEmail(principalType, mail, userPrincipalName string) string {
	switch principalType {
	case azurePrincipalUser, azurePrincipalAgentUser:
		if mail != "" {
			return mail
		}

		return userPrincipalName
	case azurePrincipalGroup, azurePrincipalForeignGroup:
		return mail
	default:
		return ""
	}
}
