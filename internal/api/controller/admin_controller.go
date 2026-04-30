package controller

import (
	"context"
	"net/http"
	"strconv"
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
	AdminController interface {
		// Course
		ListCourseGroups(ctx *gin.Context)
		ListCoursesBySemester(ctx *gin.Context)
		CreateCourse(ctx *gin.Context)
		UpdateCourse(ctx *gin.Context)
		DeleteCourse(ctx *gin.Context)

		// Schedule
		ListSchedules(ctx *gin.Context)
		CreateSchedule(ctx *gin.Context)
		UpdateSchedule(ctx *gin.Context)
		DeleteSchedule(ctx *gin.Context)

		// Lab Path
		GetAllLabPaths(ctx *gin.Context)
		CreateLabPath(ctx *gin.Context)
		UpdateLabPath(ctx *gin.Context)
		DeleteLabPath(ctx *gin.Context)

		// Prerequisite
		CreatePrerequisite(ctx *gin.Context)
		DeletePrerequisite(ctx *gin.Context)

		// Path Edge
		CreatePathEdge(ctx *gin.Context)
		DeletePathEdge(ctx *gin.Context)

		// Lecture
		GetAllLectures(ctx *gin.Context)
		CreateLecture(ctx *gin.Context)
		UpdateLecture(ctx *gin.Context)
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

// =========== COURSE ===========

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

// =========== SCHEDULE ===========

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

// =========== LAB PATH ===========

func (c *adminController) GetAllLabPaths(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	result, err := c.adminService.GetAllLabPaths(reqCtx)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_LAB_PATHS, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_LAB_PATHS, result).Send(ctx)
}

func (c *adminController) CreateLabPath(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.CreateLabPathRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.CreateLabPath(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CREATE_LAB_PATH, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_CREATE_LAB_PATH, result).Send(ctx)
}

func (c *adminController) UpdateLabPath(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	labPathID, err := uuid.Parse(ctx.Param("labPathId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPDATE_LAB_PATH, myerror.New("invalid lab path id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	var req dto.UpdateLabPathRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.UpdateLabPath(reqCtx, labPathID, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPDATE_LAB_PATH, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_UPDATE_LAB_PATH, result).Send(ctx)
}

func (c *adminController) DeleteLabPath(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	labPathID, err := uuid.Parse(ctx.Param("labPathId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_LAB_PATH, myerror.New("invalid lab path id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	if err := c.adminService.DeleteLabPath(reqCtx, labPathID); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_LAB_PATH, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_DELETE_LAB_PATH, nil).Send(ctx)
}

// =========== PREREQUISITE ===========

func (c *adminController) CreatePrerequisite(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.CreatePrerequisiteRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.CreatePrerequisite(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CREATE_PREREQUISITE, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_CREATE_PREREQUISITE, result).Send(ctx)
}

func (c *adminController) DeletePrerequisite(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	courseID, err := uuid.Parse(ctx.Param("courseId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_PREREQUISITE, myerror.New("invalid course id", http.StatusBadRequest), nil).Send(ctx)
		return
	}
	requireID, err := uuid.Parse(ctx.Param("requireId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_PREREQUISITE, myerror.New("invalid require id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	if err := c.adminService.DeletePrerequisite(reqCtx, courseID, requireID); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_PREREQUISITE, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_DELETE_PREREQUISITE, nil).Send(ctx)
}

// =========== PATH EDGE ===========

func (c *adminController) CreatePathEdge(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.CreatePathEdgeRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.CreatePathEdge(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CREATE_PATH_EDGE, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_CREATE_PATH_EDGE, result).Send(ctx)
}

func (c *adminController) DeletePathEdge(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	pathEdgeID, err := uuid.Parse(ctx.Param("pathEdgeId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_PATH_EDGE, myerror.New("invalid path edge id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	if err := c.adminService.DeletePathEdge(reqCtx, pathEdgeID); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_DELETE_PATH_EDGE, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_DELETE_PATH_EDGE, nil).Send(ctx)
}

// =========== LECTURE ===========

func (c *adminController) GetAllLectures(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	result, err := c.adminService.GetAllLectures(reqCtx)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_LECTURES, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_LECTURES, result).Send(ctx)
}

func (c *adminController) CreateLecture(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.CreateLectureRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.CreateLecture(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_CREATE_LECTURE, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_CREATE_LECTURE, result).Send(ctx)
}

func (c *adminController) UpdateLecture(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	lectureID, err := uuid.Parse(ctx.Param("lectureId"))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPDATE_LECTURE, myerror.New("invalid lecture id", http.StatusBadRequest), nil).Send(ctx)
		return
	}

	var req dto.UpdateLectureRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.adminService.UpdateLecture(reqCtx, lectureID, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPDATE_LECTURE, err, nil).Send(ctx)
		return
	}
	response.NewSuccess(dto.MESSAGE_SUCCESS_UPDATE_LECTURE, result).Send(ctx)
}
