package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Ishan-Gijavanekar/gateway-service/internal/contract"
)

var identifier = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

func Generate(spec *contract.Spec) ([]byte, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec is required")
	}
	var out bytes.Buffer
	out.WriteString("// Generated from api/openapi.json. Do not edit.\n\n")
	names := make([]string, 0, len(spec.Components.Schemas))
	for name := range spec.Components.Schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		schema := spec.Components.Schemas[name]
		if schema.Type != "object" {
			fmt.Fprintf(&out, "export type %s = %s;\n\n", name, tsType(schema))
			continue
		}
		fmt.Fprintf(&out, "export interface %s {\n", name)
		required := set(schema.Required)
		properties := make([]string, 0, len(schema.Properties))
		for property := range schema.Properties {
			properties = append(properties, property)
		}
		sort.Strings(properties)
		for _, property := range properties {
			label := property
			if !identifier.MatchString(label) {
				label = strconv.Quote(label)
			}
			optional := "?"
			if required[property] {
				optional = ""
			}
			fmt.Fprintf(&out, "  %s%s: %s;\n", label, optional, tsType(schema.Properties[property]))
		}
		out.WriteString("}\n\n")
	}
	type operation struct{ Name, Method, Path string }
	var operations []operation
	for path, item := range spec.Paths {
		for method, raw := range item {
			lower := strings.ToLower(method)
			if lower != "get" && lower != "post" && lower != "put" && lower != "patch" && lower != "delete" {
				continue
			}
			var value struct {
				OperationID string `json:"operationId"`
			}
			if json.Unmarshal(raw, &value) == nil && value.OperationID != "" {
				operations = append(operations, operation{value.OperationID, strings.ToUpper(method), path})
			}
		}
	}
	sort.Slice(operations, func(i, j int) bool { return operations[i].Name < operations[j].Name })
	out.WriteString("export const apiOperations = {\n")
	for _, operation := range operations {
		fmt.Fprintf(&out, "  %s: { method: %q, path: %q },\n", operation.Name, operation.Method, operation.Path)
	}
	out.WriteString("} as const;\n")
	return out.Bytes(), nil
}
func Check(specPath, outputPath string) error {
	spec, err := contract.LoadSpec(specPath)
	if err != nil {
		return err
	}
	generated, err := Generate(spec)
	if err != nil {
		return err
	}
	existing, err := os.ReadFile(outputPath)
	if err != nil {
		return err
	}
	if !bytes.Equal(existing, generated) {
		return fmt.Errorf("%s is stale; run go run ./cmd/openapi-types", outputPath)
	}
	return nil
}
func tsType(schema contract.Schema) string {
	if schema.Ref != "" {
		parts := strings.Split(schema.Ref, "/")
		return parts[len(parts)-1]
	}
	if len(schema.Enum) > 0 {
		values := append([]string(nil), schema.Enum...)
		sort.Strings(values)
		quoted := make([]string, len(values))
		for i, value := range values {
			quoted[i] = strconv.Quote(value)
		}
		return strings.Join(quoted, " | ")
	}
	switch schema.Type {
	case "string":
		return "string"
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		if schema.Items == nil {
			return "unknown[]"
		}
		return "(" + tsType(*schema.Items) + ")[]"
	case "object":
		if schema.AdditionalProperties != nil {
			return "Record<string, unknown>"
		}
		return "Record<string, unknown>"
	default:
		return "unknown"
	}
}
func set(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}
