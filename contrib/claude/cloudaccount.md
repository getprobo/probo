# Cloud Accounts

`Cloud accounts` are the customer-facing label for Probo's
multi-cloud connector subsystem. A `CloudAccount` is a tenant-scoped,
polymorphic record holding `(provider, credential_kind, scope,
encrypted_credentials, status)` and is the foundation for the
access-review pipeline today plus future CSPM-style security
audit modules.

This doc is the entry point for backend work on cloud accounts; it
links each concept to the file that owns the canonical implementation.

## Where things live

| Concern | Path | Notes |
|---------|------|-------|
| Entity, filter, order field | `pkg/coredata/cloud_account*.go` | All SQL stays in `coredata`; the table is `cloud_accounts`. |
| Enums | `pkg/coredata/cloud_account_{provider,credential_kind,status,scope_kind,audit_module}.go` | Iota-free string enums; `Scan`/`Value` parse-on-read. |
| Provider abstraction | `pkg/cloudaccount/{aws,gcp,azure}.go` | Builds typed SDK clients (`*iam.Client`, `*google.Credentials`, `*azidentity.ClientSecretCredential`). |
| Credentials envelope | `pkg/cloudaccount/credentials.go` | Polymorphic `Credentials` interface; `CloudAccountRecord` value type the registry consumes. **Does NOT import `pkg/coredata` entities** — the service layer maps `*coredata.CloudAccount → CloudAccountRecord`. |
| Registry | `pkg/cloudaccount/registry.go` | Three typed builders (`BuildAWSProvider` / `BuildGCPProvider` / `BuildAzureProvider`) plus a polymorphic `BuildProbeable` for the worker. |
| Install assets | `pkg/cloudaccount/install_{aws,gcp,azure}.go` | AWS: CloudFormation Quick-Create URL + content-addressed S3 template. GCP: `gcloud` shell script. Azure: structured walkthrough payload. |
| Probe helpers | `pkg/cloudaccount/probe.go` | Shared error-mapping helpers. The `Probe` method lives on each `*AWSProvider` / `*GCPProvider` / `*AzureProvider`. |
| Service | `pkg/probo/cloud_account_service.go` | TenantService sub-service: `Create`, `RotateCredentials`, `Delete`, `List`, `Get`, `Verify`, `GenerateInstallAssets`. |
| Worker | `pkg/probo/cloud_account_worker.go` | `kit/worker` poll-based; `Claim` opens its own tx with `FOR UPDATE SKIP LOCKED`; `Process` runs the probe out of any tx then writes a short status-transition tx. |
| Resolvers | `pkg/server/api/console/v1/cloud_account_*.go` | GraphQL mutations + field resolvers (incl. field-level RBAC for `scope.identifier` and `last_probe_error`). |
| Credential upload | `pkg/server/api/console/v1/cloud_account_credential_upload.go` | `POST /api/console/v1/cloud-accounts/credentials/upload`; multipart body, body never lands in any access log. |
| MCP | `pkg/server/api/mcp/v1/cloud_account.go` | Mirrors the GraphQL surface tool-by-tool. |
| CLI | `pkg/cmd/cloud-account/...` | `prb cloud-account list / get / create / verify / rotate / delete / install-assets`. |
| Access-review drivers | `pkg/accessreview/drivers/cloud_{aws,gcp,azure}.go` | One driver per cloud; consumes a typed `*cloudaccount.<X>Provider` (no http.Client). |

## Credential model

Every cloud-account row stores its credentials encrypted under the
process-wide AES-256-GCM key (`probod.EncryptionKey`, same key used
by Connector and Webhook secrets). The cleartext shape is a
**polymorphic JSON envelope**:

```json
{
  "v": 1,
  "kind": "AWS_ASSUME_ROLE" | "GCP_SERVICE_ACCOUNT_KEY" | "AZURE_CLIENT_SECRET",
  "payload": { ...per-kind body... }
}
```

- `v` is the envelope version. The `encrypted_credentials` BYTEA column
  also carries a leading version byte (`0x01`) so cipher rotation can
  add `0x02` without forcing a one-shot full-table re-encryption.
- `kind` is the only authoritative source of truth for which provider
  payload to deserialise into; do NOT branch on the `Provider` column
  in unmarshal code.
- `payload` is the per-kind body declared by the matching `*Credentials`
  struct in `pkg/cloudaccount/{aws,gcp,azure}.go`.

`pkg/cloudaccount` owns the `Credentials` interface (`json.Marshaler`,
`json.Unmarshaler`, `Provider()`, `Kind()`) and the
`UnmarshalCredentials([]byte) (Credentials, error)` round-trip. The
service layer (`pkg/probo/cloud_account_service.go`) never decodes the
envelope itself — it ships the encrypted bytes to coredata for write
and consumes the decrypted bytes (via `LoadByID(... encryptionKey)`)
when constructing a `CloudAccountRecord` for the registry.

