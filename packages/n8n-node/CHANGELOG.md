# Changelog

All notable changes to the `@probo/n8n-nodes-probo` package will be documented in this file.

## Unreleased

## [0.220.0] - 2026-08-14

### Added

- `Matrix Size` (`3×3`, `4×4`, `5×5`) option on Risk Analysis `Create` and `Update`, returned on `Get` / `Get Many`, so the likelihood/impact matrix size is explicit per analysis

### Fixed

- `Probo Trigger` no longer advertises itself as usable as an AI tool, which n8n could not honor for a trigger node

## [0.219.0] - 2026-08-13

### Added

- `AI System` resource (`Create`, `Get`, `Get Many`, `Update`, `Delete`, `Publish`), for managing an organization's AI Systems register
- `Period Start` / `Period End` fields on Risk Analysis `Create` and `Update`, and `period` on `Get` / `Get Many`
- `Compliance Portal Manager` and `Compliance Portal Access Manager` membership roles, for delegating full compliance portal management or just visitor access approval without broader admin access
- Cookie Banner `Update`: `Resource Reporting Enabled` option, to enable or disable resource detection on the banner

## [0.218.0] - 2026-08-10

### Changed

- Risk Analysis `Scope` parameters are now called `Diagram` across all operations.

## [0.217.0] - 2026-08-07

### Changed

- **Breaking**: the `Risk Assessment` resource is renamed to `Risk Analysis` (resource value `riskAssessment` -> `riskAnalysis`), along with its operations and parameters. Existing workflows using the Risk Assessment resource must reselect the Risk Analysis resource. Third-party assessments are unaffected.

### Added

- Document `Get Many`: `Published` filter option, to return only documents that have a published version

## [0.216.0] - 2026-08-06

### Added

- Access Review: `Read Campaign Access Entries` operation, to read the access entries for an access review campaign
- Compliance Portal `Update`: `Rights Requests Enabled` option, to enable or disable rights requests on the compliance portal

## [0.215.1] - 2026-08-06

### Fixed

- Organization Get Many: send `states: ['ACTIVE']` instead of the removed `state` field so the operation works again against servers that only accept the multi-value `ProfileFilter`

## [0.215.0] - 2026-08-05

### Added

- `subdivisionCode` (ISO 3166-2) returned by the Cookie Consent Record `Get` and `Get Many` operations, so workflows can branch on the detected state or province rather than country alone

## [0.214.0] - 2026-08-05

### Added

- Access Review Source `Get Many` operation to list organization access review sources

## [0.213.0] - 2026-08-05

### Added

- Business Function `Create`, `Get`, `Get Many`, `Update`, `Delete`, and `Publish` operations, for tracking DORA Critical ICT Functions

## [0.212.0] - 2026-08-04

### Added

- Cookie Consent Record `Get Many`: new `Acknowledge` action filter option, matching the new `ACKNOWLEDGE` consent action recorded for notice-only banners

## [0.211.0] - 2026-08-03

### Added

- Compliance Portal `Create`, `Delete`, `Get Many`, `Delete Audit`, `Delete Document`, `Delete Third Party`, `Update Audit Visibility`, `Update Document Visibility`, and `Update Third Party Published` operations

### Changed

- Third Party `Create`/`Update`: `Business Owner ID` and `Security Owner ID` fields replaced with a single `Administrator IDs` field (comma-separated)
- Compliance Portal `Create Commitment Group`: `Trust Center ID` field renamed to `Compliance Portal ID`

## [0.210.0] - 2026-07-31

### Added

- Device `Create`, `Get Many`, `Get`, `Revoke`, `Delete`, and `Set Owner` operations to manage ITAM devices
- Organization `Delete Horizontal Logo` operation

## [0.209.0] - 2026-07-30

### Changed

- User `Get Many` and Document `Get All Signatures` `State`/`Profile State` fields are now multi-select (`States`/`Profile States`), adding `Pending` alongside `Active` and `Deactivated`
- User `Archive` operation renamed to `Deactivate`

## [0.208.0] - 2026-07-28

### Added

- Audit `Create` and `Update` operations gained `Audit Start Date` and `Audit End Date` fields

## [0.207.0] - 2026-07-27

### Added

