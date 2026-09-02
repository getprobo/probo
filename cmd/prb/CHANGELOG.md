# Changelog

All notable changes to the `prb` CLI will be documented in this file.

## Unreleased

## [0.224.0] - 2026-09-02

### Added

- `task comment create`, `list`, `view`, `update`, and `delete` manage description-only comments on a task, owned by a membership profile (the author by default)

## [0.223.0] - 2026-09-01

### Added

- `access-review source setup-aws` returns the AWS connector setup values (issuer, audience, subject, suggested role name, Terraform snippet, and CloudFormation quick-create URL) needed to create the audit role
- `--aws-role-arn` on `access-review source create` creates an AWS workload-identity connector and the access source in one call, deleting the connector again if source creation fails

### Changed

- `access-review source view` shows the connector's connection status

## [0.222.0] - 2026-09-01

### Added

- `--as-of` on `risk-analysis view` and `treatment-plan list` reconstructs matrix cells and plans as of an RFC3339 instant; on `treatment-plan list` it requires `--risk-analysis`, and omitting it keeps reading live tables
- `--state` on `task create`, and `task create` / `task update` now validate the state against the server enum, which gained `BACKLOG`, `CANCELED`, and `DUPLICATE`

### Changed

- `treatment-plan list` and `treatment-plan view` report the plan's own category, falling back to the risk category when the plan has none

## [0.221.0] - 2026-08-31

### Added

- `risk-analysis fork <id>` copies a risk analysis's diagrams, treatment plans, and their relations into a new analysis for a later period

### Changed

- `access-review source create` reports when the source already exists for the connector instead of always claiming a fresh creation

## [0.220.0] - 2026-08-26

### Added

- `treatment-plan` command group (`create`, `list`, `view`, `update`, `delete`) for plans that hold scores, treatment, and owner per (risk, analysis)
- `measure link-treatment-plan` / `unlink-treatment-plan` to attach measures that drive plan progress
- `prb risk list --risk-analysis` lists unplanned scenario-linked risks on an analysis

### Changed

- `prb risk-analysis update` no longer accepts `--matrix-rows` / `--matrix-cols`; matrix size is fixed at create

## [0.219.0] - 2026-08-25

### Added

- `tracker-pattern list` / `view` show the catalog attribution (vendor name, first-party, or still-identifying) inherited from the linked catalog entry, for patterns with no vendor of their own

## [0.218.0] - 2026-08-24

### Changed

- `prb access-review entry list --auth-method` now accepts `OAUTH2` and `SSH`, matching GitHub credential rows (OAuth app tokens and deploy keys) that were previously indistinguishable from API keys or service accounts

## [0.217.0] - 2026-08-14

### Added

- `--matrix-rows` / `--matrix-cols` flags (required, 3, 4, or 5) on `prb risk-analysis create` and `update`, shown as a column on `list` and `view`, so the likelihood/impact matrix size is explicit per analysis

## [0.216.0] - 2026-08-13

### Added

- `COMPLIANCE_PORTAL_MANAGER` and `COMPLIANCE_PORTAL_ACCESS_MANAGER` membership roles, filterable via `prb user list --role`, for delegating full compliance portal management or just visitor access approval without broader admin access
- `prb aisystem` command group (`create`, `list`, `view`, `update`, `delete`, `publish`), for managing an organization's AI Systems register
- `--period-start` and `--period-end` flags on `prb risk-analysis create` and `update`, shown as columns on `list` and `view`
- `--resource-reporting-enabled` flag on `prb cookie-banner update`, to enable or disable resource detection on the banner

## [0.215.0] - 2026-08-11

### Changed

- `prb access-review entry list` prints the admin column as `yes`, `no`, or `unknown`, so an account the source does not report on is no longer indistinguishable from a confirmed non-admin

## [0.214.0] - 2026-08-10

### Changed

- Risk analysis scopes are now referred to as diagrams across the CLI

## [0.213.0] - 2026-08-07

### Changed

- Compliance portal document, audit, and subprocessor selection is reversed: the console now lists the organization's own documents, audits, and subprocessors with checkboxes for inline portal membership, replacing the portal-only rows plus separate add dialogs
- Risk assessments are now named risk analyses across the CLI and MCP API

