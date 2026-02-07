package apperror

// New is the generic constructor for AppError.
func New(code, message string, kind Kind, err ...error) *AppError {
	appErr := &AppError{
		Code:    code,
		Message: message,
		Kind:    kind,
	}
	if len(err) > 0 && err[0] != nil {
		appErr.Err = err[0]
	}
	return appErr
}

// NewPersistance creates an error with KindPersistance.
// Optional: Pass an existing error as the 3rd argument to wrap it.
func NewPersistance(code, message string, err ...error) *AppError {
	return New(code, message, KindPersistance, err...)
}

// NewTransient creates an error with KindTransient.
// Optional: Pass an existing error as the 3rd argument to wrap it.
func NewTransient(code, message string, err ...error) *AppError {
	return New(code, message, KindTransient, err...)
}

// NewInternal creates an error with KindInternal.
// Optional: Pass an existing error as the 3rd argument to wrap it.
func NewInternal(code, message string, err ...error) *AppError {
	return New(code, message, KindInternal, err...)
}
