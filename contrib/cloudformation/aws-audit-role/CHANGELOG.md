# Changelog

All notable changes to the CloudFormation `aws-audit-role` template will be
documented in this file.

## Unreleased

## [0.1.2] - 2026-09-01

### Changed

- Document the connect-screen flow: Probo builds the quick-create URL, and
  the customer pastes the `RoleARN` output back.
- Restore the self-hosted fallbacks: upload-a-template or Terraform, and
  that a YAML file served by probod cannot drive one-click.

## [0.1.1] - 2026-08-28

### Changed

- Maintenance release, no functional changes.

## [0.1.0] - 2026-08-27

### Added

- Initial release: CloudFormation template provisioning the OIDC provider
  and read-only `ProboAudit` role for the AWS connector, with an optional
  service-managed StackSet for organization-wide deployment.
