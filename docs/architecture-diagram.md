# Spec Kit for Codex: architecture and workflow routing

Shows the local Codex plugin's language boundaries and its greenfield and brownfield paths.

## Plugin architecture and language boundaries

<!-- mermaid:id=plugin-architecture -->
```mermaid
flowchart LR
  marketplace["Local Codex marketplace\nJSON manifest"] --> plugin["Spec Kit plugin\nportable plugin.json + mcp.json"]
  subgraph plugin_parts["Plugin package"]
    skills["Codex skills\nMarkdown workflow guidance"]
    shim["Read-only MCP shim\nGo + official MCP SDK"]
  end
  subgraph upstream["Pinned upstream submodules"]
    specify["GitHub Spec Kit\nPython CLI + templates"]
    brownfield["Brownfield extension\nMarkdown commands + YAML metadata"]
  end
  subgraph target["Selected target project"]
    init["specify init\nPython CLI via uv"]
    artifacts[".specify/ and specs/\nCodex-approved project artifacts"]
  end
  skills -->|"read-only tool calls"| shim
  shim -->|"fixed command references"| specify
  shim -->|"fixed command references"| brownfield
  skills -->|"reviewed setup command"| init
  init --> artifacts
  skills -->|"approved edits"| artifacts
```

## Greenfield and brownfield workflow gates

<!-- mermaid:id=greenfield-brownfield-routing -->
```mermaid
flowchart TD
  request["User request"] --> status["Project status\nGo MCP, read-only"]
  status --> route{"Existing source code?\n.specify present?"}
  route -->|"new or uninitialized"| green["Greenfield skill\nMarkdown"]
  green --> initReview{"Review setup targets\nand approve?"}
  initReview -->|"yes"| init["specify init\nPython CLI via uv"]
  init --> core["specify → clarify → plan → tasks → implement"]
  core --> verify["Analyze and converge\nCodex tools + project checks"]
  route -->|"existing project"| scan["Brownfield scan\nread-only"]
  scan --> profile["Review evidence-based\nproject profile"]
  profile --> existingSpec{".specify/ already present?"}
  existingSpec -->|"yes"| bootstrapGate{"Approve bootstrap\nfile plan?"}
  existingSpec -->|"no"| brownfieldInitReview{"Review init targets\nand approve?"}
  brownfieldInitReview -->|"yes"| brownfieldInit["specify init\nPython CLI via uv"]
  brownfieldInit --> bootstrapGate
  bootstrapGate -->|"yes"| bootstrap["Bootstrap Spec Kit artifacts\nCodex file tools"]
  bootstrap --> validate["Brownfield validate\nread-only"]
  validate --> migrateGate{"Select one feature\nand approve migration plan?"}
  migrateGate -->|"yes"| migrate["Migrate spec, plan, tasks\nCodex file tools"]
  migrate --> verify
```
