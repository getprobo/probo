# Probo — Go Backend — pkg/probod

**Purpose.** Single composition root. Wires every subsystem into the
`probod` daemon: builds all services, runs DB migrations, launches
the API HTTP server, the Trust HTTP/HTTPS server, every worker, every
sender, and orchestrates graceful shutdown.

> See [patterns.md § 6 Service orchestration](../patterns.md#6-service-orchestration-probod).

**Key files.**

- `probod.go` — `Implm` (`unit.Configurable` + `unit.Runnable`),
  `New()`, `Run()`, `runApiServer`, `runTrustCenterServer`, every
  `runXxxWorker` and `runXxxSender`.
- `llm.go` — `resolveAgentClient`, `buildLLMClient` (per-named-agent
  provider wiring with logger/tracer/prometheus).
- `vendor_assessor.go` — opt-in factory: returns
  `probo.DisabledVendorAssessor` if no provider is configured.

**How to extend (a new subsystem).**

1. Add config fields to the appropriate `Config` substruct (and
   propagate to all 11 config files — see
   [shared.md § 4](../../shared.md#4-configuration-propagation)).
2. Construct the subsystem inside `Run()` after migrations and before
   the goroutine launch block.
3. Launch with the standard pattern:
   ```go
   subCtx, stopSub := context.WithCancel(context.Background()) // <-- Background, not ctx
   wg.Go(func() {
       if err := sub.Run(subCtx); err != nil {
           cancel(fmt.Errorf("subsystem crashed: %w", err))
       }
   })
   ```
4. Add `stopSub()` to the shutdown sequence (after `<-ctx.Done()`,
   before `wg.Wait()`).
5. **Don't forget the stop function** — see
   [pitfalls.md § 7](../pitfalls.md).

**Top pitfalls.**

- Forgetting the `stopXxx()` call after `<-ctx.Done()` — worker keeps
  running after shutdown.
- Adopting `errgroup` at the top level — breaks the independent-lifetime
  contract. Only `runTrustCenterServer` uses errgroup, intentionally.
- Closing `pgClient` before `wg.Wait()` — abandoned in-flight queries.
  `pgClient.Close()` is the very last call.
- DB migrations are synchronous — failures abort startup. Don't try to
  parallelise them.
