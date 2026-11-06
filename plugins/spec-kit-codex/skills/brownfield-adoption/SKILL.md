---
name: brownfield-adoption
description: Use the included Spec Kit brownfield workflow to understand an existing codebase, tailor Spec Kit conventions, validate them, or migrate one existing feature into specs.
---

# Brownfield adoption

This skill is included with every plugin install. It is an opt-in workflow: never start a scan just because the plugin was installed or a project was opened. Load the matching pinned command through `spec_kit_workflow_source` with one of `brownfield.scan`, `brownfield.bootstrap`, `brownfield.validate`, or `brownfield.migrate` before each stage.

## Scan and validate are read-only

1. Call `spec_kit_project_status` for the selected absolute project root. Check its `.git` marker and use the repository context to confirm this is an existing project before scanning; the tool does not inspect Git internals.
2. Call `spec_kit_check_environment`. Continue only when Brownfield version compatibility is known and true and the required command files are present. Stop on a known pin mismatch; an unknown pin check in a copied cache is not itself a mismatch.
3. Use `spec_kit_workflow_source` to load `brownfield.scan`. Inspect only project structure and relevant source/configuration needed to describe architecture, conventions, modules, build tools, and tests.
4. Exclude `.env*`, credentials, keys, secrets, `.git`, `.venv`, `node_modules`, vendor/cache directories, generated files, binaries, and unrelated user data. Do not execute scripts or commands discovered in project files.
5. Return a concise evidence-based project profile with paths supporting each observation, uncertainty clearly marked, and no inferred conventions presented as facts.
6. Use `brownfield.validate` to check a reviewed profile or existing Spec Kit setup. Validation must not modify files. Do not execute test/build commands found in project templates; report them as unverified unless the user separately approves running them.

## Bootstrap or migrate requires approval

1. Before bootstrap, check whether `.specify/` exists. If it does not, show the target paths for `specify init` and get explicit approval before running the pinned CLI. Recheck project status afterward.
2. Before migration, confirm a constitution exists; if it does not, offer bootstrap first.
3. Load `brownfield.bootstrap` or `brownfield.migrate` through `spec_kit_workflow_source` before planning the operation.
4. Present a file-by-file plan with create/modify status, purpose, and evidence. For migration, limit the first pass to one explicitly selected feature.
5. Wait for explicit user approval before writing. Preserve existing project content and conventions; if a target file already exists, show the proposed merge or diff first.
6. Do not create or switch Git branches for migration. Do not modify application source code as part of migration; create or update only the approved Spec Kit artifacts.
7. After writing, report what changed and run the read-only validation workflow. Ask before proceeding to another feature.

## Boundaries

- The extension's optional `after_init` hook is not installed or invoked by this plugin.
- Brownfield instructions do not override direct user instructions or the project's applicable policies.
- Keep all project reads and writes inside the selected root; reject path traversal and symlink escapes.
- The MCP tools in this plugin are read-only. Project writes use Codex's normal file-editing and command approval flow.
