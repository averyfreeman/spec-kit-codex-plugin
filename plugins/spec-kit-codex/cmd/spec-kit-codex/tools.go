package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxWorkflowBytes = 64 * 1024

type pin struct {
	URL    string `json:"url"`
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
}

type pinsFile map[string]pin

type pluginManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type pluginInfo struct {
	Name                      string   `json:"name"`
	Version                   string   `json:"version"`
	UpstreamPins              pinsFile `json:"upstreamPins"`
	SpecKitVersion            string   `json:"specKitVersion"`
	BrownfieldMinimumVersion  string   `json:"brownfieldMinimumSpecKitVersion"`
	BrownfieldCompatible      bool     `json:"brownfieldCompatible"`
	VersionCompatibilityKnown bool     `json:"versionCompatibilityKnown"`
	CompatibilityDiagnostic   string   `json:"compatibilityDiagnostic,omitempty"`
	CLIProjectPath            string   `json:"cliProjectPath"`
	CLISourcePath             string   `json:"cliSourcePath"`
	InitCommand               string   `json:"initCommand"`
	BrownfieldNames           []string `json:"brownfieldWorkflowNames"`
	CoreNames                 []string `json:"coreWorkflowNames"`
}

type environmentInput struct{}

type environmentResult struct {
	GoAvailable               bool   `json:"goAvailable"`
	UVAvailable               bool   `json:"uvAvailable"`
	GitAvailable              bool   `json:"gitAvailable"`
	RunnerProjectPresent      bool   `json:"runnerProjectPresent"`
	SpecKitCLIPresent         bool   `json:"specKitCliPresent"`
	VersionCompatibilityKnown bool   `json:"versionCompatibilityKnown"`
	BrownfieldCompatible      bool   `json:"brownfieldCompatible"`
	SpecKitVersion            string `json:"specKitVersion"`
	BrownfieldMinimumVersion  string `json:"brownfieldMinimumSpecKitVersion"`
	CompatibilityDiagnostic   string `json:"compatibilityDiagnostic,omitempty"`
	OfficialPinMatches        bool   `json:"officialSubmodulePinMatches"`
	BrownfieldPinMatches      bool   `json:"brownfieldSubmodulePinMatches"`
	OfficialPinCheckKnown     bool   `json:"officialSubmodulePinCheckKnown"`
	BrownfieldPinCheckKnown   bool   `json:"brownfieldSubmodulePinCheckKnown"`
	OfficialPinDiagnostic     string `json:"officialPinDiagnostic,omitempty"`
	BrownfieldPinDiagnostic   string `json:"brownfieldPinDiagnostic,omitempty"`
	CoreWorkflowFilesPresent  bool   `json:"coreWorkflowFilesPresent"`
	BrownfieldFilesPresent    bool   `json:"brownfieldFilesPresent"`
}

type projectStatusInput struct {
	ProjectRoot string `json:"projectRoot" jsonschema:"absolute path to the user's selected project root"`
}

type projectStatus struct {
	ProjectRoot        string `json:"projectRoot"`
	GitMarkerPresent   bool   `json:"gitMarkerPresent"`
	HasSpecKitMetadata bool   `json:"hasSpecKitMetadata"`
	HasConstitution    bool   `json:"hasConstitution"`
	HasCodexSkills     bool   `json:"hasCodexSkills"`
	HasSpecsDirectory  bool   `json:"hasSpecsDirectory"`
}

type workflowInput struct {
	Name string `json:"name" jsonschema:"allowlisted workflow name, such as specify or brownfield.scan"`
}

type workflowSource struct {
	Name     string `json:"name"`
	Upstream string `json:"upstream"`
	Ref      string `json:"ref"`
	Path     string `json:"path"`
	Content  string `json:"content"`
}

type workflow struct {
	upstream string
	path     string
	pinName  string
}

