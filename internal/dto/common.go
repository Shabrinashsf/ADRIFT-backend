package dto

import (
	"net/http"

	myerror "ADRIFT-backend/internal/pkg/error"
)

const (
	// Failed Messages
	MESSAGE_FAILED_PROSES_REQUEST      = "failed to process request"
	MESSAGE_FAILED_TOKEN_NOT_FOUND     = "failed token not found"
	MESSAGE_FAILED_TOKEN_NOT_VALID     = "failed token not valid"
	MESSAGE_FAILED_GET_DATA_FROM_BODY  = "failed to get data from body"
	MESSAGE_FAILED_GET_CALLBACK_TRIPAY = "failed to get callback from tripay"
	MESSAGE_FAILED_OUT_OF_TIME         = "request made outside of allowed time frame"

	// Success Messages
	MESSAGE_SUCCESS_GET_CALLBACK_TRIPAY = "success get callback from tripay"
)

var (
	ErrRoleNotAllowed      = myerror.New("role not allowed", http.StatusForbidden)
	ErrFailedParseTime     = myerror.New("failed to parse time", http.StatusInternalServerError)
	ErrFailedProsesRequest = myerror.New("failed to process request", http.StatusInternalServerError)
	ErrOutOfTime           = myerror.New("request made outside of allowed time frame", http.StatusBadRequest)
	ErrTokenNotFound       = myerror.New("token not found", http.StatusUnauthorized)
	ErrTokenNotValid       = myerror.New("token not valid", http.StatusUnauthorized)
	ErrDeniedAccess        = myerror.New("denied access", http.StatusUnauthorized)
	ErrInvalidTypeFile     = myerror.New("invalid type file", http.StatusBadRequest)
)
