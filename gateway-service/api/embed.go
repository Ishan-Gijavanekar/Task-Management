package api

import _ "embed"

// OpenAPI is the public API contract served by the gateway.
//
//go:embed openapi.json
var OpenAPI []byte
