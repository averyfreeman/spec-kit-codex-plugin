# Troubleshooting

## Marketplace does not show the plugin

- Confirm `git submodule update --init --recursive` completed. The workflow source tool reads allowlisted Markdown files from the pinned submodules.
- If `spec_kit_check_environment` reports a known commit mismatch, run `git submodule update --init --recursive` from the marketplace checkout to restore the parent repository's recorded submodule commits. If the check is unknown in a copied plugin cache, check the original marketplace checkout's `git submodule status`.
- Register the repository root (the directory containing `.agents/plugins/marketplace.json`), not `plugins/spec-kit-codex`.
- Run `codex plugin marketplace list` and `codex plugin list` to check registration and installation. Reinstall or refresh after changing the local checkout if the host is showing a cached copy.

## MCP server fails to start

- Check `go version`; the module declares Go 1.25 or newer.
- From `plugins/spec-kit-codex`, run `go test ./...`. The suite launches the MCP server through stdio and checks initialize, tool listing, and calls. For a basic startup-only smoke check, run `go run ./cmd/spec-kit-codex --plugin-root "$PWD" </dev/null`; the empty input closes the server after startup.
- If Go cannot download the MCP SDK, check module proxy/network access and the local Go module cache. No API key is required.
- Check the host's MCP/plugin logs for stderr output. A missing or non-absolute `${PLUGIN_ROOT}` is reported on stderr and the process exits.

## Spec Kit initialization fails

- Check `uv --version` and `python --version`; the pinned Specify CLI requires Python 3.11 or newer.
- Confirm the target project is the current working directory when invoking `specify init --here` and that the `--project` argument points to the plugin root. Its `pyproject.toml` and `uv.lock` select the pinned `spec-kit-official` submodule as the CLI source.
- For network or lockfile errors, run the command again after checking access to the Python package index. `uv` may create an ignored `.venv` inside the plugin directory; the upstream submodule stays untouched.
- For a non-empty target, inspect every planned merge target. Use `--force` only after the user approves the specific merge; it can overwrite matching files.

## Brownfield workflow is unavailable or mismatched

- Confirm `plugins/spec-kit-codex/upstream/spec-kit-brownfield` is at the recorded `v1.0.0` pin and that the main Spec Kit submodule is at `v1.3.2`.
- The extension declares a minimum Spec Kit version of `0.4.0`, which the pinned CLI satisfies. If changing either pin, review `extension.yml`, the command file paths, and the workflow compatibility before updating `upstream-pins.json`.
- The optional `after_init` hook is intentionally not installed. Brownfield adoption remains selectable through the bundled skill and does not run automatically.

## Project status is incomplete

`spec_kit_project_status` reports presence of a short list of markers; it does not validate Git internals, inspect project source, or prove that an existing Spec Kit installation is healthy. Use the reported markers as routing hints and inspect only the selected project files needed for the task.
