<!--
Copyright (c) 2026 Probo Inc <hello@probo.com>.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
-->

# `getprobo/audit-role/azurerm`

Grants Probo read-only audit access to one Azure subscription through
workload identity federation. Probo holds no Azure credential. Probo
presents a short-lived signed assertion. Entra verifies that assertion
against a public key set. You revoke access when you delete the
application or the federated credential.

This module covers **one subscription**. Apply it in the subscription
you connect. It does not walk management groups or resource groups.

## Two privileged actors

The Graph app role assignments **are** admin consent. Creation of those
assignments requires a **Global Administrator** or **Privileged Role
Administrator**. A service principal with `AppRoleAssignment.ReadWrite.All`
plus `Application.Read.All` can also create them.

The subscription role assignment requires **Owner** or **User Access
Administrator** on that subscription.

In most organizations those are different people. Set
`grant_directory_read = false` so the Azure-side operator can apply the
module today. Directory grants can follow later. Until they land, Probo
resolves principals as bare GUIDs.

## Usage

Copy the issuer URL and the subject out of the connector setup screen.
Entra compares the issuer URL case-sensitively. The last path segment
is a mixed-case identifier. Paste the values. Do not retype them.

This module does not configure providers. Set `environment` on the root
`azurerm` and `azuread` providers.

```hcl
provider "azurerm" {
  features {}
}

provider "azuread" {}

module "probo_audit" {
  source = "getprobo/audit-role/azurerm"

  probo_issuer_url = "https://proboidentity.com/e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
  probo_subject    = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
  subscription_id  = "00000000-0000-0000-0000-000000000000"
}

output "probo_tenant_id" {
  value = module.probo_audit.tenant_id
}

output "probo_client_id" {
  value = module.probo_audit.client_id
}

output "probo_subscription_id" {
  value = module.probo_audit.subscription_id
}
```

Give Probo the `tenant_id`, `client_id`, and `subscription_id` outputs
when you create the connector.

## Sovereign clouds

Set `environment` on the **root** providers. This module does not set
it. GCC High and DoD both use `usgovernment` on the providers. The Graph
host is a Probo connector setting, not a Terraform setting.

### Azure Government (GCC High and DoD)

```hcl
provider "azurerm" {
  features {}
  environment = "usgovernment"
}

provider "azuread" {
  environment = "usgovernment"
}
```

### Azure China

```hcl
provider "azurerm" {
  features {}
  environment = "china"
}

provider "azuread" {
  environment = "china"
}
```

GCC as distinct from GCC High runs on the commercial cloud. Use the
default public providers.

## Verifying an install

A clean `apply` does not prove that federation works. A wrong `subject`
creates successfully and fails at token exchange. Microsoft documents
that you will not get an error at create time.

Use the Probo console probe after you apply. Isolation is the
per-organization issuer and the exact subject. Entra matches both
case-sensitively.

Federation can take a few minutes to propagate. A first probe can fail
with `AADSTS70021`. Wait, then probe again.

## What it creates

| Resource | Notes |
|---|---|
| `azuread_application_registration` | `Probo Access Review` by default. |
| `azuread_application_federated_identity_credential` | Issuer is `probo_issuer_url`. Subject is `probo_subject`. Audience is `api://AzureADTokenExchange`. |
| `azuread_service_principal` | Linked to the application. |
| `azurerm_role_assignment` | `Reader` on `/subscriptions/{subscription_id}`. |
| `azuread_app_role_assignment` ×2 | `Directory.Read.All` and `AuditLog.Read.All`, when `grant_directory_read` is true. |

App roles resolve by name through the Microsoft Graph service principal.
The role assignment uses `role_definition_name = "Reader"`. No role GUID
is hardcoded.

## Inputs and outputs

Run `terraform-docs markdown .` for the generated table. The variable and
output descriptions in [`variables.tf`](variables.tf) and
[`outputs.tf`](outputs.tf) are the source of truth.

## Notes

- **Do not add a second audience.** Entra rejects a list. The audience
  must be exactly `api://AzureADTokenExchange`.
- **Do not pass the client ID to `application_id`.** That argument takes
  the application resource ID.
- **Do not use flexible federated identity credentials.** Those
  credentials support only GitHub, GitLab, and Terraform Cloud as
  issuers.
- **Do not grant `Contributor`, `Owner`, or any write permission.**
  `Reader` plus two read-only Graph roles is the whole grant.
- **`Reader` is enough.** `Reader` is `*/read`. That grant includes
  `Microsoft.Authorization/roleAssignments/read` and
  `roleDefinitions/read`. Do not add Security Reader.
- Requires the `azuread` provider at 3.9 or later and the `azurerm`
  provider at 5.4 or later.
- **Last login and MFA need Entra ID P1 or P2** and
  `AuditLog.Read.All`. Without the licence, or when China or DoD Graph
  does not expose `signInActivity` or `userRegistrationDetails`, those
  columns stay Unknown.
