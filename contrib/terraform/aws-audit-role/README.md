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

# `aws-audit-role`

Grants Probo read-only audit access to one AWS account through OIDC web
identity federation. Probo holds no credential for the account: it presents a
short-lived signed assertion that STS verifies against a public key set, and
you revoke access by deleting the role.

Parity with
[`../../cloudformation/aws-audit-role/aws-audit-role.yaml`](../../cloudformation/aws-audit-role/aws-audit-role.yaml),
for teams that would rather not run a stack they did not write.

## Usage

Copy the two Probo values out of the connector setup screen. AWS compares the
issuer URL case-sensitively and its last path segment is a mixed-case
identifier, so paste it rather than retyping it.

```hcl
module "probo_audit" {
  source = "github.com/getprobo/probo//contrib/terraform/aws-audit-role"

  probo_issuer_url = "https://proboidentity.com/org/e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
  probo_subject    = "e5IaD7ibAAEAAAAAAZZ9aR_Oq_Npymhg"
}

output "probo_role_arn" {
  value = module.probo_audit.role_arn
}
```

Give Probo the `account_id` and `role_name` outputs when you create the
connector.

## Covering a whole organization

This module covers **one account**. Organization-wide rollout is the caller's
responsibility, because how you reach your member accounts is yours to decide
and the module cannot guess it.

Probo currently reviews only the connected account. Apply the module in other
accounts now if you want those roles ready when org-wide coverage ships.

Use one provider alias per account with a role you already have there:

```hcl
module "probo_audit_member" {
  for_each = toset(var.member_account_ids)
  source   = "github.com/getprobo/probo//contrib/terraform/aws-audit-role"

  providers = { aws = aws.member[each.key] }

  probo_issuer_url = var.probo_issuer_url
  probo_subject    = var.probo_subject

  grant_organizations_read = false
}
```

`for_each` covers the accounts that exist when you apply. If you want accounts
added later to be onboarded without a second apply, use a service-managed
`aws_cloudformation_stack_set` with auto-deployment — which is what the
CloudFormation template does — or accept re-applying.

## What it creates

| Resource | Notes |
|---|---|
| `aws_iam_openid_connect_provider` | Skipped when `create_oidc_provider = false`. AWS allows one per issuer URL per account. |
| `aws_iam_role` | `ProboAudit` by default, `max_session_duration = 3600`. |
| `SecurityAudit`, `job-function/ViewOnlyAccess` | AWS-managed, read-only. |
| `ProboAuditIdentityRead` | IAM Identity Center reads the two managed policies miss. |
| `ProboAuditOrganizationsRead` | Only with `grant_organizations_read = true`. |

The trust policy pins both `aud` and `sub` with `StringEquals`. Isolation is
the per-organization issuer: a foreign token fails at the provider-match step
before STS evaluates any condition. Those pins are IAM hygiene; Probo does not
read them back.

## Inputs and outputs

Run `terraform-docs markdown .` for the generated table; the variable and output
descriptions in [`variables.tf`](variables.tf) and [`outputs.tf`](outputs.tf)
are the source of truth.

## Notes

- **The role must have the same name in every account.** Probo stores one role
  name per connector and assumes it everywhere.
- **`thumbprint_list` is deliberately unset**, unlike the CloudFormation
  template which must supply a placeholder. AWS ignores it for publicly trusted
  certificates, so pinning one only produces a perpetual diff.
- Requires the `aws` provider at 5.0 or later, where `thumbprint_list` became
  optional.
