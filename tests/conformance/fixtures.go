//go:build integration

package conformance

type NegativeCase struct {
	Name           string
	Endpoint       string
	RequestBody    map[string]any
	WantErrorType  string
	WantHTTPStatus int
}

var Cases = []NegativeCase{
	{
		Name:           "bad-uuid",
		Endpoint:       "/v1/init_request",
		RequestBody:    map[string]any{"request_uuid": "not-a-uuid", "ttl_seconds": 60},
		WantErrorType:  "SCHEMA_VALIDATION_FAILED",
		WantHTTPStatus: 400,
	},
	{
		Name:           "ttl-too-large",
		Endpoint:       "/v1/init_request",
		RequestBody:    map[string]any{"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "ttl_seconds": 1000000},
		WantErrorType:  "SCHEMA_VALIDATION_FAILED",
		WantHTTPStatus: 400,
	},
	{
		Name:           "invalid-pii-type",
		Endpoint:       "/v1/tokenize",
		RequestBody:    map[string]any{"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "type": "BOGUS", "plaintext": "x"},
		WantErrorType:  "INVALID_PII_TYPE",
		WantHTTPStatus: 400,
	},
	{
		Name:           "tokenize-unknown-uuid",
		Endpoint:       "/v1/tokenize",
		RequestBody:    map[string]any{"request_uuid": "00000000-0000-0000-0000-000000000000", "type": "EMAIL", "plaintext": "x"},
		WantErrorType:  "SCOPE_NOT_FOUND",
		WantHTTPStatus: 404,
	},
	{
		Name:           "detokenize-bad-token",
		Endpoint:       "/v1/detokenize",
		RequestBody:    map[string]any{"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "token": "TOKEN_EMAIL_NOPE"},
		WantErrorType:  "AAD_MISMATCH",
		WantHTTPStatus: 400,
	},
	{
		Name:           "detokenize-bad-prefix",
		Endpoint:       "/v1/detokenize",
		RequestBody:    map[string]any{"request_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "token": "garbage"},
		WantErrorType:  "AAD_MISMATCH",
		WantHTTPStatus: 400,
	},
}
