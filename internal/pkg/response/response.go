package response

import (
	"net/http"

	myerror "ADRIFT-backend/internal/pkg/error"

	"github.com/gin-gonic/gin"
)

type Response struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Error      any    `json:"error,omitempty"`
	Data       any    `json:"data,omitempty"`
	Meta       any    `json:"meta,omitempty"`
}

func NewSuccess(message string, data any, meta ...any) Response {
	res := Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    message,
		Data:       data,
	}

	if len(meta) > 0 {
		res.Meta = meta[0]
	}

	return res
}

func NewFailed(message string, err error, data ...any) Response {
	res := Response{
		StatusCode: http.StatusInternalServerError,
		Success:    false,
		Message:    message,
		Error:      err.Error(),
	}

	if myErr, ok := err.(myerror.Error); ok {
		res.StatusCode = myErr.StatusCode
	}

	if len(data) > 0 {
		res.Data = data[0]
	}

	return res
}

func (r Response) ChangeStatusCode(statusCode int) Response {
	res := r
	res.StatusCode = statusCode
	return res
}

func (r Response) Send(ctx *gin.Context) {
	sendStatus := r.StatusCode
	ctx.JSON(sendStatus, r)
}

func (r Response) SendWithAbort(ctx *gin.Context) {
	sendStatus := r.StatusCode
	ctx.AbortWithStatusJSON(sendStatus, r)
}