- User `Get Many` gained `Filter` (search by name/email), `Role`, and `Kind` fields, with the default page size raised to 100

### Changed

- Access Review `Create` and `Update` operations gained a multi-select `Sources` field to configure scope sources in the same call, replacing the separate `Add Source` operation

### Fixed

- Access Review source loader now paginates through all scoped sources instead of stopping at the first 500, so organizations with more sources can select them all when creating or updating a campaign

## [0.206.2] - 2026-07-24

### Fixed

- User Get no longer queries Connect `Organization.email`, which was removed and broke the operation; organization selection now matches User list/create/update

## [0.206.1] - 2026-07-22

### Fixed

- Organization Get/Get Many/Create no longer query removed Connect `Organization` profile fields (`description`, `websiteUrl`, `email`, `headquarterAddress`), which broke Organization Get Many. Use Compliance Portal Get/Update for those fields instead
- Compliance Portal Get now returns `entityName`, `description`, `websiteUrl`, `email`, and `headquarterAddress`, matching Compliance Portal Update, and selects `logo`/`darkLogo`/`nda` as File objects instead of removed URL scalars

## [0.206.0] - 2026-07-22

### Changed

- Compliance Portal `Update` operation renamed the `Title` field to `Entity Name`, matching the compliance portal's short entity-name field

## [0.205.0] - 2026-07-21

### Added

- Compliance Portal `Update` operation gained Title, Description, Website URL, Email, and Headquarter Address fields to manage the compliance page profile
- Webhook event choices now include `RIGHT_REQUEST_CREATED`, `RIGHT_REQUEST_UPDATED`, and `RIGHT_REQUEST_DELETED`

### Changed

- Renamed the `Trust Center` resource to `Compliance Portal`, and its custom link operations to `Create Custom Link`/`Delete Custom Link`, matching the renamed GraphQL API

### Removed

- `Organization` > `Update` no longer exposes Description, Website URL, Email, and Headquarter Address fields; use `Compliance Portal` > `Update` instead

## [0.204.0] - 2026-07-20

### Added

- Trust Center commitment and commitment group operations (create, get all, update, delete) so workflows can manage compliance portal commitments end to end

## [0.203.0] - 2026-07-20

### Changed

- Renamed the risk operations' "Inherent" field label to "Initial", for consistency with the console and CLI

## [0.202.1] - 2026-07-15

### Fixed

- Probo Trigger now logs webhook subscription delete failures instead of silently swallowing them
- Corrected the Probo Trigger codex node identifier and replaced unsupported marketplace categories to satisfy the n8n community-node review

## [0.202.0] - 2026-07-15

### Added

- Probo Trigger output for `*:updated` events now includes an `updatedFrom` object alongside `data`, holding the entity snapshot from before the update so workflows can diff old vs new values (for example `{{ $json.updatedFrom.membership.role }}` vs `{{ $json.data.membership.role }}`)

## [0.201.0] - 2026-07-09

### Added

- `document.update` can now edit the current draft version's body, title, classification, and document type, creating a draft from the latest published version when none exists (empty fields are skipped)

### Changed

- Renamed the document-signature filter `state` field to `profileState`, matching the API rename and disambiguating it from the signature `states` field
- Corrected the document operation help text to note the content/body field expects a ProseMirror document JSON string, not markdown

### Removed

- `document.createDraftVersion` operation — editing the body via `document.update` now covers draft creation

## [0.200.0] - 2026-07-02

### Added

- `ProboTrigger` node that starts a workflow when Probo emits a webhook event; it owns the subscription lifecycle (creates it on activation, re-registers it when the webhook URL drifts, deletes it on deactivation) and verifies the HMAC-SHA256 signature constant-time
- Document lifecycle webhook events — document created/updated/archived/unarchived/deleted, version created/updated/published/rejected/deleted, version signature requested/signed/cancelled, and version approval quorum requested/approved/rejected/updated/voided

### Changed

- Default Probo Server credential URL is now `https://us.probo.com` (was `https://us.console.getprobo.com`)
- Update the node's codex metadata to satisfy n8n marketplace review — use supported `Development`/`Utility` categories and the scoped `@probo/n8n-nodes-probo.probo` node identifier

### Removed

- `MEETING_*` webhook event choices from the webhook create/update operations — they were never part of the backend event enum

