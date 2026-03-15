# Go Worker

Background workers follow a poll-based pattern with bounded concurrency. The struct holds a `*pg.Client`, a `*log.Logger`, and tuning knobs (`interval`, `staleAfter`, `maxConcurrency`). Use functional options (`With*` functions) for the tuning knobs with sensible defaults.

## Run loop

The `Run(ctx context.Context) error` method uses a `time.Ticker` in a `for`/`select` loop. On each tick it recovers stale rows, then drains available work via `processNext`. Work items are claimed inside a transaction with `FOR UPDATE SKIP LOCKED`, marked as processing, then handled concurrently in goroutines bounded by a semaphore channel. Use `context.WithoutCancel` for work that must complete even after shutdown, and `sync.WaitGroup` with `defer wg.Wait()` to ensure in-flight goroutines finish before `Run` returns.

```go
type (
	FooWorker struct {
		pg             *pg.Client
		logger         *log.Logger
		interval       time.Duration
		staleAfter     time.Duration
		maxConcurrency int
	}

	FooWorkerOption func(*FooWorker)
)

func NewFooWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...FooWorkerOption,
) *FooWorker {
	w := &FooWorker{
		pg:             pgClient,
		logger:         logger,
		interval:       10 * time.Second,
		staleAfter:     5 * time.Minute,
		maxConcurrency: 5,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *FooWorker) Run(ctx context.Context) error {
	var (
		wg     sync.WaitGroup
		sem    = make(chan struct{}, w.maxConcurrency)
		ticker = time.NewTicker(w.interval)
	)
	defer ticker.Stop()
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			nonCancelableCtx := context.WithoutCancel(ctx)
			w.recoverStaleRows(nonCancelableCtx)
			for {
				if err := w.processNext(ctx, sem, &wg); err != nil {
					if !errors.Is(err, coredata.ErrResourceNotFound) {
						w.logger.ErrorCtx(nonCancelableCtx, "cannot claim item", log.Error(err))
					}
					break
				}
			}
		}
	}
}
```

## processNext

Claims one work item inside a transaction, marks it as processing, then handles it in a bounded goroutine:

```go
func (w *FooWorker) processNext(ctx context.Context, sem chan struct{}, wg *sync.WaitGroup) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		item coredata.FooItem
		now  = time.Now()
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			if err := item.LoadNextPendingForUpdateSkipLocked(nonCancelableCtx, tx); err != nil {
				return err
			}
			item.Status = coredata.FooStatusProcessing
			item.UpdatedAt = now
			return item.Update(nonCancelableCtx, tx, coredata.NewNoScope())
		},
	); err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(item coredata.FooItem) {
		defer wg.Done()
		defer func() { <-sem }()

		if err := w.handle(nonCancelableCtx, &item); err != nil {
			w.logger.ErrorCtx(nonCancelableCtx, "cannot process item", log.Error(err))
		}
	}(item)

	return nil
}
```

## Key principles

- **Claim with `FOR UPDATE SKIP LOCKED`** — prevents multiple workers from picking the same row
- **Semaphore channel** — bounds goroutine concurrency to `maxConcurrency`
- **`context.WithoutCancel`** — in-flight work must complete even after shutdown
- **`defer wg.Wait()`** — `Run` blocks until all goroutines finish
- **Stale recovery** — on each tick, reset rows stuck in "processing" for longer than `staleAfter`
- **Drain loop** — keep calling `processNext` until no more pending items (`ErrResourceNotFound`)
