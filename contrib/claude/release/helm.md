# Release Helm chart (`probo`)

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `helm/v*`
- **Version source**: `contrib/helm/charts/probo/Chart.yaml` (`version` field)
- **Version bump**: Edit `version` in `contrib/helm/charts/probo/Chart.yaml`
- **Changelog**: `contrib/helm/charts/probo/CHANGELOG.md`
- **Files to stage**: `contrib/helm/charts/probo/Chart.yaml`,
  `contrib/helm/charts/probo/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-helm.yaml`
- **Path filter**: `contrib/helm`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='helm/v*')..HEAD --oneline \
  -- contrib/helm
```

If empty or non-user-facing only, do not release this track.

## Notes

The chart has its own SemVer (`version`). `appVersion` in `Chart.yaml` is
the default probod application version the chart deploys (image tag
`v<appVersion>`). Bump `appVersion` when the chart should default
to a newer probod release.

CI packages the chart and pushes it to
`oci://artifact.probo.inc/probo/probo`, then publishes a GitHub Release.
