package auth

type errorString string

func (e errorString) Error() string {
	return string(e)
}

const (
	ErrUnauthorized  = errorString("unauthorized")
	ErrUnknownDomain = errorString("unknown domain")
	ErrUnknownUser   = errorString("unknown user")
)
