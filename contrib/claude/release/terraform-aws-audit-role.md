# Release Terraform (`aws-audit-role`)

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `terraform-aws-audit-role/v*`
- **Version source**: `contrib/terraform/aws-audit-role/VERSION` (single
  `X.Y.Z` line)
- **Version bump**: Edit `contrib/terraform/aws-audit-role/VERSION` directly
- **Changelog**: `contrib/terraform/aws-audit-role/CHANGELOG.md`
- **Files to stage**: `contrib/terraform/aws-audit-role/VERSION`,
  `contrib/terraform/aws-audit-role/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-terraform-aws-audit-role.yaml`
- **Path filter**: `contrib/terraform/aws-audit-role`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='terraform-aws-audit-role/v*' 2>/dev/null)..HEAD --oneline \
  -- contrib/terraform/aws-audit-role
```

If `git describe` fails (no tag yet), the empty range lists every commit
on the path. If empty or non-user-facing only, do not release this track.

## Notes

CI checks that `VERSION` matches the tag, runs `terraform fmt -check`,
creates a GitHub Release on this repository, then copies the module into
`getprobo/terraform-aws-audit-role` and tags `v<version>` there. The
module-repo commit is created with GitHub's `createCommitOnBranch`
mutation (same as the Homebrew tap) so it satisfies the org signed-commit
rule; a local `git commit` plus `git push` would be rejected.

Consumers can pin either source:

```hcl
# Dedicated module repo (filled by the release workflow)
source = "github.com/getprobo/terraform-aws-audit-role?ref=v0.1.0"

# Monorepo subdirectory
source = "github.com/getprobo/probo//contrib/terraform/aws-audit-role?ref=terraform-aws-audit-role/v0.1.0"
```

Connecting `getprobo/terraform-aws-audit-role` to the public Terraform
Registry is a separate step after the first `v*` tag exists there.

### One-time setup (before the first release)

1. Create the public repository `getprobo/terraform-aws-audit-role`. Give
   it a one-sentence GitHub description (the registry uses it later) and
   an MIT license. An empty `main` branch is enough; the first release
   fills it.
2. Add the Actions secret `TERRAFORM_MODULE_GITHUB_TOKEN` on
   `getprobo/probo` with `contents:write` on the module repository (same
   idea as `HOMEBREW_TAP_GITHUB_TOKEN` for the Homebrew tap). The
   workflow uses that token with `gh api`, not `git push`.
