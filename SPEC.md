# la Specification

## 1. Overview

`la` is a tool for exposing workspace-specific shortcut commands in the current shell.

The intended use cases are:

- define commands per repository or working directory
- run those commands via `la run <name>`
- optionally expose them as top-level commands such as `<name>`
- expose them as `abbr` entries as well as `alias`-like commands
- enable or disable them based on the current directory or environment

Conceptually, `la` is closer to a user-specific interactive `Taskfile` than to a CI-oriented task runner.

## 2. Goals

- Provide workspace-specific command shortcuts.
- Avoid permanently polluting the global shell environment.
- Support both `la run <name>` and top-level invocation via `<name>`.
- Support both `alias` and `abbr`.
- Allow each task to define activation conditions.
- Keep configuration hand-editable and diff-friendly.

## 3. Non-Goals

- Replacing `make`, `task`, or CI job definitions.
- Embedding a general-purpose programming language into the config format.
- Automatically synchronizing config across users.
- Guaranteeing identical behavior across every shell.

## 4. Primary Use Cases

### 4.1 Explicit Invocation Through `la`

Users can run:

```sh
la run test
la run deploy
```

`la` resolves the current workspace configuration, finds the requested task, evaluates its conditions, and executes the command.

### 4.2 Top-Level Commands

Users can opt in via `.zshrc` or similar:

```sh
eval "$(la init zsh)"
```

After initialization, commands defined for the current workspace can be invoked directly:

```sh
$ test
$ deploy
```

This is meant for interactive shell usage where workspace-specific commands should feel native.

### 4.3 `abbr` Exposure

If the shell or shell plugin supports abbreviation-style expansion, a task may be exposed as `abbr` instead of a plain alias.

This is useful when:

- the user wants to edit the expanded command before execution
- alias semantics are too eager or too limited
- shell history should keep the expanded command

## 5. Terminology

- `workspace`
  The directory context used for config discovery.
- `task`
  A single command definition in configuration.
- `alias mode`
  A mode that exposes a task as a shell alias or shell function.
- `abbr mode`
  A mode that exposes a task through abbreviation expansion.
- `condition`
  A predicate that decides whether a task is active in the current context.

## 6. Command Model

### 6.1 `la run <name> [args...]`

Basic behavior:

1. Discover configuration by searching upward from the current directory.
2. Resolve the task named `<name>`.
3. Evaluate the task's activation conditions.
4. Execute the configured command, optionally forwarding extra arguments.

Expected error cases:

- the task does not exist
- the config is invalid
- the task exists but is currently disabled by its conditions
- shell integration is required but unavailable

### 6.2 `la init <shell>`

Print shell integration code for the target shell.

Examples:

```sh
eval "$(la init zsh)"
eval "$(la init bash)"
```

Responsibilities:

- define shell functions and hooks
- reevaluate workspace state as needed
- expose top-level commands as aliases, functions, or abbreviations depending on configuration
- remove previously registered definitions when they are no longer active

`la init <shell>` should be treated purely as a shell-integration code generator.
It should not directly perform workspace resolution or task registration on its own.
Instead, the emitted hook code should call back into the `la` binary during prompt-time reevaluation.

### 6.3 `la list`

List the tasks currently available in the active workspace.

Recommended output fields:

- name
- mode (`alias`, `abbr`, `command-only`)
- whether the task is enabled
- source config path
- short description

### 6.4 `la run <name> [args...]`

Execute a resolved task explicitly, without relying on top-level shell exposure.

This provides a stable execution path even when `<name>` is also exposed via `alias` or `abbr`.

### 6.5 `la doctor`

Validate the current configuration and environment, and report:

- parse errors
- naming conflicts
- unsupported shell features
- tasks that exist but are currently disabled by conditions

## 7. Shell Integration Model

### 7.1 Initialization

Recommended setup:

```sh
eval "$(la init zsh)"
```

The code emitted by `la init <shell>` should define at least a shell function named `la` and support:

- reflecting shell-state-changing subcommands in the current shell
- updating active task registrations on each prompt reevaluation
- delegating non-shell-mutating execution to the binary

Top-level task execution should primarily be implemented by pre-registering tasks as shell functions.

Recommended separation of responsibilities:

- `la init zsh` emits hook code only
- the shell hook runs before each prompt
- the hook calls back into the `la` binary to compute the current registration state
- the binary returns shell code describing the required updates
- the current shell applies those updates via `eval`

