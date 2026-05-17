# Context: afk

## Glossary

- **Agent**: An external CLI tool (e.g., Codex, Claude Code) that performs coding tasks in a directory. In this package, represented by the `Agent` interface which builds an `exec.Cmd`.
- **Run**: A single invocation of an agent process against a directory with a prompt. The atomic unit of work in the package.
- **Pipeline**: A sequence of Runs that form a logical task (e.g., implement → review → fix). Pipelines are user-defined functions, not a package-level abstraction.
- **Pool**: A concurrency-limited executor of pipeline functions with three-tier graceful shutdown.
- **Drain**: The first tier of shutdown. No new Runs start (they return `ErrDraining`), but in-flight agent processes run to completion.
- **Worktree**: A git worktree created by package helpers to allow concurrent agent work on the same repo without conflicts.
- **Output tag**: An XML-like delimiter (e.g., `<result>...</result>`) the agent uses to emit structured JSON output that the package parses into a typed struct.
