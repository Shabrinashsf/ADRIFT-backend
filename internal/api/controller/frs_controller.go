package controller

import (
	"context"
	"time"

	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/dto"
	myerror "ADRIFT-backend/internal/pkg/error"
	"ADRIFT-backend/internal/pkg/pagination"
	"ADRIFT-backend/internal/pkg/response"
	"ADRIFT-backend/internal/pkg/validate"

	"github.com/gin-gonic/gin"
)

type (
	FRSController interface {
		UploadScheduleFile(ctx *gin.Context)
		DeleteScheduleArtifacts(ctx *gin.Context)
		SubmitSchedule(ctx *gin.Context)
		ListSchedules(ctx *gin.Context)
		CreateFRSPlan(ctx *gin.Context)
		FindAlternatives(ctx *gin.Context)
		ListFRSPlans(ctx *gin.Context)
		GetFRSPlanDetail(ctx *gin.Context)
		DeleteFRSPlan(ctx *gin.Context)
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

func (c *frsController) ListSchedules(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	meta := pagination.New(ctx, "academic_year", "term", "prodi", "semester", "course_name")

	result, meta, err := c.frsService.ListSchedules(reqCtx, meta)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_LIST_SCHEDULES, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_LIST_SCHEDULES, result, meta).Send(ctx)
}

func (c *frsController) CreateFRSPlan(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userId := ctx.MustGet("user_id").(string)

	var req dto.CreateFRSPlanRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	if err := c.frsService.CreateFRSPlan(reqCtx, userId, req); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CREATE_PLAN, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_CREATE_PLAN, nil).Send(ctx)
}

func (c *frsController) FindAlternatives(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.AlternativeScheduleRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.frsService.FindAlternatives(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_FIND_ALTERNATIVE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_FIND_ALTERNATIVE, result).Send(ctx)
}

func (c *frsController) ListFRSPlans(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userId := ctx.MustGet("user_id").(string)

	result, err := c.frsService.ListFRSPlans(reqCtx, userId)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_LIST_PLANS, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_LIST_PLANS, result).Send(ctx)
}

func (c *frsController) GetFRSPlanDetail(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userId := ctx.MustGet("user_id").(string)
	planId := ctx.Param("planId")

	result, err := c.frsService.GetFRSPlanDetail(reqCtx, planId, userId)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_PLAN_DETAIL, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_PLAN_DETAIL, result).Send(ctx)
}

func (c *frsController) DeleteFRSPlan(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userId := ctx.MustGet("user_id").(string)
	planId := ctx.Param("planId")

	if err := c.frsService.DeleteFRSPlan(reqCtx, planId, userId); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_PLAN, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_DELETE_PLAN, nil).Send(ctx)
}
