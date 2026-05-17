# ADR-0002: Pool runs functions, not RunOpts

## Status

Accepted

## Context

The initial pool design accepted `RunOpts` directly — submit a task, pool runs it, collect results. This assumed each unit of work is a single agent invocation.

In practice, the primary use case is multi-step pipelines: implement → review → fix. Each step is a separate `Run()` call, and the logic between steps (check review output, decide whether to fix) is arbitrary Go code.

Two options were considered:

1. **Typed pool** — `pool.Submit(RunOpts)` — simple but forces one agent call per submission. Multi-step workflows require external coordination.
2. **Function pool** — `pool.Submit(func(ctx context.Context) error)` — the pool is a concurrency-limited task runner. The function contains arbitrary logic including multiple `Run()` calls.

## Decision

Function pool. `pool.Submit(func(ctx context.Context) error)`.

## Consequences

- The pool is agent-agnostic — it doesn't import or know about RunOpts or agents.
- Multi-step pipelines are natural: sequential Run() calls inside a function, with conditionals and error handling.
- The pool's concurrency limit applies to pipeline functions, not individual agent processes. With sequential pipelines, these are equivalent.
- Structured output handling stays at the `Run[T]()` call site, not the pool level. The pool just collects errors and durations.
- Three-tier shutdown integrates cleanly: the context passed to the function carries drain state, and `Run()` returns `ErrDraining` if called during drain.
