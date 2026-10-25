package codegen

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Ishan-Gijavanekar/gateway-service/internal/contract"
)

func TestGenerateIsDeterministicAndMarksOptionalFields(t *testing.T) {
	spec := &contract.Spec{Paths: map[string]map[string]json.RawMessage{"/things": {"get": json.RawMessage(`{"operationId":"listThings"}`)}}}
	spec.Components.Schemas = map[string]contract.Schema{"Thing": {Type: "object", Required: []string{"id"}, Properties: map[string]contract.Schema{"name": {Type: "string"}, "id": {Type: "string"}}}}
	first, err := Generate(spec)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := Generate(spec)
	if string(first) != string(second) {
		t.Fatal("generation is not deterministic")
	}
	text := string(first)
	for _, want := range []string{"id: string", "name?: string", "listThings: { method: \"GET\""} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %s in %s", want, text)
		}
	}
}
