# Security Notes

User-facing notes on security-relevant changes to Probo. For the
vulnerability reporting process, see [SECURITY.md](SECURITY.md).

## Safer Password Changes

_2026-04-29 — **IAM**_

> Changing or resetting a password now revokes old sessions automatically.

Changing a password should close the door behind it. Probo now revokes existing sessions when credentials change.

If you change your password while signed in, every other active session is expired and your current session stays open. If your password is reset, all sessions are expired.

It is a small security detail, but an important one. A password update now does what people expect: it cuts off old access immediately.

Thanks to [emimoir](https://github.com/emimoir) for reporting the security issue behind this fix.
