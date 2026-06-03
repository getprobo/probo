# Security Notes

User-facing notes on security-relevant changes to Probo. For the
vulnerability reporting process, see [SECURITY.md](SECURITY.md).

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
