# Release Terraform (`gcp-audit-role`)

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `terraform-gcp-audit-role/v*`
- **Version source**: `contrib/terraform/gcp-audit-role/VERSION` (single
  `X.Y.Z` line)
- **Version bump**: Edit `contrib/terraform/gcp-audit-role/VERSION` directly
- **Changelog**: `contrib/terraform/gcp-audit-role/CHANGELOG.md`
- **Files to stage**: `contrib/terraform/gcp-audit-role/VERSION`,
  `contrib/terraform/gcp-audit-role/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-terraform-gcp-audit-role.yaml`
- **Path filter**: `contrib/terraform/gcp-audit-role`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='terraform-gcp-audit-role/v*' 2>/dev/null)..HEAD --oneline \
  -- contrib/terraform/gcp-audit-role
```

If `git describe` fails (no tag yet), the empty range lists every commit
on the path. If empty or non-user-facing only, do not release this track.

## Notes

CI checks that `VERSION` matches the tag, runs `terraform fmt -check`,
creates a GitHub Release on this repository, then copies the module into
`getprobo/terraform-gcp-audit-role` and tags `v<version>` there. The
module-repo commit is created with GitHub's `createCommitOnBranch`
mutation (same as the Homebrew tap) so it satisfies the org signed-commit
rule; a local `git commit` plus `git push` would be rejected.

The primary consumer address is the Terraform Registry module:

```hcl
module "probo_audit" {
  source  = "getprobo/audit-role/gcp"
  version = "0.1.0"
}
```

Consumers can also pin a Git source:

```hcl
# Dedicated module repo (filled by the release workflow)
source = "github.com/getprobo/terraform-gcp-audit-role?ref=v0.1.0"

# Monorepo subdirectory
source = "github.com/getprobo/probo//contrib/terraform/gcp-audit-role?ref=terraform-gcp-audit-role/v0.1.0"
```

Connecting `getprobo/terraform-gcp-audit-role` to the public Terraform
Registry is a separate step after the first `v*` tag exists there. The
registry address is `getprobo/audit-role/gcp`.

### One-time setup (before the first release)

1. Create the public repository `getprobo/terraform-gcp-audit-role`. Give
   it a one-sentence GitHub description (the registry uses it later) and
   an MIT license. An empty `main` branch is enough; the first release
   fills it.
2. Reuse the Actions secret `TERRAFORM_MODULE_GITHUB_TOKEN` on
   `getprobo/probo`. Grant it `contents:write` on this module repository
   as well as `getprobo/terraform-aws-audit-role`. The workflow uses that
   token with `gh api`, not `git push`.
