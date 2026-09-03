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

variable "project_id" {
  type        = string
  description = "GCP project id that owns the Cloud SQL instance, bucket, and secrets."
}

variable "region" {
  type        = string
  description = "Region for the Cloud SQL instance and the GCS bucket, e.g. us-central1."
}

variable "network" {
  type        = string
  default     = "default"
  description = "VPC (by name) that the GKE cluster runs in. Cloud SQL is peered onto this network so probod reaches it over a private IP."
}

variable "sql_instance_name" {
  type        = string
  default     = "probo-db"
  description = "Cloud SQL instance name. Keep in sync with deploy-gcp.yaml, which looks the instance up by this name."
}

variable "sql_tier" {
  type        = string
  default     = "db-custom-2-8192"
  description = "Cloud SQL machine tier (2 vCPU / 8 GiB by default)."
}

variable "database_name" {
  type        = string
  default     = "probod"
  description = "Database name. Keep in sync with deploy-gcp.yaml's POSTGRES_DATABASE."
}

variable "database_user" {
  type        = string
  default     = "probod"
  description = "Database user. Keep in sync with deploy-gcp.yaml's POSTGRES_USERNAME."
}

variable "bucket_name" {
  type        = string
  description = "Globally unique GCS bucket name for probod's file storage."
}

variable "storage_service_account_id" {
  type        = string
  default     = "probo-storage"
  description = "Service account id (not email) that holds the HMAC key probod uses to talk to the bucket over the S3-compatible API."
}

variable "deploy_service_account_email" {
  type        = string
  description = "Email of the service account behind deploy-gcp.yaml's GCP_SA_KEY. Granted read access to the two secrets this module creates so the deploy workflow can fetch them without a manual copy-paste step."
}
