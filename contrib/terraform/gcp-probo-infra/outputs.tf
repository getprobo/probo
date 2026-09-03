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

output "sql_instance_name" {
  value = google_sql_database_instance.probo.name
}

output "sql_private_ip_address" {
  value = google_sql_database_instance.probo.private_ip_address
}

output "database_name" {
  value = google_sql_database.probod.name
}

output "database_user" {
  value = google_sql_user.probod.name
}

output "bucket_name" {
  value = google_storage_bucket.files.name
}

output "storage_service_account_email" {
  value = google_service_account.storage.email
}

output "s3_access_key_id" {
  value = google_storage_hmac_key.probo.access_id
}

output "db_password_secret_id" {
  value = google_secret_manager_secret.db_password.secret_id
}

output "s3_hmac_secret_secret_id" {
  value = google_secret_manager_secret.s3_hmac_secret.secret_id
}
