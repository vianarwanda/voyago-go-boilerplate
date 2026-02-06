package apperror

// Infrastructure error codes (Database, Network, External)
const (
	CodeDbConnectionFailed = "DB_CONNECTION_FAILED" // HTTP Status 500
	CodeDbTimeout          = "DB_TIMEOUT"           // HTTP Status 500
	CodeDbDeadlock         = "DB_DEADLOCK"          // HTTP Status 500
	CodeDbConstraint       = "DB_CONSTRAINT"        // HTTP Status 500
	CodeDbConflict         = "DB_CONFLICT"          // HTTP Status 500
	CodeInternalError      = "INTERNAL_ERROR"       // HTTP Status 500
)

const (
	CodeMalformedRequest              = "MALFORMED_REQUEST"               // HTTP Status 400
	CodeInvalidRequest                = "INVALID_REQUEST"                 // HTTP Status 400
	CodeValidation                    = "VALIDATION_ERROR"                // HTTP Status 400
	CodeUnauthorized                  = "UNAUTHORIZED"                    // HTTP Status 401
	CodeForbidden                     = "FORBIDDEN"                       // HTTP Status 403
	CodeNotFound                      = "NOT_FOUND"                       // HTTP Status 404
	CodeMethodNotAllowed              = "METHOD_NOT_ALLOWED"              // HTTP Status 405
	CodeNotAcceptable                 = "NOT_ACCEPTABLE"                  // HTTP Status 406
	CodeRequestTimeout                = "REQUEST_TIMEOUT"                 // HTTP Status 408
	CodeConflict                      = "CONFLICT"                        // HTTP Status 409
	CodeGone                          = "GONE"                            // HTTP Status 410
	CodeLengthRequired                = "LENGTH_REQUIRED"                 // HTTP Status 411
	CodePreconditionFailed            = "PRECONDITION_FAILED"             // HTTP Status 412
	CodePayloadTooLarge               = "PAYLOAD_TOO_LARGE"               // HTTP Status 413
	CodeURITooLong                    = "URI_TOO_LONG"                    // HTTP Status 414
	CodeUnsupportedMediaType          = "UNSUPPORTED_MEDIA_TYPE"          // HTTP Status 415
	CodeRangeNotSatisfiable           = "RANGE_NOT_SATISFIABLE"           // HTTP Status 416
	CodeExpectationFailed             = "EXPECTATION_FAILED"              // HTTP Status 417
	CodeTeapot                        = "TEAPOT"                          // HTTP Status 418
	CodeMisdirectedRequest            = "MISDIRECTED_REQUEST"             // HTTP Status 421
	CodeUnprocessableEntity           = "UNPROCESSABLE_ENTITY"            // HTTP Status 422
	CodeLocked                        = "LOCKED"                          // HTTP Status 423
	CodeFailedDependency              = "FAILED_DEPENDENCY"               // HTTP Status 424
	CodeTooEarly                      = "TOO_EARLY"                       // HTTP Status 425
	CodeUpgradeRequired               = "UPGRADE_REQUIRED"                // HTTP Status 426
	CodePreconditionRequired          = "PRECONDITION_REQUIRED"           // HTTP Status 428
	CodeTooManyRequests               = "TOO_MANY_REQUESTS"               // HTTP Status 429
	CodeRequestHeaderFieldsTooLarge   = "REQUEST_HEADER_FIELDS_TOO_LARGE" // HTTP Status 431
	CodeUnavailableForLegalReasons    = "UNAVAILABLE_FOR_LEGAL_REASONS"   // HTTP Status 451
	CodeNetworkAuthenticationRequired = "NETWORK_AUTHENTICATION_REQUIRED" // HTTP Status 511
)

