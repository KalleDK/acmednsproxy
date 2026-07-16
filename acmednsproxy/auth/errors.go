package auth

// Sentinel errors returned by Authenticator implementations.
// Callers can use errors.Is to distinguish the failure reason.
type errorString string

func (e errorString) Error() string {
	return string(e)
}

const (
	// ErrUnauthorized is returned when credentials are present but do not
	// grant access to the requested domain.
	ErrUnauthorized = errorString("unauthorized")

	// ErrUnknownDomain is returned when the requested domain is not
	// registered with the authenticator.
	ErrUnknownDomain = errorString("unknown domain")

	// ErrUnknownUser is returned when the username supplied in the
	// credentials is not registered for the requested domain.
	ErrUnknownUser = errorString("unknown user")
)
