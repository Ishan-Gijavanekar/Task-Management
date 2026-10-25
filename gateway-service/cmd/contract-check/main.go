package main

import (
	"fmt"
	"os"

	"github.com/Ishan-Gijavanekar/gateway-service/internal/contract"
)

func main() {
	spec, err := contract.LoadSpec("api/openapi.json")
	if err != nil {
		fail(err)
	}
	baseline, err := contract.LoadBaseline("api/compatibility-baseline.json")
	if err != nil {
		fail(err)
	}
	failures := contract.Check(spec, baseline)
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, failure)
	}
	if len(failures) > 0 {
		os.Exit(1)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
