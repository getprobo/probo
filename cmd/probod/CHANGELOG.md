# Changelog

All notable changes to `probod` (the server, including the bundled `@probo/console`, `@probo/trust`, and `@probo/ui` frontends) will be documented in this file.

## Unreleased

## [0.204.0] - 2026-06-05

### Added

- Dedicated error page when a magic link has already been used

### Changed

- Improved error page layout and messaging

## [0.203.0] - 2026-06-05

### Added

- Zendesk access-review connector with subdomain URL normalization
- Okta access-review connector with API-key (SSWS) authentication
- Clerk access-review connector
- SendGrid access-review connector with 2FA enforcement checks
- Datadog access-review connector with region selector and OAuth support
- PostHog access-review connector with Cloud OAuth, self-hosted OAuth, and API-key support
- Public-client (CIMD) OAuth support with auto-registration and client metadata document
- `SMTP_HELLO_NAME` environment variable to configure the EHLO/HELO hostname
- Dedicated expired magic link error page
- Audit reports are now stored as files

### Changed

- Clarify trust center access rejection emails
- Cookie banner now supports Indonesian, Italian, Japanese, Korean, Polish, Portuguese, Turkish, Ukrainian, and Chinese

### Fixed

- Fix login redirect for password-only authentication flows

## [0.202.2] - 2026-06-03

No user-facing changes; tag-only release.

## [0.202.1] - 2026-06-03

No user-facing changes; tag-only release.

## [0.202.0] - 2026-06-03

### Added

- Trigger tracker-policy document generation on banner publish; a background worker regenerates it on every snapshot
- Show tracker type in the cookie tracking policy document
- Include the website origin in the tracker policy title

### Changed

- Restrict queries and mutations to session scope
- Move the Display tab first on the cookie banner configuration page
- Link to the generated cookie policy document from tracker rows; revamp tracker row layout
- Number tracker policy section titles

### Fixed

- Use stable API URLs for vendor logo fields

## [0.201.0] - 2026-06-02

### Added

- Add async third-party vetting worker with PENDING/PROCESSING/COMPLETED/FAILED states, exposed through GraphQL and MCP; the third-party detail page polls while vetting runs
- Tune the third-party vetting worker (interval, concurrency, stale-after, agent timeout, max-turns) via config

### Changed

- Downgrade access-source instance name resolution failures from error to warning

### Fixed

- Guard the GitHub access-source name resolver against empty organization to stop the source-name worker from flooding logs with 404s

## [0.200.1] - 2026-06-01

### Fixed

- Raise tracker mapping and common-pattern enrichment agent max turns to 10 to prevent `MaxTurnsExceededError` when the tool-call budget exceeded the limit

## [0.200.0] - 2026-06-01

### Added

- Add tracker description enrichment worker
- Promote tracker patterns to organization third parties via worker, with first-party origin filtering and sibling-based mapping
- Surface third-party links on `TrackerPattern` in GraphQL, with batch loaders
- Filter banner trackers by linked third party and show third parties on the banner trackers page
- Expose HTTP cookie source through the console API
- Add document archive row action
- Add stale recovery to the tracker mapping worker
- Tune tracker workers: expose worker interval, concurrency, stale-after, agent timeout, and max-turns as config

### Changed

- Deactivate SCIM users when delete is blocked
- Rework tracker and resource row actions
- Reuse the mapping agent to attribute trackers in the enricher
- Raise default agent token budget for reasoning models (1024/512 → 4096)
- Harden catalog vendor resolution and the tracker mapping agent prompt
- Skip shared infrastructure in domain matching during tracker mapping
- Backfill tracker description from the common catalog
- Run tracker mapping outside the persist transaction to remove cross-network row locks

### Fixed

- Stop tracker agents from inventing vendors
- Drop sampling params unsupported by the model
- Tolerate source fetch failures during tracker mapping
- Skip mapping when a tracker pattern is deleted concurrently
- Guard `LinkToCommon` against overwriting an existing catalog link
- Take resolver scope from `Authorize` rather than the GID
- Copy default LLM pointers when resolving agents

