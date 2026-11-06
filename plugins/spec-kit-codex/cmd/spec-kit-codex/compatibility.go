package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type compatibilityInfo struct {
	SpecKitVersion          string `json:"specKitVersion"`
	BrownfieldMinimum       string `json:"brownfieldMinimumSpecKitVersion"`
	BrownfieldCompatible    bool   `json:"brownfieldCompatible"`
	CompatibilityKnown      bool   `json:"versionCompatibilityKnown"`
	CompatibilityDiagnostic string `json:"compatibilityDiagnostic,omitempty"`
}

type semanticVersion struct {
	major int
	minor int
	patch int
}

var (
	projectVersionPattern = regexp.MustCompile(`^version\s*=\s*"([^"]+)"\s*(?:#.*)?$`)
	minimumVersionPattern = regexp.MustCompile(`^\s*speckit_version:\s*">=\s*([^" ]+)"\s*$`)
)

func inspectCompatibility(pluginRoot string) compatibilityInfo {
	result := compatibilityInfo{}
	kitVersion, err := readProjectVersion(filepath.Join(pluginRoot, "upstream", "spec-kit-official", "pyproject.toml"))
	if err != nil {
		result.CompatibilityDiagnostic = fmt.Sprintf("cannot read pinned Spec Kit version: %v", err)
		return result
	}
	minimumVersion, err := readMinimumSpecKitVersion(filepath.Join(pluginRoot, "upstream", "spec-kit-brownfield", "extension.yml"))
	if err != nil {
		result.SpecKitVersion = kitVersion
		result.CompatibilityDiagnostic = fmt.Sprintf("cannot read Brownfield version requirement: %v", err)
		return result
	}
	current, err := parseSemanticVersion(kitVersion)
	if err != nil {
		result.SpecKitVersion = kitVersion
		result.BrownfieldMinimum = minimumVersion
		result.CompatibilityDiagnostic = err.Error()
		return result
	}
	minimum, err := parseSemanticVersion(minimumVersion)
	if err != nil {
		result.SpecKitVersion = kitVersion
		result.BrownfieldMinimum = minimumVersion
		result.CompatibilityDiagnostic = err.Error()
		return result
	}
	result.SpecKitVersion = kitVersion
	result.BrownfieldMinimum = minimumVersion
	result.BrownfieldCompatible = atLeast(current, minimum)
	result.CompatibilityKnown = true
	if !result.BrownfieldCompatible {
		result.CompatibilityDiagnostic = fmt.Sprintf("pinned Spec Kit %s does not satisfy Brownfield's >=%s requirement", kitVersion, minimumVersion)
	}
	return result
}

func readProjectVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	inProject := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inProject = line == "[project]"
			continue
		}
		if inProject {
			matches := projectVersionPattern.FindStringSubmatch(line)
			if len(matches) == 2 {
				return matches[1], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("project version was not found in %s", path)
}

func readMinimumSpecKitVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := minimumVersionPattern.FindStringSubmatch(scanner.Text())
		if len(matches) == 2 {
			return matches[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("supported Brownfield version constraint was not found in %s", path)
}

func parseSemanticVersion(value string) (semanticVersion, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return semanticVersion{}, fmt.Errorf("unsupported version format %q; expected numeric major.minor.patch", value)
	}
	components := [3]int{}
	for i, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil || parsed < 0 {
			return semanticVersion{}, fmt.Errorf("unsupported version format %q; expected numeric major.minor.patch", value)
		}
		components[i] = parsed
	}
	return semanticVersion{major: components[0], minor: components[1], patch: components[2]}, nil
}

func atLeast(current, minimum semanticVersion) bool {
	if current.major != minimum.major {
		return current.major > minimum.major
	}
	if current.minor != minimum.minor {
		return current.minor > minimum.minor
	}
	return current.patch >= minimum.patch
}
