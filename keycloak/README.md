# Keycloak 26 MFA / OIDC-IdP test deployment

Self-hosted Keycloak 26.6.1 on Render with its own PostgreSQL database and a
pre-imported `govrly-test` realm, for exercising **TOTP MFA**, **passkeys (WebAuthn)**,
and **Keycloak as an OIDC identity provider** against a small Go client.

Production-shaped (production mode, Postgres, optimized build, proxy headers) but
POC-grade: free-tier database, throwaway test user.

## Relationship to the rest of the repo

- **`compose/keycloak/`** — the *local dev* Keycloak (`start-dev`, H2, realm template
  for probo SSO). Unrelated to this directory beyond sharing the image version.
- **`keycloak/`** (here) — a *deployed* Keycloak on Render, standalone from the probo
  service: its own image, its own database, no private-network link to probo.
- **`render.yaml`** (repo root) — Render reads the Blueprint only from the repo root,
  so the `govrly-keycloak` service and `govrly-keycloak-db` database are defined
  there alongside the probo stack, pointing at `dockerfilePath: ./keycloak/Dockerfile`.

```
keycloak/
├── Dockerfile                          # 2-stage: kc.sh build --optimized, then runtime
├── entrypoint.sh                       # Render glue: hostname + postgres:// -> JDBC
├── realm-export/
│   └── govrly-test-realm.json          # realm, users, clients, MFA + passkey flow
└── test-client/                        # nested Go module, excluded from the root build
    └── main.go                         # go-oidc auth-code client on localhost:3000
```

---

## 1. Deploy

The services are already in the root `render.yaml`. Commit and push, then in Render:

- **Existing Blueprint** → the `govrly-keycloak` service and `govrly-keycloak-db`
  database show up as additions on the next sync → **Apply**.
- **No Blueprint yet** → **Dashboard → New → Blueprint**, pick this repo, **Apply**.

⚠️ Applying the Blueprint also reconciles the `probo` and `chrome-headless` services.
Review Render's diff before approving if the live probo service has drifted from
`render.yaml`.

First build takes ~4–6 min (image build + `kc.sh build`). First boot runs
`--import-realm` and creates `govrly-test`. The service is healthy once
`GET /realms/master` returns 200.

Grab the bootstrap admin password: **govrly-keycloak → Environment →
`KC_BOOTSTRAP_ADMIN_PASSWORD` → reveal**. Render generated it (`generateValue: true`),
so it is not in this repo.

Your base URL is whatever Render assigned, e.g. `https://govrly-keycloak.onrender.com`.
Nothing is hardcoded — `entrypoint.sh` reads Render's injected `RENDER_EXTERNAL_URL`.

### URLs

| What | URL |
|---|---|
| Admin console | `<BASE_URL>/admin` |
| Account console (MFA self-service) | `<BASE_URL>/realms/govrly-test/account` |
| OIDC discovery | `<BASE_URL>/realms/govrly-test/.well-known/openid-configuration` |
| Issuer | `<BASE_URL>/realms/govrly-test` |

---

## 2. Test sequence

### Step 1 — Admin login

Open `<BASE_URL>/admin`, log in as `admin` with the generated password.
Switch the realm selector (top-left) from `master` to **govrly-test** and confirm:

- **Authentication → Flows** → `browser-passkey` is bound as the browser flow.
- **Authentication → Policies → OTP Policy** → TOTP, SHA1, 6 digits, 30 s.
- **Authentication → Policies → WebAuthn Passwordless Policy** → resident key `Yes`,
  user verification `required`.
- **Users** → `amr@test.local` exists with required action *Configure OTP*.

Create yourself a proper admin user here if you plan to keep this around; the
bootstrap admin is meant to be temporary.

### Step 2 — Test user login → forced TOTP enrollment

Open the account console: `<BASE_URL>/realms/govrly-test/account`

- Username `amr@test.local`, password `ChangeMe123!`
- The password is imported as **temporary**, so Keycloak asks you to set a new one first.
- Then the `CONFIGURE_TOTP` required action fires: scan the QR with an authenticator
  app (FreeOTP / Google Authenticator / Microsoft Authenticator) and enter the 6-digit code.

From now on, that user's logins go username → password → OTP.

### Step 3 — Full OIDC flow through the Go client

Get the client secret: **Admin console → govrly-test → Clients → `govrly-api` →
Credentials → Client secret**. (Keycloak generates it at import time, so it is not
committed anywhere.)

```bash
cd keycloak/test-client
export OIDC_ISSUER="https://<your-service>.onrender.com/realms/govrly-test"
export OIDC_CLIENT_ID="govrly-api"
export OIDC_CLIENT_SECRET="<paste>"
go run .
```

PowerShell:

```powershell
cd keycloak/test-client
$env:OIDC_ISSUER        = "https://<your-service>.onrender.com/realms/govrly-test"
$env:OIDC_CLIENT_ID     = "govrly-api"
$env:OIDC_CLIENT_SECRET = "<paste>"
go run .
```

Open <http://localhost:3000/>. You get redirected to Keycloak, log in as
`amr@test.local` **including the OTP prompt**, and land back on
`localhost:3000/callback`. The client verifies the ID token signature, `aud`, and
`nonce`, then prints the claims to the terminal — check `acr`/`amr` to confirm MFA
was actually exercised.

The client sends PKCE (S256) as well as the client secret, and exits after one
successful login. Re-run it to log in again.

> To test the public SPA client instead, set `OIDC_CLIENT_ID=govrly-spa` and leave
> `OIDC_CLIENT_SECRET` unset — PKCE alone authenticates the request.

