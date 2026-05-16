# Probo — Go Backend — pkg/mail + pkg/mailer + pkg/mailman

**Purpose.** Transactional email outbox and list management.

| Package | Role |
| --- | --- |
| `pkg/mail` | Email primitives (Address, Message, attachments) and rendering helpers. |
| `pkg/mailer` | Outbox sender — drains the `emails` table via a poll worker and delivers via SMTP / SES / configured transport. |
| `pkg/mailman` | List-management: subscriber lists, unsubscribe tokens, list-unsubscribe headers. |

**Key files.**

- `pkg/mail/message.go` — `Message`, `Attachment`.
- `pkg/mailer/sender.go` — poll worker (Claim / Process / RecoverStale).
- `pkg/coredata/email.go` — outbox row.
- `pkg/mailman/service.go` — list management + unsubscribe.
- E2E: `e2e/internal/testutil/mailpit.go` reads delivered mail for
  assertions.

**How to use (canonical send).** Inside a service transaction:

```go
err := pg.WithTx(ctx, func(tx pg.Tx) error {
    // ... entity mutation ...
    return mailer.InsertEmail(ctx, tx, scope, mail.Message{
        From:    mail.Address{...},
        To:      []mail.Address{...},
        Subject: "...",
        HTMLBody: rendered,
    })
})
```

The sender drains the row asynchronously.

**Top pitfalls.**

- Sending mail outside a transaction → partial-state risk identical to
  the webhook outbox.
- Including PII in the subject line that ends up in logs (most senders
  log the subject). Keep PII in the body, not the subject.
- Mailpit assertions in e2e need polling — use
  `Client.SearchMails` with retries; don't rely on a single immediate
  fetch.
