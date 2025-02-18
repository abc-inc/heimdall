# Contributing

## Heimdall project layout

At a high level, these areas make up the `github.com/abc-inc/heimdall` project:

- [`cmd/`](../cmd) - `main` packages for building the `heimdall` executable
- [`cli/`](../cli) - most other CLI code, including different output formats and the REPL
- [`docs/`](../docs) - documentation for maintainers and contributors
- [`plugin/`](../plugin) - contains sub-commands
- [`res`](../res) - provides utilities for all kinds of file operations

## Add a new command

1. Create a new file for the new command, e.g. `plugin/<type>/custom.go`
2. The new command should expose a method, e.g. `NewCustomCmd()`, that returns a `*cobra.Command`.
   * Any logic specific to this command should be kept within the package and not added to any "global" package.
3. Use the method from the previous step to generate the command and add it to the command tree,
   typically somewhere at the root command.

## How to write tests

This task might be tricky.
Typically, Heimdall commands do things like look up information from remote files.
Moreover, one does not want to test the CLI framework, rather than the command itself.
To avoid that, you may want to define function, which do not contain any CLI-related code.
Export those functions, if necessary.

To make your code testable, write small, isolated pieces of functionality that are designed to be composed together.
Prefer table-driven tests for maintaining variations of different test inputs and expectations
when exercising a single piece of functionality.
