# Go Service Orchestration

A top-level `Run` method starts child subsystems (workers, servers) as goroutines via `sync.WaitGroup.Go`. Each child gets its own cancellable context created with `context.WithCancel(context.WithoutCancel(ctx))` so that a parent cancellation does not kill in-flight work — the parent explicitly calls each `stop*` function and then `wg.Wait()` for a controlled shutdown.

When a child crashes, it calls `cancel(fmt.Errorf("… crashed: %w", err))` to signal the parent.

```go
func (impl *Implm) Run(ctx context.Context, l *log.Logger) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	// Start a worker
	workerCtx, stopWorker := context.WithCancel(context.WithoutCancel(ctx))
	worker := NewFooWorker(pgClient, l.Named("foo-worker"))
	wg.Go(
		func() {
			if err := worker.Run(workerCtx); err != nil {
				cancel(fmt.Errorf("foo worker crashed: %w", err))
			}
		},
	)

	// Start a server
	serverCtx, stopServer := context.WithCancel(context.WithoutCancel(ctx))
	defer stopServer()
	wg.Go(
		func() {
			if err := impl.runServer(serverCtx, l); err != nil {
				cancel(fmt.Errorf("server crashed: %w", err))
			}
		},
	)

	<-ctx.Done()

	stopServer()
	stopWorker()

	wg.Wait()

	return context.Cause(ctx)
}
```

## Key principles

- **`context.WithCancelCause`** — the parent uses this to track why it's shutting down
- **`context.WithoutCancel`** — each child gets an independent context so parent cancellation doesn't kill in-flight work
- **`context.WithCancel` on the detached context** — gives the parent an explicit `stop*` function for each child
- **Crash propagation** — when a child fails, it calls `cancel(fmt.Errorf("… crashed: %w", err))` to signal the parent
- **Ordered shutdown** — `<-ctx.Done()` triggers, then each `stop*` is called, then `wg.Wait()` blocks until all children finish
- **`context.Cause(ctx)`** — returns the original error that caused shutdown
