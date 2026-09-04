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
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }
  }
}

data "google_project" "current" {}

locals {
  # The federated principal is the organization subject on this pool, not a
  # principalSet. A wildcard here would let any subject this issuer can mint
  # impersonate the service account.
  workload_identity_user = "principal://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${google_iam_workload_identity_pool.probo.workload_identity_pool_id}/subject/${var.probo_subject}"
}

resource "google_iam_workload_identity_pool" "probo" {
  workload_identity_pool_id = var.pool_id
  display_name              = "Probo"
  description               = "Workload Identity Federation pool for Probo (${var.probo_issuer_url})"
}

resource "google_iam_workload_identity_pool_provider" "probo" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.probo.workload_identity_pool_id
  workload_identity_pool_provider_id = var.provider_id
  display_name                       = "Probo"
  description                        = "OIDC provider for Probo (${var.probo_issuer_url})"

  # google.subject is the only attribute Probo needs. The condition pins the
  # subject to this organization; CEL == is StringEquals, not a prefix match.
  attribute_mapping = {
    "google.subject" = "assertion.sub"
  }

  attribute_condition = "assertion.sub == \"${var.probo_subject}\""

  oidc {
    issuer_uri = var.probo_issuer_url
    # allowed_audiences is deliberately unset. An empty list tells GCP to
    # accept the default provider URL, with or without the https: prefix.
    # Probo mints that URL as the JWT aud; the STS exchange uses the // form.
  }
}

resource "google_service_account" "probo_audit" {
  account_id   = var.service_account_name
  display_name = "Probo Audit"
  description  = "Read-only audit access for Probo (${var.probo_issuer_url})"
}

resource "google_project_iam_member" "security_reviewer" {
  project = data.google_project.current.project_id
  role    = "roles/iam.securityReviewer"
  member  = "serviceAccount:${google_service_account.probo_audit.email}"
}

resource "google_project_iam_member" "service_account_viewer" {
  project = data.google_project.current.project_id
  role    = "roles/iam.serviceAccountViewer"
  member  = "serviceAccount:${google_service_account.probo_audit.email}"
}

resource "google_project_iam_member" "logging_viewer" {
  project = data.google_project.current.project_id
  role    = "roles/logging.viewer"
  member  = "serviceAccount:${google_service_account.probo_audit.email}"
}

resource "google_project_iam_member" "policy_analyzer" {
  project = data.google_project.current.project_id
  role    = "roles/policyanalyzer.activityAnalysisViewer"
  member  = "serviceAccount:${google_service_account.probo_audit.email}"
}

resource "google_service_account_iam_member" "workload_identity_user" {
  service_account_id = google_service_account.probo_audit.name
  role               = "roles/iam.workloadIdentityUser"
  member             = local.workload_identity_user
}
