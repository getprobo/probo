# Open-source compliance workflows

This plugin wraps Probo's MCP API for Claude Code. Probo is a self-hostable
GRC platform; its MCP server exposes tools for third parties, controls,
obligations, risks, documents, audits, access reviews, and more.

## Authentication

The plugin `.mcp.json` expects:

| Variable | Purpose |
| --- | --- |
| `PROBO_BASE_URL` | Probo instance root URL (no trailing slash) |
| `PROBO_API_TOKEN` | Personal API key or OAuth access token |

MCP endpoint: `${PROBO_BASE_URL}/mcp/v1`

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
