# Security Notes

User-facing notes on security-relevant changes to Probo. For the
vulnerability reporting process, see [SECURITY.md](SECURITY.md).

## Privilege escalation to OWNER via createUser

_2026-07-09, IAM_

Promoting a membership to OWNER is an owner-only action, and
`updateMembership` enforced it. The `createUser` mutation (and its MCP
`CreateUserTool` twin) created a membership with a caller-chosen role
but only authorized `iam:membership-profile:create`, which ADMIN holds,
and never re-checked the owner-only gate. An organization ADMIN could
therefore call `createUser` with `role: OWNER` and an attacker-controlled
email, activate it through the normal invitation flow, and gain full
OWNER privileges (organization deletion, SAML/SSO and SCIM configuration,
owner promotion/demotion) that ADMIN is explicitly denied.

Both `createUser` entry points now enforce the owner-only authorization
when the requested role is OWNER, mirroring `updateMembership`. Ownership
grants were subsequently consolidated into fail-closed allow/deny IAM
policies (keyed on the requested `target_role`) so that granting OWNER
fails closed across every membership-creating and membership-updating
path.

Reported by [Pig-Tail](https://github.com/Pig-Tail).
([GHSA-cppp-g98f-gfpp](https://github.com/getprobo/probo/security/advisories/GHSA-cppp-g98f-gfpp))

## Cross-tenant and hidden-item disclosure via Query.node in the Trust Center API

_2026-07-09, Trust Center_

The public Trust Center GraphQL API exposed `Query.node(id:)` with
optional authentication. Six of its type branches (`Organization`,
`Framework`, `Audit`, `ThirdParty`, `TrustCenter`, and
`TrustCenterReference`) derived their tenant scope from the
client-supplied GID and loaded the object with no comparison to the
visited trust center's organization and no visibility filter. An
unauthenticated visitor who held a target GID could read items hidden
from that trust center, and objects belonging to any other organization
(including orgs with no published trust center), leaking names,
descriptions, contact details, and connected items.

Every trust API resolver now derives its scope from the active
compliance page's organization via `compliancepage.ScopeFromContext`,
so reads are confined to the page's tenant. Cross-tenant or unknown IDs
surface as not-found instead of leaking data.

Reported by [Pig-Tail](https://github.com/Pig-Tail).
([GHSA-w23w-f7v2-625w](https://github.com/getprobo/probo/security/advisories/GHSA-w23w-f7v2-625w))

## Broken access control in the public e-signature NDA API

_2026-07-09, Trust Center_

The public Trust Center mutations `acceptElectronicSignature` and
`recordSigningEvent` were gated only by session presence. The underlying
`esign` service built its tenant scope from the client-supplied
`signatureId` and never verified the signature belonged to the caller.
Any self-provisioned trust center visitor holding another visitor's
signature GID could complete that visitor's NDA signature (overwriting
signer name, IP, and user agent, and triggering certificate sealing) or
inject arbitrary audit-trail events, defeating the non-repudiation and
integrity guarantees of the e-signature system.

The `esign` signature operations now take a caller-provided scope rather
than deriving one from the requested ID, and `AcceptSignature` /
`RecordEvent` verify signature ownership within scope before acting. The
signer and actor emails are always derived from the verified session
identity, never client input, and are compared against the signature's
stored `SignerEmail`.

Reported by [Pig-Tail](https://github.com/Pig-Tail).
([GHSA-22xj-f767-ppw6](https://github.com/getprobo/probo/security/advisories/GHSA-22xj-f767-ppw6))

## Cross-tenant IDOR via unvalidated foreign-key references

_2026-07-03, GraphQL_

Two console GraphQL resolvers authorized the parent object and then
resolved a related object with a tenant scope taken from that related
object's own GID, instead of from the authorization result. An
authenticated member of one organization could attach another
organization's Risk or Data Protection Officer profile GID to their own
Finding or Processing Activity, then read it back:

- `Finding.risk` disclosed another organization's Risk (name,
  description, treatment, category, severity, owner).
- `ProcessingActivity.dataProtectionOfficer` disclosed another
  organization's person profile PII (full name, email addresses,
  position).

`FindingService` and `ProcessingActivityService` now validate the
referenced Risk and Data Protection Officer profile against the
caller's scope before storing the reference, and the affected
resolvers (along with every other resolver following the same
pattern) now authorize the related object's own ID rather than the
parent's.

Reported by [Pig-Tail](https://github.com/Pig-Tail).
([GHSA-c74x-79w6-63jh](https://github.com/getprobo/probo/security/advisories/GHSA-c74x-79w6-63jh))

## GraphQL alias-flooding denial of service

_2026-06-29, GraphQL_

The GraphQL endpoints (`connect`, `console`, and `trust`) built their
server with no limits on query size or complexity. A single request
containing thousands of aliased resolver calls (e.g. `a1: viewer { id
}` repeated thousands of times) was parsed, executed, and marshalled
in full, letting an attacker drive excessive CPU and memory use and
degrade service for other tenants.

The shared GraphQL handler now enforces a parser token limit that
rejects oversized queries before execution, a fixed query complexity
limit, and an LRU query cache, with limits configurable per
environment via `PROBOD_API_GRAPHQL_*` env vars and Helm values.

Reported by [Muthu-Devarajan](https://github.com/Muthu-Devarajan).
([GHSA-prh2-g8pv-m7p9](https://github.com/getprobo/probo/security/advisories/GHSA-prh2-g8pv-m7p9))

## Open redirect bypass in saferedirect

_2026-05-26, Auth_

Relative redirect URLs are now normalized before validation. Paths
containing backslashes (including percent-encoded `%5c`) are rejected,
and the cleaned path is checked for protocol-relative and backslash
prefixes.

Previously, `Validate` only inspected the second character of the raw
input. A path like `/../\evil.com` passed validation because the second
character is `.`, but Go's `http.Redirect` normalized it to `/\evil.com`,
which browsers can treat as an external redirect.

Reported by [Fushuling](https://github.com/Fushuling) and
[RacerZ](https://github.com/RacerZ-fighting).
([GHSA-x7qq-m748-8p2c](https://github.com/getprobo/probo/security/advisories/GHSA-x7qq-m748-8p2c),
[CVE-2026-49820](https://www.cve.org/CVERecord?id=CVE-2026-49820))

## Password changes invalidate existing sessions

_2026-04-29, IAM_

Password changes and resets now revoke existing sessions for the identity.

Previously, rotating a password did not touch `iam_sessions` rows: a stolen session stayed valid until its idle TTL elapsed, so a user whose account was compromised on another device could not evict that device by changing the password.

Inside the same transaction as the password update:

- A signed-in password change revokes every other active session for the identity and keeps the caller's current session.
- A forgot-password reset revokes all active sessions for the identity. The caller is anonymous (authenticated only by the reset token), so there is no current session to preserve.

The session middleware already rejects rows with `expire_reason` set, so revoked sessions are kicked out on the next request without any middleware change.

Reported by [emimoir](https://github.com/emimoir).