`test-client/` is its own Go module, so it is invisible to the repo's root
`go build ./...` and to CI.

### Step 4 — Register a passkey and log in with it

1. In the account console → **Signing in** → **Passkey** (Passwordless) → **Set up**.
2. Register a platform authenticator (Windows Hello / Touch ID) or a security key.
3. Log out, go back to the account console, enter `amr@test.local`, and pick the
   passkey option — it satisfies the first factor on its own, so no password and
   no OTP prompt.

That works because the realm binds a custom `browser-passkey` flow:

```
browser-passkey
├── Cookie                            ALTERNATIVE
├── Identity Provider Redirector      ALTERNATIVE
└── forms                             ALTERNATIVE
    ├── Username Form                 REQUIRED
    └── first factor                  REQUIRED
        ├── WebAuthn Passwordless     ALTERNATIVE   ← passkey
        └── password and otp          ALTERNATIVE
            ├── Password Form         REQUIRED
            └── conditional otp       CONDITIONAL   ← only if OTP is configured
```

Keycloak 26 ships no passwordless browser flow out of the box, so without this the
WebAuthn policy would let you *register* a passkey but never *log in* with one.

Re-run the Go client after this and you can complete the same OIDC flow with a passkey.

**Passkeys need HTTPS.** Render terminates TLS at the edge and `KC_PROXY_HEADERS=xforwarded`
lets Keycloak reconstruct the `https://` origin, so the WebAuthn RP ID resolves correctly.
The policy leaves RP ID empty, which means "derive it from the effective domain" — don't
pin it unless you move to a custom domain.

---

## What's in the realm

| Object | Detail |
|---|---|
| Realm | `govrly-test`, `sslRequired: external`, brute-force protection on |
| User | `amr@test.local` / `ChangeMe123!` (temporary), required action `CONFIGURE_TOTP` |
| Client `govrly-api` | Confidential, standard flow + **service accounts**, redirect `http://localhost:3000/*` |
| Client `govrly-spa` | Public, standard flow, **PKCE S256 enforced**, redirect `http://localhost:3000/*` |
| Realm role | `govrly-user`, assigned to the test user |
| Required actions | `CONFIGURE_TOTP`, `webauthn-register`, `webauthn-register-passwordless` all enabled |

Service-account check for `govrly-api`:

```bash
curl -s -u "govrly-api:<secret>" \
  -d grant_type=client_credentials \
  "<BASE_URL>/realms/govrly-test/protocol/openid-connect/token" | jq .
```

---

## Render specifics worth knowing

**Hostname.** `entrypoint.sh` exports `KC_HOSTNAME="$RENDER_EXTERNAL_URL"`. Render
injects that per-service at runtime, so the same image works on a preview service, a
renamed service, or a fresh deploy without edits.

**Database URL.** Render hands out `postgres://` connection strings; Keycloak wants
JDBC. Rather than string-mangling, `render.yaml` wires the five discrete values
(`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`) via `fromDatabase`, and
the entrypoint assembles `jdbc:postgresql://$DB_HOST:$DB_PORT/$DB_NAME`.

**Health check.** Keycloak 26 moved `/health` to the management port (9000). Render
routes exactly one port, so `/health` is unreachable. `healthCheckPath: /realms/master`
is used instead — unauthenticated, on the main port, and only 200s once the DB is up
and realms are loaded. `KC_HEALTH_ENABLED=true` is still set (and baked at build time)
so `/health` works if you ever shell in or expose 9000.

**Port.** `KC_HTTP_PORT=10000` matches Render's default detected port.

**Region.** `frankfurt`, and the database matches. The probo service has no `region`
set, so it stays on Render's default — the two cannot reach each other over the
private network. That is fine here; nothing links them.

**Build context.** `dockerContext: ./keycloak` — the Dockerfile's `COPY realm-export/…`
is relative to this directory, and `keycloak/.dockerignore` applies, not the root one.

**⚠️ Render's free PostgreSQL is deleted 30 days after creation.** When that happens
the service fails to start and the realm, users, TOTP secrets, and passkeys are gone.
This is a POC environment — re-apply the Blueprint (or upgrade the database to a paid
plan) and re-run the enrollment steps. Don't put anything you care about here.

---

## Local run (optional)

```bash
docker build -t govrly-keycloak ./keycloak
docker run --rm -p 8081:8081 \
  -e RENDER_EXTERNAL_URL=http://localhost:8081 \
  -e KC_HTTP_PORT=8081 \
  -e KC_HTTP_ENABLED=true \
  -e KC_BOOTSTRAP_ADMIN_USERNAME=admin \
  -e KC_BOOTSTRAP_ADMIN_PASSWORD=admin \
  -e DB_HOST=host.docker.internal -e DB_PORT=5432 \
  -e DB_NAME=keycloak -e DB_USER=keycloak -e DB_PASSWORD=keycloak \
  govrly-keycloak
```

You need a Postgres reachable at those values — the image is built
`--db=postgres --optimized`, so there is deliberately no dev-file fallback. Port 8081
avoids the probo API (8080) and the compose dev Keycloak (8082).

## Security notes

- No secrets in the repo: the admin password comes from `generateValue`, the DB
  credentials from `fromDatabase`, and the `govrly-api` client secret is generated by
  Keycloak during import.
- Rotate or delete the bootstrap admin once you have created a real admin user.
- `ChangeMe123!` is a temporary import password and is public by definition — the first
  login forces a change.
- Redirect URIs are `http://localhost:3000/*` only. Tighten to exact paths before this
  shape goes anywhere near production.
