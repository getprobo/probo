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
    organization. Copy it exactly: Entra compares it case-sensitively and the
    last path segment is a mixed-case identifier.
  EOT

  validation {
    condition     = can(regex("^https://[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?(/[A-Za-z0-9._~%-]+)*/?$", var.probo_issuer_url))
    error_message = "The issuer must be an https:// URL with no port and no query string."
  }
}

variable "probo_subject" {
  type        = string
  description = <<-EOT
    The subject claim Probo asserts, identifying your Probo organization. The
    federated credential trusts this value and no other. A wrong subject
    creates successfully and fails later at token exchange.
  EOT

  validation {
    condition     = can(regex("^[A-Za-z0-9_-]+$", var.probo_subject))
    error_message = "The subject must be the identifier Probo showed you."
  }
}

variable "subscription_id" {
  type        = string
  description = <<-EOT
    Azure subscription ID that Probo reviews. The module grants Reader on this
    subscription only.
  EOT

  validation {
    condition     = can(regex("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$", var.subscription_id))
    error_message = "The subscription ID must be a UUID."
  }
}

variable "application_display_name" {
  type        = string
  default     = "Probo Access Review"
  description = <<-EOT
    Display name of the Entra application registration. Change it only when
    this tenant already has an application with this name.
  EOT

  validation {
    condition     = length(var.application_display_name) > 0
    error_message = "The application display name must not be empty."
  }
}

variable "federated_credential_name" {
  type        = string
  default     = "probo-access-review"
  description = <<-EOT
    Name of the federated identity credential. Entra stores this as Graph
    name: 3 to 120 characters, letters, digits, dash, and underscore, first
    character alphanumeric. The name is immutable after create.
  EOT

  validation {
    condition     = can(regex("^[A-Za-z0-9][A-Za-z0-9_-]{2,119}$", var.federated_credential_name))
    error_message = "The federated credential name must be 3 to 120 characters of [A-Za-z0-9_-] and must start with a letter or digit."
  }
}

variable "grant_directory_read" {
  type        = bool
  default     = true
  description = <<-EOT
    Grant Directory.Read.All and AuditLog.Read.All on Microsoft Graph. Set
    false when the operator who applies this module cannot give admin
    consent. Principals then resolve as GUIDs until a directory
    administrator applies the grants.
  EOT
}