## [0.199.1] - 2026-05-28

### Fixed

- Fix missing icons in the UI
- Fix Metabase user listing in access reviews
- Fix PostHog resolver name

## [0.199.0] - 2026-05-28

### Added

- Add PostHog access-review connector
- Add Metabase access-review connector
- Add Grafana access-review connector
- Add Cursor access-review connector
- Support HTTP Basic auth in API-key connections
- Cancel pending signature requests when a contract ends or a connector is deactivated

### Changed

- Reject demotion of the last owner of an organization
- Scope document signatures to the major version

### Fixed

- Fix Microsoft 365 access review returning too many accounts

## [0.198.0] - 2026-05-28

### Added

- Add Tailscale connector
- Add Anthropic connector (authenticated via API key)
- Add personal account support for the Heroku connector
- Add Global region option to the vendor country picker
- Allow ordering organization members by email address

### Changed

- Connector deletion is now best-effort: remaining steps proceed even when one cleanup step fails

### Fixed

- Fix role column in the people list rendered as non-sortable to prevent runtime failures
- Surface an actionable error when a stored Sentry organization slug is no longer accessible to the connected OAuth token
- Stop the source-name worker from retrying indefinitely on a stale Sentry organization slug
- Stop the source-name worker from retrying indefinitely on a stale Heroku personal-account slug

## [0.197.0] - 2026-05-28

### Added

- Add `invitingOrganizations` field on the viewer to expose organizations that have sent a pending invitation to the current user

### Fixed

- Show SCIM error message in the connector UI

## [0.196.1] - 2026-05-27

### Fixed

- Fix serialization of SCIM bridge `SYNCING` and `DISABLED` states in the GraphQL API

## [0.196.0] - 2026-05-27

### Added

- Expose bridge sync errors in the SCIM API and on Google Workspace and Microsoft 365 connector cards
- Expose profile source field on users in the MCP API

## [0.195.0] - 2026-05-27

### Added

- Add `archiveUser` operation to deactivate a user profile while keeping them in the organization; exposed across the console UI, MCP, CLI, and n8n
- Expire pending invitations for a user when they are archived
- Grant owners full `iam:scim-bridge:*` and admins read-only SCIM bridge access in IAM policies

### Fixed

- Preserve archived and deactivated HubSpot users in access reviews instead of dropping them
- Fix common third-party logo URL returning resource-not-found in the combo box query

## [0.194.0] - 2026-05-26

### Added

- Add `probo-agent` CLI and device agent library for endpoint compliance checks
- Add screen lock detection support for i3, KDE, and more Linux desktop environments

### Fixed

- Skip unconnectable providers in provider listing
- Reject shell-unsafe paths in FreeBSD rc.d service installer
- Make Windows service uninstall idempotent
- Use platform-specific atomic key replacement on Windows
- Handle FreeBSD check command failures before reading status

## [0.193.1] - 2026-05-26

### Security

- Fix open redirect bypass in safe redirect

## [0.193.0] - 2026-05-26

### Added

- Add measure ↔ third-party many-to-many link with tabs on both detail pages
- Add self-referential third-party relations with a `first_level` filter on the third-party list
- Track source on detected storage trackers (localStorage, sessionStorage, indexedDB, cacheStorage)
- Promote tracker pattern source on detection and trigger a draft banner version when adopting uncategorised patterns

### Changed

- Allow initial minor publishing of documents
- Mark page-world extension writes (MV3 main world, userscripts with `@grant none`) with the new `EXTENSION` cookie source
- Surface the measure state as a header badge and remove the measure detail right-hand drawer

### Fixed

- Fix timing attack on signin
- Reject separator-only glob templates (e.g. `__*`) in tracker pattern analysis

## [0.192.0] - 2026-05-25

### Changed

