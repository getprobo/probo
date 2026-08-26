# Changelog

All notable changes to `probod` (the server, including the bundled `@probo/console`, `@probo/compliance-portal`, and `@probo/ui` frontends) will be documented in this file.

## Unreleased

## [0.269.0] - 2026-08-26

### Added

- Visitors get a dedicated page in the compliance portal permissions zone, replacing the Edit Access dialog: document grants have their own table and URL, changes save immediately, and multi-select with a bulk bar grants or rejects several documents at once
- NDA card accepts a PDF drop directly, with download and delete controls in its header and the signature audit trail behind an activity popover

### Changed

- Migrated the compliance portal permissions zone to UI v2, including a new native Table in `@probo/ui`, visitors restyled as a list showing NDA status and join date, and pending requests surfaced as a button
- Renamed the portal permissions route to `visitors`, so the URL, navigation, and Slack action links match the page name
- PDF watermarks now show the document's own classification instead of a hardcoded "Confidential"
- Each compliance portal page owns its H1, description, and document title instead of repeating the portal name

### Fixed

- Compliance portal document access reused source GIDs for rows that did not exist yet, so Relay saw conflicting typenames on the access list; grant and reject now upsert a real row and leave sibling documents untouched
- Visitor list row edges were not clickable because list padding sat outside the link

## [0.268.0] - 2026-08-26

### Added

- Treatment plans on risk analyses: scores, treatment, and owner live on a plan unique to one risk and one analysis (the risk must already sit on a scenario), so the same catalog risk can evolve across periods; linked measures drive progress from inherent to residual except Accepted which stays inherent
- Risk analysis pages split heatmap and plan table from diagrams
- WORKLOAD_IDENTITY connector protocol: Probo mints a short-lived OIDC assertion and the customer's STS exchanges it for temporary credentials against a role they own, so the connection stores no credential

### Changed

- Matrix size is chosen when creating a risk analysis and cannot be changed later
- Access-review connect actions use a split button when a source supports more than one method (GitHub App, OAuth, API key)

### Fixed

- Slack access-request cards still linked to pre-`/governance` document and audit paths, so reviewers hit 404s
- PDF exports of documents with omitted ProseMirror node attributes (for example a table cell without colspan) showed raw JSON instead of the rendered document

## [0.267.0] - 2026-08-25

### Added

- Catalog attribution (vendor name, first-party, or still-identifying) now shows on tracker pages in console, GraphQL, and MCP, for patterns whose catalog entry has no vendor name of its own

### Changed

- Migrated the compliance portal integrations page to UI v2, including a reworked Slack channel picker (correct icon, wider card, refreshes on open, fixed pagination)

## [0.266.0] - 2026-08-25

### Added

- ChatGPT and Codex MCP OAuth compatibility through CIMD auth-method negotiation, RFC 9207 issuer identification, and resource-bound access and refresh tokens

### Migration

Before starting this release, apply `20260824T102541Z.sql`, replace the URL
below with the exact `PROBOD_BASE_URL`, and run:

```sql
BEGIN;

UPDATE iam_oauth2_authorization_codes
SET resources = ARRAY['https://your-probo.example.com']
WHERE resources IS NULL;

UPDATE iam_oauth2_consents
SET resources = ARRAY['https://your-probo.example.com']
WHERE resources IS NULL;

UPDATE iam_oauth2_access_tokens
SET resources = ARRAY['https://your-probo.example.com']
WHERE resources IS NULL
  AND client_id IS NOT NULL;

UPDATE iam_oauth2_refresh_tokens
SET resources = ARRAY['https://your-probo.example.com']
WHERE resources IS NULL;

COMMIT;
```

Manual access tokens remain unbound because they have no `client_id`.

## [0.265.1] - 2026-08-24

### Fixed

- The GitHub App connector's authorization URL didn't pin a redirect_uri, so GitHub could send installs to the wrong callback URL on deployments with more than one registered

## [0.265.0] - 2026-08-24

### Added

- GitHub App as an access-review connector alongside OAuth and personal access tokens: users authorize an app installation, short-lived installation tokens are minted automatically, and connector health checks are protocol-aware
- OAuth2 and SSH auth methods for access-review entries, so GitHub App tokens and deploy keys are recognized instead of being reported as API keys or service accounts
- Resend as an access-review connector via CIMD authentication
- Human review tracking for third-party catalog entries: a reviewer can mark a catalog row validated or rejected, and the tracker mapping pipeline now honors a rejected verdict instead of re-attributing it to a vendor

### Changed

- Compliance portal hosting page redesigned onto the v2 UI kit: domain cards restyled around SSL status, an inline form replaces the domain dialog, and visibility is split into switch cards
- Access-review table columns widened to fit longer connector labels; GitHub App and OAuth2 connector callbacks now route through separate endpoints instead of sharing one

### Fixed

- A concurrent review update could be silently overwritten mid-mapping-run because the read wasn't locked
- The organization picker for API keys, which had regressed
- Mermaid diagrams duplicating nodes
- The Microsoft catalog entry's category and a duplicate Tawk.to catalog entry
- Upserting a common third party that already existed was reported as created instead of updated

## [0.264.1] - 2026-08-21

### Fixed

- The Cal.com and Calendly connectors were missing their client secret from required-configuration validation and `.env.example`, so a misconfigured deployment would silently fail instead of being caught at startup

## [0.264.0] - 2026-08-21

### Added

- A per-organization cloud identity issuer: probod can now act as an outbound OIDC provider, minting short-lived tokens scoped to one organization so audits inside customer AWS accounts no longer require long-lived shared credentials
- Calendly and Cal.com as access review connectors, both via OAuth (Cal.com also supports team accounts), with provider icons in the connector list

### Changed

- Webhook deliveries now run as durable jobs with processing leases, retry scheduling, stale-delivery recovery, and a terminal dead-letter state, with idempotency headers and bounded concurrent delivery
- OAuth2, SAML, ACME, and identity-federation private keys are now decoded when configuration loads instead of at server start, surfacing a malformed key immediately rather than at the next restart

### Fixed

- Fixed contract start/end dates not being updatable

## [0.263.0] - 2026-08-19

### Added

- Tracker attribution now judges third parties by data egress rather than code origin, catching vendor SDKs bundled into first-party code; artifacts belonging to a browser extension or other visitor-installed software are now marked `NOT_ATTRIBUTABLE` instead of first-party or a wrongly attributed vendor

### Changed

- Connecting Tally as an access review source now validates the API key at connect time and derives the organization id automatically; the connection probe and name resolver moved to the one Tally endpoint that accepts API-key auth, fixing every Tally connection reporting "credentials are invalid"
- The mapping worker no longer re-evaluates tracker writes already confirmed to come from a browser extension, since that evidence settles attribution on its own
- The `third_parties.common_third_party_id` index can now be built `CONCURRENTLY` ahead of a deploy; the migration is a no-op if the index already exists

### Fixed

- "Open in Probo" links from an access request and from Probot Slack notifications now land on the compliance portal's permissions page instead of a removed console route
- A tracker artifact could gain a vendor attribution after another worker had already settled it terminal (first-party / not-attributable); the terminal verdict now always wins
- The enrichment worker could re-process and re-attribute a tracker row that had already received a terminal verdict

## [0.262.0] - 2026-08-18

### Added

- authentik as an access review source, connected with an API-intent token and the instance URL. MFA status is derived from the instance's authenticator devices, and stays unknown rather than disabled when a device kind is unreadable
- Probot: a Slack bot that lets employees link their Slack identity to their Probo account (`/probot login`) and manage the link from their profile. Bound users get compliance review and approval notifications in Slack with per-item actions, and can drive Probo actions conversationally from a bound channel. Organizations manage the Slack install, channel, and identity bindings from Settings

