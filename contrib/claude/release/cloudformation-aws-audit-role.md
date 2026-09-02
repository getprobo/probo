# Release CloudFormation (`aws-audit-role`)

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `cloudformation-aws-audit-role/v*`
- **Version source**: `contrib/cloudformation/aws-audit-role/VERSION` (single
  `X.Y.Z` line)
- **Version bump**: Edit `contrib/cloudformation/aws-audit-role/VERSION`
  directly
- **Changelog**: `contrib/cloudformation/aws-audit-role/CHANGELOG.md`
- **Files to stage**: `contrib/cloudformation/aws-audit-role/VERSION`,
  `contrib/cloudformation/aws-audit-role/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-cloudformation-aws-audit-role.yaml`
- **Path filter**: `contrib/cloudformation/aws-audit-role`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='cloudformation-aws-audit-role/v*' 2>/dev/null)..HEAD --oneline \
  -- contrib/cloudformation/aws-audit-role
```

If `git describe` fails (no tag yet), the empty range lists every commit
on the path. If empty or non-user-facing only, do not release this track.

## Notes

CI checks that `VERSION` matches the tag, then creates a GitHub Release on
this repository with `aws-audit-role-<version>.yaml`. It does not upload
the template to S3 and does not push to another repository.

Download the template from the GitHub Release and create a stack from it
(`--template-body`, or upload in the console). CloudFormation quick-create
needs a public S3 `templateURL` and is a separate publishing step.
