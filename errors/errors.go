package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/spf13/cast"
)

var logger = logging.MustGetLogger("rewards.collection.errors")

// APIError defines the baas server API error
type APIError interface {
	error
	Write(c *gin.Context)
	WriteAbort(c *gin.Context)
	GetCode() Code
}

type apiErrorImp struct {
	Status    int    `json:"status"`
	Errorcode Code   `json:"code"`
	Message   string `json:"message"`
}

func (e *apiErrorImp) Error() string {
	return e.Message
}

func (e *apiErrorImp) Write(c *gin.Context) {
	logger.Errorf("errStatus: %d, errCode: %#08X, errMessage: APIError->%s", e.Status, e.Errorcode, e.Message)
	c.JSON(e.Status, e)
	err := c.Error(e)
	if err != nil {
		logger.Errorf("Api error implement write err: '%v'", err)
	}
}

func (e *apiErrorImp) WriteAbort(c *gin.Context) {
	logger.Errorf("errStatus: %d, errCode: %#08X, errMessage: APIError->%s", e.Status, e.Errorcode, e.Message)
	c.JSON(e.Status, e)
	err := c.Error(e)
	if err != nil {
		logger.Errorf("Api error implement write err: '%v'", err)
	}
	c.Abort()
}

func (e *apiErrorImp) GetCode() Code {
	return e.Errorcode
}

//BadRequestError creates a API error with StatusBadRequest http status
func BadRequestError(errorcode Code, msg interface{}) APIError {
	return apiError(http.StatusBadRequest, errorcode, cast.ToString(msg))
}

// BadRequestErrorf formats an error message and creates a API error with StatusBadRequest http status
func BadRequestErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusBadRequest, errorcode, format, arg...)
}

// ForbiddenError creates a API error with StatusForbidden http status
func ForbiddenError(errorcode Code, message interface{}) APIError {
	return apiError(http.StatusForbidden, errorcode, cast.ToString(message))
}

// ForbiddenErrorf formats an error message and creates a API error with StatusForbidden http status
func ForbiddenErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusForbidden, errorcode, format, arg...)
}

// NotFoundError creates a API error with StatusNotFound http status
func NotFoundError(errorcode Code, message interface{}) APIError {
	return apiError(http.StatusNotFound, errorcode, cast.ToString(message))
}

// NotFoundErrorf formats an error message and creates a API error with StatusNotFound http status
func NotFoundErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusNotFound, errorcode, format, arg...)
}

// UnauthorizedError creates a API error with StatusUnauthorized http status
func UnauthorizedError(errorcode Code, message interface{}) APIError {
	return apiError(http.StatusUnauthorized, errorcode, cast.ToString(message))
}

// UnauthorizedErrorf formats an error message and creates a API error with StatusUnauthorized http status
func UnauthorizedErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusUnauthorized, errorcode, format, arg...)
}

// ConflictError creates a API error with StatusConflict http status
func ConflictError(errorcode Code, message interface{}) APIError {
	return apiError(http.StatusConflict, errorcode, cast.ToString(message))
}

// ConflictErrorf formats an error message and creates a API error with StatusConflict http status
func ConflictErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusConflict, errorcode, format, arg...)
}

// TooManyRequestsError creates a API error with StatusTooManyRequests http status
func TooManyRequestsError(errorcode Code, message interface{}) APIError {
	return apiError(http.StatusTooManyRequests, errorcode, cast.ToString(message))
}

// TooManyRequestsErrorf formats an error message and creates a API error with StatusTooManyRequests http status
func TooManyRequestsErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusTooManyRequests, errorcode, format, arg...)
}

// InternalServerErrorf formats an error message and creates a API error with StatusInternalServerError http status
func InternalServerErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusInternalServerError, errorcode, format, arg...)
}

// MethodNotAllowedErrorf formats an error message and creates a API error with StatusMethodNotAllowed http status
func MethodNotAllowedErrorf(errorcode Code, format string, arg ...interface{}) APIError {
	return apiErrorf(http.StatusMethodNotAllowed, errorcode, format, arg...)
}

// ToAPIError convert an error to an unknown internal server error
func ToAPIError(err error) APIError {
	return toAPIError(UnknownError, "", err)
}

// DatabaseToAPIError converts an error caused by DB operation to API error
func DatabaseToAPIError(err error) APIError {
	return toAPIError(DatabaseError, "Database error", err)
}

func toAPIError(errcode Code, msg string, err error) APIError {
	switch e := err.(type) {
	case nil:
		panic("err is nil")
	case APIError:
		return e
	}

	errmsg := msg
	if errmsg == "" {
		errmsg = err.Error()
	}

	logger.Errorf("Unexpected server error: errorCode: %d, errorMessage: %s, detail: %s", errcode, errmsg, err)
	return &apiErrorImp{
		Status:    http.StatusInternalServerError,
		Errorcode: errcode,
		Message:   errmsg,
	}
}

func apiError(status int, errorcode Code, msg string) APIError {
	return &apiErrorImp{
		Status:    status,
		Errorcode: errorcode,
		Message:   msg,
	}
}

func apiErrorf(status int, errorcode Code, format string, arg ...interface{}) APIError {
	var msg = format
	if len(arg) > 0 {
		msg = fmt.Sprintf(format, arg...)
	}

	return &apiErrorImp{
		Status:    status,
		Errorcode: errorcode,
		Message:   msg,
	}
}
