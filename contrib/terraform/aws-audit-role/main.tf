/*
 * Copyright (c) 2026 Probo Inc <hello@probo.com>.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      # 5.x is where thumbprint_list became optional on
      # aws_iam_openid_connect_provider. See the resource below.
      version = ">= 5.0"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

locals {
  # The prefix of the trust policy's condition keys and the suffix of the OIDC
  # provider ARN are both the issuer with the scheme stripped. Deriving it once
  # is what stops the two from disagreeing.
  issuer_host_path = trimprefix(var.probo_issuer_url, "https://")

  oidc_provider_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${local.issuer_host_path}"

  # Probo always mints this audience; the connector cannot be told another.
  audience = "sts.amazonaws.com"
}

# thumbprint_list is deliberately absent, unlike the CloudFormation template
# which must supply a placeholder because its validation rejects a null list.
# AWS validates publicly trusted certificates from its own CA library and
# ignores the value, so pinning one here would only produce a perpetual diff
# every time AWS reports back what it actually uses.
resource "aws_iam_openid_connect_provider" "probo" {
  count = var.create_oidc_provider ? 1 : 0

  url            = var.probo_issuer_url
  client_id_list = [local.audience]

  tags = var.tags
}

# The provider is named by constructed ARN rather than by reference, so this
# policy is identical whether or not this module created it.
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [local.oidc_provider_arn]
    }

    # StringEquals, never StringLike. A wildcard would let any subject this
    # issuer can mint assume the role.
    condition {
      test     = "StringEquals"
      variable = "${local.issuer_host_path}:aud"
      values   = [local.audience]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.issuer_host_path}:sub"
      values   = [var.probo_subject]
    }
  }
}

resource "aws_iam_role" "probo_audit" {
  name        = var.role_name
  description = "Read-only audit access for Probo (${var.probo_issuer_url})"

  assume_role_policy = data.aws_iam_policy_document.assume_role.json

  # One hour, matching the token lifetime Probo requests. Raising it widens the
  # window a leaked session credential stays usable.
  max_session_duration = 3600

  tags = var.tags

  depends_on = [aws_iam_openid_connect_provider.probo]
}

resource "aws_iam_role_policy_attachment" "security_audit" {
  role       = aws_iam_role.probo_audit.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/SecurityAudit"
}

resource "aws_iam_role_policy_attachment" "view_only_access" {
  role       = aws_iam_role.probo_audit.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/job-function/ViewOnlyAccess"
}

data "aws_iam_policy_document" "identity_read" {
  # SecurityAudit and ViewOnlyAccess between them miss IAM Identity Center,
  # which is where the human access in an organization actually lives. Every
  # action listed is a read.
  statement {
    effect = "Allow"

    actions = [
      "identitystore:Describe*",
      "identitystore:Get*",
      "identitystore:List*",
      "sso:Describe*",
      "sso:Get*",
      "sso:List*",
      "sso-directory:Describe*",
      "sso-directory:List*",
      "sso-directory:Search*",
    ]

    resources = ["*"]
  }

  # Credential state is only readable through a generated report. The Generate*
  # calls write nothing outside IAM's own report store and are the documented
  # way to read it.
  statement {
    effect = "Allow"

    actions = [
      "iam:GenerateCredentialReport",
      "iam:GenerateServiceLastAccessedDetails",
      "iam:GetCredentialReport",
      "iam:GetServiceLastAccessedDetails",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "identity_read" {
  name   = "ProboAuditIdentityRead"
  role   = aws_iam_role.probo_audit.id
  policy = data.aws_iam_policy_document.identity_read.json
}

data "aws_iam_policy_document" "organizations_read" {
  statement {
    effect = "Allow"

    actions = [
      "organizations:DescribeAccount",
      "organizations:DescribeOrganization",
      "organizations:ListAccounts",
      "organizations:ListAccountsForParent",
      "organizations:ListOrganizationalUnitsForParent",
      "organizations:ListRoots",
      "organizations:ListTagsForResource",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "organizations_read" {
  count = var.grant_organizations_read ? 1 : 0

  name   = "ProboAuditOrganizationsRead"
  role   = aws_iam_role.probo_audit.id
  policy = data.aws_iam_policy_document.organizations_read.json
}
