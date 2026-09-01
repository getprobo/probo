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

# CloudFormation setup artifacts

[`aws-audit-role.yaml`](aws-audit-role.yaml) is the source of truth for the AWS
connector's customer-side install. The template creates an IAM OIDC provider
for the issuer of a Probo organization. It also creates a read-only
`ProboAudit` role that trusts that issuer and subject. Optionally, a
service-managed StackSet repeats both in every member account of the
organization.

Terraform parity lives in [`../../terraform/aws-audit-role`](../../terraform/aws-audit-role).

## What the customer runs

Open the AWS connect screen in Probo. The screen gives a CloudFormation link
with the issuer, subject, and role name already filled. Click that link.

Leave `DeployToOrganization` set to No. Create the stack in the AWS account
that you connect to Probo. Copy the `RoleARN` output. Paste that ARN into
Probo.

The stack creates the OIDC provider and the `ProboAudit` role in that account
only.

### Optional: cover the organization later

The template can create a service-managed StackSet. That StackSet creates the
same provider and role in every member account. Accounts that join later get
the same resources.

Probo does not use those member roles yet. If you will cover the organization
later, set `DeployToOrganization` to Yes. Then set `OrganizationalUnitIds` to
at least the organization root (`r-xxxx`). Run that stack from the management
account or from a CloudFormation delegated administrator.

A service-managed StackSet needs CloudFormation trusted access in the
organization. Without trusted access, the stack fails on the StackSet resource.

## Publishing for one-click install

Probo publishes the template at
`https://probo-cloudformation-template.s3.us-east-2.amazonaws.com/aws-audit-role.yaml`.
The CloudFormation quick-create console accepts a `templateURL` **on S3
only**. A GitHub raw URL, an arbitrary HTTPS host, and a GitHub Release asset
do not work.

CloudFormation in the customer account fetches that object. A bucket that
only Probo IAM can read does **not** work. The customer has no credentials
there, and you cannot list every customer account in a bucket policy.

The object (or its prefix) must allow anonymous `s3:GetObject`. The rest of
the bucket can stay private. Block Public Access must allow that GetObject.
The object is still public. A presigned URL can point at a private object,
but it expires. Runbooks and change tickets need a URL that does not expire.

One-click therefore requires the template in a Probo-owned bucket with
anonymous read on that key:

```bash
aws s3 cp contrib/cloudformation/aws-audit-role/aws-audit-role.yaml \
  "s3://${BUCKET}/aws-audit-role.yaml" \
  --cache-control "max-age=300"
```

Publish under a stable key. The URL is recorded in customer runbooks and
change tickets. A moved object breaks a stack update that they attempt months
later.

If a self-hosted deployment cannot publish to S3, give the customers the
upload-a-template path or the Terraform module instead. You can serve the
YAML from probod for download. That URL cannot drive one-click.

## The quick-create link

Probo builds this URL. The customer must not assemble it by hand. The link
opens the console in `us-east-1` with `stackName=probo-audit` and with
`ProboIssuerURL`, `ProboSubject`, and `RoleName` already filled. IAM is
global, so that console region is not the home region of the account.

```
https://us-east-1.console.aws.amazon.com/cloudformation/home
  ?region=us-east-1
  #/stacks/quickcreate
  ?templateURL=<url-encoded https S3 URL>
  &stackName=probo-audit
  &param_ProboIssuerURL=<url-encoded issuer>
  &param_ProboSubject=<url-encoded subject>
  &param_RoleName=ProboAudit
```

Every Probo-derived parameter must be prefilled. AWS compares the issuer URL
with case sensitivity. The last path segment of the issuer is a mixed-case
identifier. The customer must never retype it. One flipped character produces
an `AccessDenied` with nothing in it that names the cause.

If the customer changes `RoleName`, the ARN that they paste into Probo
changes with it. The audience is not a parameter. Probo always mints
`sts.amazonaws.com`. The connector cannot use another value.

## Probing an install

When the customer pastes the ARN, Probo probes the install by assuming the
role. Isolation is the per-organization issuer. A foreign token fails at the
provider-match step before STS evaluates any trust-policy condition. The
`sub` and `aud` conditions in this template are IAM hygiene. Probo does not
read them back.

If the probe fails, make sure that the stack exists in that account. Make
sure that the issuer and subject match the values on the connect screen.

## Editing the template

Two things are easy to break and produce opaque failures:

- **The trust policy is a JSON string, not a YAML mapping.** Its condition keys
  embed the issuer. CloudFormation cannot use an intrinsic function as a
  mapping key. If you convert it back to a mapping, the template compiles and
  the keys no longer get substitutions.
- **The member-account template is embedded in the StackSet's `TemplateBody`.**
  Inside that `Fn::Sub`, `${!Foo}` renders as the literal `${Foo}` for the inner
  template to resolve per account, while `${Foo}` is substituted once here.
  If you drop an escape, every member-account trust policy gets the management
  account ID.

The read-only additions policy is duplicated between the stack's own role and
the embedded template. CloudFormation offers no way to share a fragment
between the two. Change both.
