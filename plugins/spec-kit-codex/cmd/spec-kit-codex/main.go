// Command spec-kit-codex exposes a small, read-only MCP interface over pinned
// Spec Kit workflow references. Project writes remain the responsibility of
// the upstream CLI and the user's coding agent.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	pluginRoot := flag.String("plugin-root", "", "absolute path to the installed spec-kit-codex plugin")
	flag.Parse()

	root, err := resolvePluginRoot(*pluginRoot)
	if err != nil {
		log.Fatal(err)
	}

	server := newServer(root)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func resolvePluginRoot(root string) (string, error) {
	if root == "" || !filepath.IsAbs(root) {
		return "", fmt.Errorf("--plugin-root must be an absolute path")
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve plugin root: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("stat plugin root: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("plugin root is not a directory")
	}
	return resolved, nil
}

func newServer(pluginRoot string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:        "spec-kit-codex",
		Title:       "Spec Kit for Codex",
		Description: "Read-only access to pinned Spec Kit workflow references and local project setup status.",
		Version:     "0.1.0",
	}, nil)

	readOnly := &mcp.ToolAnnotations{
		ReadOnlyHint:    true,
		DestructiveHint: boolPointer(false),
		OpenWorldHint:   boolPointer(false),
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "spec_kit_plugin_info",
		Title:       "Spec Kit plugin info",
		Description: "Return this plugin's version, exact upstream pins, and the safe command shape for initializing Spec Kit in the current project. This tool does not execute the command or modify files.",
		Annotations: readOnly,
	}, pluginInfoHandler(pluginRoot))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "spec_kit_check_environment",
		Title:       "Check Spec Kit prerequisites",
		Description: "Check required local executables, the locked CLI runner project, pinned workflow files, declared Brownfield version compatibility, and exact submodule commits when Git metadata is available. This runs only fixed read-only Git revision queries; it never runs project-supplied commands.",
		Annotations: readOnly,
	}, environmentHandler(pluginRoot))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "spec_kit_project_status",
		Title:       "Check project Spec Kit status",
		Description: "Check a selected absolute project root for a .git marker, Spec Kit metadata, a constitution, Codex skills, and specs directory. Reads only these fixed path names and returns presence flags; it does not validate Git internals.",
		Annotations: readOnly,
	}, projectStatusHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "spec_kit_workflow_source",
		Title:       "Read a pinned workflow reference",
		Description: "Return one allowlisted upstream Spec Kit or Brownfield workflow reference by name. This reads only the pinned Markdown command file and cannot access arbitrary paths.",
		Annotations: readOnly,
	}, workflowSourceHandler(pluginRoot))

	return server
}

func boolPointer(value bool) *bool { return &value }
