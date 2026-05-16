# Probo — Go Backend — pkg/esign

**Purpose.** In-house e-signature workflow. Renders signing-ready PDFs
via `pkg/docgen` (chromedp + html2pdf), sends signing requests by
email, captures signatures, and emits webhook events at each stage.
Orchestrated by poll-based workers using the standard
`kit/worker` + FOR UPDATE SKIP LOCKED pattern.

**Key files.**

- `pkg/esign/service.go` — `Service`, `Sign`, `Cancel`, `LoadByID`.
- `pkg/esign/<worker>.go` — render worker, dispatch worker, reminder
  worker (each implements `worker.Handler`).
- `pkg/coredata/esignature*.go` — entity rows + status enum.

**How to extend.** Add a new state to the signature workflow:

1. Extend the status enum in `pkg/coredata`.
2. Add a migration that allows the new state.
3. If a new worker is needed, follow the
   [worker pattern](../patterns.md#3-worker-pattern-poll-based--for-update-skip-locked).
4. Update the `pkg/probo/document_service.go` error mapping if the
   custom errors change.

**Top pitfalls.**

- Document service custom errors (`ErrSignatureNotCancellable`,
  `ErrDocumentVersionNotDraft`, ...) must be mapped at every API
  boundary — missing case in resolver → leaks to client.
- Reminder cadence is bound to the dispatcher's poll interval; do not
  add a separate timer — extend the dispatcher.