### Changed

- **Operators running their own Asana OAuth app must switch it to "Full permissions" in the Asana developer console before upgrading.** Asana publishes no granular scope covering workspace memberships and rejects an authorize request that mixes `default` with granular scopes, so an app left on granular scopes refuses both new connections and reconnects. Existing Asana connectors keep their old grant and must be reconnected once
- Reorganized the console navigation into a two-level product rail and panel, replacing the flat sidebar. Settings now groups Organization, IAM (Users, Auth & Provisioning, Audit Log), Registries, and Webhooks; Access Reviews, Privacy (Processing Activities), Third-Party Risk Management, compliance portals, and cookie banners each get their own switcher or panel. Dark mode, which regressed during the rework, is restored
- Third-party and compliance-portal profile pages (Profile, Assurance, Stakeholders, branding) now save each field as it's edited (on blur) instead of via a page-level Save button
- Compliance portal review and tool actions now validate that the resource being acted on belongs to the access request, without blocking valid partial (per-item) approvals

### Fixed

- Asana access reviews failed to fetch any account. The workspace memberships endpoint the driver reads is reachable only with Asana's full `default` scope, so a connector granted `users:read` and `workspaces:read` connected successfully and then got 403 on every sync. Asana admin, guest and view-only flags now stay unknown when the API withholds them instead of being reported as false
- Fixed a crash when adding a new Statement of Applicability entry (the create mutation omitted maturity level)
- Long labels in risk analysis diagrams no longer get clipped: Mermaid flowchart and threat-hexagon labels wrap again after the Mermaid 11.13 upgrade dropped automatic wrapping
- Windows device posture (BitLocker protection, password policy, time-sync status) now reports correctly on non-English systems instead of showing Unknown

## [0.261.1] - 2026-08-17

### Fixed

- Updated Go dependencies to address a reported security vulnerability

## [0.261.0] - 2026-08-17

### Added

- Compliance portals can now run behind an externally terminated TLS load balancer, keeping strict domain routing and HTTP redirects while skipping ACME setup and avoiding dead HTTPS ports in that mode

### Fixed

- CIMD clients skipping consent now require both a verified custom domain host and identity-only scopes (`openid`/`profile`/`email`), closing a gap that could grant broader API access without a consent screen
- Agent netcheck now blocks NAT64, 6to4, Teredo, and CGNAT address ranges, closing an IPv6 SSRF gap in the agent tool guard

## [0.260.0] - 2026-08-14

### Added

- MCP `listRiskMeasures` tool to list measures linked to a risk (reverse of `listMeasureRisks`), so callers can discover links to unlink before `deleteRisk`.
- Explicit likelihood/impact matrix size (3×3, 4×4, or 5×5) on risk analyses, across GraphQL, MCP, and the console
- Device enrollment now has a dedicated "Download the Probo agent" step linking to pro.bo/install before organization selection, skippable when the agent is already installed

### Changed

- OAuth consent and token screens now translate scope labels instead of showing raw English strings
- Devices list shows the last reported agent version, and the Windows OS version column now shows the short numeric version instead of the full localized banner

### Fixed

- Deleting a risk that still has linked measures, documents, or other references now returns a clear "resource is in use" error instead of an opaque internal error.
- Linking a measure to an existing task is no longer silently dropped on update

## [0.259.0] - 2026-08-13

### Added

- `COMPLIANCE_PORTAL_MANAGER` and `COMPLIANCE_PORTAL_ACCESS_MANAGER` membership roles, for delegating full compliance portal management or just visitor access approval without broader admin access
- AI Systems register: track an organization's AI system inventory with a dedicated backend, console UI, and publish flow, alongside CLI, MCP, and n8n surfaces
- Optional period start/end fields on risk analyses
- Cookie banner resource reporting switch, so operators can disable resource detection per banner while cookies and storage keep reporting normally

### Changed

- Compliance portal tabs a role cannot access now show a no-access state inline instead of being hidden, so the header and other tabs stay usable
- Cookie banner resource reporting and compliance portal rights requests are exposed through a nested `capabilities` object on GraphQL and MCP instead of separate flat boolean fields; the CLI, n8n node, console, and visitor portal follow. The browser-facing banner config keeps its flat `resource_reporting_enabled` key so already-released cookie banner SDKs keep working
- The person profile "Type" field is free text instead of a fixed set of options
- Document content JSON now allows up to 1 MiB, up from 500,000 bytes

### Fixed

- Evidence names are shown again as a separate column in measures, so linked and uploaded evidence stays identifiable when no generated description is available
- The vetting agent no longer overwrites a third party's countries

### Removed

- Dropped the unused `generateDocumentChangelog` mutation; the console publish flow already collects a changelog by hand

## [0.258.0] - 2026-08-11

### Added

- Nuki access review integration: connect a Nuki account to review who can open which smart locks, with door grants, roles, last activity, and active state per person, and unlinked keypad codes and fobs surfaced as service accounts.

### Fixed

- Certificates of Completion now print `signed_at` at the same microsecond precision used to compute the seal, so the certificate's own Seal Verification procedure reproduces the digest instead of appearing to fail. Certificates generated before this change remain unverifiable by the printed procedure.
- Electronic signatures are rejected when the signer has no name, and compliance portal visitors without a full name are redirected to set one before they can sign the NDA instead of accepting it with a blank signature.
- Identity and membership profile names are trimmed when set, so whitespace-only names can no longer be stored.
- The console's Content-Security-Policy now allows Google favicons served from `*.gstatic.com`, and console schema validation no longer needs `eval`, so both work under the policy.
- Access review CSV sources accept multiple roles in the `role` column, separated by semicolons.

### Changed

- Access review admin status, MFA, and other provider signals distinguish unknown from false: values the source does not report, or reports in an unfamiliar form, now display as unknown instead of being shown as negative. Campaign entry rows and the additional-roles popover were reworked to present this consistently.
- Document content now allows up to 200,000 characters of text, up from 50,000.

## [0.257.0] - 2026-08-11

### Added

- Console and compliance portal now send a Content-Security-Policy header, with a matching policy applied to their local Vite dev servers, restricting scripts, styles, and connections to known origins (including object storage, Google Fonts, Google favicons, and Markdown image sources), and hardened against configuration and origin injection.

### Fixed

- Clearing a person's contract end date now updates the people list immediately instead of only on the server, and the people list's actions column no longer starves date columns of space.

### Changed

- Cookie banner presentation now matches each jurisdiction's actual disclosure duties: Japan and jurisdictions with no cookie-consent law move to the opt-out layout instead of interrupting every visitor on load, Brazil (LGPD) requires prior opt-in, and a US visitor whose state cannot be resolved falls back to CCPA rather than no regulation. Only Mexico keeps the notice-on-load presentation.
- Opt-out button copy is split into a generic label and a California-specific statutory label, so the "Do Not Sell or Share" phrasing no longer appears outside US state privacy laws.

## [0.256.0] - 2026-08-10

### Added

- People list now shows each person's contract end date, making it easier to see why someone with an ended contract is excluded from document signatures despite still appearing active.

### Fixed

- Blank SCIM position/kind values are now treated as unset instead of failing person-form validation, and the people table's empty-state row spans the correct number of columns.
- Cookie banner tracker detection no longer aborts the whole batch when an SDK-reported cookie or storage key exceeds PostgreSQL's index size limit; oversized identifiers are capped at 255 bytes and dropped instead.
- The third-party compliance document agent now downloads PDFs directly instead of navigating to them, so PDF documents are read and verified correctly.

### Changed

- SCIM audit event writes no longer block the provisioning request, and are now bounded to a fixed duration.

## [0.255.0] - 2026-08-10

### Added

- SCIM audit events now store the request and response bodies exchanged with the provisioning client, with password fields and credential attributes redacted regardless of the schema URN used to reference them.