This follows the same general integration shape used by tools such as `direnv`: shell-specific hooking lives in the emitted integration code, while state resolution stays in the binary.

### 7.2 Workspace Activation

The shell should reevaluate workspace state whenever a prompt is displayed and update registered tasks as needed.

From the user's perspective, changes to configuration should be reflected by the next prompt.

Top-level registration policy:

- `alias mode` tasks are pre-registered as shell functions or aliases
- `abbr mode` tasks are registered as abbreviations when the shell supports that capability
- conflicts are checked before registration
- shell builtins always have highest priority and must not be overridden
- existing shell definitions and commands on `PATH` must not be silently overridden

The implementation may perform full reevaluation on every prompt or use change detection and incremental updates.

Regardless of optimization strategy, the externally visible behavior should prioritize correctness and prompt-to-prompt freshness.

### 7.3 Naming Conflicts

Tasks must not silently override existing shell builtins, aliases, functions, or executables.

Top-level task resolution priority:

1. shell builtins
2. workspace tasks
3. global tasks defined in `XDG_CONFIG_HOME/la/config.toml`

This priority describes task resolution order, not blanket override behavior for existing aliases, functions, or executables on `PATH`.

Accordingly, when registering workspace or global tasks as top-level commands, `la` should check for conflicts against existing aliases, functions, and executables, and by default avoid registering conflicting names.

General policy:

- warn by default
- report conflicts in `la doctor`
- allow explicit opt-in overriding via config
- never allow shell builtins to be overridden, even when override is enabled

Suggested policy values:

- `skip`
- `warn`
- `override`

The default should be `warn`.

## 8. Configuration Format

### 8.1 Recommended Format

The primary recommended format is `TOML`.

Rationale:

- fewer implicit typing surprises than YAML
- structured enough without becoming overly complex
- supports comments well
- maps cleanly onto Go structs

YAML may still be supported for compatibility, but TOML should remain the primary documented format.

Recommended filename:

- `.la.toml`

Compatibility candidates:

- `.la.yaml`
- `.la.yml`

### 8.2 Top-Level Structure

Example:

```toml
version = 1

[settings]
conflict_policy = "warn"

[task.test]
run = "go test ./..."
mode = "alias"
runtime = "subshell"
description = "Run the full test suite"

[task.gs]
run = "git status --short"
mode = "abbr"

[task.dev]
run = "docker compose up app"
mode = "alias"
runtime = "current-shell"

[task.dev.when]
git = true
exists_any = ["docker-compose.yml", "compose.yaml"]

[task.preview]
run = "pnpm dev"
mode = "abbr"
runtime = "subshell"
when = 'inside("apps/web") && exists("package.json") && has_command("pnpm") && shell() == "zsh"'
```

### 8.3 Task Schema

Each task may define:

- `run`: required, the command string to execute
- `description`: optional, a human-readable description
- `mode`: optional, `"alias" | "abbr" | "command-only"`, default `"alias"`
- `runtime`: optional, `"current-shell" | "subshell"`, default `"subshell"`
- `shell`: optional, an explicit shell to use
- `workdir`: optional, the working directory for execution
- `args`: optional, `"append" | "ignore" | "reject"`, default `"append"`
- `override`: optional, whether conflicts with aliases, functions, and executables may be overridden; default `false`
- `when`: optional, activation conditions expressed either as a table or as a CEL string

`runtime` semantics:

- `current-shell`
  Evaluate the task in the caller's current shell context. Side effects such as `cd`, `export`, `alias`, and `source` affect the caller.
- `subshell`
  Run the task in a child shell. Environment and directory changes do not affect the caller.

The default should be `subshell` for safety.

Tasks with `override = true` may be registered even when they conflict with aliases, functions, or executables.

Conflicts with shell builtins must never be allowed.

### 8.4 Condition Schema

`when` controls whether a task is active in the current context.

`when` supports two forms:

- table form
- CEL string form

### 8.4.1 Table Form

This is intended as a compact shorthand for simple conditions.

Built-in candidates:

- `git = true`
- `inside = ["path/glob", ...]`
- `exists = ["path", ...]`
- `exists_any = ["path", ...]`
- `env = { NAME = "value" }`
- `has_command = ["mise", "docker"]`
- `os = ["darwin", "linux"]`
- `shell = ["zsh", "bash", "fish"]`

Evaluation rules:

- different condition types are ANDed together
- `inside` matches when any listed candidate matches
- `exists` requires all listed paths to exist
- `exists_any` requires at least one listed path to exist
- `has_command` requires all listed commands to be available

