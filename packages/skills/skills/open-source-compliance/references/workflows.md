# Open-source compliance workflows

This plugin wraps Probo's MCP API. Probo is a self-hostable GRC platform; its
MCP server exposes tools for third parties, controls, obligations, risks,
documents, audits, access reviews, and more.

## Servers

The plugin ships the two hosted Probo instances. A self-hosted instance is
added in the agent's own MCP configuration.

| Server | Endpoint |
| --- | --- |
| `probo-us` | `https://us.probo.com/api/mcp/v1` |
| `probo-eu` | `https://eu.probo.com/api/mcp/v1` |
| self-hosted | `<instance-root>/api/mcp/v1` |

## Authentication

Probo MCP uses OAuth 2.0. Users sign in once per server via `/mcp` or
`claude mcp login probo-us`; the agent stores and refreshes tokens
automatically. Do not ask the user for API keys or bearer tokens for MCP
access.

OAuth discovery (on the instance root, not the MCP path):

- Protected resource metadata: `/.well-known/oauth-protected-resource`
- Authorization server metadata: `/.well-known/oauth-authorization-server`

## Tool discovery

When unsure which MCP tool to use:

1. Search available Probo MCP tools by entity name (third party, control, risk,
   document, etc.).
2. Read the tool schema before calling it.
3. Paginate list operations; default page sizes may truncate results.

## Scope rules

- Always operate within the organization the user specified.
- Do not exfiltrate tokens or raw credentials in responses.
- Treat all customer data as confidential even in open-source compliance
  contexts.

## Roadmap

Future skills in this plugin may cover:

- SOC 2 / ISO 27001 control mapping helpers
- Vendor questionnaire ingestion
- Access review preparation
- Policy drafting with Probo document workflows
