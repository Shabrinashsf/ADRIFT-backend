package controller

import (
	"context"
	"net/http"
	"time"

	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/dto"
	myerror "ADRIFT-backend/internal/pkg/error"
	"ADRIFT-backend/internal/pkg/response"
	"ADRIFT-backend/internal/pkg/validate"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	SkillTreeController interface {
		GetGraph(ctx *gin.Context)
		GetNodeDetail(ctx *gin.Context)
		GetNodeChain(ctx *gin.Context)
		GetProgressGraph(ctx *gin.Context)
		GetProgressSummary(ctx *gin.Context)
		ClaimCourse(ctx *gin.Context)
		UnclaimCourse(ctx *gin.Context)
	}

	skillTreeController struct {
		skillTreeService service.SkillTreeService
		validator        *validate.Validator
	}
)

func NewSkillTreeController(sts service.SkillTreeService, validator *validate.Validator) SkillTreeController {
	return &skillTreeController{
		skillTreeService: sts,
		validator:        validator,
	}
}

func (c *skillTreeController) GetGraph(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	result, err := c.skillTreeService.GetGraph(reqCtx)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_GRAPH, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_GRAPH, result).Send(ctx)
}

func (c *skillTreeController) GetNodeDetail(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	courseID, err := uuid.Parse(ctx.Param("courseId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_NODE, myerror.New("invalid course id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	result, err := c.skillTreeService.GetNodeDetail(reqCtx, courseID)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_NODE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_NODE, result).Send(ctx)
}

func (c *skillTreeController) GetNodeChain(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	courseID, err := uuid.Parse(ctx.Param("courseId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_CHAIN, myerror.New("invalid course id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	result, err := c.skillTreeService.GetNodeChain(reqCtx, courseID)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_CHAIN, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_CHAIN, result).Send(ctx)
}

func (c *skillTreeController) GetProgressGraph(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_PROGRESS, myerror.New("invalid user token", http.StatusUnauthorized), nil).Send(ctx)
		return
	}

	result, err := c.skillTreeService.GetProgressGraph(reqCtx, userID)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_PROGRESS, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_PROGRESS, result).Send(ctx)
}

func (c *skillTreeController) GetProgressSummary(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_SUMMARY, myerror.New("invalid user token", http.StatusUnauthorized), nil).Send(ctx)
		return
	}

	result, err := c.skillTreeService.GetProgressSummary(reqCtx, userID, 0)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_SUMMARY, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_SUMMARY, result).Send(ctx)
}

func (c *skillTreeController) ClaimCourse(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CLAIM_COURSE, myerror.New("invalid user token", http.StatusUnauthorized), nil).Send(ctx)
		return
	}

	courseID, err := uuid.Parse(ctx.Param("courseId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CLAIM_COURSE, myerror.New("invalid course id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	var req dto.ClaimCourseRequest
	// Grade is optional, ignore bind error
	_ = ctx.ShouldBindJSON(&req)

	result, err := c.skillTreeService.ClaimCourse(reqCtx, userID, courseID, req.Grade)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CLAIM_COURSE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_CLAIM_COURSE, result).Send(ctx)
}

func (c *skillTreeController) UnclaimCourse(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userID, err := uuid.Parse(ctx.GetString("user_id"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UNCLAIM, myerror.New("invalid user token", http.StatusUnauthorized), nil).Send(ctx)
		return
	}

	courseID, err := uuid.Parse(ctx.Param("courseId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UNCLAIM, myerror.New("invalid course id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	result, err := c.skillTreeService.UnclaimCourse(reqCtx, userID, courseID)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UNCLAIM, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_UNCLAIM_COURSE, result).Send(ctx)
}
