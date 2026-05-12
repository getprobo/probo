# Security Notes

User-facing notes on security-relevant changes to Probo. For the
vulnerability reporting process, see [SECURITY.md](SECURITY.md).

## Password changes invalidate existing sessions

_2026-04-29, IAM_

Password changes and resets now revoke existing sessions for the identity.

Previously, rotating a password did not touch `iam_sessions` rows: a stolen session stayed valid until its idle TTL elapsed, so a user whose account was compromised on another device could not evict that device by changing the password.

Inside the same transaction as the password update:

- A signed-in password change revokes every other active session for the identity and keeps the caller's current session.
- A forgot-password reset revokes all active sessions for the identity. The caller is anonymous (authenticated only by the reset token), so there is no current session to preserve.

The session middleware already rejects rows with `expire_reason` set, so revoked sessions are kicked out on the next request without any middleware change.

Reported by [emimoir](https://github.com/emimoir).
