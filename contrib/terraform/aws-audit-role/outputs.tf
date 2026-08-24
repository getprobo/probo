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

output "role_arn" {
  value       = aws_iam_role.probo_audit.arn
  description = "ARN of the role Probo assumes in this account."
}

output "role_name" {
  value       = aws_iam_role.probo_audit.name
  description = "Name of the role. Use the same name in every account."
}

output "account_id" {
  value       = data.aws_caller_identity.current.account_id
  description = "The AWS account this module created the role in."
}

output "oidc_provider_arn" {
  value       = local.oidc_provider_arn
  description = <<-EOT
    ARN of the OIDC provider the role trusts, whether or not this module
    created it.
  EOT
}

output "issuer_host_path" {
  value       = local.issuer_host_path
  description = <<-EOT
    The issuer with the scheme stripped, which is the prefix of the trust
    policy condition keys. Useful when diagnosing an AccessDenied.
  EOT
}
