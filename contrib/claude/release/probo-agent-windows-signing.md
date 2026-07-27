# Set up Azure Trusted Signing for `probo-agent`

This guide walks through everything needed before the first
Authenticode-signed Windows release. The release workflow
(`.github/workflows/release-probo-agent.yaml`) builds on `windows-latest`,
signs `probo-agent.exe` and the MSI with **Azure Artifact Signing**
(formerly Trusted Signing), then publishes zip + MSI.

Official Microsoft docs:

- [Quickstart: Set up Artifact Signing](https://learn.microsoft.com/en-us/azure/trusted-signing/quickstart)
- [Connect GitHub Actions to Azure with OIDC](https://learn.microsoft.com/en-us/azure/developer/github/connect-from-azure)
- [Artifact Signing roles](https://learn.microsoft.com/en-us/azure/trusted-signing/concept-resources-roles)

Naming note: Azure UI/docs increasingly say **Artifact Signing**; older
pages still say **Trusted Signing**. Same service (`Microsoft.CodeSigning`).

## What you will create

| Azure / GitHub object | Purpose |
|-----------------------|---------|
| Azure subscription + billing | Pays for Artifact Signing (~Basic SKU) |
| Resource provider `Microsoft.CodeSigning` | Enables the service on the subscription |
| Artifact Signing account | Container for identity + certificate profile |
| Organization **Public** identity validation | Publisher name on the certificate (Probo Inc) |
| Certificate profile (`PublicTrust`) | Profile the CI job signs with |
| Entra app registration + service principal | Identity GitHub Actions assumes via OIDC |
| Federated credential → GitHub Environment | No client secret; tag-triggered releases can sign |
| GitHub Environment `probo-agent-release` | Scopes OIDC (tag wildcards are not supported) |
| GitHub secrets + variables | Wired into `build-windows` |

Suggested concrete names (adjust if they collide globally):

| Item | Suggested value |
|------|-----------------|
| Resource group | `rg-probo-codesigning` |
| Artifact Signing account | `probo-codesigning` (3–24 chars, globally unique) |
| Certificate profile | `probo-agent` |
| Region / endpoint | e.g. East US → `https://eus.codesigning.azure.net/` |
| Entra app | `github-probo-agent-codesign` |
| GitHub Environment | `probo-agent-release` (must match the workflow) |

Region → endpoint map (pick the region of the account):

| Region | Endpoint |
|--------|----------|
| East US | `https://eus.codesigning.azure.net/` |
| West Europe | `https://weu.codesigning.azure.net/` |
| North Europe | `https://neu.codesigning.azure.net/` |
| West US 2 | `https://wus2.codesigning.azure.net/` |
| … | See [quickstart region table](https://learn.microsoft.com/en-us/azure/trusted-signing/quickstart) |

Public Trust is available to organizations in supported countries (US,
Canada, EU, UK, and others listed in Microsoft’s docs). Confirm Probo’s
legal entity qualifies before starting.

## 1. Azure subscription and resource provider

1. Sign in to [Azure Portal](https://portal.azure.com) with an account that
   can create resources and assign roles on the target subscription.
2. Open **Subscriptions** → select the subscription → **Resource providers**.
3. Find **`Microsoft.CodeSigning`** → **Register**.
4. Note the **Subscription ID** (GUID). You will store it as
   `AZURE_SUBSCRIPTION_ID`.

Optional CLI:

```shell
az login
az account set --subscription '<subscription-id>'
az provider register --namespace Microsoft.CodeSigning
az provider show --namespace Microsoft.CodeSigning --query registrationState
az extension add --name artifact-signing   # or trustedsigning on older CLIs
```

## 2. Create the Artifact Signing account

Portal:

1. Search for **Artifact Signing Accounts** → **Create**.
2. Subscription + new resource group (e.g. `rg-probo-codesigning`).
3. Account name (e.g. `probo-codesigning`).
4. Region that supports Artifact Signing (see table above).
5. Pricing: **Basic** is enough for typical release volume; use Premium
   only if you need multiple accounts / higher quotas.
6. **Review + create** → open the resource.

CLI sketch:

```shell
az group create --name rg-probo-codesigning --location eastus
az artifact-signing create \
  -n probo-codesigning \
  -g rg-probo-codesigning \
  -l eastus \
  --sku Basic
```

Save:

- Account name → GitHub variable `AZURE_TRUSTED_SIGNING_ACCOUNT`
- Endpoint for that region → `AZURE_TRUSTED_SIGNING_ENDPOINT`

## 3. Grant yourself Identity Verifier (once)

Identity validation can only be started by a principal with the
**Artifact Signing Identity Verifier** role (portal may still show the
older “Trusted Signing Identity Verifier” name).

1. Open the Artifact Signing account → **Access control (IAM)**.
2. **Add** → **Add role assignment**.
3. Role: **Artifact Signing Identity Verifier**.
4. Assign to your user (the human who will submit Probo Inc’s docs).
5. **Review + assign**.

Without this role, **New identity** stays disabled.

## 4. Organization public identity validation (Probo Inc)

This is the long pole (often **1–20 business days**). Do it before
relying on a release tag.

1. Artifact Signing account → **Identity validations**.
2. Choose **Organization** → **New identity** → **Public**
   (required for Public Trust profiles used for public downloads).
3. Fill fields carefully — they drive the certificate subject preview:
   - Legal organization name (as it should appear as publisher)
   - Website URL owned by the entity (e.g. `https://probo.com`)
   - Primary + secondary emails (primary receives verification links;
     links expire in ~7 days)
   - Business identifier / registration details as requested
   - Legal business address
   - First/last name of the individual who will complete ID proofing
     (must match government ID)
4. Preview the certificate subject → **Create**.
5. When status is **Action Required**, complete email verification and
   the individual ID proofing flow (third-party verifier + Authenticator
   Verified ID) linked from the portal / email.
6. Wait until status is **Completed**. Failed email verification means
   start a **new** request.

Tips:

- Keep public company records and domain registration in sync with what
  you submit.
- Billing account type for the subscription should be compatible with
  **organization** validation (individual billing accounts cannot
  validate an organization identity).
- Do not expect street address on the public certificate unless you
  explicitly include it later on the profile.

## 5. Create a Public Trust certificate profile

1. Artifact Signing account → **Certificate profiles** → **Create**.
2. Type: **Public Trust** (not Private Trust; Private is for internal-only).
3. Profile name (e.g. `probo-agent`) → GitHub variable
   `AZURE_TRUSTED_SIGNING_PROFILE`.
4. Select the **Completed** organization identity for CN/O.
5. Leave street / postal off unless you intentionally want them on the
   cert.
6. Create the profile.

Leaf certificates issued under the profile are short-lived (~days). CI
always timestamps with `http://timestamp.acs.microsoft.com` so signatures
remain valid after the leaf rotates. You never download a long-lived PFX.

## 6. Entra app registration for GitHub OIDC

Prefer **federated credentials** (no long-lived client secret).

### 6.1 Register the app

1. Azure Portal → **Microsoft Entra ID** → **App registrations** →
   **New registration**.
2. Name: e.g. `github-probo-agent-codesign`.
3. Supported account types: **Single tenant**.
4. Redirect URI: leave empty.
5. **Register**.
6. Note:
   - **Application (client) ID** → `AZURE_CLIENT_ID`
   - **Directory (tenant) ID** → `AZURE_TENANT_ID`

### 6.2 Federated credential (GitHub Environment)

Tag name wildcards (`probo-agent/v*`) are **not** supported in Entra
federated credentials. The release workflow therefore uses the GitHub
Environment `probo-agent-release`.

1. On the app → **Certificates & secrets** → **Federated credentials** →
   **Add credential**.
2. Scenario: **GitHub Actions deploying Azure resources**.
3. Organization: `getprobo`
4. Repository: `probo`
5. Entity type: **Environment**
6. Environment name: `probo-agent-release` (exact match)
7. Name the credential (e.g. `github-probo-agent-release`)
8. Audience: leave default (`api://AzureADTokenExchange`)

Expected subject (for debugging):

```text
repo:getprobo/probo:environment:probo-agent-release
```

### 6.3 Grant signing permission

1. Artifact Signing account → **Access control (IAM)** → **Add role assignment**.
2. Role: **Artifact Signing Certificate Profile Signer**
   (legacy name: Trusted Signing Certificate Profile Signer).
3. Assign to **User, group, or service principal**.
4. Search by the **app registration name** (apps often do not appear until
   you type the name).
5. **Review + assign**.

Do **not** give the GitHub app the Identity Verifier role. Keep verifier
on humans only.

## 7. GitHub repository configuration

### 7.1 Environment

1. GitHub repo **Settings** → **Environments** → **New environment**.
2. Name: `probo-agent-release`.
3. Optional but recommended: required reviewers, and restrict to tags
   matching `probo-agent/v*` if your plan supports deployment branch/tag
   policies.

The `build-windows` job already sets `environment: probo-agent-release`.

### 7.2 Secrets

Repo (or org) **Settings** → **Secrets and variables** → **Actions** →
**Secrets**:

| Secret | Value |
|--------|-------|
| `AZURE_CLIENT_ID` | Entra Application (client) ID |
| `AZURE_TENANT_ID` | Entra Directory (tenant) ID |
| `AZURE_SUBSCRIPTION_ID` | Azure subscription GUID |

### 7.3 Variables

Same page → **Variables**:

| Variable | Example |
|----------|---------|
| `AZURE_TRUSTED_SIGNING_ENDPOINT` | `https://eus.codesigning.azure.net/` |
| `AZURE_TRUSTED_SIGNING_ACCOUNT` | `probo-codesigning` |
| `AZURE_TRUSTED_SIGNING_PROFILE` | `probo-agent` |

Endpoint **must** match the account’s region.

## 8. Verify end-to-end

1. Confirm identity validation status is **Completed** and the Public
   Trust profile exists.
2. Push a prerelease tag, e.g. `probo-agent/v0.0.0-rc.signing-test`
   (after the usual version/changelog process, or a disposable test tag
   on a branch that contains the workflow).
3. Watch **Release probo-agent** → **windows (amd64)** / **windows (arm64)**:
   - Azure login succeeds (OIDC)
   - Sign `probo-agent.exe` succeeds
   - WiX MSI build succeeds
   - Sign MSI succeeds
4. Download an MSI / unzip the Windows archive and check
   Properties → Digital Signatures (publisher should show the validated
   organization name; timestamp present).

### Common failures

| Symptom | Likely cause |
|---------|----------------|
| Azure login 401 / federated token rejected | Environment name mismatch, wrong org/repo, or federated credential subject |
| Artifact Signing 403 | Missing **Certificate Profile Signer** on the app, or wrong account/profile name |
| Identity validation blocked | Missing **Identity Verifier** on your user |
| Wrong endpoint / region errors | Variable endpoint does not match account region |
| Sign works but SmartScreen still warns | Expected for new publishers until reputation builds; not a signing failure |

## 9. What this does *not* cover

- Local unsigned MSI layout testing (see
  [probo-agent.md](./probo-agent.md) — WiX `build.ps1` without Azure).
- Buying a traditional DigiCert/Sectigo EV PFX (out of scope; CI uses
  Artifact Signing only).
- MSI-based auto-update (zip remains the update channel).

## Checklist

- [ ] `Microsoft.CodeSigning` registered on the subscription
- [ ] Artifact Signing account created; endpoint noted
- [ ] Identity Verifier role on the human operator
- [ ] Organization Public identity validation **Completed**
- [ ] Public Trust certificate profile created
- [ ] Entra app + federated credential for Environment `probo-agent-release`
- [ ] Certificate Profile Signer role on the Entra app
- [ ] GitHub Environment `probo-agent-release` exists
- [ ] Secrets `AZURE_CLIENT_ID` / `AZURE_TENANT_ID` / `AZURE_SUBSCRIPTION_ID`
- [ ] Vars `AZURE_TRUSTED_SIGNING_ENDPOINT` / `_ACCOUNT` / `_PROFILE`
- [ ] Prerelease tag signed successfully for both Windows arches
