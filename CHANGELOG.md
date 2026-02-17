# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

## [0.127.1] - 2026-02-17

### Changed

- Remove deprecated SOA

## [0.127.0] - 2026-02-17

### Changed

- Change single document owner to multiple approvers

## [0.126.1] - 2026-02-16

### Added

- Add delete tasks to mcp

## [0.126.0] - 2026-02-16

### Fixed

- Fix deployment

## [0.125.0] - 2026-02-16

### Added

- Add delete risks to mcp
- Add meetings to mcp

## [0.124.3] - 2026-02-16

### Fixed

- Fix missing risk validations

## [0.124.2] - 2026-02-16

### Fixed

- Display control description
- Change download button text while loading in compliance page

## [0.124.1] - 2026-02-16

### Fixed

- Fix control order in SOA
- Fix saml subject not populated

## [0.124.0] - 2026-02-13

### Added

- Add webhooks

## [0.123.3] - 2026-02-13

### Fixed

- Fix missing name id format for idp initiated SAML request

## [0.123.2] - 2026-02-13

### Fixed

- Fix SAML subject must not be updated
- Fix SAML subject not set on first login

## [0.123.1] - 2026-02-13

### Fixed

- Fix missing NameID format information in SAML metadata

## [0.123.0] - 2026-02-12

### Changed

- Upgrade Postgres to 18.1
- IAM: Migrate people into profiles

## [0.122.0] - 2026-02-12

### Added

- Redirect to previous location on authentication or assumption needed

## [0.121.1] - 2026-02-11

### Fixed

- Fix Google Workspace SCIM bridge does not set active state at creation
- Fix compliance page access request was not active by default
- Fix compliance page request access add non request file to the requested one

### Security

- Update javascript dependencies

## [0.121.0] - 2026-02-10

### Added

- Add user exclusion management to the Google Workspace bridge

### Changed

- Improve compliance page access management UX

### Fixed

- Remove noisy error log from slack queue message


## [0.120.0] - 2026-02-09

### Added

- Suport all vendors fields on mcp

### Fixed

- Fix duplicate assessments

## [0.119.1] - 2026-02-09

### Security

- Upgrade go to 1.25.7

## [0.119.0] - 2026-02-05

### Added

- Add member n8n actions

### Fixed

- Fix controls for CFR framework
- Fix missing `trace_id` on resolver logs

## [0.118.2] - 2026-02-05

### Fixed

- Noisy TLS errors are filtered from logs
- Use s3 presigned URLs for email assets

### Changed

- Safer docker ubuntu image version with digest
- Safer github actions versions with digest
- Redirect already authenticated user on compliance page home when trying to log in

## [0.118.1] - 2026-02-04

### Fixed

- Fix column reference "full_name" is ambiguous

## [0.118.0] - 2026-02-04

### Added

- Add HDS framework
- Add 21 CFR Part 11 framework

### Changed

- Serve email static assets from object store
- Rework the UI of vendor row on compliance page
- Update auth layout on console and compliance page to remove right panel

### Fixed

- Fix create organization node

### Security

- Fix npm vulnerabilities

## [0.117.3] - 2026-02-03

### Fixed

- Fix n8n node cannot fetch many organizations.
- Fix SCIM disable all non SCIM members.

## [0.117.2] - 2026-02-02

### Fixed

- Missing logo
- Static handler cache headers handling

## [0.117.1] - 2026-02-02

### Changed

- Use svg for slack logo

### Fixed

- Missing google logo
- Missing relay generated files

## [0.117.0] - 2026-02-02

### Added

- Google Workspace to SCIM bridge
- Compliance page logo branding

### Fixed

- Memberships page conditional display of search input

### Security

- Upgrade go dependencies

## [0.116.15] - 2026-01-31

### Fixed

- Console slack connection placeholder display fix

## [0.116.14] - 2026-01-30

### Fixed

- Slack compliance page access display name empty case
- Compliance page API file access check fix + granular error handling

### Security

- Upgrade go to 1.25.6

## [0.116.13] - 2026-01-30

### Fixed

- Misplaced dependabot.yaml file is making CI fail

## [0.116.12] - 2026-01-29

### Changed

- Do not display trust center subprocessors tab when there are none
- Remove query params from compliange page sidebar website displayed URLs
- Refactor trust center pages to make them more maintainable
- Rename Trust Center to Compliance Page on displayed wording
- Compliance page vite dev server proxies graphQL API calls to go server port

### Fixed

- Add noreferrer noopener to compliance page open link from console

## [0.116.11] - 2026-01-27

### Added

- CI Test analytics with junit results format
- CI performance improvements with caching
- Extend dependabot to all dependencies

### Fixed

- AWS path style s3 option for Docker image entrypoint
- Console signatures counts
- Console signatures requests notifications

## [0.116.10] - 2026-01-23

### Added

- New up to date linting rules for TS codebase

