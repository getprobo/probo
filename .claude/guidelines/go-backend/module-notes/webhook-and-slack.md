# Probo — Go Backend — pkg/webhook + pkg/slack

**Purpose.** Outbox-pattern outbound notification system. Domain
mutations write a row inside the same transaction; a background sender
drains the queue and delivers over HTTPS (webhook) or via the Slack
Web API (slack).

> See [patterns.md § 10 Webhook outbox / payload DTOs](../patterns.md#10-webhook-outbox--payload-dtos).

**Key files.**

- `pkg/webhook/data.go` — `InsertData(ctx, tx, scope, orgID, eventType,
  dto)` transactional emission API; `Payload` envelope (eventId,
  subscriptionId, organizationId, eventType, createdAt, data).
- `pkg/webhook/sender.go` — poll loop (own loop, not `kit/worker`),
  HMAC-SHA256 signing (`X-Probo-Webhook-Signature`,
  `X-Probo-Webhook-Timestamp`), per-subscription signing-secret cache
  (`sync.Map`).
- `pkg/webhook/types/<entity>.go` — public DTOs (`NewVendor`,
  `NewUser`, `NewObligation`, ...) — **the only types allowed in the
  payload `data` field**.
- `pkg/coredata/webhook_subscription.go` — encrypted signing secret
  (`EncryptedSigningSecret []byte`, AES-256-GCM).
- `pkg/coredata/webhook_data.go` — outbox row, `LoadNextUnprocessedForUpdate`
  (FOR UPDATE SKIP LOCKED).
- `pkg/coredata/webhook_event.go` — per-subscription delivery attempt.
- `pkg/slack/sender.go` — sibling outbox for Slack messages.

**How to extend (a new event).**

1. Add an event-type constant to `pkg/coredata/webhook_event_type.go`
   (`namespace:action`).
2. Add a DTO in `pkg/webhook/types/<entity>.go` if it doesn't exist —
   `NewXxx(coredata.Xxx) Xxx` constructor.
3. Call `webhook.InsertData(ctx, tx, scope, orgID, "namespace:action",
   webhooktypes.NewXxx(entity))` **inside** the entity-mutation
   transaction.
4. Add an e2e test that exercises the GraphQL mutation and asserts the
   webhook payload (use the Mailpit pattern but for webhooks — capture
   the delivery via a fake HTTP server).

**Top pitfalls.**

- Passing a raw `coredata` struct as the DTO. Compiles, runs, leaks
  internals. Frequency-2 reviewer rule. See
  [pitfalls.md § 14](../pitfalls.md).
- Calling `webhook.InsertData` outside the transaction.
- Sender is **not** `kit/worker`-based — don't try to use stale-recovery
  primitives from there. The Sender's own loop handles retries via
  the `WebhookEvent` status column.
