# Security Notes

User-facing notes on security-relevant changes to Probo. For the
vulnerability reporting process, see [SECURITY.md](SECURITY.md).

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

## Password changes invalidate existing sessions

_2026-04-29, IAM_

Password changes and resets now revoke existing sessions for the identity.

Previously, rotating a password did not touch `iam_sessions` rows: a stolen session stayed valid until its idle TTL elapsed, so a user whose account was compromised on another device could not evict that device by changing the password.

Inside the same transaction as the password update:

- A signed-in password change revokes every other active session for the identity and keeps the caller's current session.
- A forgot-password reset revokes all active sessions for the identity. The caller is anonymous (authenticated only by the reset token), so there is no current session to preserve.

The session middleware already rejects rows with `expire_reason` set, so revoked sessions are kicked out on the next request without any middleware change.

Reported by [emimoir](https://github.com/emimoir).
