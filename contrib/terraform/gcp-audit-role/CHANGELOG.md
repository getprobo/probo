# Changelog

All notable changes to the Terraform `getprobo/audit-role/gcp` module will
be documented in this file.

## Unreleased

### Changed

- Replace project-wide `roles/logging.viewer` with
  `roles/logging.viewAccessor` on the `_Required` bucket `_AllLogs`
  view. The impersonated token can read Admin Activity and the other
  `_Required` audit logs, not application logs in `_Default`. Apply a
  Probo release that queries that view before or with this module;
  applying the module first leaves last login unknown until Probo is
  upgraded. If Cloud Logging default resource settings moved `_Required`
  off `global`, set `required_bucket_location`.

## [0.1.1] - 2026-09-07

### Changed

- Document the optional Google Admin console Users-read role on
  `probo-audit` so the same WIF token can read 2-Step Verification
  enrollment. Cloud IAM cannot grant that; skip it and MFA stays unknown.

## [0.1.0] - 2026-09-03

### Added

- Initial module: Workload Identity Federation pool, OIDC provider, and
  read-only `probo-audit` service account for the GCP connector.
