# Go Worker

Background workers use `go.gearno.de/kit/worker`. The kit handles the polling loop, semaphore-based concurrency, graceful shutdown, non-cancellable contexts, Prometheus metrics, and OpenTelemetry tracing. You only implement a **handler**.

## Handler interface

Implement `worker.Handler[T]` with `Claim` and `Process`. Optionally implement `worker.StaleRecoverer` for stale row recovery.

```go
type fooHandler struct {
	pg         *pg.Client
	logger     *log.Logger
	staleAfter time.Duration
}

func (h *fooHandler) Claim(ctx context.Context) (coredata.FooItem, error) {
	var item coredata.FooItem

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := item.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			item.Status = coredata.FooStatusProcessing
			item.UpdatedAt = now
			return item.Update(ctx, tx, coredata.NewNoScope())
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.FooItem{}, worker.ErrNoTask
		}
		return coredata.FooItem{}, err
	}

	return item, nil
}

func (h *fooHandler) Process(ctx context.Context, item coredata.FooItem) error {
	if err := h.handle(ctx, &item); err != nil {
		if failErr := h.fail(ctx, &item, err); failErr != nil {
			h.logger.ErrorCtx(ctx, "cannot fail item", log.Error(failErr))
		}
		return err
	}
	return nil
}

func (h *fooHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleFooItems(ctx, conn, h.staleAfter)
		},
	)
}
```

## Constructor

The constructor builds the handler and returns `*worker.Worker[T]`. Use `worker.WithInterval` and `worker.WithMaxConcurrency` for tuning. Keep domain-specific options (e.g. timeouts, staleAfter) on the handler struct.

```go
func NewFooWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.FooItem] {
	h := &fooHandler{
		pg:         pgClient,
		logger:     logger,
		staleAfter: 5 * time.Minute,
	}

	return worker.New(
		"foo-worker",
		h,
		logger,
		opts...,
	)
}
```

## Key principles

- **Claim with `FOR UPDATE SKIP LOCKED`** — prevents multiple workers from picking the same row
- **Return `worker.ErrNoTask`** from `Claim` when no work is available (not the coredata sentinel)
- **Context is non-cancellable** — the kit provides `context.WithoutCancel` to both `Claim` and `Process`
- **Process handles its own failures** — update DB status on error, the kit only logs and records metrics
- **Stale recovery is optional** — implement `worker.StaleRecoverer` if the worker marks rows as "processing"
- **Kit provides observability** — Prometheus metrics (`worker_tasks_total`, `worker_task_duration_seconds`, etc.) and OTel traces are automatic
