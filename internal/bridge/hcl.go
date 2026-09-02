package bridge

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var resourceHeader = regexp.MustCompile(`^resource[ \t]+"([^"]+)"[ \t]+"([^"]+)"[ \t]*\{[ \t]*$`)
var attributeLine = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)[ \t]*=[ \t]*(.+)$`)

func ReadBoundedOpenTofu(path string) (InfraInput, error) {
	files, err := inputFiles(path, func(name string) bool {
		return strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tofu")
	})
	if err != nil {
		return InfraInput{}, err
	}
	input := InfraInput{Resources: []InfraResource{}, Files: []FileDigest{}}
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			return InfraInput{}, err
		}
		resources, err := parseHCLFile(file, raw)
		if err != nil {
			return InfraInput{}, err
		}
		input.Resources = append(input.Resources, resources...)
		input.Files = append(input.Files, digestFor(file, raw))
	}
	if len(input.Resources) == 0 {
		return InfraInput{}, errors.New("bounded OpenTofu reader found no resource blocks")
	}
	sort.Slice(input.Resources, func(i, j int) bool {
		return resourceKey(input.Resources[i]) < resourceKey(input.Resources[j])
	})
	return input, nil
}

func inputFiles(path string, allowed func(string) bool) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !allowed(path) {
			return nil, fmt.Errorf("input %s is outside the bounded file set", path)
		}
		return []string{filepath.Clean(path)}, nil
	}
	files := []string{}
	err = filepath.WalkDir(path, func(candidate string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == ".terraform" || entry.Name() == ".tofu" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type().IsRegular() && allowed(candidate) {
			files = append(files, filepath.Clean(candidate))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no bounded input files found below %s", path)
	}
	return files, nil
}

func parseHCLFile(path string, raw []byte) ([]InfraResource, error) {
	resources := []InfraResource{}
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	lineNumber := 0
	depth := 0
	var current *InfraResource
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(stripHCLComment(scanner.Text()))
		if line == "" {
			continue
		}
		if current == nil {
			if match := resourceHeader.FindStringSubmatch(line); match != nil {
				current = &InfraResource{Type: match[1], Name: match[2], File: path, Line: lineNumber, Attributes: map[string]string{}}
				depth = 1
				continue
			}
			return nil, fmt.Errorf("%s:%d: only resource blocks are supported by the bounded reader", path, lineNumber)
		}
		if line == "}" {
			depth--
			if depth != 0 {
				return nil, fmt.Errorf("%s:%d: malformed resource closure", path, lineNumber)
			}
			resources = append(resources, *current)
			current = nil
			continue
		}
		if strings.HasSuffix(line, "{") {
			return nil, fmt.Errorf("%s:%d: nested blocks are outside the bounded reader", path, lineNumber)
		}
		match := attributeLine.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("%s:%d: only literal single-line arguments are supported", path, lineNumber)
		}
		if _, exists := current.Attributes[match[1]]; exists {
			return nil, fmt.Errorf("%s:%d: duplicate argument %s", path, lineNumber, match[1])
		}
		value, err := literalValue(match[2])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNumber, err)
		}
		current.Attributes[match[1]] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if current != nil || depth != 0 {
		return nil, fmt.Errorf("%s: unterminated resource block", path)
	}
	seen := map[string]bool{}
	for _, resource := range resources {
		key := resourceKey(resource)
		if seen[key] {
			return nil, fmt.Errorf("%s: duplicate resource %s", path, key)
		}
		seen[key] = true
	}
	return resources, nil
}

func stripHCLComment(line string) string {
	var quote rune
	escaped := false
	for index, r := range line {
		if escaped {
			escaped = false
			continue
		}
		if quote != 0 {
			if r == '\\' {
				escaped = true
			} else if r == quote {
				quote = 0
			}
			continue
		}
		if r == '"' {
			quote = r
			continue
		}
		if r == '#' || r == '/' && index+1 < len(line) && line[index+1] == '/' {
			return line[:index]
		}
	}
	return line
}

func literalValue(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, `"`) {
		value, err := strconv.Unquote(raw)
		if err != nil {
			return "", fmt.Errorf("invalid string literal %q", raw)
		}
		return value, nil
	}
	if raw == "true" || raw == "false" {
		return raw, nil
	}
	if _, err := strconv.ParseFloat(raw, 64); err == nil {
		return raw, nil
	}
	return "", fmt.Errorf("expression %q is outside the bounded literal subset", raw)
}

func resourceKey(resource InfraResource) string {
	return resource.Type + "." + resource.Name
}

