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
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 3.9"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 5.4"
    }
  }
}

data "azuread_client_config" "current" {}

data "azuread_application_published_app_ids" "well_known" {}

data "azuread_service_principal" "msgraph" {
  client_id = data.azuread_application_published_app_ids.well_known.result.MicrosoftGraph
}

resource "azuread_application_registration" "probo" {
  display_name = var.application_display_name
}

resource "azuread_application_federated_identity_credential" "probo" {
  application_id = azuread_application_registration.probo.id
  display_name   = var.federated_credential_name
  issuer         = var.probo_issuer_url
  subject        = var.probo_subject
  audiences      = ["api://AzureADTokenExchange"]
}

resource "azuread_service_principal" "probo" {
  client_id = azuread_application_registration.probo.client_id
}

resource "azurerm_role_assignment" "reader" {
  scope                            = "/subscriptions/${var.subscription_id}"
  role_definition_name             = "Reader"
  principal_id                     = azuread_service_principal.probo.object_id
  skip_service_principal_aad_check = true
}

resource "azuread_app_role_assignment" "directory_read" {
  count = var.grant_directory_read ? 1 : 0

  app_role_id         = data.azuread_service_principal.msgraph.app_role_ids["Directory.Read.All"]
  principal_object_id = azuread_service_principal.probo.object_id
  resource_object_id  = data.azuread_service_principal.msgraph.object_id
}

resource "azuread_app_role_assignment" "audit_log_read" {
  count = var.grant_directory_read ? 1 : 0

  app_role_id         = data.azuread_service_principal.msgraph.app_role_ids["AuditLog.Read.All"]
  principal_object_id = azuread_service_principal.probo.object_id
  resource_object_id  = data.azuread_service_principal.msgraph.object_id
}
