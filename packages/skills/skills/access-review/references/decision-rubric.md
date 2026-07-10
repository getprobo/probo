# Semi-auto decision rubric

Classify each `PENDING` entry before writing to Probo.

## Auto — record without asking

Apply the **first matching rule** (top to bottom). Always set `decision_note`
for non-`APPROVED` decisions.

| Condition | Decision | decision_note template |
| --- | --- | --- |
| `TERMINATED_USER` flag | `REVOKE` | Terminated user — access no longer required |
| `CONTRACTOR_EXPIRED` flag | `REVOKE` | Contractor engagement ended |
| `active === false` and no `NEW` tag | `REVOKE` | Account inactive at source |
| `ORPHANED` flag and `active === false` | `REVOKE` | Orphaned inactive account |
| `ORPHANED` flag only, `active !== false` | `ESCALATE` | Orphaned account still active — needs owner |
| `SHARED_ACCOUNT` flag | `ESCALATE` | Shared account — assign individual owner |
| `SOD_CONFLICT` flag | `ESCALATE` | Segregation of duties conflict |
| `PRIVILEGED_ACCESS` or `ROLE_CREEP` flag | `ESCALATE` | Privileged access requires explicit approval |
| `is_admin === true` and (`DORMANT` flag or last_login very stale) | `ESCALATE` | Admin access dormant — confirm business need |
| `account_type === SERVICE_ACCOUNT`, active, no danger flags | `APPROVED` | Service account with expected access |
| Active user, no flags (or only `NONE`), not admin, `incremental_tag !== NEW` | `APPROVED` | Routine access reaffirmed |

"Very stale" last_login: no login in 90+ days when `last_login` is present.

## Ambiguous — show user, do not write

| Condition | Suggested default | Why ambiguous |
| --- | --- | --- |
| `incremental_tag === NEW` | `ESCALATE` | New access since last campaign |
| `NO_BUSINESS_JUSTIFICATION` flag | `ESCALATE` or `REVOKE` | Needs human judgment |
| `OUT_OF_DEPARTMENT` flag | `ESCALATE` | Role/department mismatch |
| `EXCESSIVE` or `ROLE_MISMATCH` flag | `ESCALATE` | Role change needs context |
| `is_admin === true` without dormant signals | `ESCALATE` | Admin approvals need explicit sign-off |
| `mfa_status === DISABLED` and (`is_admin` or privileged flags) | `ESCALATE` | MFA gap on sensitive access |
| `auth_method === API_KEY` or `PASSWORD` on production-like roles | `ESCALATE` | Non-SSO auth on sensitive access |
| Multiple conflicting flags | `ESCALATE` | Rubric rules disagree |
| `active === null` with revoke-leaning flags | `ESCALATE` | Unknown activity state |

Present the suggested decision; wait for explicit user confirmation.

## Flagging before decision

Do not auto-flag unless the user asks. When reviewing ambiguous entries, you
may **suggest** `flagAccessReviewEntry` if Probo shows `NONE` but signals are
obvious (e.g. admin + 180d no login → suggest `DORMANT`).

## Decision notes

- `APPROVED` — `decision_note` optional
- `REVOKE`, `DEFER`, `ESCALATE` — `decision_note` **required** (MCP rejects empty)
- Keep notes short, factual, auditable. Reference flags and activity signals.

## DEFER vs ESCALATE

- `ESCALATE` — needs another reviewer or manager (security, HR, app owner)
- `DEFER` — modify access (role change, downgrade) before final approval; use
  when the user indicates access should change rather than fully revoke

Default to `ESCALATE` when unsure.