**Do not log the envelope.** Log only the opaque GID and an error
category derived from the typed `cloudaccount.Err*` sentinels. The
`last_probe_error` column is the operator-facing diagnostic surface.

## AWS — AssumeRole + ExternalId flow

1. Customer clicks **Connect AWS** in the console; the wizard
   POSTs `generateCloudAccountInstallAssets` with provider=AWS.
2. Backend mints a fresh 64-hex `external_id` (`crypto/rand`),
   assembles the **content-addressed** CloudFormation Quick-Create
   URL pointing at `s3://<bucket>/cloud-account/access-review-<sha256>.yml`.
   The URL pre-fills the `ExternalId` parameter so the customer
   only confirms and clicks Create. The S3 object key embeds the
   SHA-256 of the YAML so the URL is the integrity pin: a
   customer's stack stays bound to the exact bytes they reviewed.
   Probo's deployment publishes new objects under new hashes
   (never overwrites) and the bucket is configured with S3 Object
   Lock + a deny-overwrite bucket policy.
3. Customer's CloudFormation stack creates an IAM role whose trust
   policy is `Principal: AWS = <Probo's STS identity>` plus
   `Condition: StringEquals: sts:ExternalId = <ExternalId>`.
4. Customer pastes the resulting Role ARN back into Probo; the
   wizard calls `createCloudAccount` with `awsRoleArn` and
   `awsExternalId` echoed from step 2.
5. Backend persists the row in `PENDING_VERIFICATION`, then
   synchronously calls `Verify`:
   - `sts.GetCallerIdentity` — confirm the assume-role works.
   - `iam.ListUsersInput{MaxItems:1}` — confirm the policy grants
     the access-review module's required actions.
6. On success the row flips to `VERIFIED`; on failure it stays
   `PENDING_VERIFICATION` (never auto-promote a never-verified row)
   with `last_probe_error` set.

