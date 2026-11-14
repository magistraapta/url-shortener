package model

// Response is the envelope returned by every API endpoint. T is the type of the
// success payload; exactly one of Data or Error is set.
//
//	{"success": true,  "data": {...}}
//	{"success": false, "error": {"code": "...", "message": "..."}}
type Response[T any] struct {
	Success bool       `json:"success"`
	Data    *T         `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

// ErrorBody describes a failed request. Code is a stable, machine-readable
// identifier; Message is for humans.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success wraps data in a successful response.
func Success[T any](data T) Response[T] {
	return Response[T]{Success: true, Data: &data}
}

// Failure builds an error response. It has no payload, so its T is struct{}.
func Failure(code, message string) Response[struct{}] {
	return Response[struct{}]{Error: &ErrorBody{Code: code, Message: message}}
}