### Changed

- Refactor SOA
- Dropped prettier
- Updated eslint related dependencies
- Ignore new TLS errors in logs

### Fixed

- otel utf8 errors

## [0.116.9] - 2026-01-21

### Fixed

- Document signing authentication is still done with token
- SOA permissions handling + tenant scoping

## [0.116.8] - 2026-01-20

### Changed

- Revert revert console graphql endpoint

## [0.116.7] - 2026-01-20

### Fixed

- n8n app calls by reverting console graphql endpoint

## [0.116.6] - 2026-01-20

### Fixed

- n8n http request options URL

## [0.116.5] - 2026-01-20

### Fixed

- Console invitationResolver.Organization authorize check

## [0.116.4] - 2026-01-19

### Changed

- Clean child sessions in IAM memberships migration

## [0.116.3] - 2026-01-19

### Fixed

- IAM memberships migration for entity ID

## [0.116.2] - 2026-01-19

### Changed

- Membership Profile authz done from membership in membershipResolver (it's a 1:1 association)
- Match keycloak URL ports with default base URL one for local dev
- Drop the authentication dialog in favor of a dedicated auth page for compliance page
- Order memberships by organization name on console / and memberships dropdown
- Update kit

### Fixed

- Missing console react-pdf dependency in package.json
- On sign out, clear cookie along with the existing session expiration
- Remove conditional rendering of org search input on console
- Fix organizations page layout vertical alignment

## [0.116.1] - 2026-01-18

### Fixed

- 5xx on profile loading.

## [0.116.0] - 2026-01-18

### BREAKING CHANGES

- API keys generated with previous versions are no longer compatible.

### Added

- Add SCIM provisioning support.
- Add magic link authentication.
- Add membership disable state.
- Add ABAC policy.

### Changed

- Filter junk HTTP TLS server errors.
- Change API token format.
- Add session support to compliance page.

### Fixed

- Fix overly strict obligation validation.

## [0.115.0] - 2026-01-15

### Added

- Add processing activity exports

## [0.114.0] - 2026-01-15

### Added

- Add proxy protocol v2 support.

## [0.113.0] - 2026-01-15

### Added

- Add right requests

## [0.112.4] - 2026-01-12

### Fixed

- Fix people in mcp
- Fix task display

## [0.112.3] - 2026-01-05

### Fixed

- Fix code blocks in documents
- Fix cancel signature permissions

## [0.112.2] - 2025-12-31

### Fixed

- Fix people deletion

## [0.112.1] - 2025-12-31

### Changed

- Change minutes max length

## [0.112.0] - 2025-12-22

### Added

- Add new GDPR registries

### Fixed

- Create Slack notification when updating trust center access if no existing message found

## [0.111.2] - 2025-12-19

### Fixed

- Fix trust center nil pointer dereference

## [0.111.1] - 2025-12-19

### Fixed

- Console: framework logo import

## [0.111.0] - 2025-12-18

### Added

- Add ISO 27701 (2025) framework.
- Add ISO 42001 (2023) framework.
- Add GDPR framework.
- Add CCPA framework.
- Add NIS2 framework.
- Add DORA framework.

## [0.110.2] - 2025-12-17

### Fixed

- Fix azure blob storage

## [0.110.1] - 2025-12-17

### Fixed

- Frameworks logo SVG colors
- Framework name displayed in trust center org sidebar

## [0.110.0] - 2025-12-17

### Added

- Add risk vendor risk assesments to n8n

## [0.109.0] - 2025-12-16

### Added

- Add organization filtering
- Use in-house logos when importing framework

### Fixed

- Unblock ACME provision queue on error
- GQLGen version handling with go tool
- Fix GraphQL types

## [0.108.0] - 2025-12-15

### Added

- Blacklist emails from trust requests

### Fixed

- ACME cert renewing

## [0.107.1] - 2025-12-15

### Fixed

- Fix infinit loop when renew ACME TLS certificate.

## [0.107.0] - 2025-12-15

### Added

- Add service to vendor on n8n

## [0.106.0] - 2025-12-12

### Added

- Add auditor role

### Fix

- Fix the “no change” error display in documents bulk update
- Update the front end after a document is published

## [0.105.0] - 2025-12-11

### Added

- Update trust center access slack message on console actions

## [0.104.0] - 2025-12-11

### Added

- Add risk management to N8N

## [0.103.0] - 2025-12-11

### Added

- Reject/Revoke trust center document accesses via slack app or console

## [0.102.0] - 2025-12-10

### Added

- New mime types for truct center files

## [0.101.1] - 2025-12-10

### Fix

- Fix missing validation on relation existence
- Fix permissions for trust center access

## [0.101.0] - 2025-12-10

### Added

- Enable svg support for company logos

## [0.100.0] - 2025-12-09

### Changed

- Allow non conformities without audit

### Fix

- Fix audit and framework deletion
- Make employee role assignable

## [0.99.0] - 2025-12-09

### Added

- Add employee page
- Add people management to N8N

### Fixed

- Fix invations never deleted when organization is deleted.
- Fix otel network error locally.

### Security

- Update kit.

## [0.98.1] - 2025-12-04

### Fixed

- Missing organization_id on Report
- ESLint issues

## [0.98.0] - 2025-12-03

### Added

- Add n8n vendor operations.

### Fixed

- Fix n8n node always returns success.

### Security

- Upgrade golang to 1.25.5

## [0.97.0] - 2025-12-02

### Added

- @probo/node-n8n-probo Meeting operations

### Fixed

- Console permissions initialisation
- Probod dev config values

## [0.96.1] - 2025-12-02

### Fixed

- Fix missing n8n placeholder.

## [0.96.0] - 2025-12-02

### Added

- Add n8n-node package.

## [0.95.0] - 2025-12-02

### Added

- New UI EditableTable component + implementation on assets page.

## [0.94.2] - 2025-11-27

### Fixed

- Fix missing compliance page permission again.

## [0.94.1] - 2025-11-27

### Fixed

- Fix missing compliance page permission.

## [0.94.0] - 2025-11-26

### Added

- Add updatePeople MCP tool.
- Add `SMTP_USER` and `SMTP_PASSWORD` to entrypoint.sh.

### Fixed

- MCP permission tools.

## [0.93.0] - 2025-11-25

### Added

- Add document MCP tools.
- Add document version MCP tools.
- Add document version signature MCP tools.
- Add MCP tools annotation hints.

## [0.92.0] - 2025-11-24

### Added

- Add snapshot MCP tools.
- Add task MCP tools.
- Add control MCP tools.
- Add control mapping MCP tools.

### Fixed

- Fix 5xx on vendor snapshot.

## [0.91.0] - 2025-11-24

### Added

- Add audit MCP tools.

### Fixed

- Fix 5xx when create vendor snapshot.
- Fix compliance page http to https redirect.
- Fix cannot create continious improvment via MCP.

## [0.90.1] - 2025-11-23

### Fixed

- Fix missing permission to delete custom domain.

## [0.90.0] - 2025-11-23

### Added

- MCP tools for many new objects.

### Fixed

- HTTP to HTTPS redirect for trust center.

## [0.89.1] - 2025-11-21

### Fixed

- MCP client always lost their session.

## [0.89.0] - 2025-11-21

### Added

- New MCP tools.

## [0.88.8] - 2025-11-20

- Fix SAML entrypoint config.
- Fix missing permission to verify SAML domain.

### Changed

- Update go dependencies

## [0.88.7] - 2025-11-19

### Fixed

- Fix snapshot creation

## [0.88.6] - 2025-11-18

### Fixed

- Fix asset permissions

## [0.88.5] - 2025-11-17

### Fixed

- Fix invitation permissions

## [0.88.4] - 2025-11-14

### Fixed

- Fix ca-cert-bundle entrypoint.sh

## [0.88.3] - 2025-11-14

### Fixed

- Fix document permissions

## [0.88.2] - 2025-11-14

### Fixed

- Fix missing healthcheck for postgres docker compose prod.
- Fix missing `AUTH_COOKIE_SECURE` support in entrypoint.sh.

## [0.88.1] - 2025-14-07

## Fixed

- Fix support PostgreSQL CA bundle in Helm charts with file path option

## [0.88.0] - 2025-11-13

### Added

- Add role management.
- Enable IdP-initiated SAML

### Changed

- API keys now have access to the organization they just created.

## [0.87.0] - 2025-11-13

### Added

- Add beta MCP server.
- Add official Kubernetes HEML chart.
- Add SAML IDP initiated flow support.
- Add meeting object.

### Changed

- SAML role is not mandatory anymore.

## [0.86.4] - 2025-11-07

### Fixed

- Test a fix of the deletion freeze

## [0.86.3] - 2025-11-07

### Fixed

- Fix document validations

## [0.86.2] - 2025-11-06

### Fixed

- Fix vendor url validations

## [0.86.1] - 2025-11-06

### Fixed

- Fix vendor validations

## [0.86.0] - 2025-11-06

### Added

- Add field validation system

## [0.85.0] - 2025-11-04

### Changed

- Make secure cookie configurable

## [0.84.0] - 2025-11-04

### Added

- Update documentation

### Fixed

- Fix download button in pdf preview

## [0.83.0] - 2025-10-31

### Added

- Add clearer error messages

### Fixed

- Fix organization creation
- Fix organization display order

## [0.82.0] - 2025-10-31

### Added

- Add SAML support
- Add vendors to processing activities

## [0.81.0] - 2025-10-29

### Added

- Add custom order ranking to trust center references

## [0.80.2] - 2025-10-28

### Fixed

- Fix create report access query

## [0.80.1] - 2025-10-28

### Fixed

- Fix report list in trust center update access modal

## [0.80.0] - 2025-10-28

### Changed

- Change trust center console UX/UI

## [0.79.0] - 2025-10-28

### Added

- Add trust center files

### Fixed

- Fix HTML entities displaying incorrectly in PDF exports (e.g., "&" showing as "&amp;")

## [0.78.0] - 2025-10-23

### Added

- Add slack integration for trust center access management

### Changed

- Markdown links now open in new tab with security attributes

## [0.77.0] - 2025-10-23

### Added

- Add invitation filtering by multiple states (PENDING/ACCEPTED/EXPIRED)

### Changed

- Settings page now shows only pending and expired invitations (accepted invitations are hidden)

### Fixed

- Fix document classification not being passed when creating documents
- Fix document classification changes not syncing to draft versions
- Fix organization invitations filter not being applied at database level

## [0.76.0] - 2025-10-22

### Added

- Add customizable document classification

## [0.75.1] - 2025-10-22

### Fixed

- Invitation status not updated on the UI
- Inconsistent updates of nullable values

## [0.75.0] - 2025-10-21

### Added

- Add signature filtering by state (REQUESTED/SIGNED) on document signatures tab
- Add HTTP to HTTPS redirect for custom domain 404 pages

### Changed

- Optimize document PDF export signature loading with single query instead of N+1 queries

## [0.74.7] - 2025-10-20

### Fixed

- Fix compliance website wording

## [0.74.6] - 2025-10-20

### Fixed

- Fix ordered list display in documents
- Fix 5xx on risk measures resolver

## [0.74.5] - 2025-10-16

### Fixed

- Broken document scroll when document list unfollded

## [0.74.4] - 2025-10-15

### Fixed

- React pdf race condition

## [0.74.3] - 2025-10-15

### Changed

- Increase signature link period from 7 days to 30 days

## [0.74.2] - 2025-10-15

### Fixed

- Not all signature are visible

## [0.74.1] - 2025-10-15

### Changed

- Use hosted png for logo in emails

## [0.74.0] - 2025-10-15

### Fixed

- Fix audit update in trust center settings

### Changed

- New signatures page design
- New emails design

## [0.73.1] - 2025-10-15

### Fixed

- Fix organization logo update

## [0.73.0] - 2025-10-14

### Added

- Add invitation management
- Bootstrap role management

### Changed

- Refactor of authentication and authorization

## [0.72.0] - 2025-10-14

### Added

- Add retry tracking for certificate provisioning and renewal.
- Add automatic cleanup of stale provisioning attempts (4+ hours old)
- Add max retry limit (3 attempts) before marking domains as failed
- Add distinction between fatal and transient ACME errors

### Changed

- Silently reject TLS connections without SNI (health checks, scanners)

### Fixed

- Fix stale certificate provisioning attempts blocking the queue

## [0.71.0] - 2025-10-14

### Fixed

- Fix SQL measure queries

### Changed

- Remove trust center slug config UI
- Add EU as possible contry code for vendor

## [0.70.0] - 2025-10-14

### Added

- Add horizontal logo to documents

### Fixed

- Fix file download Content-Disposition header format

## [0.69.0] - 2025-10-13

### Added

- Add document version on document list
- Add procedure document type
- Allow to filter measures by state

### Changed

- Support ID-based trust center URLs with slug fallback
- Show custom domain URL on trust center page when configured
- Update framework icons

### Removed

- Remove verifiedAt field from CustomDomain
- Remove criticity on assets

## [0.68.3] - 2025-10-11

### Fixed

- Fix trust center design on custom domain
- Send trust center invitation with custom domain when available
- Fix evidence deletion
- Fix dead ACME challenge

## [0.68.2] - 2025-10-10

### Added

- Socket binding for trust center

## [0.68.1] - 2025-10-09

### Fixed

- Fix filename content type regression
- Add missing permission to binary in docker image

## [0.68.0] - 2025-10-09

### Added

- Add custom domain to trust centers

## [0.67.2] - 2025-10-09

### Fixed

- Fix vendor compliance reports files migration

## [0.67.1] - 2025-10-09

### Fixed

- Fix evidence files migration

## [0.67.0] - 2025-10-09

## Changed

- Store all file data in one table

### Fixed

- Display more measures and tasks

# [0.66.1] - 2025-10-06

### Fixed

- Fix access to public documents for unauthenticated users

# [0.66.0] - 2025-10-06

### Added

- Store id of accepted nda
- Add public documents on trust centers

## Changed

- Allow missing NDA

### Fixed

- Fix trust center v2 design

# [0.65.1] - 2025-10-03

### Fixed

- Restrict deletion of users who have assets

# [0.65.0] - 2025-10-03

### Added

- Add access by document on trust center
- Add ordering measures by name in the API

### Fixed

- Fix resetting state during measure editing
- Handle document mapping conflict error

# [0.64.1] - 2025-09-30

### Fixed

- Remove document description and footer in template

# [0.64.0] - 2025-09-30

### Added

- Add trust center v2
- Add optional watermark and signatures in pdf export

### Changed

- Remove description from pdf export

# [0.63.1] - 2025-09-26

### Fixed

- Fix request document signature

# [0.63.0] - 2025-09-25

### Changed

- Allow `.csv` file as evidences.
- Remove section id from obligations
- Change obligations status enum

### Added

- Add link between obligations and risks

# [0.62.0] - 2025-09-24

### Added

- Add reference companies for trust center trusted by section

### Fixed

- Fix watermark display in trust center

# [0.61.1] - 2025-09-23

### Added

- Display countries on trust centers

# [0.61.0] - 2025-09-23

### Added

- Decouple users from people
- Add more details to organization

# [0.60.0] - 2025-09-18

### Added

- Add bulk export documents
- Add bulk delete documents
- Display risk description
- Uniformize date diplay
- Add risk order by owner full name

# [0.59.1] - 2025-09-16

### Fixed

- Build trust center in the make file

# [0.59.0] - 2025-09-16

### Added

- Add nda to trust centers
- Add confidential watermark to trust center documents
- Add countries to vendors

### Fixed

- Add category back to vendors

# [0.58.4] - 2025-09-12

### Fixed

- Fix framework export email

## [0.58.3] - 2025-09-12

### Fixed

- Fix release workflow

## [0.58.2] - 2025-09-12

### Added

- Send framework export by email
- Order organization by name

## [0.58.1] - 2025-09-10

### Chore

- Rename registries

## [0.58.0] - 2025-09-08

### Added

- Soft delete documents
- Store sidebar state

## [0.57.1] - 2025-09-04

### Fixed

- Fix framework export for evidence link

## [0.57.0] - 2025-09-04

### Added

- Add framework exports

## [0.56.0] - 2025-09-03

### Added

- Add risks snapshots

### Fixed

- Fix tabs counter
- Fix password page redirection

## [0.55.0] - 2025-09-01

### Added

- Add vendor snapshots

### Fixed

- Fix document deletion and update errors
- Remove signature block from trust center documents
- Fix non mandatory fields on vendor

## [0.54.0] - 2025-09-01

### Added

- Add compliance registry snapshots
- Add continual improvement snapshots
- Add processing activity registry snapshots
- Add assets snapshots

## [0.53.0] - 2025-08-29

### Added

- Add processing activity registries
- Add continual improvement registries
- Add noncoformity registry snapshots

### Fixed

- Probo instance allow crawling bot to index.

## [0.52.0] - 2025-08-27

### Added

- Add data snapshot

## [0.51.1] - 2025-08-23

### Fixed

- Fix query loops in public trust center
- Fix button display when disconected in public trust center

## [0.51.0] - 2025-08-22

### Added

- Add trust center access requests

## [0.50.1] - 2025-08-21

### Fixed

- Fix authentification token error

## [0.50.0] - 2025-08-21

### Added

- Add compliance registries
- Add vendor services

### Chore

- Replace mailhog

## [0.49.0] - 2025-08-20

### Added

- Add nonconformity registries

### Fixed

- Fix trust center dark mode

## [0.48.1] - 2025-08-14

### Fixed

- Fix display of download buttons in the public trust center

## [0.48.0] - 2025-08-14

### Added

- Add baa to vendors
- Add dpa to vendors
- Add contacts to vendors
- Add name to audits
- Add audits to controls

## [0.47.0] - 2025-08-13

### Added

- Add document draft deletion
- People now have contract start and end dates in the UI and API.
- Lists can filter out people whose contracts have ended.

### Fixed

- Fix closing of document deletion pop up
- Fix creation of empty draft without save

## [0.46.2] - 2025-08-10

### Fixed

- Fix various SQL queries failures due to trust center
- Fix internal information leaking to API

## [0.46.1] - 2025-08-10

### Fixed

- 5xx on risk show page

## [0.46.0] - 2025-08-08

### Added

- Add organization deletion
- Add Probo by default

## [0.45.1] - 2025-08-06

### Fixed

- Fix data page display

## [0.45.0] - 2025-08-06

### Added

- Add trust center
- Add edition of document fields

## [0.44.0] - 2025-07-23

### Fixed

- Fix PDF tables
- Fix display issue on control and framework
- Fix control creation

## [0.43.1] - 2025-07-22

### Fixed

- Fix document draft creation

## [0.43.0] - 2025-07-21

### Added

- Add control exclusion

### Fixed

- Fix small issues on SOA

## [0.42.1] - 2025-07-16

### Fixed

- Fix missing document download button

## [0.42.0] - 2025-07-16

### Fixed

- Fix document version selector
- Fix duplicate people

## [0.41.0] - 2025-07-16

### Changed

- Revision of multiple UI elements

### Added

- Add document version selector on details page
- Add document bulk publication
- Add document bulk signature request
- Add cancel signature request

## [0.40.0] - 2025-07-11

### Added

- Add cancel request mutation
- Add bulk publish document version mutation
- Add bulk request signature mutation

## [0.39.0] - 2025-06-03

### Added

- Add policy PDF export

### Security

- Update go dependencies
- Update node dependencies

## [0.38.1] - 2025-07-03

### Fixed

- Fix 5xx on document type order

## [0.38.0] - 2025-07-03

### Added

- Allow to change doucment order in the UI

### Change

- Change default document sorting order

## [0.37.5] - 2025-06-30

### Fixed

- Fix missing risk score on detail risk page
- Fix matrix risk score color on risk matrix

## [0.37.4] - 2025-06-30

### Fixed

- Fix SOA with risk

## [0.37.3] - 2025-06-30

### Fixed

- Fix missing framework controls

## [0.37.2] - 2025-06-30

### Changed

- Generate excel in memory instead of using fs

## [0.37.1] - 2025-06-30

### Added

- Add updated at and created at order for vendor

### Fixed

- Fix SOA filename

## [0.37.0] - 2025-06-30

### Added

- Add SOA generator
- Show last assessment date

## [0.36.0] - 2025-06-29

### Added

- Add URI evidence type
- Add link dialog for measure evidences
- Add default security header to API
- Add support for extra header

### Fixed

- Fix tasks deadline
- Fix order people by kind
- Fix missing people role order

### Security

- Remove all data after logout
- Enforce maximum password limit
- Mitigate timing attack on signin

### Changed

- Use httplogger on GraphQL error
- Returns internal error when error is known

## [0.35.0] - 2025-06-20

### Added

- Add forgot password pages

## [0.34.0] - 2025-06-17

### Added

- Pagination for people, vendors, documents, data and assets

### Fixed

- Fix 404 on email confirmation page
- Fix 404 on invitation confirmation page
- Fix login redirection
- Fix form not reset after submit

## [0.33.6] - 2025-06-15

### Fixed

- Fix filedrop upload too small file size

## [0.33.5] - 2025-06-14

### Fixed

- Fix framework view too many queries
- Fix image upload failed

## [0.33.4] - 2025-06-13

- Fix measure count

## [0.33.3] - 2025-06-13

### Fixed

- Fix 5xx on document count for risk
- Fix leaking pg connections

## [0.33.2] - 2025-06-13

### Fixed

- Fix API path contain undefined

## [0.33.1] - 2025-06-13

### Fixed

- Fix localhost enforce at build time

## [0.33.0] - 2025-06-13

### Added

- Add backend fulltext search on controls, documents, risks and measures
- Add `totalCount` field in `Connection` object
- New console design

### Fixed

- Use new enum for data classification

## [0.32.0] - 2025-06-07

### Changed

- Prevent publishing of document versions with no changes
- Update AI prompt used for changelog generation

### Added

- Add deadline on tasks
- Add controls manual create, update, and delete

## [0.31.0] - 2025-06-04

### Added

- Add assets inventory
- Add data inventory
- Add title and owner id to document versions
- Add automatic changelog

## [0.30.1] - 2025-06-02

### Added

- Added sort key `updated_at` for vendors

### Fixed

- Fix 5xx on signature request
- Fix sort key 5xx

## [0.30.0] - 2025-05-31

### Changed

- Change policies to documents

### Added

- Add type to documents

### Fixed

- Fix HIPAA import

## [0.29.0] - 2025-05-29

### Added

- HIPAA releated risks
- Add framework import from json

### Changed

- New add vendor UI

### Fixed

- Add url input type for vendor assessement

## [0.28.0] - 2025-05-29

### Added

- Add automatic vendor assessment
- Add vendor category fields
- Add vendor business associate agreement url and subprocessors list url

### Fixed

- Fix 5xx when create new vendor
- Fix 5xx when update a vendor

## [0.27.1] - 2025-05-25

### Fixed

- Fix missing `position` field migration

## [0.27.0] - 2025-05-25

### Changed

- Rename severity into risk score
- Add contract dates fields
- Add people position field

### Fixed

- Fix conflict http header fields

### [0.26.0] - 2025-05-20

### Added

- Add docker image security scan
- Show latest risk updated date
- Log error when GraphQL resolver failed

### Security

- Update Golang dependencies
- Update to latest Ubuntu LTS
- Update to latest Golang version

### [0.25.0] - 2025-05-13

### Changed

- Allow data and text file for evidences

### Fixed

- Fix missing people when inviting user already in other organization
- Fix cannot upload organization logo

### [0.24.0] - 2025-05-06

### Added

- Task page list

### Changed

- Task is now linked to organization

### Fixed

- Fix cannot see vendor assessment note

### Security

- Add filetype validation for end-user upload

## [0.23.1] - 2025-05-04

### Fixed

- Fix http cache etag
- Fix cannot delete measure

## [0.23.0] - 2025-05-04

### Added

- Add ISO 27001 document header
- Add policy downlaod

### Changed

- Enable HTML support in Markdown renderer

## [0.22.0] - 2025-05-04

### Added

- Show owner of the policy in list
- Show number of singatures in the policy list

### Changed

- Link evidences to measure

## [0.21.0] - 2025-05-01

### Changed

- Add markdown table support
- Explicit risk score calcul

## [0.20.1] - 2025-05-01

### Changed

- New vendors in the built-in lists

## [0.20.0] - 2025-05-01

### Added

- Add end-user confirmation before sending policy sign notification
- Add assessed at in the vendor list

### Fixed

- Fix not aligned button on policy list view

### Removed

- Remove start and end service date of vendor

## [0.19.2] - 2025-05-01

### Fixed

- Fix 5xx when invite user in an organization

## [0.19.1] - 2025-05-01

### Fixed

- Evidence URL not set

## [0.19.0] - 2025-04-30

### Changed

- New vendors in the built-in lists

### Security

- Update javascript dependencies
- Fix open redirect when the redirect url use `//`

## [0.18.1] - 2025-04-30

### Fixed

Fix typo `mesure` instead of `measure`

## [0.18.0] - 2025-04-30

### Added

- Static files are served using GZip
- Static fiels are served with ETag and Cache header fields

### Fixed

- Entrypoint JS/CSS has no chunk hash

## [0.17.0] = 2025-04-29

### Added

- Add policy unlogged sign

## [0.16.0] = 2025-04-29

### Added

- Policy history
- Policy signature
- New vendors in the built-in lists

### Fixed

- Fix cannot delete measure with linked risk

## [0.15.1] - 2025-04-27

### Fixed

- Fix SQL syntax error

## [0.15.0] - 2025-04-27

### Added

- Add delete measure in the UI and GraphQL API

### Removed

- Remove `importance` field from measure as it's not used anymore

### Fixed

- Fix delete evidence from task list does not work
- Fix cannot load attached measure risks

## [0.14.0] - 2025-04-24

### Added

- Risk can have note
- Cache static assets

## [0.13.2] - 2025-04-24

### Fixed

- Fix psql `generated_gid` returns padded base64
- Fix `user_id` not set when create new organization
- Fix `additional_email_addresses` not set when invite in organization

## [0.13.0] - 2025-04-23

### Added

- New "Risk assessments" tab for vendors that allows you to:
    - View all risk assessments for a vendor in one place
    - Create new risk assessments with data sensitivity and business impact ratings
    - Track assessment expiration dates
- Automatic people record creation when accepting invitations
- New vendors in the built-in lists
- Introduced a connector framework enabling integration with external
  services:
    - Add OAuth2 connector implementation

### Changed

- Completely redesigned vendor list page
- Completely redesigned vendor detail page
- Improved compliance reports table with better file size formatting and date display
- People may be linked to user

## [0.12.0] - 2025-04-20

### Added

- New vendors in the built-in lists

### Changed

- Update risk library with new risks

### Security

- Upgrade Golang dependencies
- Upgrade Node dependencies

## [0.11.1] - 2025-04-15

### Changed

- New vendors in the built-in vendors list

## [0.11.0] - 2025-04-15

### Added

- New vendors in the built-in vendors list

### Changed

- More explicit scale, legend and score for risk matrix

## [0.10.1] - 2025-04-14

### Fixed

- Fix grammar for "people" in the navigation bar
- Fix editor change cursor position at each keystroke
- Fix editor does not display list icon

## [0.10.0] - 2025-04-14

### Changed

- Improve UI of the risk matrix
- Rename "Mitigation" in "Mesure"

## [0.9.0] - 2025-04-12

### Added

- Added business owner and security owner fields to vendors

### Changed

- Improved vendor detail page with organized sections
    - Split information into logical sections (Basic Information, Ownership, Risk & Service, Documentation)
    - Better visual organization of vendor information

## [0.8.0] - 2025-04-12

### Added

- New risk treatment strategy options: Mitigate, Accept, Avoid, Transfer
- Risk ownership functionality

## [0.7.0] - 2025-04-12

### Added

- Enhanced risk management with inherent and residual risk assessment capabilities
    - Added new fields to track both inherent and residual likelihood/impact values
    - Introduced risk severity calculation as the product of likelihood and impact
    - Added visual risk matrix to view risk distribution by severity
- New risk-policy mapping functionality allowing risks to be linked to policies
- New risk-control mapping functionality enabling risks to be linked to controls
- Added edit functionality for risks with a new edit page
- New popover components for mitigation information on the mitigations list view
- Pre-populated risk templates from a JSON data source

### Changed

- Update vendors catalog.
- Updated risk creation form to include both inherent and residual risk parameters
- Improved risk list view with risk matrix visualization
- Enhanced breadcrumb navigation for risk detail pages
- Refactored risk-mitigation mapping to remove redundant probability/impact fields
- Renamed probability field to likelihood for better alignment with risk management terminology

### Fixed

- Improved license file formatting in vendors and risks data directories
- Fixed URL in attribution text (`getprobo.com` → `www.getprobo.com`)

## [0.6.0] - 2025-04-10

### Added

- Added vendors.json data file under Creative Commons Attribution-ShareAlike 4.0 license`
- New vendor data management system with comprehensive vendor information
- Pre-populated vendor database with 12 common SaaS vendors and their certifications
- Vendor details page with extended fields for improved vendor management:
    - Legal name and headquarters address
    - Website URL
    - Certification tracking with tag-based interface
    - Links to important vendor documents (SLA, DPA, security pages)
    - Support for multiple compliance certifications per vendor

### Fixed

- Fix cannot create vendor when the name is too similar to suggested one
- Fix UI showing double button to close evidence preview modal
- Fix cannot delete vendor with compliance reports (added cascade delete constraint)

## [0.5.0] - 2025-04-10

### Added

- Add vendor compliance reports UI
- Controls can now be linked to policies, enabling better organization of compliance documentation and clearer traceability between policies and security controls
- New UI for viewing and managing policies related to a specific control

## [0.4.2] - 2025-04-09

### Changed

- Simplified policy data model by removing version field and optimistic concurrency
- Refactored policy update flow to load-modify-save pattern

### Fixed

- Added user-friendly error messages when importing frameworks that already exist

## [0.4.1] - 2025-04-09

### Changed

- Update ISO 27001 and SOC2 framework definition.

## [0.4.0] - 2025-04-09

### BREAKING CHANGES

- **BREAKING:** Renamed GraphQL mutations for control-mitigation mappings:
    - `createControlMapping` → `createControlMitigationMapping`
    - `deleteControlMapping` → `deleteControlMitigationMapping`
    - Input and payload types have been updated accordingly

### Added

- Add import control <> mitigation mapping.
- Add mitigation tasks import.
- Add auto-scroll to opened category.
- Added support for mapping controls to policies:
    - New GraphQL mutations `createControlPolicyMapping` and `deleteControlPolicyMapping`
    - Controls can now be associated with both mitigations and policies
    - New bidirectional relationships:
        - Control objects now expose a `policies` field to list associated policies
        - Policy objects now expose a `controls` field to list associated controls
- Added vendor compliance reports:
    - New GraphQL types `VendorComplianceReport` and related connection types
    - New GraphQL mutations `uploadVendorComplianceReport` and `deleteVendorComplianceReport`
    - New `complianceReports` field on the Vendor type
    - Support for uploading, viewing, and managing vendor compliance documentation
- Added pre-configured frameworks:
    - Added ISO/IEC 27001:2022 and SOC 2 framework templates
    - Improved framework import interface with dropdown menu for template selection
    - Support for one-click import of standard compliance frameworks

### Changed

- Evidence can now be requested.

### Fixed

- Fix unfoldable mitigation category when open via the URI fragment.
- Fix ctrl+click on mitigation does not open new tab.
- Fix error handling in framework view when no controls are available.

## [0.3.0] - 2025-04-01

### Added

- Add sidebar to show a task.
- Add task estimate edition.
- Add control+framework auditor views.
- Add import mitigations support.
- Add import framework support.
- Add risk object management.
- Add risk template.
- Add mapping between control and risk.

### Changed

- Rename control in mitigation.
- Home page is now mitigations page.

### Fixed

- Fix panic in GraphQL resolver are not reported.
- Fix otal trace never started.
- Fix React.lazy chunck error.
- Fix login page show `unauthorized` error.
- Fix cannot delete task with evidences.
- Fix cannot download file with non-ASCII filename.

## [0.2.0] - 2025-03-24

### Added

- Add forget password.
- Allow evidence to be a link.
- Add task import support.
- Allow to create vendor when it not exist in the auto-complete.
- Add service account people kind.

### Changed

- Make task time estimate optional.
- Set invitation token to 12 hours.
- Order people by fullname.
- Order vendor by name.
- Allow to edit control state without going to edit page.
- Redirect on people list after people creation.
- New UI for the framework overview page.

### Fixed

- Fix flickering on hover on categories.
- Fix control order under a category.
- Fix UI does not refresh after importing a framework.
- Fix cannot create control.
- Fix missing include cookie on confirmation invit.
- Fix sign-in does not include cookie.
- Fix missing version when create task.
- Fix random order on framework overview.
- Fix change task state not visible on UI.
- Fix control card items alignement.
- Fix cannot delete task.
- Fix password managers misidentifying token fields as usernames in reset password and invitation confirmation forms.

## [0.1.0] - 2025-03-14

Initial release.
