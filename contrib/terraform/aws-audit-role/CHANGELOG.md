# Changelog

All notable changes to the Terraform `getprobo/audit-role/aws` module will
be documented in this file.

## Unreleased

### Changed

- Document the public registry address `getprobo/audit-role/aws` as the
  module source. The connector now records a role ARN, not an account id
  and a role name.

## [0.1.3] - 2026-08-28

### Changed

- Maintenance release, no functional changes.

## [0.1.2] - 2026-08-28

### Changed

- Maintenance release, no functional changes.

## [0.1.1] - 2026-08-28

### Changed

- Maintenance release, no functional changes.

## [0.1.0] - 2026-08-27

### Added

- Initial release: Terraform module provisioning the OIDC provider and
  read-only `ProboAudit` role for the AWS connector, with an optional
  service-managed StackSet for organization-wide deployment.
