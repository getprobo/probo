# Changelog

All notable changes to the Probo Helm chart will be documented in this file.

## Unreleased

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
