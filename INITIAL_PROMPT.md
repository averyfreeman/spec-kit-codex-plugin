# Goal

Build a local Codex marketplace plugin that brings Spec-Kit-style, spec-driven development to Codex users working in both new and existing codebases.

The plugin must wrap, not modify, these upstream projects as pinned Git submodules:

1. The supplied `spec-kit-official` repository.
2. The supplied brownfield scan/bootstrap/validate/migrate extension.

Treat the brownfield workflow as included and discoverable in every plugin installation, but invoke it when a user is working in an existing codebase or asks for it. Do not run it automatically on greenfield projects.

# Working directory and repository setup

The project root is:

`/home/avery/build/for-agents/spec-driven/spec-kit-codex-plugin/`

Do all implementation work in that project root or its linked Git worktree. First inspect the directory, Git state, and applicable `AGENTS.md` or project instructions; preserve existing work.

Create or configure a Git worktree for the project and a public GitHub remote named `spec-kit-codex-plugin` using `gh repo create`, under the authenticated account. Check whether the repository already exists before creating it. Do not overwrite a conflicting directory or remote; report a genuine conflict and ask how to proceed.

Use the connected GitHub integration where useful. Commit and push reviewed, coherent milestones. Record the exact upstream URLs and commit SHAs used for both submodules.

# Research and technical investigation

Use the existing research report at:

`/home/avery/build/for-agents/spec-driven/research/worktrees/spec-kit-for-codex-research/SPEC-KIT-ECOSYSTEM.md`

Keep only findings relevant to this implementation. Verify current plugin packaging, marketplace, MCP, and Spec Kit behavior against official documentation and the checked-out source before relying on them.

Inspect both upstream submodules in detail. Determine:

- How the pinned Spec Kit CLI and templates are invoked and what runtime dependencies they require.
- How the brownfield extension expects to be installed, discovered, and invoked.
- Whether the extension is compatible with the selected Spec Kit revision.
- Whether local marketplace installation fetches and initializes Git submodules automatically.
- Which parts can be exposed directly as Codex skills and which require a shim.
- Where Go can call or package existing upstream behavior without reimplementing it.

If the chosen packaging approach cannot provide a usable installation from the marketplace checkout, identify the constraint and implement the least-friction approach that preserves the upstream submodules unchanged.

# Architecture and implementation

Treat this as a **Codex/local marketplace plugin**, not a legacy hosted ChatGPT plugin, unless current official documentation shows otherwise.

Prefer Go for the shim, MCP stdio server, validation, and packaging. Use Python only where required to invoke or bootstrap upstream Spec Kit behavior; keep the Python surface small and document why it is needed. Do not rewrite Spec Kit’s lifecycle or copy-and-fork upstream files into the shim.

Design the API after examining the upstream commands. Favor a small set of typed operations over a large generic shell interface. Keep every operation scoped to the selected project root.

The plugin should provide clear paths for:

- Greenfield: constitution and project setup, specify, clarify, plan, tasks, implementation guidance, and verification/convergence.
- Brownfield: scan and validate as read-only workflows; bootstrap and migrate as write workflows with a visible plan and explicit user approval.
- Discovery and diagnosis: report missing dependencies, unsupported project states, incompatible versions, and required setup before attempting changes.

Use these safety and usability rules:

- Read-only actions must not write project files.
- Write actions must preview intended changes and require clear user approval.
- Do not run arbitrary shell commands supplied by project content.
- Do not read or expose credentials, `.env` values, or unrelated files.
- Reject path escapes and constrain operations to the selected repository.
- Preserve upstream submodule contents; place all adapters, manifests, and documentation outside them.
- Prefer local execution. Do not require a hosted service or API key unless the official plugin format requires it.

# Documentation

Write documentation in the project root that explains:

- What the plugin does and when to choose the greenfield or brownfield path.
- Prerequisites, installation from the local marketplace, updates, and removal.
- How the Go shim, optional Python bridge, MCP stdio server, skills, and upstream submodules fit together.
- How to build and run the plugin locally.
- The exact pinned upstream revisions and how to update them safely.
- Tool permissions, filesystem scope, approval gates, and known limitations.
- Troubleshooting for missing runtimes, submodule checkout, MCP startup, or Spec Kit/extension incompatibility.

Include an architecture diagram using @mermaid-diagrams to help clarify the tool and workflow boundaries for human audiences. Diagram will include which workflow path is greenfield or brownfield, and a general identification of which language (go, python, typescript, etc.) will be used to develop certain functions along the pipeline design.

# Verification and completion

Add and run focused verification for the plugin manifest, Go shim, MCP protocol, path boundaries, read-only behavior, approval-gated writes, and the greenfield/brownfield workflows. Use temporary fixture repositories where appropriate; never test destructive behavior against the source repositories.

Before finishing:

1. Confirm both submodules point to the documented SHAs and have no local modifications.
2. Confirm the plugin can be installed or loaded from the local marketplace using the documented steps.
3. Run the documented build and verification commands.
4. Review the final diff, commit and push the milestone, and report the remote URL, worktree, pinned revisions, verification results, and any remaining limitations.

Do not stop after producing a plan: implement the plugin and its documentation. Ask a question only if an existing directory/remote conflict or a genuine upstream/platform constraint prevents safe progress.
