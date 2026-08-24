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
connector's customer-side install. It creates an IAM OIDC provider for a Probo
organization's issuer, a read-only `ProboAudit` role trusting exactly that
issuer and subject, and — optionally — a service-managed StackSet that repeats
both in every member account of the organization.

Terraform parity lives in [`../../terraform/aws-audit-role`](../../terraform/aws-audit-role).

## What the customer runs

Deploy one stack in the AWS account that you connect to Probo.

Leave `DeployToOrganization` set to No. The stack creates the OIDC provider
and the `ProboAudit` role in that account only.

### Optional: cover the organization later

The template can create a service-managed StackSet. That StackSet creates the
same provider and role in every member account, including accounts added later.

Probo does not use those member roles yet. Set `DeployToOrganization` to Yes
only if you will cover the organization later. Then set
`OrganizationalUnitIds` to at least the organization root (`r-xxxx`). Run
that stack from the management account or from a CloudFormation delegated
administrator.

A service-managed StackSet needs CloudFormation trusted access in the
organization. Without trusted access, the stack fails on the StackSet resource.

## Publishing for one-click install

CloudFormation's quick-create console accepts a `templateURL` **on S3 only** —
not a GitHub raw URL, not an arbitrary HTTPS host, not a GitHub Release
asset.

The customer's CloudFormation fetches that object. A bucket that only Probo
IAM can read does **not** work: the customer has no credentials there, and
you cannot list every customer account in a bucket policy.

The object (or its prefix) must allow anonymous `s3:GetObject`. The rest of
the bucket can stay private. Block Public Access must allow that GetObject.
The object is still public. A presigned URL can point at a private object,
but it expires and is a poor fit for runbooks and change tickets.

One-click therefore requires the template in a Probo-owned bucket with
anonymous read on that key:

```bash
aws s3 cp contrib/cloudformation/aws-audit-role/aws-audit-role.yaml \
  "s3://${BUCKET}/aws-audit-role.yaml" \
  --cache-control "max-age=300"
```

Publish under a stable key. The URL ends up in customer runbooks and change
tickets, and a moved object breaks a stack update they attempt months later.

Self-hosted deployments that cannot publish to S3 give their customers the
upload-a-template path or the Terraform module instead. Serving the YAML from
probod is fine for download but **cannot** drive one-click.

## The quick-create link

```
https://<region>.console.aws.amazon.com/cloudformation/home
  ?region=<region>
  #/stacks/quickcreate
  ?templateURL=<url-encoded https S3 URL>
  &stackName=probo-audit
  &param_ProboIssuerURL=<url-encoded issuer>
  &param_ProboSubject=<url-encoded subject>
```

Every Probo-derived parameter must be prefilled. AWS compares the issuer URL
case-sensitively, and the issuer's last path segment is a mixed-case identifier,
so **the customer must never have to retype it** — a single flipped character
produces an `AccessDenied` with nothing in it that names the cause.

`RoleName` is left to its default unless the customer has a reason to change
it; if they do, the same name must be recorded on the connector. The audience
is not a parameter: Probo always mints `sts.amazonaws.com` and the connector
cannot be told another value.

## Verifying an install

Probo probes the install by assuming the role. Isolation is the
per-organization issuer: a foreign token fails at the provider-match step
before STS evaluates any trust-policy condition. The `sub` and `aud`
conditions in this template are IAM hygiene; Probo does not read them back.

## Editing the template

Two things are easy to break and produce opaque failures:

- **The trust policy is a JSON string, not a YAML mapping.** Its condition keys
  embed the issuer, and CloudFormation cannot use an intrinsic function as a
  mapping key. Converting it back to a mapping compiles and silently stops
  templating the keys.
- **The member-account template is embedded in the StackSet's `TemplateBody`.**
  Inside that `Fn::Sub`, `${!Foo}` renders as the literal `${Foo}` for the inner
  template to resolve per account, while `${Foo}` is substituted once here.
  Dropping an escape bakes the management account's ID into every member
  account's trust policy.

The read-only additions policy is duplicated between the stack's own role and
the embedded template; CloudFormation offers no way to share a fragment between
the two. Change both.