## [0.212.0] - 2026-08-06

### Added

- `--rights-requests-enabled` flag on `prb compliance-portal update`, to enable or disable the public Requests surface on the compliance portal

## [0.211.0] - 2026-08-05

### Added

- `SUBDIVISION` column on `prb consent-record list` and the subdivision code on `prb consent-record view`, showing the ISO 3166-2 state or province detected for each consent record

## [0.210.0] - 2026-08-05

### Added

- `prb businessfunction` commands: `create`, `update`, `delete`, `list`, `view`, and `publish`, for tracking DORA Critical ICT Functions

## [0.209.0] - 2026-08-03

### Added

- `prb compliance-portal` commands: `create`, `update`, `delete`, `list`, and `view`, plus `audit`, `document`, and `third-party` `update`/`delete` subcommands, and portal-scoped `file list` / `reference create`/`list`
- `--administrator-id` flag on `prb thirdparty create`/`update`, and administrators shown on `prb thirdparty view`

## [0.208.0] - 2026-07-31

### Added

- `prb device list`, `view`, `create`, `revoke`, `delete`, and `set-owner` commands to manage ITAM devices

## [0.207.1] - 2026-07-30

### Added

- `prb login` now requests the `v1:itam`/`v1:itam:read` OAuth2 scopes, in anticipation of upcoming ITAM commands in the CLI

## [0.207.0] - 2026-07-30

### Changed

- `prb user list --state` now accepts multiple values (repeat the flag or comma-separate) and filters by `PENDING`, `ACTIVE`, or `DEACTIVATED` instead of `ACTIVE`/`INACTIVE`
- `prb user archive` renamed to `prb user deactivate`
- `prb audit-log export` and `prb scim event export` now emit CSV instead of JSONL

## [0.206.0] - 2026-07-29

### Added

- `prb audit-log export` and `prb scim event export` commands to request CSV exports of audit log entries and SCIM events

## [0.205.0] - 2026-07-28

### Added

- `prb audit create`/`update` gained `--audit-start-date` and `--audit-end-date` flags, and `prb audit list`/`view` show the new AUDIT START/AUDIT END columns
- `prb audit list --sort` now accepts `AUDIT_START_DATE` and `AUDIT_END_DATE`

## [0.204.0] - 2026-07-27

### Added

- `prb user list` gained `--filter` (search by name/email), `--role`, and `--kind` flags, with the default page size raised to 100

## [0.203.0] - 2026-07-22

### Changed

- `prb compliance-portal update` renamed the `--title` flag to `--entity-name`, matching the compliance portal's short entity-name field

## [0.202.0] - 2026-07-21

### Added

- `prb compliance-portal update` gained `--title`, `--description`, `--website-url`, `--email`, and `--headquarter-address` flags to manage the compliance page profile
- `prb webhook` event choices now include `RIGHT_REQUEST_CREATED`, `RIGHT_REQUEST_UPDATED`, and `RIGHT_REQUEST_DELETED`

### Changed

- Renamed the `prb trust-center` command tree to `prb compliance-portal` to match the product and GraphQL rename

## [0.201.0] - 2026-07-20

### Added

- `prb trust-center commitment` and `prb trust-center commitment-group` commands to create, list, update, and delete compliance portal commitments and their groups

## [0.200.0] - 2026-07-20

### Changed

- Renamed the risk overview's "Inherent" label to "Initial" in command help text, for consistency with the console

## [0.199.0] - 2026-07-02

### Added

- `prb webhook` event choices now cover the full document lifecycle: `document.created`/`updated`/`archived`/`unarchived`/`deleted`, the `document.version.*` events, and their `signature.*` and `approval.*` sub-events

### Changed

- `prb auth login` defaults and region prompts now use `eu.probo.com` and `us.probo.com` instead of `*.console.getprobo.com`

### Removed

- `MEETING_*` webhook event choices removed from `prb webhook` create/update; they were never valid backend event types and only produced API rejections

## [0.198.0] - 2026-06-30

### Removed

- Access review campaigns no longer expose a framework-controls field

## [0.197.1] - 2026-06-25

### Fixed

- `prb login` now requests the correct OAuth2 scopes (removed redundant `:read` suffix variants)

## [0.197.0] - 2026-06-22

### Added