Example:

```toml
[task.preview]
run = "pnpm dev"
mode = "abbr"

[task.preview.when]
exists = ["package.json"]
has_command = ["pnpm"]
shell = ["zsh"]
```

### 8.4.2 CEL String Form

For more expressive conditions, `when` may be written as a CEL expression string.

Example:

```toml
[task.preview]
run = "pnpm dev"
mode = "abbr"
when = 'inside("apps/web") && exists("package.json") && has_command("pnpm") && shell() == "zsh"'
```

The expression must ultimately evaluate to `bool`.

### 8.4.3 CEL Built-In Functions

CEL should expose built-in functions with semantics aligned to the table form.

Expected function set:

- `inside(path)`
- `exists(path1, path2, ...)`
- `exists_any(path1, path2, ...)`
- `has_command(name)`
- `env(name)`
- `shell()`
- `os()`
- `git()`

Example:

```toml
when = 'inside("apps/api") && has_command("mise") && env("CI") != "true"'
```

### 8.4.4 Shared Evaluation Model

Table-form conditions and CEL-form conditions should share the same underlying evaluation logic.

Principles:

- predicates such as `inside()`, `exists()`, and `has_command()` should have a single shared implementation
- table form should call that shared implementation directly
- CEL form should expose the same logic through CEL built-ins
- the implementation should not rely on fragile string conversion from table form into CEL source

This keeps the simple shorthand and the expressive form semantically aligned.

### 8.4.5 Exposure Model for Condition Primitives

Each condition primitive may be exposed in one of three ways:

- table form only
- CEL form only
- both table form and CEL form

General policy:

- common primitives should be available in both forms whenever practical
- table form is shorthand
- CEL form is the full expressive form
- the supported exposure model should be defined explicitly per primitive

For example:

- `inside`
  available in both table form and CEL form
- `exists`
  available in both table form and CEL form
- `exists_any`
  available in both table form and CEL form
- `has_command`
  available in both table form and CEL form
- future advanced predicates
  may be CEL-only

This cleanly separates condition semantics from the syntax used to expose them.

### 8.4.6 Implementation Guidance for Condition Primitives

Internally, each condition primitive should separate at least the following responsibilities:

- the predicate's core evaluation logic
- how table-form input is decoded
- how the primitive is exposed to CEL

This makes it easier to:

- vary table/CEL support per primitive
- report capabilities in `la doctor`
- generate docs or completion metadata from condition definitions

### 8.4.7 Path Resolution Semantics

Path-based predicates should interpret paths relative to the workspace root by default.

For example, `inside("apps/web")` checks whether the current directory is located under `apps/web` within the workspace root.

## 9. Execution Semantics

### 9.1 Command Execution

Unless an explicit `shell` is configured, `run` is evaluated using the current shell.

General policy:

- top-level `alias` and `abbr` exposure should preserve natural interactive shell behavior
- `la run` should execute via a shell, not via naive whitespace splitting

### 9.2 Runtime Modes

Task execution is controlled by `runtime`.

- `runtime = "subshell"`
  Run in a child shell. This should be the default for normal command execution.
- `runtime = "current-shell"`
  Evaluate in the current shell. This is intended for tasks with shell-state side effects.

Typical `current-shell` examples:

- `cd apps/web`
- `export FOO=bar`
- `source ./scripts/env.sh`
- commands that rely on aliases or functions defined in `.zshrc`

### 9.3 Inheriting Existing Shell State

All task executions should inherit environment variables from the caller's shell.

Tasks with `runtime = "current-shell"` should additionally be able to use aliases and functions already defined in the caller's shell context.

For example, if the user has `g='git'` defined in their shell, then `run = "g status"` is considered valid for a `current-shell` task.

For `runtime = "subshell"`, environment variables are inherited, but aliases and functions are not guaranteed to be available.

### 9.4 Argument Handling

By default, extra CLI arguments are appended to `run`.

Example:

```toml
[task.k]
run = "kubectl"
mode = "abbr"
args = "append"
```

Then:

```sh
la k get pods
```

behaves like:

```sh
kubectl get pods
```

### 9.5 Exit Status

The exit code of the final process must be propagated back to the caller.

## 10. `alias` and `abbr`

### 10.1 Alias Mode

Best suited for:

- short, fixed commands
- commands that should execute as defined without pre-editing
- shells where alias-like exposure is broadly supported

### 10.2 Abbr Mode

Best suited for:

