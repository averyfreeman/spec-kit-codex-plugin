# Upstream sources and pin updates

## Pinned sources

| Submodule | Upstream | Ref | Commit |
| --- | --- | --- | --- |
| Spec Kit | [`github/spec-kit`](https://github.com/github/spec-kit) | `v1.3.2` | `55264c2f2f2d1146f02abc543cb913dcdf1f9342` |
| Brownfield extension | [`Quratulain-bilal/spec-kit-brownfield`](https://github.com/Quratulain-bilal/spec-kit-brownfield) | `v1.0.0` | `7479c5206a573e95890cbc252dd30806121a44bb` |

The exact pins are duplicated in `plugins/spec-kit-codex/upstream-pins.json` for diagnostics. Git's recorded submodule commits remain authoritative. Both repositories are MIT licensed. The Brownfield manifest requires Spec Kit `>=0.4.0`, which is satisfied by Spec Kit 1.3.2.

## Update process

1. Select a tagged upstream release; do not move to an unreviewed branch head by default.
2. Update one submodule at a time using an explicit tag, then record its peeled commit SHA in `upstream-pins.json`.
3. When changing Spec Kit, update the exact `specify-cli` version in the plugin-owned `pyproject.toml` to match the submodule's `[project].version`, then regenerate and review `plugins/spec-kit-codex/uv.lock` with `uv lock`. The lock file uses the local submodule path and leaves its files untouched.
4. Review the upstream diff and extension compatibility metadata. Keep local modifications out of both submodules.
5. Check the workflow allowlist and run the package, Go, MCP, and CLI smoke checks in `docs/verification.md`.
6. Update this table and the diagrams only if the architecture or workflow paths changed; commit the parent repository's gitlink update and metadata together.

Example:

```sh
git -C plugins/spec-kit-codex/upstream/spec-kit-official fetch --tags origin
git -C plugins/spec-kit-codex/upstream/spec-kit-official checkout --detach v1.3.2
```

Do not edit, patch, or commit changes inside either upstream submodule.