- Enforce IAM authorization on all console resolvers — every data-bearing field now goes through the policy engine and produces an audit log entry; adds `ActionCommonThirdPartyGet`, `ActionCommonThirdPartyList`, and `ActionElectronicSignatureGet` actions wired into Viewer and Auditor policies

### Fixed

- Fix signature count mismatch between the document version badge and the signatures tab — both now filter by `activeContract: true` and `state: ACTIVE`, so deactivated signers and ended-contract signers are consistently excluded
- Fix MCP server resolvers after the signature filter and authorization changes

## [0.191.0] - 2026-05-22

### Added

- Add a tracker pattern detail page in the console with a properties section and a list of detected tracker resources

### Fixed

- Strip empty ProseMirror text nodes from third-party list documents (and migrate existing `document_versions.content` to drop them) so Tiptap renders them instead of erroring with "Empty text nodes are not allowed"
- Tailor signature certificate email copy for document approvals — store the per-signature email subject on creation so the certificate worker uses "Your approved <Title> - Certificate of Completion" for approvals and the existing default for other flows
- Return a CONFLICT error instead of an opaque Internal error when deleting a membership profile that is still referenced as owner, approver, or assignee, by detecting the Postgres foreign-key violation in the coredata Delete path
- Always instantiate the coredata `CookieCategoryFilter` in the cookie banner queries to avoid nil-pointer risks

## [0.190.1] - 2026-05-20

### Fixed

- Fix the snapshot-cleanup migration to delete from `processing_activity_third_parties` (the table was renamed from `processing_activity_vendors` in 0.189.0), so the migration runs on databases upgraded past the rename

## [0.190.0] - 2026-05-20

### Added

- Add a hierarchical risk assessment system with Risk Assessment, Scope, Node (ENTITY / BOUNDARY / ASSET / DATA), Process, Threat, and Risk Scenario entities, and render a Mermaid data-flow diagram per scope (nodes typed by shape, threats attached as dashed edges)
- Add 13 access-review connector providers (with PKCE, token-body extras, and `AuthURL` templating support in the OAuth2 driver), and wire them through the review engine, the name worker, and the Helm chart
- Add a tracker mapping worker that resolves detected trackers to third parties using initiator domain extraction (eTLD+1), pattern-glob analysis, and a Firecrawl-backed LLM agent fallback for unmapped patterns
- Add a shared `common_third_parties` / `common_third_party_domains` catalog with slug-based deduplication, allow a single domain to be associated with multiple third parties, and auto-create entries from OCD imports
- Introduce the `proboctl` CLI (replaces the standalone `common-third-parties-import` and `common-tracker-patterns-import` commands as `proboctl seed ...`), with `data.json` embedded in the binary

### Changed

- Move the Firecrawl API key from the top-level config into `Agents.Tools`, hardcode the Firecrawl API endpoint (drop `FIRECRAWL_ENDPOINT`), and replace the SearXNG search backend with Firecrawl
- Split cookie names on both `_` and `-` separators so cookies like `__Secure-1PSID` no longer collapse into a bogus `___*` heuristic pattern

### Fixed

- Filter SCIM-deactivated (INACTIVE) people from signature request recipient lists in both the multi-select dialog and the document signatures page

### Removed

- Remove the deprecated snapshot system (the register/document model fully replaces it)
- Remove backend inactive-profile validation that incorrectly rejected newly-created users on first login

## [0.189.0] - 2026-05-15

### Changed

- Rename `vendor` to `third party` across the API surface (GraphQL, MCP), database schema (migration), webhook event types (`vendor:*` → `third_party:*`), snapshot type (`VENDORS` → `THIRD_PARTIES`), and console / trust URL paths (breaking)
- Log `identity_id` on every authenticated request (cookie session, API key, OAuth2 access token) so operators can correlate a request back to its user and credential

## [0.188.0] - 2026-05-13

### Changed

- Derive cookie consent mode dynamically from the visitor's country and applicable regulation at consent-recording time; the `consent_mode` column is dropped from `cookie_banners` and persisted on `cookie_consent_records` instead, defaulting to `OPT_OUT` when no regulation matches (breaking)
- Capture `X-SDK-Version` via middleware and include it as `sdk_version` on all cookie banner request logs
- Use distinct badge colors per resource type and tracker type instead of only highlighting scripts

