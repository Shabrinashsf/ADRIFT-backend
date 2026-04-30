package controller

import (
	"context"
	"strconv"
	"time"

	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/dto"
	myerror "ADRIFT-backend/internal/pkg/error"
	"ADRIFT-backend/internal/pkg/response"
	"ADRIFT-backend/internal/pkg/validate"

	"github.com/gin-gonic/gin"
)

type (
	AdminController interface {
		ListCourseGroups(ctx *gin.Context)
		ListCoursesBySemester(ctx *gin.Context)
		CreateCourse(ctx *gin.Context)
		UpdateCourse(ctx *gin.Context)
		DeleteCourse(ctx *gin.Context)
		ListSchedules(ctx *gin.Context)
		CreateSchedule(ctx *gin.Context)
		UpdateSchedule(ctx *gin.Context)
		DeleteSchedule(ctx *gin.Context)
	}

	adminController struct {
		adminService service.AdminService
		validator    *validate.Validator
	}
)

func NewAdminController(ads service.AdminService, validator *validate.Validator) AdminController {
	return &adminController{
		adminService: ads,
		validator:    validator,
	}
}

func (c *adminController) ListCourseGroups(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	result, err := c.adminService.ListCourseGroups(reqCtx)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_LIST_COURSE_GROUPS, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_LIST_COURSE_GROUPS, result).Send(ctx)
}

func (c *adminController) ListCoursesBySemester(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	semesterStr := ctx.Param("semester")
	semester, err := strconv.Atoi(semesterStr)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_LIST_COURSES_SEMESTER, dto.ErrInvalidSemester, nil).Send(ctx)
		return
	}

	courseName := ctx.Query("course_name")

	result, err := c.adminService.ListCoursesBySemester(reqCtx, semester, courseName)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_LIST_COURSES_SEMESTER, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_LIST_COURSES_SEMESTER, result).Send(ctx)
}

func (c *adminController) CreateCourse(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.AdminCreateCourseRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	if err := c.adminService.CreateCourse(reqCtx, req); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CREATE_COURSE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_CREATE_COURSE, nil).Send(ctx)
}

func (c *adminController) UpdateCourse(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	courseID := ctx.Param("courseId")

	var req dto.AdminUpdateCourseRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.UpdateCourse(reqCtx, courseID, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPDATE_COURSE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_UPDATE_COURSE, result).Send(ctx)
}

func (c *adminController) DeleteCourse(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	courseID := ctx.Param("courseId")

	if err := c.adminService.DeleteCourse(reqCtx, courseID); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_COURSE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_DELETE_COURSE, nil).Send(ctx)
}

func (c *adminController) ListSchedules(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	academicYear := ctx.Query("academic_year")
	term := ctx.Query("term")
	prodi := ctx.Query("prodi")
	semester := ctx.Query("semester")
	courseName := ctx.Query("course_name")

	if academicYear == "" && term == "" && prodi == "" && semester == "" {
		result, err := c.adminService.ListScheduleGroups(reqCtx)
		if err != nil {
			response.NewFailed(dto.MESSAGE_FAILED_LIST_SCHEDULE_GROUPS, myerror.FromDBError(err), nil).Send(ctx)
			return
		}
		response.NewSuccess(dto.MESSAGE_SUCCESS_LIST_SCHEDULE_GROUPS, result).Send(ctx)
		return
	}

	if academicYear == "" || term == "" || prodi == "" || semester == "" {
		response.NewFailed(dto.MESSAGE_FAILED_LIST_SCHEDULES_FILTER, dto.ErrScheduleFilterMissing, nil).Send(ctx)
		return
	}

	result, err := c.adminService.ListSchedulesByFilter(reqCtx, academicYear, term, prodi, semester, courseName)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_LIST_SCHEDULES_FILTER, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_LIST_SCHEDULES_FILTER, result).Send(ctx)
}

func (c *adminController) CreateSchedule(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.AdminCreateScheduleRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	if err := c.adminService.CreateSchedule(reqCtx, req); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CREATE_SCHEDULE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_CREATE_SCHEDULE, nil).Send(ctx)
}

func (c *adminController) UpdateSchedule(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	scheduleID := ctx.Param("scheduleId")

	var req dto.AdminUpdateScheduleRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.UpdateSchedule(reqCtx, scheduleID, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPDATE_SCHEDULE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_UPDATE_SCHEDULE, result).Send(ctx)
}

func (c *adminController) DeleteSchedule(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	scheduleID := ctx.Param("scheduleId")

	if err := c.adminService.DeleteSchedule(reqCtx, scheduleID); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_SCHEDULE, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_DELETE_SCHEDULE, nil).Send(ctx)
}