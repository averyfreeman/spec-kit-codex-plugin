# Spec Kit for Codex

A local Codex marketplace plugin for using GitHub Spec Kit in both new and existing codebases. The plugin packages two unmodified, pinned upstream repositories, two Codex skills, and a small Go MCP server that serves allowlisted workflow instructions and read-only project status.

The Go server does not edit projects or execute project commands. Spec Kit's Python CLI remains the authority for initialization; a plugin-owned `pyproject.toml` and `uv.lock` run the pinned CLI without changing the upstream submodule. The Codex skills guide the agent through specification, planning, implementation, and brownfield adoption with review gates.

## Install from this local marketplace

Requirements: Codex CLI or Codex desktop with local plugin marketplaces enabled, Git with submodule support, Go 1.25 or newer, and `uv` plus Python 3.11 or newer when initializing a project with the pinned Specify CLI.

```sh
git clone --recurse-submodules https://github.com/averyfreeman/spec-kit-codex-plugin.git
cd spec-kit-codex-plugin
codex plugin marketplace add "$PWD"
codex plugin add spec-kit-codex@spec-kit-codex-marketplace
codex plugin list
```

`--recurse-submodules` is required because the workflow references point into the upstream submodules. If you already cloned without it, run `git submodule update --init --recursive` before installing or refreshing the plugin.

To remove the plugin and marketplace later:

```sh
codex plugin remove spec-kit-codex@spec-kit-codex-marketplace
codex plugin marketplace remove spec-kit-codex-marketplace
```

For the exact Codex marketplace format and install behavior, see [Installation](docs/installation.md).

## Choose a workflow

- **New project or new feature:** invoke `$spec-driven-development`. The skill guides constitution, specify, clarify, plan, tasks, implementation, and convergence using the pinned Spec Kit command templates.
- **Existing codebase adoption:** invoke `$brownfield-adoption`. Scan and validate are read-only; bootstrap and migrate require a file-by-file plan and your approval. The skill is available in every installation but never runs automatically.

The plugin does not install the brownfield extension into each project or enable its optional `after_init` hook. It exposes the original command instructions through the plugin so the extension remains available without copying or altering its source.

## Initialize Spec Kit in a target project

After reviewing the target root and approving the intended changes, run the pinned CLI from that root:

```sh
uv run --locked --project "/path/to/spec-kit-codex-plugin/plugins/spec-kit-codex" \
  specify init --here --integration codex --integration-options=--skills
```

For a non-empty project, inspect the CLI's merge targets first. Use `--force --non-interactive` only after approving the merge; `--force` can overwrite matching template files. The plugin's skill will route an existing codebase through Brownfield scan and review before bootstrap or migration.

## What's included

- [Architecture](docs/architecture.md) and [architecture diagram](docs/architecture-diagram.md)
- [Security and approval boundaries](docs/security.md)
- [Pinned upstream revisions and update procedure](docs/upstream.md)
- [Build and verification](docs/verification.md)
- [Troubleshooting](docs/troubleshooting.md)
- The original project brief in [`INITIAL_PROMPT.md`](INITIAL_PROMPT.md)
