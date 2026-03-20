# Release

This guide describes how to cut a new release. The outcome is a single
commit on `main` plus a Git tag; CI handles everything else (binaries,
Docker images, npm packages, signatures).

## Steps

### 1. List changes since the last release

Find the latest tag and review every commit since then:

```shell
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

### 2. Write the changelog entry

Create a new version section in `CHANGELOG.md` from those commits.
Keep the empty `## Unreleased` heading above it.

**Categorize** entries under Keep-a-Changelog sections:

| Section       | Use for                                        |
|---------------|------------------------------------------------|
| `### Added`   | New features, new commands, new endpoints      |
| `### Changed` | Behavioral changes, refactors visible to users |
| `### Fixed`   | Bug fixes                                      |
| `### Removed` | Removed features or deprecated code            |

**Skip** commits that are not user-facing:

- Style / formatting (`Style`, `Run go fmt/fix`)
- CI-only changes (`Add reviewdog`, `Cache Go modules`)
- Internal refactors (`Move X to contrib/claude`, `Remove deadcode`)
- Documentation-only changes
- Release commits (`Release v…`)

**Only list fixes for pre-existing bugs.** If a "fix" commit repairs something
introduced by another commit in the same release cycle, do NOT list it as a
separate fix. To verify, check whether the affected file or feature existed
at the previous tag:

```shell
git ls-tree <previous-tag> -- path/to/file
```

If the file did not exist at the previous tag, the fix is part of the new
feature and should not appear in `### Fixed`.

**Summarize** related commits into a single line when appropriate.
For example a series of `Add proboctl X commands` commits becomes
`Add CLI`.

Format: `## [X.Y.Z] - YYYY-MM-DD` (today's date).

Example result:

```markdown
## Unreleased

## [0.144.0] - 2026-03-17

### Added

- Add document viewer with 404 handling for trust center

## [0.143.0] - 2026-03-16
```

### 3. Decide the version bump

The project is in the **0.x** series. Never bump MAJOR.

- Bug fixes only → bump **PATCH**
- New features or non-breaking changes → bump **MINOR**

### 4. Bump version in `GNUmakefile`

Update the `VERSION` variable at the top of `GNUmakefile`:

```makefile
VERSION=	0.144.0
```

### 5. Review with the user

Before committing, show the user the full `CHANGELOG.md` entry and
the new `VERSION` value. Ask them to confirm everything looks good.
Only proceed once they approve.

### 6. Create the release commit

Stage only `CHANGELOG.md` and `GNUmakefile`. The commit message
**must** follow this exact format:

```
Release v<VERSION>
```

No body is needed.

### 7. Create the tag

Tag the release commit with an **annotated** tag. The tag **must**
match `v<VERSION>`:

```shell
git tag -a v<VERSION> -m "v<VERSION>"
```

### 8. Push

Push both the commit and the tag:

```shell
git push origin main --follow-tags
```

CI (`.github/workflows/release.yaml`) triggers on `v*` tags and takes
care of:

- Building binaries via GoReleaser (probod, probod-bootstrap, prb)
- Publishing multi-arch Docker images to `ghcr.io/getprobo/probo`
- Publishing the npm package `@probo/n8n-nodes-probo`
- Generating SBOMs, attestations, and Cosign signatures

## Checklist

1. [ ] Reviewed commits since last tag
2. [ ] `CHANGELOG.md` — new version section with categorized entries
3. [ ] `GNUmakefile` — `VERSION` bumped
4. [ ] User confirmed changelog and version look good
5. [ ] Commit message is `Release v<VERSION>`
6. [ ] Annotated tag `v<VERSION>` on the release commit
7. [ ] Push commit and tag