var (
	ErrCodeDbConnectionFailed = NewTransient(CodeDbConnectionFailed, "Database connection failed", nil)
	ErrCodeDbTimeout          = NewTransient(CodeDbTimeout, "Database timeout", nil)
	ErrCodeDbDeadlock         = NewTransient(CodeDbDeadlock, "Database deadlock", nil)
	ErrCodeDbConstraint       = NewPermanent(CodeDbConstraint, "Database constraint violation", nil)
	ErrCodeDbConflict         = NewPermanent(CodeDbConflict, "Database conflict", nil)
	ErrCodeInternalError      = NewInternal(CodeInternalError, "Internal error", nil)
)

var (
	ErrCodeMalformedRequest              = NewPermanent(CodeMalformedRequest, "Invalid JSON format or data type", nil)
	ErrCodeInvalidRequest                = NewPermanent(CodeInvalidRequest, "Invalid request", nil)
	ErrCodeValidation                    = NewPermanent(CodeValidation, "Validation error", nil)
	ErrCodeUnauthorized                  = NewPermanent(CodeUnauthorized, "Unauthorized", nil)
	ErrCodeForbidden                     = NewPermanent(CodeForbidden, "Forbidden", nil)
	ErrCodeNotFound                      = NewPermanent(CodeNotFound, "Not found", nil)
	ErrCodeMethodNotAllowed              = NewPermanent(CodeMethodNotAllowed, "Method not allowed", nil)
	ErrCodeNotAcceptable                 = NewPermanent(CodeNotAcceptable, "Not acceptable", nil)
	ErrCodeRequestTimeout                = NewPermanent(CodeRequestTimeout, "Request timeout", nil)
	ErrCodeConflict                      = NewPermanent(CodeConflict, "Conflict", nil)
	ErrCodeGone                          = NewPermanent(CodeGone, "Gone", nil)
	ErrCodeLengthRequired                = NewPermanent(CodeLengthRequired, "Length required", nil)
	ErrCodePreconditionFailed            = NewPermanent(CodePreconditionFailed, "Precondition failed", nil)
	ErrCodePayloadTooLarge               = NewPermanent(CodePayloadTooLarge, "Payload too large", nil)
	ErrCodeURITooLong                    = NewPermanent(CodeURITooLong, "URI too long", nil)
	ErrCodeUnsupportedMediaType          = NewPermanent(CodeUnsupportedMediaType, "Unsupported media type", nil)
	ErrCodeRangeNotSatisfiable           = NewPermanent(CodeRangeNotSatisfiable, "Range not satisfiable", nil)
	ErrCodeExpectationFailed             = NewPermanent(CodeExpectationFailed, "Expectation failed", nil)
	ErrCodeTeapot                        = NewPermanent(CodeTeapot, "Teapot", nil)
	ErrCodeMisdirectedRequest            = NewPermanent(CodeMisdirectedRequest, "Misdirected request", nil)
	ErrCodeUnprocessableEntity           = NewPermanent(CodeUnprocessableEntity, "Unprocessable entity", nil)
	ErrCodeLocked                        = NewPermanent(CodeLocked, "Locked", nil)
	ErrCodeFailedDependency              = NewPermanent(CodeFailedDependency, "Failed dependency", nil)
	ErrCodeTooEarly                      = NewPermanent(CodeTooEarly, "Too early", nil)
	ErrCodeUpgradeRequired               = NewPermanent(CodeUpgradeRequired, "Upgrade required", nil)
	ErrCodePreconditionRequired          = NewPermanent(CodePreconditionRequired, "Precondition required", nil)
	ErrCodeTooManyRequests               = NewPermanent(CodeTooManyRequests, "Too many requests", nil)
	ErrCodeRequestHeaderFieldsTooLarge   = NewPermanent(CodeRequestHeaderFieldsTooLarge, "Request header fields too large", nil)
	ErrCodeUnavailableForLegalReasons    = NewPermanent(CodeUnavailableForLegalReasons, "Unavailable for legal reasons", nil)
	ErrCodeNetworkAuthenticationRequired = NewPermanent(CodeNetworkAuthenticationRequired, "Network authentication required", nil)
)
