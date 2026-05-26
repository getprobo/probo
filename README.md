# Probo

[![License](https://img.shields.io/github/license/getprobo/probo)](LICENSE)
[![Build](https://img.shields.io/github/actions/workflow/status/getprobo/probo/make.yaml)](https://github.com/getprobo/probo/actions)
[![Discord](https://img.shields.io/discord/1326589224811757568?color=7289da&label=Discord&logo=discord&logoColor=ffffff)](https://discord.gg/8qfdJYfvpY)

**Open-source GRC platform for engineers.**

Probo is a self-hostable governance, risk, and compliance (GRC) platform built for engineering and security teams. It covers the full GRC lifecycle: risk identification, control tracking, vendor risk, data privacy, access reviews, audit programs, and document approval workflows. Every entity is accessible through a web console, a CLI, a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) API, and a GraphQL API, so you can automate compliance work from code, scripts, or any LLM agent.

## Why Probo?

- **AI-native by design.** 270+ MCP tools expose every entity and operation. Any MCP-compatible LLM agent can read and write your GRC data, draft policies, run risk assessments, and generate evidence packs.
- **Full GRC coverage.** Risk management, controls, vendor risk, data privacy (DPIA/TIA), access reviews, audit programs.
- **Multiple interfaces.** Web console, `prb` CLI (44+ command groups), MCP API, GraphQL, and an n8n community node for no-code automation.
- **Open source and self-hostable.** ISC licensed. Run it on your own infrastructure with Docker.
- **Audit-ready.** Policy-based RBAC, immutable audit logs, electronic document sign-off workflows, and evidence chains.

## Capabilities

| Domain | Features |
|--------|----------|
| **Risk Management** | Risk register, inherent/residual scoring, treatment strategies (mitigate, accept, avoid, transfer), threat-based risk assessments |
| **Controls & Frameworks** | Control library with maturity levels, custom framework import/export, Statement of Applicability (SoA) |
| **Vendor / Third-Party Risk** | Vendor inventory, automated website risk assessment, DPA/BAA tracking, subprocessor discovery |
| **Data Privacy** | DPIA, Transfer Impact Assessments, processing activity records, data inventory, rights requests (SAR/erasure) |
| **Access Reviews** | Campaign management, per-entry access decisions, integration with SaaS, cloud infra, and source code sources |
| **Audit Programs** | Audit scoping, control mapping, finding tracking, report generation |
| **Evidence & Measures** | Evidence collection (files and URLs), implementation state tracking, task assignment |
| **Document Management** | Versioned documents, approval quorums, electronic signatures, PDF export, bulk operations |
| **Compliance Page** | Public compliance portal, NDA management, certification publishing, custom domain support |
| **Cookie & Consent** | Cookie banner management, tracker detection, consent records |

## Interfaces

### Web console

The primary interface for day-to-day GRC work. Runs at `http://localhost:8080` in development.

### CLI (`prb`)

A fully-featured command-line client for scripting, automation, and CI/CD integration. Covers all 44+ resource types available in the web console.

```sh
# Authenticate
prb auth login

# List open risks
prb risk list

# Create a measure and link evidence
prb measure create --name "MFA enforced on all production systems"
prb evidence create --measure <id> --file screenshot.png

# Manage vendor compliance
prb thirdpartymgmt vendor list
prb thirdpartymgmt risk-assessment create --vendor <id>
```

Run `prb help` for the full command reference.

### MCP API

Probo exposes 270+ [MCP](https://modelcontextprotocol.io) tools covering every entity and operation in the platform. Any MCP-compatible LLM agent (Claude, Cursor, Continue, and others) can connect directly and interact with your compliance data.

The full MCP specification is at [`pkg/server/api/mcp/v1/specification.yaml`](pkg/server/api/mcp/v1/specification.yaml).

### n8n node

The [`@probo/n8n-nodes-probo`](packages/n8n-node/) community node brings Probo into n8n workflows for no-code automation of compliance tasks over the GraphQL API.

## Quick Start

**Prerequisites**

| Tool | Version |
|------|---------|
| Go | 1.26+ |
| Node.js | 22+ |
| Docker | latest |
| mkcert | latest |

**Steps**

```sh
# 1. Clone with submodules
git clone --recurse-submodules https://github.com/getprobo/probo.git
cd probo

# 2. Install dependencies
go mod download
npm ci

# 3. Start infrastructure services (PostgreSQL, object storage, etc.)
make stack-up

# 4. Build
make build

# 5. Generate the local dev config
make dev-config

# 6. Run the server
bin/probod -cfg-file cfg/dev.yaml
```

The web console is available at `http://localhost:8080`.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development environment walkthrough, including the frontend dev server and code generation steps.

## Tech Stack

| Layer | Technologies |
|-------|-------------|
| Backend | Go, PostgreSQL |
| API | GraphQL, MCP |
| Frontend | React, TypeScript, Relay, TailwindCSS |
| Infrastructure | Docker, GitHub Actions |
| Observability | OpenTelemetry, Grafana, Prometheus, Loki, Tempo |

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. All commits require a Developer Certificate of Origin (DCO) sign-off (`git commit -s`). No CLA required.

To report a security vulnerability, email [security@getprobo.com](mailto:security@getprobo.com) rather than opening a public issue. See [SECURITY.md](SECURITY.md) for the full disclosure policy.

## Community

- [Discord](https://discord.gg/8qfdJYfvpY) - Get help, share feedback, and talk to the team
- [Documentation](https://www.getprobo.com/docs)
- [Blog](https://www.getprobo.com/blog)
- [Twitter / X](https://twitter.com/getprobo)
- [LinkedIn](https://www.linkedin.com/company/getprobo)
- [Website](https://www.getprobo.com)

## License

Probo is [ISC licensed](LICENSE).
