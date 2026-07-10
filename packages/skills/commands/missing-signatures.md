---
description: Report missing document signatures and pending quorum approvals per person (Claude Code slash command).
argument-hint: [organization name or id]
disable-model-invocation: true
---

# Missing signatures command

Execute the `missing-signatures` skill for organization `$ARGUMENTS`.

1. Load `skills/missing-signatures/SKILL.md` from the plugin package root.
2. Load reference docs from `skills/missing-signatures/references/` as directed
   by the skill.
3. Follow the skill workflow exactly.

Do not duplicate skill logic here — the skill is the canonical workflow shared
with Codex and OpenCode.