### Fixed

- Eliminate deadlocks when concurrent `ReportDetectedTrackers` calls update `tracker_patterns.last_matched_at` by replacing per-row updates with a single bulk update
- Stop generating bare `*` tracker patterns from separator-less cookie names; such names are kept as individual exact-match patterns for triage

### Removed

- Drop the legacy `cookies` and `cookie_patterns` tables (superseded by `tracker_patterns` and `detected_trackers`)

## [0.187.0] - 2026-05-12

### Added

- Add a shared `common_third_parties` reference catalog, seeded from `packages/vendors/data.json` via a one-shot `common-third-parties-import` CLI, and back the `CreateVendorDialog` autocomplete with a new `commonThirdParties(name)` GraphQL query (server-side `ILIKE` search) instead of shipping the full vendor JSON bundle to the browser
- Self-host vendor logos in S3: at import time, fetch each site's HTML and pick the best icon (SVG, `apple-touch-icon`, large PNG, `msapplication-TileImage`) via the new `pkg/webinspect` package, then serve through the existing `/api/files/v1/{id}` endpoint instead of calling Google's favicon service per page load

### Changed

- Sanitize MCP error responses so internal details (stack traces, wrapped errors) are no longer leaked to clients

### Fixed

- Return a clean not-found error instead of a 500 when a membership lookup misses
- Upgrade `mermaid` to 11.15.0 to address GHSA-6m6c-36f7-fhxh (Gantt infinite-loop DoS), GHSA-xcj9-5m2h-648r and GHSA-87f9-hvmw-gh4p (CSS injection via `classDef`/configuration), and GHSA-ghcm-xqfw-q4vr (HTML injection via `classDef` in state diagrams)

## [0.186.1] - 2026-05-12

### Fixed

- Fix wrong entity types in `tracker_patterns` and `detected_trackers` GIDs: rows carried entity types of removed `CookiePatternEntityType` / `CookieEntityType` instead of `TrackerPatternEntityType` / `DetectedTrackerEntityType`

## [0.186.0] - 2026-05-12

### Changed

- Update kit package

## [0.185.0] - 2026-05-12

### Added

- Add `TrackerResource` entity for detected scripts, iframes, images, beacons, fonts, fetches, media, and service workers, with full GraphQL, MCP, CLI, and frontend surface (list, view, create, update, delete, move-to-category); new "Resources" page under the cookie banner configuration tab
- Add `GLOB` match type for tracker patterns supporting prefix, suffix, and sandwich patterns (e.g. `ph_phc_*_posthog`), with duration-aware merging so trackers with materially different lifetimes are no longer collapsed into a single pattern
- Detect HTTP-header cookies via the Chromium `CookieStore` change event and expose a new `http` cookie source
- Add tracker-type filter and color-coded badges on the trackers page for quick visual scanning across Cookie / localStorage / sessionStorage / IndexedDB / Cache Storage
- Capture script initiator URL on detected trackers to enable per-vendor attribution for cookies and storage writes (column captured now, surfaced later)

### Changed

- Replace `PREFIX` tracker pattern match type with `GLOB` across GraphQL, MCP, and the frontend; existing `PREFIX` rows are migrated to `GLOB` with a trailing `*` (breaking)
- Make tracker pattern `displayName` read-only across GraphQL, MCP, and the frontend — it is now derived from pattern + match type (breaking)
- Pattern analysis worker now detects UUID-like, hash-like, and long numeric tokens as variable parts even from a single observation, so site-specific identifiers no longer get treated as static text
- Rename the cookie banner "Detection" page to "Trackers" and drop the `SCRIPT` / `IFRAME` tracker types (replaced by `TrackerResource`) (breaking)
- Agent runs now treat ctx cancellation as a graceful suspend signal: supervisor shutdown maps to run ctx cancellation, and the previous `WithStopSignal` API is removed (breaking for in-process callers)

