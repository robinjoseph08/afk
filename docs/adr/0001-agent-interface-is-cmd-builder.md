# ADR-0001: Agent interface is a Cmd builder

## Status

Accepted

## Context

The package needs an Agent abstraction to support multiple CLI tools (Codex, Claude Code, etc.). The question is how much responsibility the interface carries.

Two options were considered:

1. **Thick interface** — `Run(ctx, opts) (*Result, error)` — each agent implementation owns process execution, output capture, signal handling, and timeout behavior.
2. **Thin interface** — `Cmd(opts) *exec.Cmd` — each agent implementation only builds the command. The package owns all process lifecycle.

## Decision

Thin interface: `Cmd(RunOpts) *exec.Cmd`.

## Consequences

- Process lifecycle (start, capture output, send signals, kill) is implemented once in the package, not duplicated across agents.
- Adding a new agent is trivial — just map RunOpts fields to CLI flags.
- Agents cannot customize execution behavior (e.g., an agent that communicates over a socket instead of stdio). This is acceptable because all target agents are CLI tools with the same stdin/stdout model.
- If an agent ever needs richer lifecycle control, the interface would need to change — but that's unlikely given the target use cases.
