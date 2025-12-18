package model

// GenericResponse for all API communications

type GenericRequest[T any] struct {
	Data *T `json:"data,omitempty"`
}
type GenericResponse[T any] struct {
	Success bool         `json:"success"`
	Data    *T           `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target,omitempty"`
}