var workflows = map[string]workflow{
	"analyze":              {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/analyze.md"},
	"checklist":            {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/checklist.md"},
	"clarify":              {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/clarify.md"},
	"constitution":         {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/constitution.md"},
	"converge":             {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/converge.md"},
	"implement":            {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/implement.md"},
	"plan":                 {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/plan.md"},
	"specify":              {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/specify.md"},
	"tasks":                {upstream: "spec-kit-official", pinName: "spec-kit-official", path: "upstream/spec-kit-official/templates/commands/tasks.md"},
	"brownfield.scan":      {upstream: "spec-kit-brownfield", pinName: "spec-kit-brownfield", path: "upstream/spec-kit-brownfield/commands/speckit.brownfield.scan.md"},
	"brownfield.bootstrap": {upstream: "spec-kit-brownfield", pinName: "spec-kit-brownfield", path: "upstream/spec-kit-brownfield/commands/speckit.brownfield.bootstrap.md"},
	"brownfield.validate":  {upstream: "spec-kit-brownfield", pinName: "spec-kit-brownfield", path: "upstream/spec-kit-brownfield/commands/speckit.brownfield.validate.md"},
	"brownfield.migrate":   {upstream: "spec-kit-brownfield", pinName: "spec-kit-brownfield", path: "upstream/spec-kit-brownfield/commands/speckit.brownfield.migrate.md"},
}

func pluginInfoHandler(root string) mcp.ToolHandlerFor[struct{}, pluginInfo] {
	return func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, pluginInfo, error) {
		var manifest pluginManifest
		if err := readJSON(filepath.Join(root, "plugin.json"), &manifest); err != nil {
			return nil, pluginInfo{}, fmt.Errorf("read plugin manifest: %w", err)
		}
		var pins pinsFile
		if err := readJSON(filepath.Join(root, "upstream-pins.json"), &pins); err != nil {
			return nil, pluginInfo{}, fmt.Errorf("read upstream pins: %w", err)
		}
		core, brownfield := workflowNames()
		compatibility := inspectCompatibility(root)
		return nil, pluginInfo{
			Name: manifest.Name, Version: manifest.Version, UpstreamPins: pins,
			SpecKitVersion:            compatibility.SpecKitVersion,
			BrownfieldMinimumVersion:  compatibility.BrownfieldMinimum,
			BrownfieldCompatible:      compatibility.BrownfieldCompatible,
			VersionCompatibilityKnown: compatibility.CompatibilityKnown,
			CompatibilityDiagnostic:   compatibility.CompatibilityDiagnostic,
			CLIProjectPath:            filepath.Clean(root),
			CLISourcePath:             filepath.Join(root, "upstream", "spec-kit-official"),
			InitCommand:               `uv run --locked --project "<plugin-root>" specify init --here --integration codex --integration-options=--skills`,
			CoreNames:                 core, BrownfieldNames: brownfield,
		}, nil
	}
}

func environmentHandler(root string) mcp.ToolHandlerFor[environmentInput, environmentResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ environmentInput) (*mcp.CallToolResult, environmentResult, error) {
		result := environmentResult{
			GoAvailable:          executableAvailable("go"),
			UVAvailable:          executableAvailable("uv"),
			GitAvailable:         executableAvailable("git"),
			RunnerProjectPresent: regularFile(filepath.Join(root, "pyproject.toml")) && regularFile(filepath.Join(root, "uv.lock")),
			SpecKitCLIPresent:    regularFile(filepath.Join(root, "upstream", "spec-kit-official", "pyproject.toml")),
		}
		compatibility := inspectCompatibility(root)
		result.VersionCompatibilityKnown = compatibility.CompatibilityKnown
		result.BrownfieldCompatible = compatibility.BrownfieldCompatible
		result.SpecKitVersion = compatibility.SpecKitVersion
		result.BrownfieldMinimumVersion = compatibility.BrownfieldMinimum
		result.CompatibilityDiagnostic = compatibility.CompatibilityDiagnostic
		var pins pinsFile
		if err := readJSON(filepath.Join(root, "upstream-pins.json"), &pins); err != nil {
			result.OfficialPinDiagnostic = fmt.Sprintf("cannot read upstream pins: %v", err)
			result.BrownfieldPinDiagnostic = result.OfficialPinDiagnostic
		} else {
			result.OfficialPinMatches, result.OfficialPinCheckKnown, result.OfficialPinDiagnostic = pinnedSubmoduleMatches(ctx, root, "spec-kit-official", pins["spec-kit-official"])
			result.BrownfieldPinMatches, result.BrownfieldPinCheckKnown, result.BrownfieldPinDiagnostic = pinnedSubmoduleMatches(ctx, root, "spec-kit-brownfield", pins["spec-kit-brownfield"])
		}
		result.CoreWorkflowFilesPresent = true
		result.BrownfieldFilesPresent = true
		for name, item := range workflows {
			if !regularFile(filepath.Join(root, filepath.FromSlash(item.path))) {
				if strings.HasPrefix(name, "brownfield.") {
					result.BrownfieldFilesPresent = false
				} else {
					result.CoreWorkflowFilesPresent = false
				}
			}
		}
		return nil, result, nil
	}
}

func projectStatusHandler(_ context.Context, _ *mcp.CallToolRequest, input projectStatusInput) (*mcp.CallToolResult, projectStatus, error) {
	root, err := resolveProjectRoot(input.ProjectRoot)
	if err != nil {
		return nil, projectStatus{}, err
	}
	git, err := pathExistsWithin(root, ".git", pathAny)
	if err != nil {
		return nil, projectStatus{}, err
	}
	metadata, err := pathExistsWithin(root, ".specify", pathDirectory)
	if err != nil {
		return nil, projectStatus{}, err
	}
	constitution, err := pathExistsWithin(root, ".specify/memory/constitution.md", pathFile)
	if err != nil {
		return nil, projectStatus{}, err
	}
	skills, err := pathExistsWithin(root, ".agents/skills", pathDirectory)
	if err != nil {
		return nil, projectStatus{}, err
	}
	specs, err := pathExistsWithin(root, "specs", pathDirectory)
	if err != nil {
		return nil, projectStatus{}, err
	}
	return nil, projectStatus{
		ProjectRoot: root, GitMarkerPresent: git, HasSpecKitMetadata: metadata,
		HasConstitution: constitution, HasCodexSkills: skills, HasSpecsDirectory: specs,
	}, nil
}

func workflowSourceHandler(root string) mcp.ToolHandlerFor[workflowInput, workflowSource] {
	return func(_ context.Context, _ *mcp.CallToolRequest, input workflowInput) (*mcp.CallToolResult, workflowSource, error) {
		item, ok := workflows[input.Name]
		if !ok {
			return nil, workflowSource{}, fmt.Errorf("unknown workflow %q; use spec_kit_plugin_info to list allowed names", input.Name)
		}
		path := filepath.Join(root, filepath.FromSlash(item.path))
		resolved, err := resolveInside(root, path)
		if err != nil {
			return nil, workflowSource{}, fmt.Errorf("resolve workflow reference: %w", err)
		}
		file, err := os.Open(resolved)
		if err != nil {
			return nil, workflowSource{}, fmt.Errorf("open workflow reference: %w", err)
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, maxWorkflowBytes+1))
		if err != nil {
			return nil, workflowSource{}, fmt.Errorf("read workflow reference: %w", err)
		}
		if len(data) > maxWorkflowBytes {
			return nil, workflowSource{}, fmt.Errorf("workflow reference exceeds %d bytes", maxWorkflowBytes)
		}
		var pins pinsFile
		if err := readJSON(filepath.Join(root, "upstream-pins.json"), &pins); err != nil {
			return nil, workflowSource{}, fmt.Errorf("read upstream pins: %w", err)
		}
		pinned, ok := pins[item.pinName]
		if !ok {
			return nil, workflowSource{}, fmt.Errorf("upstream pin %q is missing", item.pinName)
		}
		return nil, workflowSource{
			Name: input.Name, Upstream: item.upstream, Ref: pinned.Ref,
			Path: item.path, Content: string(data),
		}, nil
	}
}

func workflowNames() ([]string, []string) {
	var core, brownfield []string
	for name := range workflows {
		if strings.HasPrefix(name, "brownfield.") {
			brownfield = append(brownfield, name)
		} else {
			core = append(core, name)
		}
	}
	sortStrings(core)
	sortStrings(brownfield)
	return core, brownfield
}

func sortStrings(values []string) {
	sort.Strings(values)
}

func resolveProjectRoot(input string) (string, error) {
	if input == "" || !filepath.IsAbs(input) {
		return "", fmt.Errorf("projectRoot must be an absolute path")
	}
	for _, part := range strings.Split(input, string(filepath.Separator)) {
		if part == ".." {
			return "", fmt.Errorf("projectRoot must not contain parent-directory traversal")
		}
	}
	root, err := filepath.EvalSymlinks(input)
	if err != nil {
		return "", fmt.Errorf("resolve projectRoot: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("stat projectRoot: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("projectRoot is not a directory")
	}
	return root, nil
}

type expectedKind int

const (
	pathAny expectedKind = iota
	pathFile
	pathDirectory
)

func pathExistsWithin(root, relative string, want expectedKind) (bool, error) {
	path := filepath.Join(root, filepath.FromSlash(relative))
	resolved, err := resolveInside(root, path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	info, err := os.Stat(resolved)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect project path %q: %w", relative, err)
	}
	switch want {
	case pathFile:
		return info.Mode().IsRegular(), nil
	case pathDirectory:
		return info.IsDir(), nil
	default:
		return true, nil
	}
}

func resolveInside(root, path string) (string, error) {
	relative, err := filepath.Rel(root, filepath.Clean(path))
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("path resolves outside its allowed root")
	}
	if relative == "." {
		return root, nil
	}

	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			return current, nil
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", fmt.Errorf("resolve symlink %q: %w", current, err)
			}
			resolved, err = filepath.Abs(resolved)
			if err != nil {
				return "", err
			}
			within, err := filepath.Rel(root, resolved)
			if err != nil {
				return "", err
			}
			if within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) || filepath.IsAbs(within) {
				return "", fmt.Errorf("path resolves outside its allowed root")
			}
			current = resolved
		}
	}
	return current, nil
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func executableAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func pinnedSubmoduleMatches(ctx context.Context, pluginRoot, moduleName string, expected pin) (bool, bool, string) {
	if expected.Commit == "" {
		return false, false, fmt.Sprintf("pin for %s is missing", moduleName)
	}
	if _, err := hex.DecodeString(expected.Commit); err != nil {
		return false, false, fmt.Sprintf("pin for %s is not a hexadecimal commit: %v", moduleName, err)
	}
	path := filepath.Join(pluginRoot, "upstream", moduleName)
	command := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--verify", "HEAD^{commit}")
	output, err := command.Output()
	if err != nil {
		return false, false, fmt.Sprintf("cannot inspect %s submodule revision: %v", moduleName, err)
	}
	actual := strings.TrimSpace(string(output))
	if _, err := hex.DecodeString(actual); err != nil || len(actual) != len(expected.Commit) {
		return false, false, fmt.Sprintf("git returned an invalid revision for %s", moduleName)
	}
	if actual != expected.Commit {
		return false, true, fmt.Sprintf("%s is at %s, expected %s", moduleName, actual, expected.Commit)
	}
	return true, true, ""
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
