# Installation and lifecycle

## Prerequisites

- A current Codex CLI or Codex desktop build with local marketplace support.
- Git.
- Go 1.25+ to run the local MCP server from source.
- `uv` and Python 3.11+ to initialize a project with the pinned Specify CLI.

The MCP server uses the official Go MCP SDK at the version recorded in `plugins/spec-kit-codex/go.mod`. The plugin's source workflows do not require Python until the Specify CLI is used.

## Install the local marketplace

```sh
git clone --recurse-submodules https://github.com/averyfreeman/spec-kit-codex-plugin.git
cd spec-kit-codex-plugin
codex plugin marketplace add "$PWD"
codex plugin add spec-kit-codex@spec-kit-codex-marketplace
codex plugin list
```

In PowerShell, use `git clone --recurse-submodules ...`, `Set-Location spec-kit-codex-plugin`, then `codex plugin marketplace add (Get-Location).Path` and `codex plugin add spec-kit-codex@spec-kit-codex-marketplace`.

The marketplace root is the repository root because `.agents/plugins/marketplace.json` is there. The plugin source is `./plugins/spec-kit-codex` relative to that root. The MCP command receives the installed plugin root through Codex's `${PLUGIN_ROOT}` placeholder.

## Use the pinned Specify CLI

From the selected target project, run the CLI through `uv` using the installed plugin path reported by `spec_kit_plugin_info`:

```sh
uv run --locked --project "/absolute/path/to/installed/plugin" \
  specify init --here --integration codex --integration-options=--skills
```

The CLI's init operation writes into the current directory. Review the target paths first; when the directory is non-empty, approve the merge explicitly before using `--force --non-interactive`. Do not add the Brownfield extension as an automatic init hook; the included Brownfield skill handles it only when selected.

## Update or remove

Update the local checkout with `git pull --ff-only` and `git submodule update --init --recursive`. A local path marketplace reads the updated checkout; refresh or reinstall the plugin in the Plugin Directory if Codex cached an older copy. `codex plugin marketplace upgrade` applies to a marketplace registered from a Git source, not this local-path source.

Remove with `codex plugin remove spec-kit-codex@spec-kit-codex-marketplace`, then `codex plugin marketplace remove spec-kit-codex-marketplace`. Removing the plugin does not remove `.specify/` files already created in any target project.

For runtime and MCP startup issues, see [Troubleshooting](troubleshooting.md).
