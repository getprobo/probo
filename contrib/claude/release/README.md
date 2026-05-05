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

After choosing the track and reviewing its commits:

1. Bump the version in the track's source of truth (see the per-track
   entrypoint).
2. Write the changelog entry in the track's `CHANGELOG.md`.
3. For npm tracks, run the workspace `build` script after the version
   bump (see the per-track entrypoint for why).
4. Show the user the proposed `CHANGELOG.md` diff and the new version.
   Wait for confirmation.
5. Commit only the files modified by the version bump and changelog
   edit. Subject: `Release <pkg>/v<version>`. No body.
6. Annotated tag: `git tag -a <pkg>/v<version> -m "<pkg>/v<version>"`.
7. Push: `git push origin main --follow-tags`.

CI handles the rest: binary builds, npm publish, Docker image, Homebrew
formula, SBOMs, attestations, GitHub Release.

## Checklist (every track)

1. [ ] Pulled latest `main`
2. [ ] Confirmed the track has user-facing commits since its last tag
3. [ ] Reviewed track-specific commits
4. [ ] Track CHANGELOG entry written, categorized, summarized
5. [ ] Version bumped in the track's source of truth
6. [ ] (npm tracks) Workspace `build` script run successfully after bump
7. [ ] User confirmed changelog and version
8. [ ] Commit message is `Release <pkg>/v<version>`
9. [ ] Annotated tag `<pkg>/v<version>` on the release commit
10. [ ] Pushed commit and tag
