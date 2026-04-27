package middleware

import (
	"time"

	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// Reject if request is made before limit. FORMAT: YYYY-MM-DD hh:mm:ss
func NotBefore(limit string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		waktu, err := time.Parse("2006-01-02 15:04:05", limit)
		if err != nil {
			res := response.NewFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.ErrFailedParseTime, nil)
			res.SendWithAbort(ctx)
			return
		}

		now := time.Now()
		if now.Before(waktu) {
			res := response.NewFailed(dto.MESSAGE_FAILED_OUT_OF_TIME, dto.ErrFailedProsesRequest, nil)
			res.SendWithAbort(ctx)
			return
		}

		ctx.Next()
	}
}

// Reject if request is made after limit. FORMAT: YYYY-MM-DD hh:mm:ss
func NotAfter(limit string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		waktu, err := time.Parse("2006-01-02 15:04:05", limit)
		if err != nil {
			response.NewFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.ErrFailedParseTime, nil).SendWithAbort(ctx)
			return
		}

		now := time.Now()
		if now.After(waktu) {
			response.NewFailed(dto.MESSAGE_FAILED_OUT_OF_TIME, dto.ErrOutOfTime, nil).SendWithAbort(ctx)
			return
		}

		ctx.Next()
	}
}
