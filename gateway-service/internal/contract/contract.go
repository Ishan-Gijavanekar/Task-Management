package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Schema struct {
	Type                 string            `json:"type"`
	Required             []string          `json:"required"`
	Properties           map[string]Schema `json:"properties"`
	Enum                 []string          `json:"enum"`
	Ref                  string            `json:"$ref"`
	Items                *Schema           `json:"items"`
	AdditionalProperties any               `json:"additionalProperties"`
}
type Operation struct {
	OperationID string `json:"operationId"`
}
type Spec struct {
	OpenAPI    string                                `json:"openapi"`
	Security   []map[string][]string                 `json:"security"`
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]Schema `json:"schemas"`
	} `json:"components"`
}
type Baseline struct {
	Version    string              `json:"version"`
	Operations []string            `json:"operations"`
	Schemas    map[string][]string `json:"schemas"`
	Enums      map[string][]string `json:"enums"`
}

var methods = map[string]bool{"get": true, "post": true, "put": true, "patch": true, "delete": true}

func LoadSpec(path string) (*Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec Spec
	if err = json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("parse OpenAPI: %w", err)
	}
	if spec.OpenAPI != "3.1.0" {
		return nil, fmt.Errorf("OpenAPI 3.1.0 is required")
	}
	return &spec, nil
}
func LoadBaseline(path string) (*Baseline, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var baseline Baseline
	if err = json.Unmarshal(raw, &baseline); err != nil {
		return nil, fmt.Errorf("parse baseline: %w", err)
	}
	return &baseline, nil
}

func Check(spec *Spec, baseline *Baseline) []error {
	var failures []error
	if spec == nil || baseline == nil {
		return []error{fmt.Errorf("spec and baseline are required")}
	}
	seen := map[string]bool{}
	for path, item := range spec.Paths {
		for method, raw := range item {
			if !methods[strings.ToLower(method)] {
				continue
			}
			var operation Operation
			if err := json.Unmarshal(raw, &operation); err != nil || operation.OperationID == "" {
				failures = append(failures, fmt.Errorf("%s %s is missing operationId", strings.ToUpper(method), path))
				continue
			}
			if seen[operation.OperationID] {
				failures = append(failures, fmt.Errorf("duplicate operationId %s", operation.OperationID))
			}
			seen[operation.OperationID] = true
		}
	}
	for _, operation := range baseline.Operations {
		if !seen[operation] {
			failures = append(failures, fmt.Errorf("breaking change: operation %s was removed", operation))
		}
	}
	for name, required := range baseline.Schemas {
		schema, ok := spec.Components.Schemas[name]
		if !ok {
			failures = append(failures, fmt.Errorf("breaking change: schema %s was removed", name))
			continue
		}
		current := set(schema.Required)
		for _, field := range required {
			if !current[field] {
				failures = append(failures, fmt.Errorf("breaking change: required field %s.%s was removed", name, field))
			}
		}
	}
	for key, values := range baseline.Enums {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) != 2 {
			failures = append(failures, fmt.Errorf("invalid baseline enum %s", key))
			continue
		}
		schema, ok := spec.Components.Schemas[parts[0]]
		if !ok {
			failures = append(failures, fmt.Errorf("breaking change: enum schema %s was removed", parts[0]))
			continue
		}
		property, ok := schema.Properties[parts[1]]
		if !ok {
			failures = append(failures, fmt.Errorf("breaking change: enum property %s was removed", key))
			continue
		}
		current := set(property.Enum)
		for _, value := range values {
			if !current[value] {
				failures = append(failures, fmt.Errorf("breaking change: enum value %s.%s was removed", key, value))
			}
		}
	}
	if len(spec.Security) == 0 {
		failures = append(failures, fmt.Errorf("global bearer authentication requirement is missing"))
	}
	sort.Slice(failures, func(i, j int) bool { return failures[i].Error() < failures[j].Error() })
	return failures
}
func set(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}
