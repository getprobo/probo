---
description: Run a semi-automated access review on a Probo campaign (Claude Code slash command).
argument-hint: [campaign name or id]
disable-model-invocation: true
---

# Access review command

Execute the `access-review` skill for campaign `$ARGUMENTS`.

1. Load `skills/access-review/SKILL.md` from the plugin package root.
2. Load reference docs from `skills/access-review/references/` as directed by
   the skill.
3. Follow the skill workflow exactly.

Do not duplicate skill logic here — the skill is the canonical workflow shared
with Codex and OpenCode.
