# Architecture

## Shape of the plugin

The repository is a Codex marketplace root. `.agents/plugins/marketplace.json` points to `plugins/spec-kit-codex`, which contains the portable plugin manifest, MCP configuration, skills, Go server, and the two upstream Git submodules.

The plugin uses the current portable Agent Plugins package shape: root `plugin.json`, root `mcp.json`, and `skills/`. Codex's repo-scoped marketplace manifest lives at `.agents/plugins/marketplace.json`. A local MCP server runs over stdio and needs a local executor. See the official [plugin packaging guide](https://developers.openai.com/plugins/build/plugins), [Codex marketplace format](https://learn.chatgpt.com/docs/enterprise/plugin-management), and [desktop-only plugin guidance](https://developers.openai.com/plugins/deploy/desktop-only).

## Why the shim is small

The upstream Spec Kit release includes its own CLI and an experimental stdio MCP server. The current server exposes version information, artifact inspection, and workflow listing; it does not expose the nine core SDD command templates or the four Brownfield commands as a unified Codex workflow. This plugin therefore uses a small Go MCP server to:

- report plugin/upstream metadata and local runtime availability;
- report a bounded, read-only status for an explicitly selected project root;
- return an allowlisted upstream command document from one of the pinned submodules.

The Go server does not run commands found in project files and has no project-write tool. Initialization uses the pinned upstream `specify` CLI; the skills guide the model through the command templates and require review before project edits. This avoids maintaining a second implementation of Spec Kit's lifecycle.

## Languages and responsibilities

| Pipeline part | Language or format | Responsibility |
| --- | --- | --- |
| Local marketplace/plugin metadata | JSON | Connect the repo marketplace, plugin, skills, and stdio server. |
| MCP shim | Go | Typed workflow lookup, installation diagnostics, version compatibility checks, and read-only project status. |
| Spec Kit CLI runtime | Python project metadata + `uv.lock` | Resolves the pinned upstream Specify CLI as a local dependency and locks Python dependencies without editing the submodule. |
| Spec Kit CLI | Python | Project initialization and upstream CLI behavior, invoked through the plugin-owned `uv` project. |
| Codex workflow guidance | Markdown | Greenfield and Brownfield routing, user questions, review gates, and verification steps. |
| Brownfield extension | Markdown and YAML | Original scan/bootstrap/validate/migrate prompts and extension metadata. |

No TypeScript or custom Python bridge is needed in this design. Go is used for new executable behavior; Python remains the upstream runtime because the pinned Specify CLI is Python. The declarative `pyproject.toml` and `uv.lock` are outside the submodule and use a local path dependency, so runtime setup does not change upstream files. The Go MCP SDK is the only direct Go dependency.

## Workflow boundaries

See the [greenfield and brownfield workflow diagram](architecture-diagram.md). In brief: skills select the path; the Go server provides source instructions and read-only status; the upstream CLI initializes `.specify/` and Codex skills; Codex performs approved file changes. Brownfield scan and validation are read-only, while bootstrap and migration require an explicit approval step.

## Submodule and package behavior

The Go server maps each allowed workflow name to a fixed Markdown file inside the two submodules. A clone without initialized submodules cannot serve these files. Clone with `--recurse-submodules`; the Codex local marketplace installs the plugin package from the checked-out repository, including the pinned source files.

The local MCP server runs `go run` from the plugin root and therefore needs Go 1.25+ on first launch. Its module cache is managed by Go. The pinned upstream CLI is selected through a plugin-owned `pyproject.toml` and `uv.lock`; this keeps the upstream submodule unchanged and resolves dependencies into the plugin's ignored `.venv/` on first use. The environment check reads Spec Kit's declared version and the extension's minimum version; it also checks the two submodule HEADs against their pins when Git metadata is available. Neither runtime needs an API key or an application-hosted service.