### Fixed

- The compliance portal's unsigned-NDA banner no longer appears before a visitor has requested access to a private document, and an admin-granted access no longer triggers it as if the visitor had requested access themselves.
- Locked document viewer CTA now names the NDA instead of reusing the banner's generic "Review and sign" label.
- Compliance portal mailing list update detail pages no longer 404.

## [0.254.0] - 2026-08-10

### Changed

- Organization creation is now restricted to owners when signup is disabled.
- Console access is now restricted when the user has no organization membership.
- Risk analysis scopes are now referred to as diagrams across the API and console UI.

### Fixed

- Resolved npm audit vulnerabilities in `dompurify`, `js-yaml`, and `mermaid` dependencies.

## [0.253.0] - 2026-08-07

### Fixed

- Evidence uploads now display in the PDF viewer instead of a degraded fallback when the display mode was changed or the browser navigated away and back

## [0.252.0] - 2026-08-07

### Changed

- Organization risk assessments are now named risk analyses across the database, console API, MCP API, CLI, and console UI. Third-party assessments and Statement of Applicability control flags keep their existing naming.
- Compliance portal document, audit, and subprocessor selection is reversed: the console now lists the organization's own documents, audits, and subprocessors with checkboxes for inline portal membership, replacing the portal-only rows plus separate add dialogs.
- Compliance portal catalog document, audit, and third-party rows are now exposed as Relay Nodes in the console API.

### Fixed

- Unchecking a compliance portal document no longer reports a remove error after the deletion actually succeeded.
- `probod` now persists its local ACME account key across restarts, so pending certificate orders no longer fail with 401 after a restart; unauthorized order polls now clear resumable state and retry provisioning instead of stalling.
- Compliance portal document table now spaces the alias and visibility controls instead of rendering them flush against each other.

## [0.251.0] - 2026-08-06

### Added

- Compliance portal Requests surface (rights requests) can now be enabled or disabled per organization from the console overview; when disabled, the portal hides the nav item and treats `/requests` as not found.
- Document export watermarks now accept custom text instead of only an email address, with a 64-byte size limit and validation; existing `watermarkEmail` usage keeps working.
- Access review campaign entries can now be filtered by account status (Active, Disabled, Unknown), mirroring the existing MFA filter.

### Changed

- Portal "Get Access" CTA is now labeled "Request Access" throughout the TopBar and document flows.
- Access review campaign entries now show admin status as Yes/No text instead of a check/X icon, so admins stand out instead of non-admins looking flagged.

### Fixed

- Locked document viewer now shows the correct CTA — Sign NDA takes priority over Request Access when the portal NDA is incomplete — instead of always showing Get Access regardless of a pending request.
- Compliance portal asset URLs on nested routes (e.g. `/en/documents/...`) no longer resolve under the route path and serve HTML instead of the expected chunk.

## [0.250.0] - 2026-08-06

### Added

- HubSpot access-review connector now derives last login from successful events on the Account Activity login API (past 90 days). Existing HubSpot connectors must reconnect (or private apps must grant `account-info.security.read`) before last-login population works; missing scope leaves last login unset rather than failing the fetch.
- Access review campaign entries now show each account's authentication method (e.g. SSO or password), with a matching filter in the campaign toolbar.
- Obligations linked to a control now expose a total count in the console API.
- Mailing list subscribers and updates are now resolvable as GraphQL nodes in the console API.

### Fixed

- HubSpot access-review connector now marks deactivated/archived users as inactive by consulting the Owners API (`archived=true`) instead of treating every Settings Users row as active. Existing HubSpot connectors must reconnect (or private apps must grant `crm.objects.owners.read`) before inactive detection works.
- Closed an open-redirect bypass where a control byte (tab, newline, etc.) before a path segment could collapse to an off-origin protocol-relative URL, evading the fix for CVE-2026-49820.
- Access review campaign entries no longer silently truncate for sources with more than 500 entries; entries are now paginated instead of clamped to the API's page limit.
- Third-party business associate agreement, compliance report, and data privacy agreement resolvers now return the correct linked third party instead of erroring or resolving the wrong one.
- Creating a compliance portal reference without a logo file no longer fails; a default placeholder logo is used instead.
- The MCP `SendMailingListUpdateTool` now checks the send permission instead of the update permission.
- Evidence uploads now reject empty files, and evidence without an associated task no longer returns an internal server error.
- MCP-created tracker patterns now default to the script source instead of leaving it unset.

## [0.249.1] - 2026-08-06

### Fixed

- Compliance portal PDF preview no longer fails to load when export/download still worked: the viewer worker is now bundled from the same pdf.js version as react-pdf

## [0.249.0] - 2026-08-05

### Added

- Subdivision-aware geolocation: IP lookups now resolve an ISO 3166-2 subdivision alongside the country, and the resolved jurisdiction is persisted on consent records. Bundled country blocks are replaced by provider-neutral ranges so licensed importers can supply subdivision data
- US visitors are mapped to their own state privacy law (VCDPA, TDPSA, CPA, and so on) instead of being labeled CCPA. California keeps the statutory Privacy Choices experience; other mapped states use a generic opt-out banner with US-specific copy
- Canadian visitors are mapped per province: Quebec to Law 25, Alberta to PIPA_AB, and British Columbia to PIPA_BC, with PIPEDA retained elsewhere. Canadian opt-out banner copy applies to PIPEDA and provincial PIPA; `PIPA_CA` remains valid for legacy records
- `subdivisionCode` on the consent record list and detail pages, so operators can verify the detected jurisdiction

### Changed

- Vetting agent notes in the third-party register are parsed into structured document blocks instead of being embedded as raw Markdown, so headings, tables, and bold markers render correctly in generated documents and PDFs

### Fixed

- MS365 OAuth reconnects no longer appear incomplete: a refresh token is treated as `offline_access` even though Microsoft's v2 token response omits that scope
- A Markdown conversion failure in a single risk-assessment note no longer aborts the whole third-party register publish; the note falls back to a plain-text paragraph

## [0.248.0] - 2026-08-05

### Changed

- Restructured the access review console list UI: dedicated list item components for campaigns, sources, and entry sections; replaced the add-source dialog with the provider list flow; consolidated fetch-status UI, shared the campaign deletable-status helper, and now load remaining sources while searching so matches beyond the first page aren't missed

### Fixed

- Presigned S3 requests (`PresignGetObject`) no longer fail with a misleading "not found" error
- A failed page load during source search no longer hides "Load more", so later pages stay reachable

## [0.247.0] - 2026-08-05

### Added

- Organizations can now track DORA Critical ICT Functions via a new Business Functions register — link assets, third parties, and people to a business function, with matching CLI, MCP, and n8n operations
- Dutch (nl-NL) language support extended to the compliance portal's documents, NDA, requests, subprocessors, and updates pages

### Changed

- Revamped the access review campaign bulk-decision UI: selection now survives partial and stale failures (retaining failed entries so bulk actions can be resubmitted), and accessibility and narrow-layout behavior were hardened

### Fixed

- Cloudflare access reviews are now scoped to a single account instead of merging members across every account reachable by the API token, which produced duplicate emails
- Asana access review roles and admin status are now populated from membership data instead of always showing empty
- Fixed the OpenAI connector probe endpoint
- Uploads to Google Cloud Storage's S3-interoperability endpoint no longer fail with `SignatureDoesNotMatch`; the SDK no longer signs the `Accept-Encoding` header, which GCS rewrites in transit
- Device codes are now kept for 15 minutes past expiry before garbage collection, so a client polling right at expiry gets `expired_token` instead of `invalid_grant`

## [0.246.0] - 2026-08-04

### Added