## [0.199.0] - 2026-06-30

### Added

- `document.getSignature` now returns the document version ID alongside the signature record

## [0.198.0] - 2026-06-30

### Added

- `document.getLatestPublishedVersionId` operation to retrieve the ID of the latest published version of a document

### Fixed

- `organization.getMany` failing when the organization has pending user invitations

## [0.197.0] - 2026-06-22

### Added

- `resourceAlias` resource with `setAlias` and `removeAlias` operations on the trust center node

## [0.196.0] - 2026-06-19

### Changed

- Internal toolchain maintenance (build/lint dependency bumps); no change to node behavior or operations

## [0.195.0] - 2026-06-19

### Removed

- `document.sendSigningNotifications` operation (replaced by an automatic debounced notification worker on the server)

## [0.194.0] - 2026-06-11

### Changed

- Connect org logo fields now return File download URLs
- `document publishVersion` now requires explicit `approverIds` when publishing a major version
- References updated to probo.com

## [0.193.0] - 2026-06-10

### Added

- Expose `regulationSource` (`DETECTED`/`DEFAULT`) on cookie consent record operations to indicate whether the regulation came from geolocation or fell back to GDPR
- `parentThirdPartyId` and `level` fields on third-party `create`/`getAll` operations for sub-third-party scoping under a parent

## [0.192.0] - 2026-06-09

### Added

- Add risk assessment `boundary` resource (`create`, `get`, `getAll`, `update`, `delete`) and `boundaryId` field on node create/update to group risk assessment nodes within a scope
- Add `cookieBanner regeneratePolicy` operation to re-trigger tracker policy generation for a banner that already has a published version
- Expose `commonTrackerPatternId` on tracker pattern `get`/`getAll` operations to indicate whether a pattern is linked to the common tracker catalog

## [0.191.0] - 2026-06-02

### Added

- Add `thirdParty vet` operation to enqueue async third-party vetting

## [0.190.0] - 2026-05-28

### Added

- Add Global region option to the vendor country picker

## [0.189.0] - 2026-05-27

### Added

- Add `user archiveUser` operation to deactivate a user profile while keeping them in the organization

### Changed

- Sort user operation options alphabetically

## [0.188.0] - 2026-05-26

### Added

- Add `measure linkThirdParty`/`unlinkThirdParty` operations
- Add `thirdParty linkThirdParty`/`unlinkThirdParty`/`listChildThirdParties` operations for self-referential relations

### Changed

- Allow initial minor publishing of documents

## [0.187.1] - 2026-05-25

### Fixed

- Fix signature count mismatch in `getAllSignatures` — add a `state` filter to `DocumentVersionSignatureFilter` so results match the console's signatures tab

## [0.187.0] - 2026-05-22

### Added

- Add a `riskAssessment` resource exposing the full risk assessment hierarchy — assessments, scopes, nodes, processes, threats, and scenarios — with CRUD operations, scenario-to-risk and scenario-to-threat link/unlink, and scope Mermaid chart retrieval

## [0.186.0] - 2026-05-15

### Changed

- Rename the `vendor` resource and its operations to `thirdParty` across all node actions (breaking)

## [0.185.0] - 2026-05-13

### Changed

- Drop the `consentMode` field from cookie banner create/update operations and remove `consent_mode` from cookie banner outputs — consent mode is now derived from the visitor's geolocation at consent time (breaking)

## [0.184.0] - 2026-05-12

### Changed

- Replace `PREFIX` with `GLOB` in tracker pattern match type options (breaking)
- Drop `displayName` from tracker pattern update operations — it is now derived from pattern + match type (breaking)

## [0.183.0] - 2026-05-07

### Added

- Add `regulation` and `countryCode` fields on cookie consent record operations

## [0.182.0] - 2026-05-06

### Changed

- Replace `publishMinor`, `publishMajor`, and `requestApproval` document operations with a unified `publish` accepting a `minor` flag and required `changelog` (breaking)
- Rename `cookiePattern` operations to `trackerPattern` with new `trackerType` field (breaking)

### Removed

- Remove legacy `cookiePattern` operations

## [0.0.1] - 2026-04-27

### Changed

- First per-package release. Prior history is in the archived monorepo [CHANGELOG.archive.md](../../CHANGELOG.archive.md).
