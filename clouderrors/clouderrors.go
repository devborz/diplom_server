package clouderrors

type Error interface {
	Error() map[string]any
}

type CloudError struct {
	code    int
	message string
}

func (e CloudError) Error() map[string]any {
	return map[string]any{
		"error": map[string]any{
			"code":    e.code,
			"message": e.message,
		},
	}
}

var (
	ErrInvalidData                = CloudError{1, "invalid data"}
	ErrShortPassword              = CloudError{2, "the password must be longer than 8 characters"}
	ErrWrongPasswordPolicy        = CloudError{3, "the password must include uppercase and lowercase letters, numbers, and special characters"}
	ErrEmailIsAlreadyTaken        = CloudError{4, "email is already taken"}
	ErrInvalidEmail               = CloudError{5, "invalid email"}
	ErrRegistration               = CloudError{6, "registration failed"}
	ErrInvalidAuthenticationToken = CloudError{7, "invalid authentication token"}
	ErrMissingAuthenticationToken = CloudError{8, "missing authentication token"}
	ErrInvalidCredentials         = CloudError{9, "invalid credentials"}
	ErrLogin                      = CloudError{10, "login failed"}
	ErrMissingFilePath            = CloudError{11, "missing filepath"}
	ErrInvalidFilePath            = CloudError{12, "invalid filepath"}
	ErrResourceExists             = CloudError{13, "resource with the same path already exists"}
)
