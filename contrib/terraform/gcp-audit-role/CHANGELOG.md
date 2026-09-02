# Changelog

All notable changes to the Terraform `getprobo/audit-role/gcp` module will
be documented in this file.

## Unreleased

### Changed

- Document the optional Google Admin console Users-read role on
  `probo-audit` so the same WIF token can read 2-Step Verification
  enrollment. Cloud IAM cannot grant that; skip it and MFA stays unknown.

## [0.1.0] - 2026-09-03

### Added

- Initial module: Workload Identity Federation pool, OIDC provider, and
  read-only `probo-audit` service account for the GCP connector.
