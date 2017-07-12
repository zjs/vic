package util


// TODO: rewrite with https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully in mind
type HttpError struct {
	code    int
	message string
}

func (e *HttpError) Code() int {
	return e.code
}

func (e *HttpError) Error() string {
	return e.message
}

func NewHttpError(code int, message string) *HttpError {
	return &HttpError{code: code, message: message}
}
