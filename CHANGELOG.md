# Changelog Index

Each release track now keeps its own changelog. The history below 0.173.0 of the unified monorepo changelog is preserved in [CHANGELOG.archive.md](CHANGELOG.archive.md).

## Per-track changelogs

- `prb` (CLI) — [cmd/prb/CHANGELOG.md](cmd/prb/CHANGELOG.md)
- `probod` (server, including bundled `@probo/console`, `@probo/trust`, `@probo/ui`) — [cmd/probod/CHANGELOG.md](cmd/probod/CHANGELOG.md)
- `probod-bootstrap` — [cmd/probod-bootstrap/CHANGELOG.md](cmd/probod-bootstrap/CHANGELOG.md)
- `@probo/n8n-nodes-probo` — [packages/n8n-node/CHANGELOG.md](packages/n8n-node/CHANGELOG.md)
- `@probo/cookie-banner` — [packages/cookie-banner/CHANGELOG.md](packages/cookie-banner/CHANGELOG.md)
- Helm chart (`probo`) — [contrib/helm/charts/probo/CHANGELOG.md](contrib/helm/charts/probo/CHANGELOG.md)

## Tag scheme

Each track is published under its own annotated tag of the form `<track>/v<version>`:

- `prb/vX.Y.Z`
- `probod/vX.Y.Z` (also tags the `artifact.probo.inc/probo/probo` Docker image)
- `probod-bootstrap/vX.Y.Z`
- `helm/vX.Y.Z` (also publishes the chart to `oci://artifact.probo.inc/probo/probo`)
- `@probo/n8n-nodes-probo/vX.Y.Z`
- `@probo/cookie-banner/vX.Y.Z`

The legacy `vX.Y.Z` tag scheme is retired at `v0.173.0` (2026-04-24).
