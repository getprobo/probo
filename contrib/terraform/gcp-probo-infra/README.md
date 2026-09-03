# gcp-probo-infra

Provisions the stateful GCP infrastructure `probod` needs beyond the GKE
cluster itself: a Cloud SQL Postgres instance (private IP, peered onto the
cluster's VPC), a GCS bucket for file storage, a service account + HMAC key
pair so `probod` can talk to that bucket over its S3-compatible API, and two
Secret Manager secrets (the generated DB password and the HMAC secret key)
that [`deploy-gcp.yaml`](../../../.github/workflows/deploy-gcp.yaml) reads at
deploy time.

Applied via [`provision-gcp-infra.yaml`](../../../.github/workflows/provision-gcp-infra.yaml),
a manual (`workflow_dispatch`) workflow — deliberately not tied to every push,
since this touches a stateful database.

## One-time bootstrap (before the first run)

Terraform can't create its own state bucket, and the APIs this module
enables can't be enabled by a plan that needs them already enabled. Run once,
by hand, against the target project:

```bash
gcloud config set project YOUR_PROJECT_ID

gcloud services enable \
  sqladmin.googleapis.com \
  storage.googleapis.com \
  secretmanager.googleapis.com \
  servicenetworking.googleapis.com \
  compute.googleapis.com

# Terraform state bucket — name it, and set TF_STATE_BUCKET to it.
gcloud storage buckets create gs://YOUR_TF_STATE_BUCKET --location=YOUR_REGION
```

Then add these repository secrets (in addition to the ones `deploy-gcp.yaml`
already needs — `GCP_SA_KEY`, `GCP_PROJECT_ID`):

| Secret | Example |
|---|---|
| `TF_STATE_BUCKET` | `probo-tfstate` |
| `GCP_REGION` | `us-central1` |
| `GCS_BUCKET_NAME` | `probo-production-files` (must be globally unique) |

Run `provision-gcp-infra.yaml` with `action: plan` first to review, then
again with `action: apply`.

## What it doesn't do

- Doesn't create the GKE cluster or the VPC — `var.network` (default
  `"default"`) must name a VPC that already exists and that the cluster's
  nodes are on.
- Doesn't rotate the DB password or HMAC key. Re-running `apply` after a
  manual rotation in the console will fight the state; rotate through
  Terraform (`terraform taint random_password.db`, or the HMAC key
  resource) instead.
- `deletion_protection = true` on the Cloud SQL instance is a guard against
  `terraform destroy`, not a backup strategy — point-in-time recovery is
  enabled, but test a real restore before you need one.
