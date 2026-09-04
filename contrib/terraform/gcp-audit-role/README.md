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

# `getprobo/audit-role/gcp`

Grants Probo read-only audit access to one GCP project through Workload
Identity Federation. Probo holds no credential for the project: it presents a
short-lived signed assertion that STS verifies against a public key set, and
you revoke access by deleting the pool or the service account.

This module covers **one project**. Apply it in the project you are
connecting. It does not walk folders or the organization.

## Usage

Copy the issuer URL and the subject out of the connector setup screen. GCP
compares the issuer URL case-sensitively and its last path segment is a
mixed-case identifier, so paste it rather than retyping it.

The Google provider must target the project you want to connect. Set
`project` on the provider, or export `GOOGLE_CLOUD_PROJECT`.

Enable `iam.googleapis.com`, `cloudresourcemanager.googleapis.com`,
`sts.googleapis.com`, and `iamcredentials.googleapis.com` in that
project before you apply. Terraform uses the first two. Probo uses STS
and IAM Credentials to exchange a token and impersonate the service
account.

```hcl
module "probo_audit" {
  source = "getprobo/audit-role/gcp"

  probo_issuer_url     = "https://proboidentity.com/e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
  probo_subject        = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
  service_account_name = "probo-audit"
}

output "probo_workload_identity_provider" {
  value = module.probo_audit.workload_identity_provider
}

output "probo_service_account_email" {
  value = module.probo_audit.service_account_email
}
```

Give Probo the `workload_identity_provider` and `service_account_email`
outputs when you create the connector.

## Verifying an install

Probo probes the install by exchanging a token and impersonating the service
account. Isolation is the per-organization issuer: a foreign token fails at
the provider-match step before GCP evaluates the attribute condition. The
`assertion.sub` condition in this module is IAM hygiene; Probo does not read
it back.

## What it creates

| Resource | Notes |
|---|---|
| `google_iam_workload_identity_pool` | `probo` by default. |
| `google_iam_workload_identity_pool_provider` | OIDC. Issuer is `probo_issuer_url`. `allowed_audiences` is unset. |
| `google_service_account` | `probo-audit` by default. |
| `roles/iam.securityReviewer` | Project IAM, additive. |
| `roles/iam.serviceAccountViewer` | Project IAM, additive. |
| `roles/logging.viewer` | Project IAM, additive. |
| `roles/policyanalyzer.activityAnalysisViewer` | Project IAM, additive. |
| `roles/iam.workloadIdentityUser` | On the service account, for `principal://…/subject/{probo_subject}` only. |

The attribute condition pins `assertion.sub` with CEL `==`. Isolation is the
per-organization issuer: a foreign token fails at the provider-match step
before GCP evaluates the condition. That pin is IAM hygiene; Probo does not
read it back.

## Inputs and outputs

Run `terraform-docs markdown .` for the generated table. The variable and
output descriptions in [`variables.tf`](variables.tf) and
[`outputs.tf`](outputs.tf) are the source of truth.

## Notes

- **Do not set `allowed_audiences`.** An empty list tells GCP to accept the
  default provider URL, with or without the `https:` prefix. Probo mints that
  URL as the JWT `aud`.
- **The subject condition is exact equality.** A `startsWith` wildcard would
  let any subject this issuer can mint impersonate the service account.
- **The four project roles are additive members**, not bindings. A binding
  would replace every other member of that role in the project.
- Requires the `google` provider at 5.0 or later.
