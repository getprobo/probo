---
name: commit
description: Write Probo commit messages following the seven-rules style and signed-off-by conventions. Use when committing, amending commits, or drafting commit messages for this repository.
---

# Probo commit messages

Follow the seven rules of a great Git commit message:

1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain *what* and *why* vs. *how*

The subject must complete: "If applied, this commit will …"

## Do not use Conventional Commits

This repository does **not** use `type(scope): summary` prefixes. Write a plain
imperative subject instead.

```text
# GOOD
Disconnect observer in cookie-banner load() error path

# BAD
fix(cookie-banner): disconnect observer in load() error path
```

## Signing

Every commit must be signed with both flags:

- `-s` — `Signed-off-by` trailer (DCO)
- `-S` — GPG/SSH signature

```bash
git commit -s -S -m "$(cat <<'EOF'
Subject line in imperative mood

Body explaining what and why, wrapped at 72 chars.
EOF
)"
```

The commit author must be the human responsible for the change. Do not add
`Co-Authored-By` trailers crediting bots.

Verify with `git log -1 --show-signature`.

## Examples

Single-line when self-explanatory:

```
Fix typo in third-party assessment prompt
```

With body when the why matters:

```
Add third-party assessment agent for third-party reviews

The existing changelog generator only covers internal changes.
This introduces a dedicated agent that evaluates third parties
against our compliance criteria, producing a structured risk report.
```

For the full guide, see `references/commit.md` in this skill directory.
