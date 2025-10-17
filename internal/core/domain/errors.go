package domain

import (
	"errors"
	"fmt"
)

type UserError string

const (
	UserErrorInvalidEmail       UserError = "invalid_email"
	UserErrorInvalidPhoneNumber UserError = "invalid_phone_number"
	UserErrorUserNotFound       UserError = "user_not_found"
	UserErrorUserAlreadyExists  UserError = "user_already_exists"
	UserErrorInvalidUserType    UserError = "invalid_user_type"
	UserErrorDatabaseError      UserError = "database_error"
	UserErrorUnauthorized       UserError = "unauthorized"
	UserErrorInvalidInput       UserError = "invalid_input"
	EmailAlreadyInUseError      UserError = "email_already_in_use"
	LoginNotFoundError          UserError = "login_not_found"

	// Http Status Codes
	UserErrorBadRequest          UserError = "bad_request"
	UserErrorNotFound            UserError = "not_found"
	UserErrorInternalServerError UserError = "internal_server_error"
	UserErrorMethodNotAllowed    UserError = "method_not_allowed"
	UserErrorConflict            UserError = "conflict"
	UserErrorTooManyRequests     UserError = "too_many_requests"
	UserErrorServiceUnavailable  UserError = "service_unavailable"
	InvalidBodyError             UserError = "invalid_body_error"
	UserNotFoundError            UserError = "user_not_found_error"
	UserNotActiveError           UserError = "user_not_active_error"

	// Auth
	InvalidOtpError              UserError = "invalid_otp_error"
	InvalidTokenError            UserError = "invalid_token_error"
	TokenExpiredError            UserError = "token_expired_error"
	AccessDeniedError            UserError = "access_denied_error"
	InvalidCredentialsError      UserError = "invalid_credentials_error"
	PasswordTooWeakError         UserError = "password_too_weak_error"
	PasswordMismatchError        UserError = "password_mismatch_error"
	EmailAlreadyVerifiedError    UserError = "email_already_verified_error"
	PhoneAlreadyVerifiedError    UserError = "phone_already_verified_error"
	OtpSendFailedError           UserError = "otp_send_failed_error"
	OtpExpiredError              UserError = "otp_expired_error"
	OtpNotFoundError             UserError = "otp_not_found_error"
	RefreshTokenNotFoundError    UserError = "refresh_token_not_found_error"
	RefreshTokenExpiredError     UserError = "refresh_token_expired_error"
	InvalidRefreshTokenError     UserError = "invalid_refresh_token_error"
	AuthTokenInvalidError        UserError = "auth_token_invalid_error"
	AuthTokenExpiredError        UserError = "auth_token_expired_error"
	InsufficientPermissionsError UserError = "insufficient_permissions_error"
	InvalidAuthTokenFormatError  UserError = "invalid_auth_token_format_error"
	InvalidAuthTokenTypeError    UserError = "invalid_auth_token_type_error"
	UnauthorizedError            UserError = "unauthorized_error"

	// Resource
	ResourceNotFoundError UserError = "resource_not_found_error"
	ResourceConflictError UserError = "resource_conflict_error"
	InvalidResourceError  UserError = "invalid_resource_error"
	UnableToProcessError  UserError = "unable_to_process_error"
	UnableToUpdateError   UserError = "unable_to_update_error"
	UnableToDeleteError   UserError = "unable_to_delete_error"
	UnableToCreateError   UserError = "unable_to_create_error"
	UnableToFetchError    UserError = "unable_to_fetch_error"

	// Form validation
	InvalidInputError      UserError = "invalid_input_error"
	UnableToMarshalError   UserError = "unable_to_marshal_error"
	UnableToUnmarshalError UserError = "unable_to_unmarshal_error"

	// Connection
	DatabaseConnectionError UserError = "database_connection_error"
	CacheConnectionError    UserError = "cache_connection_error"
	ExternalServiceError    UserError = "external_service_error"
	BucketConnectionError   UserError = "bucket_connection_error"

	/// File upload
	UnableToUploadError UserError = "unable_to_upload_error"
	UnableToDownloadError UserError = "unable_to_download_error"
)

type DomainError struct {
	Code    UserError `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s → %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *UserError) Error() string {
	return string(*e)
}

func NewDomainError(code UserError, message string, err error) *DomainError {
	if err == nil {
		return &DomainError{
			Code:    code,
			Message: message,
			Err:     errors.New("Asset not found"),
		}
	}
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