### Fixed

- Fix empty country code being persisted on cookie consent records when IP geolocation returns no matching CIDR block
- Fix SQL corruption (HTTP 500 on `/report`) in `FindMatchingPattern` caused by `fmt.Sprintf` interpreting `%` characters in the LIKE escape clause
- Use `@deleteEdge` on the access review campaign delete mutation so the cached connection no longer surfaces a missing-data error when reopening the access reviews tab

## [0.184.2] - 2026-05-08

### Security

- Upgrade go to 1.26.3

## [0.184.1] - 2026-05-08

### Changed

- Microsoft 365 access review driver now fetches only internal members from Microsoft Graph (`$filter=userType eq 'Member'`), so guest (B2B) accounts are no longer pulled into access review
- SCIM settings page now hides the other IdP connector card once a bridge is connected; both remain listed when nothing is configured

### Fixed

- Fix cookie banner opt-out button opening the preference panel instead of performing a one-click reject in OPT_OUT regulations

## [0.184.0] - 2026-05-07

### Added

- Allow editing approvers inline on SOA generated documents from the Statement of Applicability detail page (visible after first publish)

### Fixed

- Fix Microsoft 365 SCIM bridge: register the `MICROSOFT_365` connector provider, scope each Identity Provider card to its own bridge type so connecting one provider no longer marks others as connected, and filter Microsoft Graph users to home-tenant members (skip B2B guests)
- Fix cookie banner REST config endpoint compatibility for SDK versions ≤ 0.2.0
- Fix geolocation IP-to-country block imports

## [0.183.0] - 2026-05-07

### Added

- Add IP-to-country geolocation service with shadow-table swap import and CIDR-based lookups
- Detect the visitor's privacy regulation (GDPR, UK GDPR, FADP, CCPA, PIPEDA, LGPD, LFPDPPP, POPIA, PDPA, PIPL, PIPA, APPI, DPDP, PDPL) on the cookie banner config endpoint and adapt the banner UI and texts accordingly (opt-out notice for CCPA, simple notice when no regulation applies)
- Store regulation and country code on cookie consent records and expose both across GraphQL, MCP, CLI, and n8n
- Allow deleting access review campaigns from the UI (DRAFT or CANCELLED only, gated on `core:access-review-campaign:delete`)
- Support Google Cloud Identity in the SCIM bridge (in addition to Google Workspace)

### Changed

- Access review campaigns no longer transition to `FAILED` when individual sources fail to fetch; the failure stays surfaced on the source fetch (status + last error) and reviewers can proceed on the sources that succeeded (breaking: removed `FAILED` from `AccessReviewCampaignStatus`)
- Allow editing metadata (title, document type, classification) on generated document versions; only content edits remain rejected

### Fixed

- Fix cookie banner docs link to `www.getprobo.com/docs`

## [0.182.0] - 2026-05-06

### Added

- Add Microsoft 365 SCIM bridge and access review driver
- Add unified tracker detection backend with `tracker_patterns` and `detected_trackers` schema
- Add `trackerType` field on patterns to support tracking technologies beyond cookies

### Changed

- Replace `publishMajor`, `publishMinor`, and `requestDocumentVersionApproval` mutations with a unified `publishDocument` and `bulkPublishDocuments` accepting `minor: Boolean!` and a required `changelog: String!` (breaking)
- Rename cookie pattern API surfaces to tracker patterns across GraphQL, MCP, CLI, and n8n (breaking)

### Removed

- Remove legacy `cookie_patterns` GraphQL schema, MCP tools, CLI commands, and n8n operations

### Fixed

- Restore MCP cross-origin protection after go-sdk v1.6.0 bump

## [0.181.0] - 2026-05-05

### Added

- Add SCIM tools to MCP API
- Add SCIM commands to CLI
- Add cookie banner detection page for uncategorised patterns
- Add `last_detected_at` and `last_matched_at` tracking on cookie patterns
- Add `uncategorisedPatterns` GraphQL connection on `CookieBanner`

