# Build and verification

From `plugins/spec-kit-codex`:

```sh
uv lock --check
go test ./...
mkdir -p bin
go build -o bin/spec-kit-codex ./cmd/spec-kit-codex
```

The Go suite checks JSON manifests, submodule revisions and cleanliness, pinned workflow files, skill approval guidance, MCP tool schemas and read-only annotations, project-root validation, symlink escape rejection, and a stdio MCP initialize/list/call round trip. It uses only temporary fixture projects for path and behavior checks.

After local marketplace installation, exercise the exact cached package copy by setting `SPEC_KIT_INSTALLED_PLUGIN_ROOT` to the installed plugin root printed by `codex plugin add`, then rerun `go test ./...`. The stdio test starts the server from that copy and reads both upstream command sources; it tolerates unavailable Git metadata in the cache while still requiring the files and version compatibility data.

From the repository root, run the CLI smoke check only from a disposable temporary project, never from the plugin or either submodule. Python's temporary-directory context cleans up the fixture even if the CLI fails:

```python
import subprocess
import tempfile

plugin_root = "/absolute/path/to/spec-kit-codex-plugin/plugins/spec-kit-codex"
command = [
    "uv", "run", "--locked", "--project", plugin_root, "specify", "init",
    "--here", "--non-interactive", "--force", "--ignore-agent-tools",
    "--integration", "codex", "--integration-options=--skills",
]
with tempfile.TemporaryDirectory(prefix="spec-kit-codex-smoke-") as fixture:
    subprocess.run(command, cwd=fixture, check=True)
```

The fixture is empty, so `--force` only suppresses the initializer's prompt in this isolated smoke test. Do not use it in the plugin checkout, either upstream submodule, or a user project without first showing the write plan and receiving approval. After the smoke check, verify the temporary directory is removed and both submodules remain clean.

Finally, register and install the local marketplace as documented in `docs/installation.md`, then confirm Codex lists both skills and the MCP server starts. Static manifest checks and MCP transport tests do not prove every Codex UI version renders the plugin identically; record the tested CLI/app version.
