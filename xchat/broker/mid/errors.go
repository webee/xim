package mid

// APIError is api error.
type APIError interface {
	Code() int
	error
}

// APIErrorToRPCResult generate rpc error result from api error.
func APIErrorToRPCResult(e APIError) []interface{} {
	return []interface{}{false, e.Code(), e.Error()}
}

type defaultAPIError struct {
	code int
	msg  string
}

func (e *defaultAPIError) Code() int {
	return e.code
}

func (e *defaultAPIError) Error() string {
	return e.msg
}

func newDefaultAPIError(msg string) *defaultAPIError {
	return &defaultAPIError{0, msg}
}

// errors.
var (
	// return errors.
	InvalidArgumentError    = &defaultAPIError{1, "invalid argument"}
	SessionExceptionError   = &defaultAPIError{2, "session exception"}
	MsgSizeExceedLimitError = &defaultAPIError{3, "msg size exceed limit"}
	InvalidChatTypeError    = &defaultAPIError{4, "invalid chat type"}
)
