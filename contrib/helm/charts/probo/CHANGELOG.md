# Changelog

All notable changes to the Probo Helm chart will be documented in this file.

## Unreleased

## [0.1.0] - 2026-05-25

### Added

- Initial Helm chart for deploying Probo (`probod v0.192.0`) with configurable PostgreSQL, SeaweedFS object storage, ingress, SAML, SMTP, and connector OAuth credentials

### Changed

- Container images are pulled from the `artifact.probo.inc` OCI registry
- Firecrawl API key is now configured under `agents.tools.firecrawl.apiKey`; `FIRECRAWL_ENDPOINT` is no longer configurable
- Access-review connectors now require `clientSecret`

### Removed

- SearXNG search backend — Firecrawl is the only supported web search provider
