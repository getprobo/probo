# Commit Messages

Follow the [seven rules of a great Git commit message](https://cbea.ms/git-commit/):

1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain *what* and *why* vs. *how*

The subject line should complete the sentence: "If applied, this commit will *your subject line here*".

```
Add third-party assessment agent for third-party reviews

The existing changelog generator only covers internal changes.
This introduces a dedicated agent that evaluates third-party
thirdParties against our compliance criteria, producing a structured
risk report.
```

Not every commit needs a body -- a single line is fine when the change is self-explanatory:

```
Fix typo in third-party assessment prompt
```

## No Conventional Commits

This repository **does not** use Conventional Commits (`type(scope): summary`). The seven-rules style above is the only accepted format. Existing Conventional-Commits-style messages in the history are drift and must not be used as precedent.

```text
# GOOD
Disconnect observer in cookie-banner load() error path

# BAD -- Conventional Commits prefix
fix(cookie-banner): disconnect observer in load() error path
```

The project does not consume the `type(scope):` prefix for any tooling (no changelog generator, no semantic-release, no commit-lint), so the prefix only adds noise. If a future need for machine-readable commit types arises, raise it in a separate change that updates this document first.

## Signing and Authorship

All commits **must** be signed (`-s -S`):

- `-s` adds a `Signed-off-by` trailer (DCO).
- `-S` creates a GPG/SSH signature.

The commit author must be the **human responsible for the change**, not the AI assistant. Do **not** add `Co-Authored-By` trailers crediting bots.
