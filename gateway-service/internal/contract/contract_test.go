package contract

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Ishan-Gijavanekar/gateway-service/internal/gateway"
)

func TestCommittedContractIsCompatible(t *testing.T) {
	spec, err := LoadSpec("../../api/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := LoadBaseline("../../api/compatibility-baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	if failures := Check(spec, baseline); len(failures) > 0 {
		t.Fatalf("compatibility failures: %v", failures)
	}
}

func TestEveryContractOperationHasAGatewayRoute(t *testing.T) {
	spec, err := LoadSpec("../../api/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	for path, item := range spec.Paths {
		for method := range item {
			if !methods[strings.ToLower(method)] {
				continue
			}
			if _, ok := gateway.Match(method, "/api/v1"+path); !ok {
				t.Errorf("%s %s is documented but is not routed", strings.ToUpper(method), path)
			}
		}
	}
}

func TestDetectsBreakingChanges(t *testing.T) {
	spec := &Spec{Paths: map[string]map[string]json.RawMessage{}, Security: []map[string][]string{{"bearerAuth": {}}}}
	baseline := &Baseline{Operations: []string{"getTask"}, Schemas: map[string][]string{"Task": {"id"}}, Enums: map[string][]string{"Task.status": {"done"}}}
	failures := Check(spec, baseline)
	joined := ""
	for _, failure := range failures {
		joined += failure.Error() + "\n"
	}
	for _, want := range []string{"operation getTask", "schema Task", "enum schema Task"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %s", want, joined)
		}
	}
}
