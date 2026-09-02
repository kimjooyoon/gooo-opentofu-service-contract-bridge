package bridge

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

var openAPIMethods = map[string]bool{
	"GET": true, "PUT": true, "POST": true, "DELETE": true,
	"OPTIONS": true, "HEAD": true, "PATCH": true, "TRACE": true,
}

func ReadBoundedOpenAPI(path string) (OpenAPIInput, error) {
	info, err := os.Stat(path)
	if err != nil {
		return OpenAPIInput{}, err
	}
	if info.IsDir() {
		return OpenAPIInput{}, errors.New("bounded OpenAPI reader accepts one YAML or JSON document")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return OpenAPIInput{}, err
	}
	version, operations, err := parseOpenAPI(raw, path)
	if err != nil {
		return OpenAPIInput{}, err
	}
	if !strings.HasPrefix(version, "3.") || len(operations) == 0 {
		return OpenAPIInput{}, errors.New("bounded OpenAPI reader requires OpenAPI 3.x with at least one operation")
	}
	sort.Slice(operations, func(i, j int) bool {
		left := operations[i].Path + " " + operations[i].Method
		right := operations[j].Path + " " + operations[j].Method
		return left < right
	})
	return OpenAPIInput{Version: version, Operations: operations, File: path, Lines: lineCount(raw)}, nil
}

func parseOpenAPI(raw []byte, path string) (string, []OpenAPIOperation, error) {
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "{") {
		return parseOpenAPIJSON(raw, path)
	}
	return parseOpenAPIYAML(raw, path)
}

func parseOpenAPIJSON(raw []byte, path string) (string, []OpenAPIOperation, error) {
	var document map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	if err := decoder.Decode(&document); err != nil {
		return "", nil, fmt.Errorf("%s: invalid OpenAPI JSON: %w", path, err)
	}
	version, ok := document["openapi"].(string)
	if !ok || version == "" {
		return "", nil, fmt.Errorf("%s: OpenAPI JSON must contain a string openapi version", path)
	}
	paths, ok := document["paths"].(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("%s: bounded reader requires a paths object", path)
	}
	pathNames := make([]string, 0, len(paths))
	for pathName := range paths {
		pathNames = append(pathNames, pathName)
	}
	sort.Strings(pathNames)
	operations := []OpenAPIOperation{}
	for _, pathName := range pathNames {
		item, ok := paths[pathName].(map[string]interface{})
		if !ok {
			return "", nil, fmt.Errorf("%s: path %s is not an object", path, pathName)
		}
		methodNames := make([]string, 0, len(item))
		for name := range item {
			methodNames = append(methodNames, strings.ToUpper(name))
		}
		sort.Strings(methodNames)
		for _, method := range methodNames {
			if !openAPIMethods[method] {
				continue
			}
			operation, ok := item[strings.ToLower(method)].(map[string]interface{})
			if !ok {
				return "", nil, fmt.Errorf("%s: operation %s %s is not an object", path, method, pathName)
			}
			operationID, _ := operation["operationId"].(string)
			operations = append(operations, OpenAPIOperation{Method: method, Path: pathName, OperationID: operationID, File: path})
		}
	}
	return version, operations, nil
}

func parseOpenAPIYAML(raw []byte, path string) (string, []OpenAPIOperation, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	version := ""
	sawPaths := false
	currentPath := ""
	currentMethod := ""
	currentOperationID := ""
	operationLine := 0
	operations := []OpenAPIOperation{}
	flush := func() {
		if currentPath != "" && currentMethod != "" {
			operations = append(operations, OpenAPIOperation{
				Method: currentMethod, Path: currentPath, OperationID: currentOperationID,
				File: path, Line: operationLine,
			})
		}
		currentMethod = ""
		currentOperationID = ""
		operationLine = 0
	}
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		withoutComment := strings.TrimRight(stripYAMLComment(scanner.Text()), " \t\r")
		if strings.TrimSpace(withoutComment) == "" {
			continue
		}
		indent := leadingSpaces(withoutComment)
		content := strings.TrimSpace(withoutComment)
		if indent == 0 && strings.HasPrefix(content, "openapi:") {
			version = yamlScalar(strings.TrimSpace(strings.TrimPrefix(content, "openapi:")))
			continue
		}
		if indent == 0 && content == "paths:" {
			sawPaths = true
			continue
		}
		if !sawPaths {
			continue
		}
		if indent == 2 && strings.HasSuffix(content, ":") && strings.HasPrefix(content, "/") {
			flush()
			currentPath = yamlScalar(strings.TrimSuffix(content, ":"))
			continue
		}
		if indent == 4 && strings.HasSuffix(content, ":") {
			method := strings.ToUpper(strings.TrimSuffix(content, ":"))
			if openAPIMethods[method] {
				flush()
				currentMethod = method
				operationLine = lineNumber
			}
			continue
		}
		if indent >= 6 && currentMethod != "" && strings.HasPrefix(content, "operationId:") {
			currentOperationID = yamlScalar(strings.TrimSpace(strings.TrimPrefix(content, "operationId:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil, err
	}
	flush()
	if version == "" || !sawPaths {
		return "", nil, fmt.Errorf("%s: bounded YAML reader requires openapi and paths", path)
	}
	return version, operations, nil
}

func stripYAMLComment(line string) string {
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
		if r == '"' || r == '\'' {
			quote = r
			continue
		}
		if r == '#' {
			return line[:index]
		}
	}
	return line
}

func leadingSpaces(value string) int {
	count := 0
	for _, r := range value {
		if r != ' ' {
			break
		}
		count++
	}
	return count
}

func yamlScalar(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
		return value[1 : len(value)-1]
	}
	return value
}

func findResource(resources []InfraResource, locator string) *InfraResource {
	parts := strings.SplitN(locator, ".", 3)
	if len(parts) != 3 || parts[0] != "resource" {
		return nil
	}
	for index := range resources {
		if resources[index].Type == parts[1] && resources[index].Name == parts[2] {
			return &resources[index]
		}
	}
	return nil
}

func findOperation(operations []OpenAPIOperation, locator string) *OpenAPIOperation {
	parts := strings.SplitN(locator, ".", 3)
	if len(parts) != 3 || parts[0] != "path" {
		return nil
	}
	method := strings.ToUpper(parts[1])
	for index := range operations {
		if operations[index].Method == method && operations[index].Path == parts[2] {
			return &operations[index]
		}
	}
	return nil
}
