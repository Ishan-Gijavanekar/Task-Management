package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Ishan-Gijavanekar/gateway-service/internal/codegen"
	"github.com/Ishan-Gijavanekar/gateway-service/internal/contract"
)

func main() {
	check := flag.Bool("check", false, "fail when generated types are stale")
	flag.Parse()
	if *check {
		if err := codegen.Check("api/openapi.json", "api/generated.ts"); err != nil {
			fail(err)
		}
		return
	}
	spec, err := contract.LoadSpec("api/openapi.json")
	if err != nil {
		fail(err)
	}
	generated, err := codegen.Generate(spec)
	if err != nil {
		fail(err)
	}
	if err = os.WriteFile("api/generated.ts", generated, 0644); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
