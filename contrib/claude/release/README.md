# Release

The repository ships five independently-versioned tracks. Each has its own
version source, its own `CHANGELOG.md`, its own tag pattern, and its own
release workflow. Cutting a release means: bump the version, write a
changelog entry, commit, tag, push.

| Track                   | Tag pattern                    | Entrypoint                       |
| ----------------------- | ------------------------------ | -------------------------------- |
| CLI (`prb`)             | `prb/v*`                       | [prb.md](./prb.md)               |
| Server (`probod` group) | `probod/v*`                    | [probod.md](./probod.md)         |
| `probod-bootstrap`      | `probod-bootstrap/v*`          | [probod-bootstrap.md](./probod-bootstrap.md) |
| `@probo/n8n-nodes-probo` | `@probo/n8n-nodes-probo/v*`   | [n8n-nodes-probo.md](./n8n-nodes-probo.md) |
| `@probo/cookie-banner`  | `@probo/cookie-banner/v*`      | [cookie-banner.md](./cookie-banner.md) |

When the user asks for a release **without specifying a track**, follow
[Step 1](#1-decide-which-tracks-to-release) below to detect which tracks
have user-facing changes since their last tag, then ask the user which of
those tracks to release. **Only release tracks that actually have
user-facing changes.** Never release a track that has no commits since its
last tag.

When the user asks for a release **for a specific track** (e.g. "release
the CLI", "release probod"), open the corresponding entrypoint above and
follow it.

Versions are SemVer in the **0.x** series. Never bump MAJOR.
Bug fixes only -> bump PATCH; new features or non-breaking changes -> bump
MINOR.

## 1. Decide which tracks to release

Before any release, identify which tracks have user-facing commits since
their last tag. A track with zero commits, or only non-user-facing
commits (style, CI, internal refactors, doc-only, release commits) must
**not** be released.

Run this from a clean `main`:

```shell
git checkout main && git pull origin main
```

Then for each track, list commits since its last tag, scoped to that
track's paths:

```shell
# prb
git log $(git describe --tags --abbrev=0 --match='prb/v*')..HEAD --oneline \
  -- cmd/prb pkg/cli pkg/cmd

# probod (server group: probod + console + trust + ui)
git log $(git describe --tags --abbrev=0 --match='probod/v*')..HEAD --oneline \
  -- cmd/probod apps/console apps/trust packages/ui pkg

# probod-bootstrap
git log $(git describe --tags --abbrev=0 --match='probod-bootstrap/v*')..HEAD --oneline \
  -- cmd/probod-bootstrap

# @probo/n8n-nodes-probo
git log $(git describe --tags --abbrev=0 --match='@probo/n8n-nodes-probo/v*')..HEAD --oneline \
  -- packages/n8n-node

# @probo/cookie-banner
git log $(git describe --tags --abbrev=0 --match='@probo/cookie-banner/v*')..HEAD --oneline \
  -- packages/cookie-banner
```

If a track returns no commits, skip it. If all commits for a track are
non-user-facing, skip it (and tell the user). For each remaining track,
proceed with its entrypoint.

## 2. Writing a changelog entry

Categorize entries under Keep-a-Changelog sections in the relevant track's
`CHANGELOG.md`:

| Section       | Use for                                        |
| ------------- | ---------------------------------------------- |
| `### Added`   | New features, new commands, new endpoints      |
| `### Changed` | Behavioral changes, refactors visible to users |
| `### Fixed`   | Bug fixes                                      |
| `### Removed` | Removed features or deprecated code            |

**Skip** non-user-facing commits (style/formatting, CI-only, internal
refactors, doc-only, release commits).

**Only list fixes for pre-existing bugs.** If a "fix" commit repairs
something introduced earlier in the same release cycle, do NOT list it as
a separate fix.

**Summarize** related commits into a single line when appropriate.

Format: `## [X.Y.Z] - YYYY-MM-DD` (today's date). Always keep an
`## Unreleased` heading above the latest version.

## 3. Common steps (every track)

Open the per-track entrypoint and run its **Detect commits** command.
If the track has no user-facing commits, stop. Otherwise:

1. Decide the version bump (PATCH for fixes, MINOR for features).
2. Bump the version using the track's **Version bump** method.
3. If the track lists a **Build** command, run it now to catch compile
   errors before tagging.
4. Write the changelog entry in the track's **Changelog** file following
   the rules in [section 2](#2-writing-a-changelog-entry).
5. Show the user the proposed changelog diff and the new version. Wait
   for confirmation.
6. Stage only the files listed in the track's **Files to stage**.
   Commit subject: `Release <pkg>/v<version>`. No body.
7. Annotated tag: `git tag -a <pkg>/v<version> -m "<pkg>/v<version>"`.
8. Push the commit: `git push origin main`.
9. Push the tag **separately**: `git push origin <pkg>/v<version>`.

Each tag must be pushed in its own `git push` command so the
corresponding GitHub Actions workflow fires reliably. Pushing multiple
tags at once (e.g. `--follow-tags`) can silently skip workflow triggers.

When releasing multiple tracks in one session, push each track's tag
immediately after creating it, before moving on to the next track.

## Checklist (every track)

1. [ ] Pulled latest `main`
2. [ ] Track has user-facing commits since its last tag
3. [ ] Version bumped per track's **Version bump** method
4. [ ] (npm tracks) **Build** command succeeded
5. [ ] Changelog entry written, categorized, summarized
6. [ ] User confirmed changelog and version
7. [ ] Commit: `Release <pkg>/v<version>`
8. [ ] Annotated tag `<pkg>/v<version>`
9. [ ] Pushed commit to `main`
10. [ ] Pushed tag independently
