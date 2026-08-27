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
- **Path filter**: `contrib/cloudformation/aws-audit-role`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='cloudformation-aws-audit-role/v*' 2>/dev/null)..HEAD --oneline \
  -- contrib/cloudformation/aws-audit-role
```

If `git describe` fails (no tag yet), the empty range lists every commit
on the path. If empty or non-user-facing only, do not release this track.

## Notes

This track has no GitHub Actions workflow. After the common tag push,
extract the changelog entry and create the GitHub Release by hand. Do not
upload the template to S3 as part of the release.

```shell
VERSION="$(cat contrib/cloudformation/aws-audit-role/VERSION)"
awk -v ver="$VERSION" '
  /^## \[/ { if (found) exit; if ($0 ~ "\\[" ver "\\]") found=1 }
  found
' contrib/cloudformation/aws-audit-role/CHANGELOG.md > release-notes.md

gh release create "cloudformation-aws-audit-role/v${VERSION}" \
  --title "cloudformation-aws-audit-role/v${VERSION}" \
  --notes-file release-notes.md \
  contrib/cloudformation/aws-audit-role/aws-audit-role.yaml
```
