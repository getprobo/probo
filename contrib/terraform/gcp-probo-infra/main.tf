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

locals {
  apis = [
    "sqladmin.googleapis.com",
    "storage.googleapis.com",
    "secretmanager.googleapis.com",
    "servicenetworking.googleapis.com",
    "compute.googleapis.com",
  ]
}

resource "google_project_service" "this" {
  for_each = toset(local.apis)

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

data "google_compute_network" "vpc" {
  name = var.network
}

# Cloud SQL's private-IP path requires a reserved peering range and a
# service-networking connection onto the same VPC the GKE cluster uses.
resource "google_compute_global_address" "private_ip_range" {
  name          = "probo-db-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = data.google_compute_network.vpc.id

  depends_on = [google_project_service.this]
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = data.google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_range.name]
}

resource "random_password" "db" {
  length  = 32
  special = false
}

resource "google_sql_database_instance" "probo" {
  name             = var.sql_instance_name
  project          = var.project_id
  region           = var.region
  database_version = "POSTGRES_16"

  depends_on = [google_service_networking_connection.private_vpc_connection]

  settings {
    tier = var.sql_tier

    ip_configuration {
      ipv4_enabled    = false
      private_network = data.google_compute_network.vpc.id
    }

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
    }
  }

  # A safety net, not a substitute for a backup/restore drill: this only
  # stops `terraform destroy` from taking the instance out from under you.
  deletion_protection = true
}

resource "google_sql_database" "probod" {
  name     = var.database_name
  project  = var.project_id
  instance = google_sql_database_instance.probo.name
}

resource "google_sql_user" "probod" {
  name     = var.database_user
  project  = var.project_id
  instance = google_sql_database_instance.probo.name
  password = random_password.db.result
}

resource "google_storage_bucket" "files" {
  name                        = var.bucket_name
  project                     = var.project_id
  location                    = var.region
  uniform_bucket_level_access = true
  force_destroy               = false

  versioning {
    enabled = true
  }
}

# HMAC keys are only issued for service accounts, which is why probod talks
# to GCS through a dedicated one rather than the workload's own identity.
resource "google_service_account" "storage" {
  account_id   = var.storage_service_account_id
  project      = var.project_id
  display_name = "probod S3-compatible access to ${var.bucket_name}"
}

resource "google_storage_bucket_iam_member" "storage_access" {
  bucket = google_storage_bucket.files.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.storage.email}"
}

resource "google_storage_hmac_key" "probo" {
  service_account_email = google_service_account.storage.email
  project                = var.project_id
}

resource "google_secret_manager_secret" "db_password" {
  secret_id = "probo-db-password"
  project   = var.project_id

  replication {
    auto {}
  }

  depends_on = [google_project_service.this]
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db.result
}

resource "google_secret_manager_secret" "s3_hmac_secret" {
  secret_id = "probo-s3-hmac-secret"
  project   = var.project_id

  replication {
    auto {}
  }

  depends_on = [google_project_service.this]
}

resource "google_secret_manager_secret_version" "s3_hmac_secret" {
  secret      = google_secret_manager_secret.s3_hmac_secret.id
  secret_data = google_storage_hmac_key.probo.secret
}

# Lets deploy-gcp.yaml read both secrets with the same service account it
# already authenticates as, instead of a manual copy-paste into GitHub Secrets.
resource "google_secret_manager_secret_iam_member" "deploy_reads_db_password" {
  secret_id = google_secret_manager_secret.db_password.id
  project   = var.project_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.deploy_service_account_email}"
}

resource "google_secret_manager_secret_iam_member" "deploy_reads_s3_hmac_secret" {
  secret_id = google_secret_manager_secret.s3_hmac_secret.id
  project   = var.project_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.deploy_service_account_email}"
}
