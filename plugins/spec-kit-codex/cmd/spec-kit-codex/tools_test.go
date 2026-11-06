package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func fixturePluginRoot(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller did not return a test source path")
	}
	root, err := filepath.EvalSymlinks(filepath.Join(filepath.Dir(source), "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestPinnedManifestsAndSubmodules(t *testing.T) {
	pluginRoot := fixturePluginRoot(t)
	repositoryRoot := filepath.Clean(filepath.Join(pluginRoot, "..", ".."))

	var manifest pluginManifest
	if err := readJSON(filepath.Join(pluginRoot, "plugin.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "spec-kit-codex" || manifest.Version != "0.1.0" {
		t.Fatalf("unexpected plugin identity: %+v", manifest)
	}
	var presentation struct {
		Extensions struct {
			OpenAI struct {
				Interface struct {
					ShortDescription string   `json:"shortDescription"`
					Category         string   `json:"category"`
					Capabilities     []string `json:"capabilities"`
					DefaultPrompt    []string `json:"defaultPrompt"`
				} `json:"interface"`
			} `json:"com.openai"`
		} `json:"extensions"`
	}
	if err := readJSON(filepath.Join(pluginRoot, "plugin.json"), &presentation); err != nil {
		t.Fatal(err)
	}
	iface := presentation.Extensions.OpenAI.Interface
	if len(iface.ShortDescription) > 30 || iface.Category != "Developer Tools" || len(iface.Capabilities) == 0 || len(iface.DefaultPrompt) > 3 {
		t.Errorf("invalid plugin presentation metadata: %+v", iface)
	}
	for _, path := range []string{
		filepath.Join(pluginRoot, "mcp.json"),
		filepath.Join(repositoryRoot, ".agents", "plugins", "marketplace.json"),
	} {
		var value map[string]any
		if err := readJSON(path, &value); err != nil {
			t.Errorf("parse %s: %v", path, err)
		}
	}

	var pins pinsFile
	if err := readJSON(filepath.Join(pluginRoot, "upstream-pins.json"), &pins); err != nil {
		t.Fatal(err)
	}
	for name, expected := range pins {
		modulePath := filepath.Join(pluginRoot, "upstream", name)
		cmd := exec.Command("git", "-C", modulePath, "rev-parse", "HEAD")
		actual, err := cmd.Output()
		if err != nil {
			t.Fatalf("read %s revision: %v", name, err)
		}
		if strings.TrimSpace(string(actual)) != expected.Commit {
			t.Errorf("%s revision = %s, want %s", name, strings.TrimSpace(string(actual)), expected.Commit)
		}
		cmd = exec.Command("git", "-C", modulePath, "status", "--porcelain")
		status, err := cmd.Output()
		if err != nil {
			t.Fatalf("check %s cleanliness: %v", name, err)
		}
		if len(strings.TrimSpace(string(status))) != 0 {
			t.Errorf("upstream submodule %s has local changes:\n%s", name, status)
		}
	}

}

func TestProjectRootValidationAndSymlinkBoundary(t *testing.T) {
	fixture := t.TempDir()
	project := filepath.Join(fixture, "project")
	outside := filepath.Join(fixture, "outside")
	if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(project, ".agents", "skills")); err != nil {
		t.Fatal(err)
	}

	if _, err := resolveProjectRoot("relative/path"); err == nil {
		t.Fatal("relative project root unexpectedly accepted")
	}
	if _, err := resolveProjectRoot(filepath.Join(fixture, "..", "project")); err == nil {
		t.Fatal("project root containing parent traversal unexpectedly accepted")
	}
	_, _, err := projectStatusHandler(context.Background(), nil, projectStatusInput{ProjectRoot: project})
	if err == nil {
		t.Fatal("symlink escaping project root unexpectedly accepted")
	}
	projectWithSymlinkedParent := filepath.Join(fixture, "project-with-symlinked-parent")
	if err := os.Mkdir(projectWithSymlinkedParent, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(projectWithSymlinkedParent, ".agents")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := projectStatusHandler(context.Background(), nil, projectStatusInput{ProjectRoot: projectWithSymlinkedParent}); err == nil {
		t.Fatal("symlinked parent escaping project root unexpectedly accepted")
	}

	inside := filepath.Join(fixture, "inside")
	if err := os.Mkdir(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := resolveProjectRoot(inside)
	if err != nil || root != inside {
		t.Fatalf("valid root resolution got %q, %v", root, err)
	}
	exists, err := pathExistsWithin(inside, "missing", pathAny)
	if err != nil || exists {
		t.Fatalf("expected a missing fixed path, got exists=%v err=%v", exists, err)
	}
}

func TestWorkflowNamesAndSourceAllowlist(t *testing.T) {
	core, brownfield := workflowNames()
	if len(core) != 9 || len(brownfield) != 4 {
		t.Fatalf("workflow counts = %d core, %d brownfield", len(core), len(brownfield))
	}
	for i := 1; i < len(core); i++ {
		if core[i-1] > core[i] {
			t.Fatalf("core workflow names are not sorted: %v", core)
		}
	}
	for i := 1; i < len(brownfield); i++ {
		if brownfield[i-1] > brownfield[i] {
			t.Fatalf("brownfield workflow names are not sorted: %v", brownfield)
		}
	}

	root := fixturePluginRoot(t)
	handler := workflowSourceHandler(root)
	_, source, err := handler(context.Background(), nil, workflowInput{Name: "specify"})
	if err != nil {
		t.Fatal(err)
	}
	if source.Ref != "v1.3.2" || !strings.Contains(source.Content, "Create or update the feature specification") {
		t.Fatalf("unexpected pinned specify source: ref=%s, bytes=%d", source.Ref, len(source.Content))
	}
	_, _, err = handler(context.Background(), nil, workflowInput{Name: "../../etc/passwd"})
	if err == nil {
		t.Fatal("arbitrary path workflow unexpectedly accepted")
	}

	tempRoot := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("do not read"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(tempRoot, filepath.FromSlash(workflows["analyze"].path))
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	_, _, err = workflowSourceHandler(tempRoot)(context.Background(), nil, workflowInput{Name: "analyze"})
	if err == nil || !strings.Contains(err.Error(), "outside its allowed root") {
		t.Fatalf("escaping workflow symlink was not rejected: %v", err)
	}
}

func TestPinnedBrownfieldVersionCompatibility(t *testing.T) {
	compatibility := inspectCompatibility(fixturePluginRoot(t))
	if !compatibility.CompatibilityKnown || !compatibility.BrownfieldCompatible || compatibility.SpecKitVersion != "1.3.2" || compatibility.BrownfieldMinimum != "0.4.0" {
		t.Fatalf("unexpected Brownfield compatibility result: %+v", compatibility)
	}
	current, err := parseSemanticVersion("0.3.9")
	if err != nil {
		t.Fatal(err)
	}
	minimum, err := parseSemanticVersion("0.4.0")
	if err != nil {
		t.Fatal(err)
	}
	if atLeast(current, minimum) {
		t.Fatal("version below extension minimum was considered compatible")
	}
	pluginRoot := fixturePluginRoot(t)
	var pins pinsFile
	if err := readJSON(filepath.Join(pluginRoot, "upstream-pins.json"), &pins); err != nil {
		t.Fatal(err)
	}
	for name, expected := range pins {
		matches, known, diagnostic := pinnedSubmoduleMatches(context.Background(), pluginRoot, name, expected)
		if !matches || !known || diagnostic != "" {
			t.Errorf("submodule %s did not match its pin: %s", name, diagnostic)
		}
	}
}

func TestSkillsKeepWritesBehindApproval(t *testing.T) {
	pluginRoot := fixturePluginRoot(t)
	checks := map[string][]string{
		"skills/spec-driven-development/SKILL.md": {
			"Get explicit user approval before running `specify init`",
			"Ask before writing any project artifact",
			"Keep reads and edits inside the selected project root",
			"Call `spec_kit_project_status` with that root and `spec_kit_check_environment`",
			"If a known submodule pin mismatch appears, stop",
		},
		"skills/brownfield-adoption/SKILL.md": {
			"Wait for explicit user approval before writing",
			"Do not modify application source code as part of migration",
			"The MCP tools in this plugin are read-only",
			"Brownfield version compatibility is known and true",
			"Before bootstrap, check whether `.specify/` exists",
		},
	}
	for path, required := range checks {
		content, err := os.ReadFile(filepath.Join(pluginRoot, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		for _, phrase := range required {
			if !strings.Contains(string(content), phrase) {
				t.Errorf("%s is missing required safety guidance %q", path, phrase)
			}
		}
	}
}

func TestStdioMCPInitializeListAndCalls(t *testing.T) {
	pluginRoot := fixturePluginRoot(t)
	if installedRoot := os.Getenv("SPEC_KIT_INSTALLED_PLUGIN_ROOT"); installedRoot != "" {
		var err error
		pluginRoot, err = resolvePluginRoot(installedRoot)
		if err != nil {
			t.Fatalf("resolve installed plugin root: %v", err)
		}
	}
	project := t.TempDir()
	for _, path := range []string{".git", ".specify/memory", ".agents/skills", "specs"} {
		if err := os.MkdirAll(filepath.Join(project, path), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(project, ".specify", "memory", "constitution.md"), []byte("# fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := snapshotTree(t, project)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	client := mcp.NewClient(&mcp.Implementation{Name: "spec-kit-tests", Version: "0.0.1"}, nil)
	cmd := exec.Command("go", "run", "./cmd/spec-kit-codex", "--plugin-root", pluginRoot)
	cmd.Dir = pluginRoot
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 4 {
		t.Fatalf("listed %d tools, want 4", len(listed.Tools))
	}
	allowedTools := map[string]bool{
		"spec_kit_plugin_info":       true,
		"spec_kit_check_environment": true,
		"spec_kit_project_status":    true,
		"spec_kit_workflow_source":   true,
	}
	for _, tool := range listed.Tools {
		if !allowedTools[tool.Name] {
			t.Errorf("unexpected MCP tool %q", tool.Name)
		} else {
			delete(allowedTools, tool.Name)
		}
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint || tool.Annotations.OpenWorldHint == nil || *tool.Annotations.OpenWorldHint {
			t.Errorf("tool %s lacks read-only/closed-world annotations: %+v", tool.Name, tool.Annotations)
		}
		if tool.InputSchema == nil {
			t.Errorf("tool %s has no input schema", tool.Name)
		}
	}
	if len(allowedTools) != 0 {
		t.Errorf("MCP tool list is missing expected tools: %v", allowedTools)
	}

	infoResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "spec_kit_plugin_info"})
	if err != nil || infoResult.IsError {
		t.Fatalf("plugin info call: result=%+v err=%v", infoResult, err)
	}
	var info pluginInfo
	decodeStructured(t, infoResult.StructuredContent, &info)
	if info.Version != "0.1.0" || info.UpstreamPins["spec-kit-official"].Ref != "v1.3.2" || !strings.Contains(info.InitCommand, "--locked") || !info.VersionCompatibilityKnown || !info.BrownfieldCompatible {
		t.Fatalf("unexpected plugin info: %+v", info)
	}

	environmentCall, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "spec_kit_check_environment"})
	if err != nil || environmentCall.IsError {
		t.Fatalf("environment check call: result=%+v err=%v", environmentCall, err)
	}
	var environment environmentResult
	decodeStructured(t, environmentCall.StructuredContent, &environment)
	if !environment.GoAvailable || !environment.UVAvailable || !environment.GitAvailable || !environment.RunnerProjectPresent || !environment.SpecKitCLIPresent || !environment.CoreWorkflowFilesPresent || !environment.BrownfieldFilesPresent || !environment.VersionCompatibilityKnown || !environment.BrownfieldCompatible {
		t.Fatalf("fixture environment did not report required tools and sources: %+v", environment)
	}
	if environment.OfficialPinCheckKnown && !environment.OfficialPinMatches {
		t.Fatalf("official submodule pin mismatch: %+v", environment)
	}
	if environment.BrownfieldPinCheckKnown && !environment.BrownfieldPinMatches {
		t.Fatalf("Brownfield submodule pin mismatch: %+v", environment)
	}

	statusResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "spec_kit_project_status",
		Arguments: map[string]any{"projectRoot": project},
	})
	if err != nil || statusResult.IsError {
		t.Fatalf("project status call: result=%+v err=%v", statusResult, err)
	}
	var status projectStatus
	decodeStructured(t, statusResult.StructuredContent, &status)
	if !status.GitMarkerPresent || !status.HasSpecKitMetadata || !status.HasConstitution || !status.HasCodexSkills || !status.HasSpecsDirectory {
		t.Fatalf("fixture status did not report expected presence: %+v", status)
	}

	workflowResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "spec_kit_workflow_source",
		Arguments: map[string]any{"name": "brownfield.scan"},
	})
	if err != nil || workflowResult.IsError {
		t.Fatalf("workflow source call: result=%+v err=%v", workflowResult, err)
	}
	var source workflowSource
	decodeStructured(t, workflowResult.StructuredContent, &source)
	if source.Ref != "v1.0.0" || source.Name != "brownfield.scan" || !strings.Contains(source.Content, "Analyze an existing codebase") {
		t.Fatalf("unexpected Brownfield source: %+v", source)
	}

	errorResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "spec_kit_workflow_source",
		Arguments: map[string]any{"name": "../../etc/passwd"},
	})
	if err != nil || !errorResult.IsError {
		t.Fatalf("arbitrary workflow request should return a tool error: result=%+v err=%v", errorResult, err)
	}
	if after := snapshotTree(t, project); !reflect.DeepEqual(before, after) {
		t.Fatalf("read-only MCP calls changed the fixture project:\nbefore=%v\nafter=%v", before, after)
	}
}

func decodeStructured(t *testing.T, value any, target any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, target); err != nil {
		t.Fatalf("decode structured MCP result %s: %v", encoded, err)
	}
}

func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		value := info.Mode().String() + ":" + info.ModTime().UTC().Format(time.RFC3339Nano)
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += ":" + string(data)
		} else if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			value += ":" + target
		}
		result[relative] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
