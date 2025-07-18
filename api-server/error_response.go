package apiserver

type ErrorResponseConstructor func(error) any

type errorResponse struct {
	Error string `json:"error"`
}

func newErrorResponse(err error) any {
	return errorResponse{Error: err.Error()}
}