### Changed

- Accept CIDR ranges in proxy `trusted-proxies` configuration
- Rename `categories` to `consentCategories` on cookie banner API surfaces
- Move cookie management from separate Cookies tab into the Display page
- Filter uncategorised category from cookie banner config and version snapshots

## [0.180.0] - 2026-05-04

### Fixed

- Use natural sort for SOA document export rows

### Added

- Add risk publish to document system

## [0.179.1] - 2026-05-02

### Fixed

- Fix n8n cookieConsentRecord getAll operation

## [0.179.0] - 2026-05-02

### Added

- Add cookie banner operations to n8n node
- Add `excluded` flag to cookie patterns (GraphQL/MCP/CLI/n8n) with source badge in category table
- Validate cookie policy link in banner description

### Changed

- Skip draft cookie banner version for uncategorised-only merges
- Exclude uncategorised category from consent contract
- Run cookie detection regardless of banner state
- Stop bumping cookie banner version on no-op updates
- Exclude translations from cookie banner version snapshots
- Allow clearing optional fields in n8n cookie updates
- Bump `@probo/cookie-banner` to 0.2.0

### Fixed

- Clear pending cookie-consent queue before stopping on 404

## [0.178.0] - 2026-05-01

### Added

- Add MCP tools for cookie banner, category, pattern, version, and consent records
- Add CLI commands for cookie banner, category, pattern, and consent records

### Fixed

- Fix auditor access to processing activities
- Fix contract end date field cut off in Add Person dialog

## [0.177.1] - 2026-04-30

### Fixed

- Reveal cookie banner sidebar entry in IAM organizations
- Render cookie-consent placeholders when no prior consent exists
- Fix cookie-consent placeholder sizing for absolutely or sticky positioned elements
- Allow OIDC and magic-link sessions to assume password-only organizations

## [0.177.0] - 2026-04-30

### Added

- Add cookie patterns to group detected cookies by URL prefix, with auto-detection worker and console management
- Add `DurationInput` component to `@probo/ui`

### Changed

- Refactor cookie banner forms to react-hook-form
- Store cookie durations as `max_age_seconds`
- Update `@probo/cookie-banner` public exports and bump to 0.1.0

### Fixed

- Filter browser-extension cookies from detection

## [0.176.1] - 2026-04-29

### Fixed

- Fix empty text nodes in generated documents

## [0.176.0] - 2026-04-29

### Added

- Add vendor publish to document system, replacing snapshot mode

## [0.175.0] - 2026-04-29

### Added

- Add processing activity, DPIA and TIA publish to document system, replacing snapshot mode

### Changed

- Introspect OAuth2 refresh tokens per RFC 7662, honoring `token_type_hint`
- Invalidate other sessions on password change and all sessions on password reset
- Use forwarded headers for SCIM event client IP when running behind a load balancer
- Extract client IP from rightmost entry of `X-Forwarded-For` and `Forwarded` headers
- Update avatar initials colors

## [0.174.0] - 2026-04-28

### Added

- Add agent run supervisor with checkpoint persistence and resume across restarts
- Add finding and obligation publish to document system, replacing snapshot mode
- Add `--state` and `--contract-ended` filters to CLI/MCP/GraphQL user list
- Add Notion workspace name resolver for access review
- Add `X-SDK-Version` header to cookie banner SDK requests

### Changed

- Rename `excludeContractEnded` to `contractEnded` (two-way) across MCP, GraphQL, CLI, frontend
- Remove auditor's ability to publish SoA
- Request Google customer directory scope for access-review name sync

### Fixed

- Fix copy-paste in rich editor
- Fix long cookie name display and label colors in cookie banner
- Fix suspension checkpoint fallback in nested and parallel agent execution

## [0.173.0] - 2026-04-27

### Changed

- First per-package release. Prior history is in the archived monorepo [CHANGELOG.archive.md](../../CHANGELOG.archive.md).