- Cookie banners can now present a dedicated CCPA "Your Privacy Choices" panel (right to opt out of sale/sharing, right to limit sensitive personal information) and a first-class notice-only presentation for jurisdictions with no cookie-consent law, driven by an additive `layout` field in the banner config
- Consent records can now capture a distinct `ACKNOWLEDGE` action for notice-only banners instead of always recording an accept-all; exposed through GraphQL, MCP, the console consent-records view, and n8n
- MCP access review entries now include the campaign source name and connector ID, so agents no longer have to infer the source tool from account metadata

### Changed

- Generated tracker policy documents no longer show a "Last updated" timestamp in the body, since it churned on every automatic regeneration

### Fixed

- Dutch (nl-NL) cookie banner translations were missing the Privacy Choices panel copy and the notice dismiss label added after the nl-NL backfill; both are now included

## [0.245.0] - 2026-08-04

### Added

- Dutch (nl-NL) language support across the console, compliance portal, and cookie banner
- MCP `tools/list` responses now include a cache TTL hint so well-behaved clients can skip refetching the static tool set every session

### Fixed

- The compliance portal subprocessor picker no longer lists non-top-level third parties; only level-one third parties can be selected as subprocessors
- Uploads to S3-compatible endpoints that don't support AWS's checksum extensions (e.g. Google Cloud Storage's S3-interop endpoint) no longer fail with a misleading `SignatureDoesNotMatch: Access denied` error when checksum calculation is explicitly configured

## [0.244.0] - 2026-08-03

### Added

- Organizations can now operate multiple compliance portals: portal-scoped catalogs, access, visitor experience, and console management (including portal selection and catalog publication controls) replace the single organization-wide portal, with matching CLI, MCP, and n8n operations for managing portals and their audit, document, third-party, file, and reference catalogs. The legacy "trust center" terminology and database tables are renamed to "compliance portal" throughout.
- Deployments can now override compiled-in connector endpoints (e.g. to point a connector at a vendor's sandbox), validated at boot with SSRF-hardened checks — atomic per-provider overrides, host/scheme/credential validation, and pagination guarded against off-host or path-traversal redirection so a spoofed next-page link cannot exfiltrate a connection's bearer token
- Expose mailing lists, detected trackers, compliance portal frameworks, and document/control/framework mutations via MCP that were previously console-only

### Changed

- Third-party business and security owners are now migrated into a single shared administrators list across GraphQL, MCP, CLI, n8n, and the console
- Access review source rows now explain why a provider returned no organizations (app not approved, personal account, or provider without a picker) instead of always falling back to a free-text slug input
- Providers whose organization is captured during the OAuth callback (PagerDuty, Vercel, Datadog, Zendesk) now show it read-only instead of an editable slug input that always failed to save

### Fixed

- Access review connector rows now show which OAuth scopes are missing after a reconnect, instead of leaving "Reconnect required" unexplained
- Microsoft 365 access review no longer fails an entire source fetch when `AuditLog.Read.All` is unavailable on the tenant; affected accounts import with MFA left unknown instead
- Linear connector errors no longer echo the provider's raw error message, which could leak tenant identifiers or query fragments
- A source with a missing or deleted connector is no longer reported the same way as a provider that legitimately returned no organizations
- Long process names on risk assessment Mermaid flowchart edges no longer get clipped to a single line
- Compliance portal document filtering is now consistent across the catalog, access, and visitor services

## [0.243.0] - 2026-07-31

### Added

- ITAM devices are now exposed via MCP and n8n (in addition to GraphQL): list, get, create, revoke, delete, and set-owner, with latest postures nested on list/get responses

### Changed

- Compliance portal locale-mismatch banner now falls back to a supported browser-language match for visitors without a saved locale, and stays silent for unsupported languages instead of assuming English is preferred

### Fixed

- Manual agent install instructions now point to `probo-agent/v*` releases instead of the repo-wide `/releases/latest`, which pointed at the wrong assets when other tracks published more recently
- SCIM event CSV exports no longer fail with a "cannot parse address" error caused by a missing email address in the underlying membership profile query

## [0.242.0] - 2026-07-30

### Added

- Organization owners and admins can now delete the organization's horizontal logo; the underlying file is soft-deleted so existing download URLs stop serving the image, and the operation is exposed in the n8n organization node
- Compliance portal shows an unsigned-NDA banner directly in the portal shell, with a "Sign" CTA and a Documents back link on the NDA page, instead of only surfacing the requirement when a document export fails
- Access-review connector rows show a "reconnect" prompt when the connector needs additional OAuth scopes
- Devices can now be soft-deleted after being revoked; ITAM garbage collection also hard-deletes orphaned pending/revoked devices with no API key, postures, or valid enrollment token

### Changed

- NDA signature is now required only for exporting a private (protected) document, report, or file — requesting access and exporting public files no longer require signing first
- Locale-mismatch and unsigned-NDA banners in the compliance portal are now visually distinguished (info vs warning) so they don't read as a single band when both are shown
- MCP third-party list now defaults to level 1 (direct third parties only) when no level filter is given
- Document signing/approval reminders due on a weekend now defer to Monday instead of sending over the weekend or spending an escalation step unused

### Fixed

- Microsoft 365 access review lists home-tenant organization members only, using the same Graph `/users?$filter=userType eq 'Member'` call as the SCIM bridge
- Microsoft 365 access review now resolves actual MFA status for users instead of always reporting it as unknown
- Portal mutations now consistently enforce sign-in, full-name, and NDA gates instead of relying on each call site to reimplement the redirect logic
- Guests who sign in from a shared-locale portal link no longer land back on it showing a stale locale-mismatch banner; the redirect now adopts the identity's saved locale
- ITAM GraphQL actions are now registered under the `v1:itam`/`v1:itam:read` OAuth2 scopes, fixing bearer-token callers that were failing closed despite having role-level permission
- Device enrollment token deletes are now scoped to the owning tenant
- SCIM audit-log CSV exports collected by ID no longer fail to load activation/deactivation timestamps

## [0.241.0] - 2026-07-30

### Added

- MCP tools now carry titles and complete annotation hints (`readOnlyHint`, `destructiveHint`, `idempotentHint`, `openWorldHint`) so clients like Claude present reads, writes, and deletes accurately

### Changed

- Member profiles now have a `PENDING` state distinct from `ACTIVE`/`DEACTIVATED`, so invited-but-not-yet-activated members remain assignable to assets, data, and risks; owner pickers and state filters across the console, CLI, MCP, and n8n now accept multiple states
- "Archive" user action renamed to "Deactivate" across the API, CLI, MCP, n8n, and console
- People list status filter is now a multi-select dropdown instead of cramped checkboxes, and reflects `PENDING` immediately after an activation email is sent
- Audit log and SCIM export CSV files now escape leading `=`, `+`, `-`, `@`, tab, and carriage-return characters to prevent spreadsheet formula injection

### Fixed

- Asset table's inline owner picker no longer offers deactivated profiles as owners
- People list search no longer briefly reverts to stale results when a debounced search resolves after a filter change
- Disabled dropdown checkbox items are now fully non-interactive (opacity, cursor, hover) instead of only dimming the checkbox glyph

## [0.240.0] - 2026-07-29

### Added

- Log export for audit logs and SCIM events: request a CSV export job from the console, Connect, MCP, or CLI, streamed and uploaded to S3 by a concurrent export worker
- Auth cookie `SameSite` attribute is now configurable (`PROBOD_AUTH_COOKIE_SAMESITE`, defaults to `Lax`), rejecting `None` unless `Secure` is enabled
- In-tab light/dark display mode toggle for signed-in compliance portal guests, independent of the OS color scheme
- Subprocessor cards truncate long country/region lists behind a "+N" popover

### Changed

- Access review campaigns can now be deleted in any status, except while a campaign is still fetching sources
- Document share action relabeled "Copy link" and exposed on each row of the documents list, not just the viewer
- Locked document viewer keeps showing the document title and an access-request CTA instead of withholding the whole page
- Guest language selector shows the locale code instead of the full language name to save space next to "Get access"
- Compliance portal TopBar stacks the brand name over the tagline and shortens nav labels on narrower viewports
- html2pdf-rendered PDFs (e.g. signature certificates) now request a tagged structure tree and document outline from Chrome for accessibility
- `safeRedirect` now accepts configured CORS origins as valid continue-URL hosts, alongside verified custom domains

### Fixed

- Document soft-delete now clears associated control/risk/measure mappings, fixing risk deletion when the linked document was already removed
- Console document `lang` attribute now follows the active i18next locale instead of always reporting English
- "Trusted by" logo tiles now stretch to match row height when a caption wraps to two lines
- Compliance portal drawer's dropdown/select menus are clickable again when the swipe drawer's transform was blocking pointer events underneath it
- Ukrainian "+N more regions" label now uses the correct one/few/many plural forms

## [0.239.0] - 2026-07-29

### Added

- Device page shows a formatted value per posture check alongside current postures; the Postures tab is replaced with paginated report history grouped by agent push time

## [0.238.0] - 2026-07-28

### Added

- Optional `auditStartDate`/`auditEndDate` on audits, exposed through GraphQL, MCP, CLI, n8n, and console, with new sortable order fields `AUDIT_START_DATE`/`AUDIT_END_DATE`

### Fixed

- Unverified password identities could open a session after signing out; password sign-in is now rejected with `EMAIL_NOT_VERIFIED` until the address is confirmed, with a resend-confirmation email flow
- Completing a magic link for an existing identity now marks the address verified, matching OIDC behavior
- Mermaid diagrams that fail to render now show a toast with a safe inline fallback instead of leaking raw Mermaid error markup into documents

## [0.237.0] - 2026-07-27

### Added

- Compliance portal documents page supports bulk requesting access: select several locked documents, reports, or files and request access to all of them in one call
- `updateAccessReviewCampaign` now accepts `accessReviewSourceIds` to sync a campaign's scope sources alongside other updates in the same call
- People list gained search and status/role/type filters across GraphQL, MCP, and CLI, with the default page size raised to 100
- Third-party selector fields on measures and other forms now show only first-level third parties

### Fixed

- Fixed a crash on the third-party detail page when a third party's country was set to the global pseudo-region
- Empty toast viewport no longer blocks pointer events (hover/click) over content underneath it, such as bottom action bar buttons

### Security

- Bumped `postcss` to 8.5.23 and `react-router` to 8.3.0 to remediate high-severity npm audit findings

## [0.236.0] - 2026-07-27

### Added

- Five new access-review connectors: Google Analytics (GA4), Dotfile, Segment, Square, and UpCloud, with account listing, role/status resolution, and connector logos
- French translation for the console and compliance portal, backed by a new `react-i18next` setup
- Documentation links for the Dotfile, Segment, and 19 other API-key access-review connectors' setup dialogs
- Restored the "Risk Accepted" status on finding create/update forms, with a searchable paginated risk picker to supply the required linked risk

### Fixed

- Preserve the post-login destination through OIDC, magic-link, and SAML auth-error redirects instead of dropping it on `/auth/error`
- Custom-domain CNAME verification no longer fails on multi-hop DNS alias chains
- 1Password and Langfuse access-review connectors could not be connected: required extra settings (SCIM bridge URL, base URL) were silently dropped by the console

## [0.235.0] - 2026-07-25

### Added

- Device management: enroll employee devices via the new `probo-agent`, view posture and admin the device fleet with owner assignment from the console, and let employees confirm enrollment and check their own device status without an assumed session
- "Documentation" links for connector setup: the API-key connect dialog and access-review Add Source dialog now link to a connector's docs page when available

### Changed

- Renamed the "My Signatures" console menu item to "Employee Portal"
- Auth failures (SAML, OIDC, magic-link) now redirect to a shared `/auth/error` page with a stable error code and a clear explanation instead of raw JSON or one-off pages; SAML failures always show a generic reason to avoid leaking account or org state

### Fixed

- Hardened custom-domain verification: CNAME/TXT checks now require the record owner to match the verified hostname, and CAA checks climb to the registrable domain and validate RFC 8659 syntax, closing a bypass where an apex record could satisfy verification for a subdomain
- Third-party category is no longer reset to "Other" when omitted from a partial update via MCP, GraphQL, or CLI

## [0.234.0] - 2026-07-24

### Changed

- Render longer vetting notes as markdown, keeping more of the orchestrator assessment text and skipping profile fields already on the third party

### Fixed

- Reject empty SAML NameIDs during assertion validation, preventing duplicate-key failures on later logins, and return a clear error when a NameID is already linked to another account
- Require portal redirect hosts to have a verified certificate (Active/Renewing) before allowing OIDC, magic-link, and compliance-portal OAuth `continue` redirects, closing an open redirect on newly claimed custom domains
- Fixed PostHog access-review sources being marked disconnected for EU cloud OAuth tokens by delegating region discovery to the shared resolver used by the access-review driver

## [0.233.0] - 2026-07-22

### Added

- Added pointer parallax to logo backdrops (`MediaTile`/`BackdropCard`), with eased entry and a frozen pose when the pointer leaves

### Fixed

- Split certificate provisioning into separate begin-challenge and poll-order workers so the global ACME rate-limit cooldown no longer blocks polling of in-flight orders
- Fixed several access-review connector defects that silently synced zero users or hammered vendor APIs (Sentry trailing-slash routing, Vercel OAuth team parameter, Brex company-read scope, Cloudflare `per_page` floor, picker providers now auto-select their first workspace, org defaulting now also applied via MCP, and the source-name worker no longer re-claims permanently-failed sources or loses a resolved name after reconnect)

### Removed

- Disabled the Clerk access-review connector: its Backend API only exposes application end-users, not workspace admins, so reviews targeted the wrong population

## [0.232.0] - 2026-07-22

### Added

- Exposed a gauge for the ACME rate-limit cooldown end time (`certmanager_certificate_acme_cooldown_until_timestamp_seconds`) and expanded provisioning failure logs with the full ACME error detail (status code, instance, link, subproblems, retry-after), making cooldowns and CA errors diagnosable

## [0.231.0] - 2026-07-22

### Added

- Compliance portal URLs are now locale-prefixed (e.g. `/fr/documents`), with self-canonical and hreflang SEO tags and a persisted identity locale that reconciles the browser locale against a saved preference via a mismatch banner

### Changed

- Compliance portal home page now shows a short entity name instead of the full portal title as its heading (API field renamed `title` -> `entityName`), restoring the localized hero composition and using the entity name as the OAuth client name

### Fixed

- Hardened automatic TLS certificate provisioning: fixed several race conditions and correctness gaps in the ACME worker (challenge acceptance ordering, write-back locking, Retry-After parsing, stale-order reset scoping, poll lease sizing, CAA error classification, and duplicate metrics registration) that could leave certificates stuck, unissued, or crash the provisioning worker

## [0.230.0] - 2026-07-21

### Added

- Webhook events now include `RIGHT_REQUEST_CREATED`, `RIGHT_REQUEST_UPDATED`, and `RIGHT_REQUEST_DELETED` for rights requests created via console or compliance portal
- Compliance portal visitors can now sign in via OAuth/OIDC, with sign-in and consent screens showing the requesting client's branding (logo and name)
- Compliance portal home page now displays the full custom portal title as its heading, instead of composing it from the organization name
- Added a configurable trust center base domain (default `probopage.com`) so a managed default domain and certificate are provisioned for every compliance portal at organization creation

### Changed

- Renamed "Trust Center" to "Compliance Portal" across the GraphQL API and MCP tools — notably `ComplianceExternalURL` is now `ComplianceCustomLink` and the visibility field is `compliancePortalVisibility`
- Moved public-facing profile fields (title, description, website, email, headquarters address) off the organization and onto the compliance portal
- Custom domains and their certificates now belong to the compliance portal rather than the whole organization, with issuance and renewal handled by a dedicated poll-based worker
- Console: restructured the brand page into profile, domains, visual identity, and custom link sections, and moved frameworks and the NDA card onto the overview page

### Removed

- Removed the one-time session-transfer authentication flow for custom-domain SSO cookies, now that portal visitors authenticate via OAuth

## [0.229.1] - 2026-07-21

### Fixed

- Mapped commitment and commitment group actions to the `v1:compliance-page` OAuth2 scope so tokens with that scope can create and manage commitments via MCP/API

## [0.229.0] - 2026-07-20

### Added

- MCP commitment and commitment group tools (create, list, update, delete) so automation can manage compliance portal commitments end to end

### Changed

- Redirect the legacy compliance portal `/overview` URL to home
- Introduced a v2 z-index scale so portaled menus and popups sit above in-page media

### Fixed

- Aligned the compliance portal PDF preview and desktop gutter with the page header, and matched list skeletons to the card surface

## [0.228.0] - 2026-07-20

### Added

- The compliance portal is now served in production for the `/trust` path and custom-domain sites, replacing the previous trust center frontend
- Added Data Requests pages to the compliance portal: data subjects can submit and track GDPR/CCPA rights requests (including rectification, objection, and complaint types) after magic-link sign-in, scoped to their verified email
- Added subscribe-to-updates in the compliance portal, with an Updates CTA and a sign-out action in the user menu
- Added eleven new compliance-portal locales: German, Spanish, Indonesian, Italian, Japanese, Korean, Polish, Portuguese, Turkish, Ukrainian, and Simplified Chinese

### Changed

- Renamed the document approval button to "Approve" and updated the consent text to match, since the action only approves
- Made the compliance portal layout responsive
- Standardized copy to say "compliance portal" instead of "trust center" throughout

### Fixed

- Fixed document row accessibility for mobile screen readers and corrected PDF page sync after fit-to-width reflow
- Preserved `Field` `aria-describedby` when showing form errors

## [0.227.0] - 2026-07-20

### Added

- Added a Trust Center documents page with category grouping, a Public/Private visibility filter, and a full document viewer for PDFs, images, and other file downloads
- Added configurable, reorderable commitment cards to the compliance portal home page, replacing the hardcoded placeholder content
- Added a sign-in dialog (magic link + OIDC) that gates trust-center access requests

### Changed

- Renamed risk overview labels for consistency: "Severity" to "Score" and "Inherent" to "Initial"
- Archiving a document now voids pending approval quorums and cancels requested signatures instead of leaving them dangling

### Fixed

- Fixed an ambiguous tenant filter that caused signature load queries to fail

## [0.226.1] - 2026-07-15

### Fixed

- Updated `golang.org/x/text` to v0.39.0 to remediate CVE-2026-56852

## [0.226.0] - 2026-07-15

### Added

- Added a v2 `ErrorBoundary` component (with companion `ErrorState` and `InlineError` primitives) to `@probo/ui` for graceful React error handling
- Webhook deliveries for `*:updated` events now include a top-level `updatedFrom` field alongside `data`, carrying a full snapshot of the entity as it was before the update (e.g. the prior membership role on `user:updated`). The field is omitted for non-update events

### Fixed

- SCIM user provisioning: deleting a user whose profile is still referenced (e.g. by signed document versions) now archives the user instead of returning a 500 and disabling the connector

## [0.225.0] - 2026-07-13

### Added

- Added managed access-review connectors for Scaleway, Yousign, Railway, and Crisp, including provider logos and website/ownership verification before a connection is established; managed connectors stay hidden until they are fully configured
- Added server-side filtering for trust center subprocessors, exposing facet-driven filters through the trust API
- Added trust center Updates list and detail pages with pagination and a mailing-list subscribe action
- Added a Risk Assessments tab to the risks list page in the console

### Changed

- The OAuth2 consent page now shows a submit loader and a redirect screen after the user grants consent

### Fixed

- Microsoft OIDC sign-in now requires verified domain ownership before linking an account

## [0.224.1] - 2026-07-09

### Fixed

- Built with Go 1.26.5 (up from 1.26.4) to pick up the standard-library security fixes for CVE-2026-42505 (ECH handshake de-anonymization) and CVE-2026-39822 (`os.Root` symlink following on Unix)

## [0.224.0] - 2026-07-09

### Added

- Added FERPA and PCI DSS framework datasets (controls and logos), selectable in the framework import selector alongside the existing frameworks

### Changed

- Consolidated ownership-grant authorization into policy: granting OWNER (via `createUser` or `updateMembership`) is now restricted to organization owners through role-scoped allow policies conditioned on the assigned role, replacing the per-resolver custom checks and the now-removed `iam:membership-role:set-owner` action. The `permission` field gained an optional generic `attributes` key/value argument so the console can refine dry-run checks (e.g. by target role) without loosening the base grants
- Renamed the `DocumentVersionSignatureFilter` `state` field to `profileState` (GraphQL) / `profile_state` (MCP) to disambiguate it from the signature `states` field, since it filters on the signatory's profile state

### Fixed

- Enforced owner-only member removal: `removeUser` (API resolver and MCP `RemoveUserTool`) now requires the owner-only `iam:membership:delete` gate instead of the weaker `iam:membership-profile:delete`, so an organization ADMIN can no longer remove members (including OWNERs)
- Fixed a hole (GHSA-22xj-f767-ppw6) where a self-provisioned trust center visitor could accept or inject audit-trail events into another visitor's NDA signature by supplying its ID; esign now verifies signature ownership against the verified session identity in the accept and record-event flows
- Confined all public trust API reads and electronic-signature operations to the requesting compliance page's tenant, closing a cross-tenant access gap where a visitor could resolve nodes, export audit-report PDFs, and read or mutate signatures belonging to another organization
- Guarded the third-party vetting agent's HTTP tools against SSRF: outbound requests now route through an SSRF-protected client that rejects loopback, private, CGNAT, link-local, and reserved addresses on every redirect hop
- Excluded documents with no published version from the trust center "Grant All" available-access list so it matches the request-all filter and Slack notification
- Hid draft and hidden documents from the trust center access-request Slack notification so it only lists documents a requester could actually be granted
- Fixed Anthropic thinking budgets so `budget_tokens` stays below `max_tokens`; thinking is now omitted when the configured budget is too small to meet Anthropic's minimum

## [0.223.3] - 2026-07-06

### Fixed

- Fixed IP address recording for NDA acceptance, document signing/approval events, and session creation behind a layer-7 proxy; affected endpoints now read the real client IP from `Forwarded` / `X-Forwarded-For` headers

## [0.223.2] - 2026-07-03

### Fixed

- Fixed a privilege-escalation gap where an organization ADMIN could mint an OWNER membership through `createUser`, bypassing the owner-only authorization enforced elsewhere; `createUser` (both the API resolver and the MCP `CreateUserTool`) now requires set-owner authorization when the requested role is OWNER

## [0.223.1] - 2026-07-03

### Fixed

- Fixed a cross-tenant IDOR where a Finding's linked Risk or a Processing Activity's Data Protection Officer could disclose another organization's data (GHSA-c74x-79w6-63jh): affected resolvers now authorize the referenced object itself instead of its parent, and the write paths validate the reference against the caller's organization

## [0.223.0] - 2026-07-02

### Added

- Served files now support HTTP range requests, enabling seeking and resumable downloads
- Document lifecycle webhook events (`document.*`, including the `version`, `signature`, and `approval` sub-events) can now be subscribed to

### Changed

- Public files are served from a stable URL so CDN infrastructure can cache them properly
- Compliance report upload limit raised to 30MB
- OAuth token and consent UIs now display friendly names for the `v1:resource-alias` scopes instead of the raw scope string

### Fixed

- Compliance page no longer treats every Slack connector as connected
- Long Mermaid flowchart labels now wrap instead of being clipped in risk assessment diagrams
- Dialogs no longer close when dismissing a nested dropdown or select whose pointer lands inside the dialog, preserving form state

## [0.222.2] - 2026-07-01

### Changed

- Bootstrap config output now omits empty fields and unset LLM provider blocks, producing cleaner generated YAML

## [0.222.1] - 2026-07-01

### Changed

- String configuration defaults now come from `probod`'s built-in values when the corresponding environment variable is unset, ensuring bootstrap-generated and directly-configured deployments use the same defaults

## [0.222.0] - 2026-06-30

### Added

- OAuth2 loopback redirect URIs now match regardless of port, enabling native OAuth clients such as Claude Code that use ephemeral ports at authorization time (RFC 8252 section 7.3)
- Access review campaigns can now be closed when entries are in a failed state

### Changed

- Third-party risk assessment vetting notes now persist the full structured breakdown (risk classification, per-category analysis, privacy and data-processing practices, AI governance, contractual clauses, professional standing) instead of a short summary only

### Fixed

- Deleting a user still referenced elsewhere (e.g. as an asset owner) now returns a 409 Conflict instead of an internal error
- GraphQL endpoint is now protected against alias-flooding DoS (GHSA-prh2-g8pv-m7p9): parser token limit, field complexity cap, LRU query cache, and field suggestion suppression added to all three GraphQL handlers

### Removed

- Access review campaigns no longer expose a framework-controls field
- `pendingEntryCount` field removed from access review campaigns

## [0.221.0] - 2026-06-30

### Added

- Compliance portal home page sections
- v2 UI component library: Text, Heading, Avatar, Button, IconButton, Badge, Callout, Dropdown menu, Card, Anchor, and Link components
- Webhook sender now runs on the kit worker framework

## [0.220.0] - 2026-06-25

### Added

- Four new API-key access-review connectors: Pylon, OpenRouter, incident.io, and Brevo

### Fixed

- Advertised scopes for OAuth2 protected resources

## [0.219.0] - 2026-06-24

### Added

- DocuSign partner OAuth2 with PKCE: full authorization-code flow with account picker; the selected account is persisted and its data-center base URI resolved from /oauth/userinfo
- Five new API-key access-review connectors: Mercury, Apollo.io, Deepgram, ClickHouse Cloud, and Langfuse
- APIKeyBasicAuthUserPass auth mode for API-key connectors supporting username:password credentials
- Read actions on all unprefixed OAuth scopes

### Changed

- Pending signature requests on a superseded version are moved to the newly published minor version, preserving the notification schedule

### Fixed

- Signature requests are now restricted to the current published version
- Heroku connection probe sends the versioned `Accept: application/vnd.heroku+json; version=3` header, correctly detecting revoked tokens
- Third party assessment header display

## [0.218.1] - 2026-06-23

### Fixed

- Bump `golang.org/x/image` to v0.43.0, remediating CVE-2026-33813 (denial of service via malformed WEBP parsing) and CVE-2026-46602 (missing tile-size limit in `x/image/tiff`)

## [0.218.0] - 2026-06-23

### Added

- RFC 6750 `WWW-Authenticate` challenges on OAuth bearer APIs (MCP, Console and Connect GraphQL, Files, OAuth2 userinfo): responses now advertise `resource_metadata`, `invalid_token`, and `insufficient_scope` with the required scopes

## [0.217.0] - 2026-06-22

### Added

- Resource aliases: trust center entries support a custom URL slug; alias field and set/remove mutations exposed in the console API, trust API, and MCP tools, with alias-based navigation in the trust center

### Changed

- Agent tool JSON schemas normalize required fields for OpenAI compatibility

### Fixed

- Alias resolver, field blur, and sitemap URL generation

## [0.216.1] - 2026-06-19

### Fixed

- OAuth2 scope registration for CIMD client identifiers was missing; CIMD actions are now correctly gated by their corresponding scopes

## [0.216.0] - 2026-06-19

### Added

- OAuth2 Client ID Metadata Document (CIMD) support: MCP connectors such as ChatGPT and Claude can now register via HTTPS client_id URLs instead of pre-provisioned GIDs; metadata documents are fetched and cached, clients are upserted on first use, and CIMD is advertised in OIDC discovery when allowed URLs are configured

## [0.215.1] - 2026-06-19

### Fixed

- Tracker-mapping no longer reprocesses sibling patterns O(N^2) times per banner; the re-enqueue now skips siblings already linked to a common third party or marked first-party, and routine mapping logs are demoted from INFO to Debug

## [0.215.0] - 2026-06-19

### Added

- Tracker pattern category is now editable from the pattern detail page (matching the table-row behaviour)
- Trackers page filter now offers the HTTP cookie source, and extension-sourced rows render a proper source badge

### Changed

- Local storage, IndexedDB, and cache-storage trackers without an expiry now display as "persistent" rather than "session"

### Fixed

- Fixed a deadlock between concurrent tracker-mapping workers processing sibling patterns on the same banner
- An HTTP server-set cookie now outranks a pre-existing detection, re-arming mapping so the pattern is identified instead of being skipped

## [0.214.0] - 2026-06-19

### Added

- OAuth2 API scope enforcement: v1:* scopes registered and advertised in OIDC discovery and protected-resource metadata, enforced in the IAM authorizer before policy evaluation
- Identity-scoped OAuth token management: users can create, list, and revoke manual bearer tokens from `/me/oauth-tokens` and the console UI
- Auditor role now includes the `v1:iam:read` scope

### Changed

- OAuth consent screen groups API scopes under an accordion

### Fixed

- MCP API now accepts OAuth bearer tokens (was previously rejected)
- Notifications are skipped for inactive users

## [0.213.0] - 2026-06-18

### Added

- Tracker-pattern catalog rows now carry a terminal attribution verdict (UNDETERMINED, THIRD_PARTY, FIRST_PARTY); FIRST_PARTY short-circuits the mapping pipeline so first-party and generic artifacts are never re-attributed

### Changed

- Document signing and approval emails are now batched per recipient by a debounced worker that sends one consolidated email and widening reminders, replacing the immediate per-document approval email and the manual "send signing notifications" action
- Deterministic vendor adoption in tracker mapping is gated behind a confidence/trust bar; lower-confidence rows are reused as hints and re-confirmed by an independent agent, and attributions must cite concrete evidence

### Fixed

- Tracker-pattern attribution is kept consistent with the vendor link: a first-party reclassification clears the stale org vendor link, and FIRST_PARTY rows are excluded from the enrichment requeue

## [0.212.0] - 2026-06-18

### Added

- OIDC authentication now opens a child session when assuming an organization

### Fixed

- OIDC organization access errors now return a 404 instead of an internal error

## [0.211.2] - 2026-06-18

### Fixed

- Exit codes

## [0.211.1] - 2026-06-18

### Fixed

- Missing OS exit code on error

## [0.211.0] - 2026-06-18

### Added

- Document delete confirmation dialog in the console

### Changed

- Error responses during a server panic are now always serialized as JSON
- Enrichment tracking unified with outcome-based status; enrichment state, attempts, and run outcome now recorded per field

### Fixed

- Enrichment re-arm and migration backfill gaps corrected
- Google Workspace access review source-name resolution no longer loops on 403 responses

## [0.210.0] - 2026-06-16

### Added

- Electronic signature on employee document signings: the signed PDF is generated and an esign record is created and accepted (capturing signer IP and user agent), mirroring the document approval flow; consent wording is now a single backend source of truth rendered consistently across the signing, approval, and NDA pages
- Structured authorization decision logging: every authorizer evaluation (allow, deny, no_match, assumption error) emits a decision line with policy id and reason using opaque ids

### Changed

- Access review campaign sources reworked: sources are first-class with a per-campaign snapshot (name, connector) taken at start time, fetch attempts recorded as an append-only log, the unused source category removed, and the deleted-source badge dropped from the campaign detail
- Access review roles render as up to three badges with a "+X more" popover instead of one long comma-separated string

### Fixed

- Access review connection status now probes all providers (static, dynamic, and custom) so bad API keys and expired OAuth tokens no longer show as Connected
- Cursor access review driver marks an account inactive when either `isRemoved` or role `removed` is set, fixing accounts reported active despite removal
- MCP profile output no longer fails schema validation when a profile has no additional email addresses

## [0.209.0] - 2026-06-12

### Added

- Common third-party enricher worker that fills the global catalog (legal name, headquarters, canonical website, compliance docs, certifications, logo, owned domains) with per-field provenance and confidence thresholds; opt-in, no-ops without an agent provider
- Tracker mapping and common-pattern enrichment agents can now open pages with a read-only headless browser (gated on Chrome endpoint) to read setters from cookie-database and policy pages
- Discovery and persistence of common third-party owned domains, used to re-resolve previously unmapped tracker patterns
- `Cache-Control` and `ETag` on `/api/files/v1/static` brand assets; startup validation of required assets

### Changed

- Tracker enrichment agent now restates source-page descriptions in its own words rather than copying them verbatim
- Common third-party enrichment agents run in parallel after website resolution; prompts rewritten in role/task/instructions XML style with a calibrated confidence rubric
- Oversized logo responses are rejected instead of truncated; ownership substring matching tightened with a length-ratio guard; per-agent error text sanitized and bounded before persistence

### Fixed

- Activate-login path
- `find_links_matching` browser tool double-encoded its pattern, starving any agent using it
- Worker confidence threshold of `0` no longer dropped by Helm falsy-numeric truthiness

## [0.208.1] - 2026-06-12

### Fixed

- Trust center file creation in console
- S3 filename header escaping

## [0.208.0] - 2026-06-11

### Added

- `active` status field on access entries
- Import action for catalog vendors in trackers; `importThirdPartyFromCommon`
  mutation to pull a catalog vendor into an org
- Catalog vendors surfaced in tracker policy documents
- File download URLs for console file fields

### Changed

- Trust and MCP connector logos now use the File type
- Third parties deduplicated by name; unique index enforced per org
- Tracker mapping no longer auto-creates org third parties; explicit import required
- Tracker row and category select restyled; move-to-category confirm dialog removed
- Tracker mapping restored to link existing patterns; "create only" mode removed
- Document major version publishing requires explicit `approver_ids`
- References updated to probo.com

### Removed

- Third-party disambiguation agent and automatic matching removed

### Fixed

- DNS TXT lookup retried over TCP on truncated UDP response

## [0.207.0] - 2026-06-10

### Added

- Neon access-review connector (organization members via Neon API, API-key auth)
- Render access-review connector (workspace members, API-key + Workspace ID)
- Qovery access-review connector (organization members, configurable `Token` Authorization scheme)
- API-key connector providers can now declare a custom Authorization token scheme (defaults to `Bearer`)
- `regulationSource` (`DETECTED`/`DEFAULT`) on cookie consent records, with GDPR/OPT_IN applied as the safe default when geolocation does not resolve a known regulation
- `--keyword` scoping on the banner tracker-reset operator path: rebuilds only patterns whose pattern or display name contains the substring
- `parent_third_party_id` foreign key on third parties for arbitrary sub-third-party nesting depth; `level` (int, 1+) replaces the `firstLevel` boolean

### Changed

- Tracker-mapping agent ignores cookie-database/consent-directory operators (Cookipedia, cookiedatabase.org, CookieServe, …) as vendor attributions; CMP own-cookie attributions (OneTrust, Cookiebot, …) still survive
- Tracker-mapping agent ignores own-domain tracker attributions (patterns embedding the scanned site's own eTLD+1) with a deterministic backstop
- Relinking a common tracker pattern to a different third party now updates the confidence on linked org patterns
- Rename console label "Detected Count" to "Distinct Trackers Detected"
- `proboctl common-tracker-pattern reenrich` now accepts catalog-wide filters with no selection anchor (e.g. `--without-description` re-enriches every pattern lacking a description)

### Fixed

- Null out stale `initiator_url`/`initiator_domain` rows on `detected_trackers` that point at the @probo/cookie-banner bundle, so genuine third-party initiators repopulate on next detection
- Cookie-database denylist now matches domain and URL forms (e.g. `cookiedatabase.org`, `https://www.cookiepedia.co.uk/list`), not just bare brand names

### Removed

- `createThirdPartyThirdPartyMapping` and `deleteThirdPartyThirdPartyMapping` mutations and MCP tools; create a child third party by passing `parentThirdPartyId` on `createThirdParty`

## [0.206.0] - 2026-06-09

### Added

- `RiskAssessmentBoundary` first-class entity to group nodes within a risk assessment scope, with self-nesting parent boundary, scope-membership validation, nested-subgraph Mermaid rendering, and dedicated IAM actions
- `regenerateCookieBannerTrackerPolicy` mutation/MCP tool to re-trigger tracker policy generation on a banner that already has a published version, gated by a dedicated `regenerate-policy` action
- Better Stack access-review connector (Uptime API team members + pending invitations)
- SigNoz access-review connector (organization members, region/tenant or self-hosted base URL)
- `commonTrackerPatternId` field on `TrackerPattern` to indicate whether a pattern is linked to the global common-tracker catalog
- Files API: public endpoint `GET /api/files/v1/public/{fileID}` (unauthenticated, public files only) and private endpoint `GET /api/files/v1/{fileID}` (session/API key/OAuth2, `core:file:get` enforced); IAM and not-found errors both return 404
- Static brand assets served via `/api/files/v1/static` instead of S3

### Changed

- Connector provider infos promoted from `Organization.connectorProviderInfos` to a root-level `accessReviewDrivers` query, listable by any authenticated identity
- Tracker-mapping, common-pattern enrichment, and third-party disambiguation agents each get their own config (own timeout, own max-turns, own optional provider slot, with fallback to the tracker-mapping slot when unset)
- Console: tracker pages now surface common-tracker/third-party links with a "common" badge and updated pattern properties display

### Fixed

- Cookie tracker pattern analysis: removed unused sync re-enrich path and tightened reset/remap scoping

### Removed

- `ActionFileDownloadUrl` (replaced by `ActionFileGet`) and the standalone `pkg/filesign` package (folded into `file.Service`)

## [0.205.0] - 2026-06-08

### Added

- `submitAgentRunApproval` mutation to merge human approval decisions into an interrupted agent run and resume it
- Suspendable agent-tool subtrees: nested agent runs can now checkpoint and restore across multi-level tool calls

### Changed

- Agent-run worker no longer relies on leases and heartbeats: a graceful suspend returns the run to `PENDING`, an approval interruption parks it in `AWAITING_APPROVAL`, and crashed runs are left `RUNNING` for manual recovery
- AWS credentials now resolve through the full standard AWS SDK credential chain

### Fixed

- Auditors can now read the organization context and see the Context page in the console
- NDA upload now correctly sets the organization ID
- Logo updates no longer wipe unspecified fields on partial update
- Cookie tracker pattern analysis now splits on `:` and `.` so UUID-bearing keys collapse to a single template

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
