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

variable "probo_issuer_url" {
  type        = string
  description = <<-EOT
    The issuer URL Probo mints its assertions under, unique to your Probo
    organization. Copy it exactly: AWS compares it case-sensitively and the
    last path segment is a mixed-case identifier.
  EOT

  validation {
    condition     = can(regex("^https://[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?(/[A-Za-z0-9._~%-]+)*/?$", var.probo_issuer_url))
    error_message = "The issuer must be an https:// URL with no port and no query string."
  }

  validation {
    # AWS rejects an OIDC provider URL longer than this.
    condition     = length(var.probo_issuer_url) <= 255
    error_message = "The issuer must be at most 255 characters."
  }
}

variable "probo_subject" {
  type        = string
  description = <<-EOT
    The subject claim Probo asserts, identifying your Probo organization. The
    role trusts this value and no other.
  EOT

  validation {
    condition     = can(regex("^[A-Za-z0-9_-]+$", var.probo_subject))
    error_message = "The subject must be the identifier Probo showed you."
  }
}

variable "role_name" {
  type        = string
  default     = "ProboAudit"
  description = <<-EOT
    Name of the role Probo assumes. Use the same name in every account, and
    record it on the connector if you change it.
  EOT

  validation {
    condition     = can(regex("^[\\w+=,.@-]{1,64}$", var.role_name))
    error_message = "The role name must be a valid IAM role name."
  }
}

variable "create_oidc_provider" {
  type        = bool
  default     = true
  description = <<-EOT
    Create the IAM OIDC provider for the issuer. Set false only when this
    account already has a provider for this exact URL: AWS allows one provider
    per issuer URL per account and a second create fails with
    EntityAlreadyExists. The role's trust policy names the provider by
    constructed ARN, so it is correct either way.
  EOT
}

variable "grant_organizations_read" {
  type        = bool
  default     = false
  description = <<-EOT
    Grant the Organizations reads Probo needs to enumerate the organization.
    Set true for the management account (or the delegated administrator) and
    false for member accounts, which cannot make these calls at all.
  EOT
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Tags to apply to the role and the OIDC provider."
}