External-id generation is per-row, never reused. The
[confused-deputy](https://docs.aws.amazon.com/IAM/latest/UserGuide/confused-deputy.html)
mitigation requires the customer's trust policy to assert the
external_id; an attacker who copies an external_id across orgs
cannot impersonate because the AWS trust policy itself is bound to
the original org's role.

## GCP — Service-account JSON key flow

Two scope variants split the role grants:

| Scope | Roles granted | Probe call |
|-------|---------------|-----------|
| `GCP_PROJECT` | `roles/iam.securityReviewer` at the project | `cloudresourcemanager.Projects.GetIamPolicy(project)` |
| `GCP_ORGANIZATION` | `roles/iam.securityReviewer` at the project + `roles/cloudasset.viewer` at the org | `cloudresourcemanager.Organizations.GetIamPolicy(org)` |

`roles/cloudasset.viewer` is granted only at organization scope
because `cloudasset.SearchAllIamPolicies` (the v1 access-review
enrichment path) requires org or folder scope; granting it at
project scope is a no-op. Project-scope installs deliberately
omit it.

Install flow:

1. Wizard posts `generateCloudAccountInstallAssets` with provider=GCP.
2. Backend returns a self-contained `gcloud` script that creates
   the dedicated `probo-scanner` project, the service account, the
   custom `ProboCloudScanner` role, enables the relevant APIs
   (`cloudresourcemanager.googleapis.com`, `iam.googleapis.com`,
   plus `cloudasset.googleapis.com` only for org scope), and
   prints the SA email + JSON key path.
3. Customer runs the script in Cloud Shell or locally with the
   `gcloud` CLI installed.
4. Customer uploads the JSON key body via
   `POST /api/console/v1/cloud-accounts/credentials/upload`
   (NOT a GraphQL variable — secret bytes never travel over
   GraphQL). The handler encrypts the body, writes it onto the
   row, and triggers `Verify`.

## Azure — App Registration flow

Single-tenant App Registration created by the customer. Probo
grants nothing on its own — the customer assigns the `Reader`
role at the chosen Management Group.

1. Wizard posts `generateCloudAccountInstallAssets` with provider=AZURE
   and `scopeKind=AZURE_MANAGEMENT_GROUP` (or `AZURE_TENANT` /
   `AZURE_SUBSCRIPTION`).
2. Backend returns a structured walkthrough payload (`steps[]` with
   title/body/code) plus the required RBAC roles + Microsoft Graph
   permissions list. There is **no Quick-Deploy in v1** because
   ARM-templating an App Registration cleanly requires extra
   ceremony.
3. Customer follows the walkthrough: creates the App Registration,
   grants admin consent for `Directory.Read.All`, assigns Reader at
   the chosen MG.
4. Customer pastes `tenant_id` + `client_id` via the GraphQL
   `createCloudAccount` mutation, then uploads the
   `client_secret` via the credential-upload endpoint.
5. Backend probes via:
   - `armsubscriptions.NewClient.NewListPager` — confirm the
     credential authenticates and reaches the management plane.
   - `armauthorization.NewRoleAssignmentsClient` — confirm the
     `Reader` assignment exists at scope.

## Install-assets entry points

`pkg/probo.CloudAccountService.GenerateInstallAssets` is the only
public entry point; the resolver and the CLI both go through it.
The deployment-side knobs (`AWSAssumerARN`, `AWSTemplateURL`,
`AWSTemplateSHA256`, plus per-cloud feature flags) live in
`pkg/probodconfig.CloudAccountConfig` and are wired into the
service in `pkg/probod/probod.go`.

## Probe lifecycle

A cloud account moves through four statuses:

```
PENDING_VERIFICATION  → VERIFIED  → ERRORED  → DISCONNECTED
       ↑                  ↑                       ↑
   row created        probe ok            3 strikes over ≥1h
```

- **PENDING_VERIFICATION**: row created, awaiting first probe. A
  never-verified row never auto-promotes to ERRORED on probe
  failure — failure keeps it at PENDING_VERIFICATION with
  `last_probe_error` set so the operator gets an actionable hint.
- **VERIFIED**: last probe succeeded.
- **ERRORED**: one or more recent probe failures, customer hasn't
  acted, transient. Each failure increments
  `consecutive_probe_failures` and stamps `first_probe_failure_at`
  (only the first failure of a streak).
- **DISCONNECTED**: escalated after **3 consecutive failures over
  ≥1h**. Requires manual reconnect (RotateCredentials).

The transition algebra is implemented in
`pkg/probo/cloud_account_worker.go`'s `computeCloudAccountTransition`,
covered exhaustively by `pkg/probo/cloud_account_worker_test.go`.

The probe worker is built on `go.gearno.de/kit/worker`:

- **Default poll interval**: 15 minutes (matches `pkg/iam/scim/bridge_runner.go`).
- **Default `MaxConcurrency`**: 4 (cloud APIs are rate-limited;
  conservative cap to avoid AWS STS / GCP API throttling).
- Both tunable via `probodconfig.CloudAccount.{ProbeInterval,
  ProbeMaxConcurrency}`.
- `Claim` uses `FOR UPDATE SKIP LOCKED` so multiple worker
  instances can run safely. The shape mirrors
  `pkg/probo/evidence_description_worker.go`.

The worker NEVER opens a long-running tx — `Probe` runs out of any
DB tx, then a short tx writes the status transition.

## Field-level RBAC

Two columns are credential-adjacent and require
`core:cloud-account:rotate-credentials` (OWNER / ADMIN) to read:

- `scope.identifier` — AWS account id / GCP project or org id /
  Azure subscription or MG id. Reconnaissance value.
- `last_probe_error` — raw SDK errors frequently embed role ARNs,
  GCP SA emails, Azure tenant/client IDs. Same gate.

The resolver (`pkg/server/api/console/v1/cloud_account_resolvers.go`)
consults `r.authorizeCached` per request so a 25-row page render
does not open 25 `pg.WithTx` authorize calls.

## Where to extend

### Adding a new provider

1. New enum value in `pkg/coredata/cloud_account_provider.go`.
2. New `<X>Credentials` struct + `<X>Provider` in `pkg/cloudaccount/<x>.go`,
   each with a `Probe(ctx) error` method.
3. New `BuildXProvider(rec)` on the registry; extend
   `BuildProbeable` switch.
4. New `pkg/cloudaccount/install_<x>.go` for install-assets
   generation; extend the GraphQL `CloudAccountInstallAssets` union
   and the resolver's discriminated-union mapping
   (`newAWSInstallAssets` / etc. in `cloud_account_helpers.go`).
5. New driver in `pkg/accessreview/drivers/cloud_<x>.go`; wire into
   `pkg/probo/cloud_account_driver.go` (driver factory closure).

### Adding a new audit module

1. New enum value in
   `pkg/coredata/cloud_account_audit_module.go`.
2. New per-(provider, scope) action list in
   `pkg/cloudaccount/permissions.go`.
3. Update install-assets templates so the customer's IAM policy
   grants the new actions.
4. Add a probe / fetch path in the relevant driver.

## Pattern references

- Probe worker shape: `pkg/probo/evidence_description_worker.go`.
- Connector reconnect (analogous to RotateCredentials):
  `pkg/probo/connector_service.go`.
- File-upload route: existing avatar / evidence endpoints in
  `pkg/server/api/console/v1`.
- Field-level RBAC: `cloud_account_resolvers.go` (the
  `LastProbeError` resolver).
