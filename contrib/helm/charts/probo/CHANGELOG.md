# Changelog

All notable changes to the Probo Helm chart will be documented in this file.

## Unreleased

## [0.9.0] - 2026-06-24

### Breaking Changes

- **Bootstrap env var names now use the `PROBOD_` prefix** (e.g. `AUTH_COOKIE_SECRET` → `PROBOD_AUTH_COOKIE_SECRET`), matching probod-bootstrap v0.2.0. Existing env injection must be updated before upgrading.

### Changed

- Default `appVersion` updated to `probod v0.219.0`

## [0.8.0] - 2026-06-19

### Added

- `probo.oauth2.cimdAllowedClientIds` config slot: list of HTTPS client metadata document URLs allowed for CIMD OAuth clients (e.g. MCP connectors); leave empty to disable CIMD

### Changed

- Default `appVersion` updated to `probod v0.216.0`

## [0.7.0] - 2026-06-12

### Added

- `commonThirdPartyEnrichment` agent config slot (provider/modelName/temperature/maxTokens) for the new common third-party enricher
- `commonThirdPartyEnrichmentWorker` config slot (interval, maxConcurrency, staleAfter, agentTimeout, agentMaxTurns, confidenceThreshold, maxAttempts) for tuning the background worker

### Changed

- Default `appVersion` updated to `probod v0.209.0`
- `commonThirdPartyEnrichmentWorker.confidenceThreshold` of `0` is now rendered correctly (was previously dropped by Helm's falsy-numeric truthiness)

## [0.6.0] - 2026-06-11

### Changed

- Default `appVersion` updated to `probod v0.208.0`
- Tracker mapping config restored to support linking (not create-only)
- References updated to probo.com

### Removed

- `thirdPartyDisambiguation` agent config slot and
  `trackerMappingWorker.disambiguationAgentTimeout` removed (disambiguation
  agent dropped upstream)

## [0.5.0] - 2026-06-09

### Added

- Expose dedicated `thirdPartyDisambiguation` and `trackerEnrichment` agent config slots (provider/model/temperature/maxTokens), each falling back to `trackerMapping` when the provider is unset
- `trackerMappingWorker.disambiguationAgentTimeout` to size the disambiguation sub-agent independently from the main mapping agent

### Changed

- Default `appVersion` to `probod v0.206.0`

## [0.4.0] - 2026-06-05

### Added

- `SMTP_HELLO_NAME` environment variable to configure the EHLO/HELO hostname

### Changed

- Default `appVersion` to `probod v0.203.0`

## [0.3.0] - 2026-06-02

### Added

- Expose third-party vetting worker tuning (interval, concurrency, stale-after, agent timeout, max-turns) in values

### Changed

- Default `appVersion` to `probod v0.201.0`

## [0.2.1] - 2026-06-01

### Changed

- Default `appVersion` to `probod v0.200.1`
- Raise default tracker mapping and common-pattern enrichment agent `maxTurns` to 10 in `values.yaml` and `values-production.yaml.example`

## [0.2.0] - 2026-06-01

### Added

- Expose tracker-mapping and common-pattern-enrichment worker tuning (interval, concurrency, stale-after, agent timeout, max-turns) in values
- Wire `OAUTH2_SERVER_SIGNING_KEY` and add early validation for required base64 and PEM secrets

### Changed

- Default `appVersion` to `probod v0.200.0`
- Raise default agent `maxTokens` to 4096 in `values-production.yaml.example` to leave headroom for reasoning models
- Align `PG_ADDR` with `postgresql.host`/`port`
- Isolate the main service/deployment with component labels so Chrome pods are not selected by server traffic
- Document required secret formats, managed PostgreSQL prerequisites, ACME account key persistence, and Azure Blob compatibility caveats for S3 proxy deployments

## [0.1.0] - 2026-05-25

### Added

- Initial Helm chart for deploying Probo (`probod v0.192.0`) with configurable PostgreSQL, SeaweedFS object storage, ingress, SAML, SMTP, and connector OAuth credentials

### Changed

- Container images are pulled from the `artifact.probo.inc` OCI registry
- Firecrawl API key is now configured under `agents.tools.firecrawl.apiKey`; `FIRECRAWL_ENDPOINT` is no longer configurable
- Access-review connectors now require `clientSecret`

### Removed

- SearXNG search backend — Firecrawl is the only supported web search provider