- `prb resource-alias` commands to set and remove aliases on trust center entries

## [0.196.0] - 2026-06-19

### Added

- `prb login` now requests all v1 API scopes so device tokens work under OAuth2 scope enforcement

## [0.195.0] - 2026-06-16

### Changed

- `prb access-review` campaign sources are now first-class: each campaign captures a source snapshot (name, connector) at start time so a review stays coherent when the underlying source is edited or deleted, and fetch attempts are tracked as an append-only log instead of a single denormalized status field

## [0.194.0] - 2026-06-11

### Added

- `active` status field on access entries exposed in `prb access-entry` commands

### Changed

- References updated to probo.com

## [0.193.0] - 2026-06-10

### Added

- Expose `regulation_source` (`detected`/`default`) on `prb consent-record list`/`view` to show whether the regulation was resolved from geolocation or fell back to GDPR

### Changed

- `prb third-party list --first-level` replaced by `--level <N>` (1 = direct, 2+ = indirect) to support arbitrary nesting depth

### Removed

- `prb third-party link`/`unlink` commands; sub-third-parties are now scoped by a `parent_third_party_id` on the third-party itself

## [0.192.0] - 2026-06-09

### Added

- Add `prb risk-assessment boundary` command group (`create`, `list`, `view`, `update`, `delete`) and `--boundary-id` flag on `node create`/`update` to group risk assessment nodes within a scope
- Add `prb cookie-banner regenerate-policy` command to re-trigger tracker policy generation for a banner that already has a published version
- Expose `common_tracker_pattern_id` on `prb tracker-pattern list`/`view` to show whether a pattern is linked to the common tracker catalog

## [0.191.0] - 2026-06-02

### Changed

- Replace `prb third-party assess` with `prb third-party vet` to enqueue async third-party vetting; the command now returns immediately after queuing the job instead of waiting for the report

## [0.190.0] - 2026-05-27

### Added

- Add `prb user archive` command to deactivate a user profile while keeping them in the organization

## [0.189.0] - 2026-05-26

### Added

- Add `prb third-party link`/`unlink` commands for self-referential third-party relations
- Add `prb measure link-third-party`/`unlink-third-party` commands

### Changed

- Allow initial minor publishing of documents

## [0.188.0] - 2026-05-22

### Added

- Add `prb risk-assessment` command group with nested `scope`, `node`, `process`, `threat`, and `scenario` subcommands for managing the hierarchical risk assessment system, including scenario-to-risk and scenario-to-threat link/unlink and Mermaid chart retrieval

## [0.187.0] - 2026-05-15

### Changed

- Rename `prb vendor*` command group to `prb third-party*` (breaking)

## [0.186.0] - 2026-05-13

### Changed

- Drop `--consent-mode` flag from `prb cookie-banner create`/`update` and remove the `consent_mode` column from `cookie-banner` outputs — consent mode is now derived from the visitor's geolocation at consent time (breaking)

## [0.185.0] - 2026-05-12

### Changed

- Update kit package

## [0.184.0] - 2026-05-12

### Added

- Add `prb tracker-resource` command group (`list`, `view`, `create`, `update`, `delete`, `move`) for managing detected scripts, iframes, and other tracker resources

### Changed

- Replace `PREFIX` match type with `GLOB` in `prb tracker-pattern` interactive prompts (breaking)
- Drop `--display-name` from `prb tracker-pattern update` — display names are now derived from pattern + match type (breaking)

## [0.183.1] - 2026-05-08

### Security

- Upgrade go to 1.26.3

## [0.183.0] - 2026-05-07

### Added

- Add `regulation` and `country code` fields on cookie consent records, plus the `STATEMENT_OF_APPLICABILITY` document type on `prb document update`

### Fixed

- Allow editing metadata (title, document type, classification) on generated document versions

## [0.182.0] - 2026-05-06

### Added

- Add `--minor` flag to generated-document publish commands

### Changed

- Replace `prb document publish-major` and `publish-minor` with `prb document publish [--minor]`
- Rename `prb cookie-pattern` command group to `prb tracker-pattern`

## [0.173.0] - 2026-04-27

### Changed

- First per-package release. Prior history is in the archived monorepo [CHANGELOG.archive.md](../../CHANGELOG.archive.md).