- commands the user wants to edit before execution
- argument-heavy commands
- workflows where the expanded command should appear in shell history

Because `abbr` support is shell-specific, it should be treated as a capability with fallbacks:

- preferred: native abbreviation support from the shell or plugin
- fallback: an approximation using shell functions or similar mechanisms
- final fallback: report the feature as unsupported while keeping the task available through `la run`

## 11. Config Discovery

Configuration should be discovered by searching upward from `$PWD`.

If multiple supported config files exist in the same directory, the precedence is:

1. `.la.toml`
2. `.la.yaml`
3. `.la.yml`

The first match wins.

### 11.1 Global Config

User-level global configuration should be supported at `XDG_CONFIG_HOME/la/config.toml`.

This file is intended for common tasks not tied to any specific workspace.

### 11.2 Relationship Between Workspace and Global Config

Task resolution priority:

1. shell builtins
2. tasks from the current workspace
3. tasks from `XDG_CONFIG_HOME/la/config.toml`

This applies to both `la run <name>` and top-level command exposure.

If a task name exists in both workspace and global config, the workspace definition wins.

## 12. Security and Safety

- `la` is a local convenience tool; commands are assumed to be user-authored and trusted
- `la init` must only emit shell integration code and must not execute workspace commands during initialization
- entering a directory must not automatically execute tasks
- prompt-time updates must be limited to registration and deregistration

## 13. Suggested MVP

The first usable version should support at least:

- `la run <name>`
- `la init zsh`
- upward config discovery
- `.la.toml`
- `run`
- `mode = "alias" | "abbr"`
- `runtime = "current-shell" | "subshell"`
- table-form `when`
- CEL-form `when`
- condition primitive exposure metadata
- shared evaluation primitives such as `inside()`, `exists()`, and `has_command()`
- prompt-based reevaluation and task re-registration

Can be deferred:

- robust multi-shell support
- complex boolean syntax beyond the current condition model
- remote or shared config sources

## 14. Open Questions

- whether `abbr` should use shell-native support, plugin integration, or both
- whether top-level commands should ultimately be implemented as aliases, shell functions, or generated dispatchers
- whether non-string command templates or placeholder expansion should be supported
- how much caller shell state, beyond environment variables, should be preserved for `subshell`

## 15. Internal Model

The implementation should be centered on an internal model that is independent from config file syntax.

### 15.1 Task

`Task` represents one executable unit.

Minimum responsibilities:

- name
- command string
- exposure mode (`alias` / `abbr` / `command-only`)
- runtime mode (`current-shell` / `subshell`)
- condition
- override policy
- description

### 15.2 Condition

`Condition` is the abstraction that determines whether a task is active in the current context.

Design principles:

- conditions should share a common evaluable interface
- table form and CEL form should both normalize into the same evaluation model
- implementations such as `TableCondition`, `CELCondition`, and `TrueCondition` are acceptable

### 15.3 EvalContext

`EvalContext` represents runtime context needed for condition evaluation.

Minimum fields:

- workspace root
- current working directory
- shell
- environment variables

It may later grow to include OS information, Git state, and other derived context.

### 15.4 Registry

`Registry` stores tasks for resolution.

At minimum it should support:

- workspace tasks
- global tasks

Resolution must follow the precedence rules defined by this specification.

### 15.5 Runner

`Runner` is responsible for task execution.

Minimum responsibilities:

- branching between `current-shell` and `subshell`
- inheriting environment variables
- applying extra arguments
- propagating exit status

### 15.6 Shell Integration

Shell integration is responsible for registration and synchronization of top-level task exposure.

Minimum responsibilities:

- prompt-time reevaluation
- top-level task registration
- conflict checks against existing definitions
- deregistration of previously registered tasks

### 15.7 Architectural Separation

The implementation should separate at least these three layers:

- config parsing
- normalization into the internal model
- task resolution, condition evaluation, and shell integration

With that separation in place, config file syntax can be replaced or extended later without forcing a redesign of core behavior.

## 16. Implementation Direction

Recommended order of implementation:

1. define the internal model around `Task`, `Condition`, `EvalContext`, `Registry`, and `Runner`
2. implement condition primitives and the shared evaluation layer
3. implement runtime mode handling
4. implement `la run`, `la list`, and `la doctor` against the internal model
5. implement zsh prompt reevaluation and top-level task registration
6. implement conditional activation
7. add config parsing and upward config discovery
8. add `abbr` support and fallback behavior
9. expand shell support and config layering as needed
