package controller

import (
	"context"
	"time"

	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/pkg/response"
	"ADRIFT-backend/internal/pkg/validate"

	"github.com/gin-gonic/gin"
)

type (
	FRSController interface {
		UploadScheduleFile(ctx *gin.Context)
		DeleteScheduleArtifacts(ctx *gin.Context)
		SubmitSchedule(ctx *gin.Context)
	}

	frsController struct {
		frsService service.FRSService
		validator  *validate.Validator
	}
)

func NewFRSController(frsService service.FRSService, validator *validate.Validator) FRSController {
	return &frsController{
		frsService: frsService,
		validator:  validator,
	}
}

func (c *frsController) UploadScheduleFile(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.FRSUploadRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.frsService.UploadScheduleFile(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_PROCESS_FRS, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_PROCESS_FRS, result).Send(ctx)
}

func (c *frsController) DeleteScheduleArtifacts(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.FRSUploadDeleteRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	if err := c.frsService.DeleteScheduleArtifacts(reqCtx, req); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_FRS, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_DELETE_FRS, nil).Send(ctx)
}

func (c *frsController) SubmitSchedule(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.FRSSubmitRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.frsService.SubmitSchedule(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_SUBMIT_FRS, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_SUBMIT_FRS, result).Send(ctx)
}
