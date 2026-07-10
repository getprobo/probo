---
name: open-source-compliance
description: This skill should be used when the user wants to assess, track, or automate open-source compliance work in Probo — vendor reviews, control mapping, evidence collection, risk registers, or policy workflows using Probo MCP tools.
---

# Open-source compliance with Probo

Use the Probo MCP server bundled with this plugin to read and write GRC data.
Do not guess entity IDs or organization scope — discover them with MCP tools
first.

## Before you start

1. Confirm the Probo MCP server is connected. If not, ask the user to run
   `/mcp` or `claude mcp login probo` to complete the OAuth 2.0 sign-in.
2. Identify the target organization. List organizations if the user did not
   provide one.
3. Prefer MCP tools over manual API calls. The Probo MCP API mirrors the
   platform's GraphQL surface.

## Common workflows

### Vendor / third-party review

1. List or search third parties for the organization.
2. Pull existing risk assessments and contacts.
3. Record findings and update assessment status through MCP mutations.
4. Summarize residual risk and recommended follow-ups for the user.

### Control and obligation tracking

1. List controls, measures, or obligations relevant to the user's question.
2. Link evidence (documents, audits) where appropriate.
3. Report gaps between required and implemented controls.

### Evidence and documentation

1. Locate the relevant document or audit in Probo.
2. Fetch version history or published versions as needed.
3. Draft updates; use MCP upload tools when the user asks to attach files.

## Output expectations

- Cite Probo entity IDs (GIDs) for anything you create or update.
- Separate facts pulled from Probo from your analysis.
- Flag missing data instead of inventing compliance status.
- Keep recommendations actionable and mapped to Probo entities where possible.

## References

See `references/workflows.md` for extended workflow notes.
