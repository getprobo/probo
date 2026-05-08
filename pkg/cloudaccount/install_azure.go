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

package cloudaccount

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// AzureInstallStep is one rung of the structured Azure install
	// walkthrough. Title is rendered as a heading; Body is markdown;
	// Code (when set) is a CLI snippet the customer copies as-is.
	AzureInstallStep struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Code  string `json:"code,omitempty"`
	}

	// AzureInstallGuide is the structured walkthrough payload
	// returned to the frontend. v1 ships no auto-generated ARM
	// template (App Registration cannot be ARM-templated cleanly
	// without extra ceremony); the guide steps walk the customer
	// through the portal manually.
	AzureInstallGuide struct {
		Steps                    []AzureInstallStep                 `json:"steps"`
		RequiredRBACRoles        []string                           `json:"required_rbac_roles"`
		RequiredGraphPermissions []string                           `json:"required_graph_permissions"`
		Modules                  []coredata.CloudAccountAuditModule `json:"modules"`
	}
)

// BuildAzureInstallGuide renders the structured walkthrough payload
// the frontend uses to drive the App Registration install wizard.
// The "Grant admin consent" step is mandatory: Directory.Read.All
// is an admin-consent permission and the install cannot complete
// without it.
func BuildAzureInstallGuide(
	modules []coredata.CloudAccountAuditModule,
	scopeKind coredata.CloudAccountScopeKind,
) (AzureInstallGuide, error) {
	if scopeKind != coredata.CloudAccountScopeKindAzureSubscription &&
		scopeKind != coredata.CloudAccountScopeKindAzureManagementGroup &&
		scopeKind != coredata.CloudAccountScopeKindAzureTenant {
		return AzureInstallGuide{}, fmt.Errorf("cannot build azure install guide: unsupported scope kind %q", scopeKind)
	}

	rbacRoles := AzureRBACRolesForModules(modules)
	graphPerms := AzureGraphPermissionsForModules(modules)
	if len(rbacRoles) == 0 || len(graphPerms) == 0 {
		return AzureInstallGuide{}, fmt.Errorf("cannot build azure install guide: no rbac or graph permissions for modules %v", modules)
	}

	scopeLabel := azureScopeLabel(scopeKind)

	steps := []AzureInstallStep{
		{
			Title: "Create the App Registration",
			Body:  "In the Microsoft Entra admin center, open *App registrations* and create a new single-tenant registration named **Probo Cloud Scanner**. Probo never asks for a redirect URI -- leave it blank.",
		},
		{
			Title: "Generate a client secret",
			Body:  "From your new app's *Certificates & secrets* blade, create a client secret. Copy its *Value* immediately -- it is shown only once. You will paste it back into Probo at the end of this wizard.",
		},
		{
			Title: fmt.Sprintf("Assign the Reader role at the %s scope", scopeLabel),
			Body:  fmt.Sprintf("Open the %s you want Probo to inspect, then go to *Access control (IAM)* -> *Add role assignment*. Pick **%s**, set *Members* to the App Registration above, and confirm.", scopeLabel, strings.Join(rbacRoles, ", ")),
			Code:  fmt.Sprintf("az role assignment create \\\n    --assignee \"<app-client-id>\" \\\n    --role \"Reader\" \\\n    --scope \"<%s-id>\"", strings.ToLower(strings.ReplaceAll(scopeLabel, " ", "-"))),
		},
		{
			Title: "Add the Microsoft Graph permissions",
			Body:  fmt.Sprintf("Back in the App Registration, open *API permissions* -> *Add a permission* -> *Microsoft Graph* -> *Application permissions*, then add: %s.", strings.Join(graphPerms, ", ")),
		},
		{
			Title: "Grant admin consent",
			Body:  "Still under *API permissions*, click **Grant admin consent for <tenant>**. This step is mandatory -- the Microsoft Graph permissions above are admin-consent only and Probo's probe will fail with `ErrInsufficientPermissions` until consent is recorded. The button is greyed out for non-admin users; if so, hand the registration to a Global Administrator.",
		},
		{
			Title: "Paste the credentials back into Probo",
			Body:  "Copy your tenant id, the App Registration's client id, and the client secret value into Probo's Cloud Accounts dialog. Probo immediately probes the credentials by listing one page of subscriptions; on success the account moves to **VERIFIED**.",
		},
	}

	return AzureInstallGuide{
		Steps:                    steps,
		RequiredRBACRoles:        rbacRoles,
		RequiredGraphPermissions: graphPerms,
		Modules:                  modules,
	}, nil
}

func azureScopeLabel(scopeKind coredata.CloudAccountScopeKind) string {
	switch scopeKind {
	case coredata.CloudAccountScopeKindAzureSubscription:
		return "Subscription"
	case coredata.CloudAccountScopeKindAzureManagementGroup:
		return "Management Group"
	case coredata.CloudAccountScopeKindAzureTenant:
		return "Tenant Root Group"
	default:
		return string(scopeKind)
	}
}
