package controller

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/dto"
	myerror "ADRIFT-backend/internal/pkg/error"
	"ADRIFT-backend/internal/pkg/response"
	"ADRIFT-backend/internal/pkg/storage"
	"ADRIFT-backend/internal/pkg/validate"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	UserController interface {
		RegisterUser(ctx *gin.Context)
		Login(ctx *gin.Context)
		SendVerificationEmail(ctx *gin.Context)
		VerifyEmail(ctx *gin.Context)
		ForgotPassword(ctx *gin.Context)
		ResetPassword(ctx *gin.Context)
		MeAuth(ctx *gin.Context)
		UpdateUser(ctx *gin.Context)
	}

	userController struct {
		userService service.UserService
		validator   *validate.Validator
		storage     storage.FileSystemStorage
	}
)

func NewUserController(us service.UserService, validator *validate.Validator, storage storage.FileSystemStorage) UserController {
	return &userController{
		userService: us,
		validator:   validator,
		storage:     storage,
	}
}

func (c *userController) RegisterUser(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.UserRegistrationRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.userService.RegisterUser(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_REGISTER_USER, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_REGISTER_USER, result).Send(ctx)
}

func (c *userController) Login(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.UserLoginRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.userService.Login(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_LOGIN_USER, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_LOGIN_USER, result).Send(ctx)
}

func (c *userController) SendVerificationEmail(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.SendVerificationEmailRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	err := c.userService.SendVerificationEmail(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SEND_VERIFICATION_EMAIL_SUCCESS, nil).Send(ctx)
}

func (c *userController) VerifyEmail(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	token := ctx.Query("token")

	if token == "" {
		response.NewFailed(dto.MESSAGE_FAILED_TOKEN_NOT_FOUND, dto.ErrTokenNotFound, nil).Send(ctx)
		return
	}

	req := dto.VerifyEmailRequest{
		Token: token,
	}

	if err := ctx.ShouldBind(&req); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, myerror.New(myerror.FormatValidationError(err), http.StatusBadRequest), nil).Send(ctx)
		return
	}

	result, err := c.userService.VerifyEmail(reqCtx, req)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_VERIFY_EMAIL, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_VERIFY_EMAIL, result).Send(ctx)
}

func (c *userController) ForgotPassword(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	var req dto.ForgotPasswordRequest
	if !c.validator.Bind(ctx, &req) {
		return
	}

	if err := c.userService.ForgotPassword(reqCtx, req); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_FORGET_PASSWORD, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_FORGET_PASSWORD, nil).Send(ctx)
}

func (c *userController) ResetPassword(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	token := ctx.Query("token")
	var req dto.ResetPasswordRequest

	if !c.validator.Bind(ctx, &req) {
		return
	}

	if err := c.userService.ResetPassword(reqCtx, token, req.Password); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_RESET_PASSWORD, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_RESET_PASSWORD, nil).Send(ctx)
}

func (c *userController) MeAuth(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()
	userId := ctx.MustGet("user_id").(string)

	result, err := c.userService.GetUserByID(reqCtx, uuid.MustParse(userId))
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_GET_USER, result).Send(ctx)
}

func (c *userController) UpdateUser(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 20*time.Second)
	defer cancel()

	userId := ctx.MustGet("user_id").(string)
	userUUID := uuid.MustParse(userId)
	var req dto.UserUpdateRequest

	if !c.validator.Bind(ctx, &req) {
		return
	}

	result, err := c.userService.UpdateUser(reqCtx, userUUID, req)
	if err != nil {
		c.rollbackIfNeeded()
		response.NewFailed(dto.MESSAGE_FAILED_UPDATE_USER, myerror.FromDBError(err), nil).Send(ctx)
		return
	}

	c.commitIfNeeded()
	response.NewSuccess(dto.MESSAGE_SUCCESS_UPDATE_USER, result).Send(ctx)
}

func (c *userController) handleProfileUpload(file *multipart.FileHeader, email string) (*string, error) {
	mime := []string{"image/jpeg", "image/jpg", "image/png"}

	_, ext, err := validate.ValidateFile(file, 1*1024*1024, mime...)
	if err != nil {
		return nil, err
	}

	c.storage.Begin()

	uploadFilename := fmt.Sprintf("%s%s", email, ext)
	objectKey, err := c.storage.UploadFile(uploadFilename, file, "profiles")
	if err != nil {
		c.storage.Rollback()
		return nil, err
	}

	publicLink := c.storage.GetPublicLink(objectKey)
	return &publicLink, nil
}

func (c *userController) handleProfileReplacement(file *multipart.FileHeader, currentProfile, email string) (*string, error) {
	mime := []string{"image/jpeg", "image/jpg", "image/png"}

	_, ext, err := validate.ValidateFile(file, 1*1024*1024, mime...)
	if err != nil {
		return nil, err
	}

	c.storage.Begin()

	oldObjectKey := c.extractProfileObjectKey(currentProfile)
	if oldObjectKey != "" {
		if err := c.storage.DeleteFile(oldObjectKey); err != nil {
			c.storage.Rollback()
			return nil, err
		}
	}

	uploadFilename := fmt.Sprintf("%s%s", email, ext)
	objectKey, err := c.storage.UploadFile(uploadFilename, file, "profiles")
	if err != nil {
		c.storage.Rollback()
		return nil, err
	}

	publicLink := c.storage.GetPublicLink(objectKey)
	return &publicLink, nil
}

func (c *userController) extractProfileObjectKey(profile string) string {
	if profile == "" {
		return ""
	}

	if parsedURL, err := url.Parse(profile); err == nil && parsedURL.Path != "" {
		profile = parsedURL.Path
	}

	profile = strings.TrimPrefix(profile, "/")
	profile = strings.TrimPrefix(profile, "api/assets/")

	return profile
}

func (c *userController) rollbackIfNeeded() {
	c.storage.Rollback()
}

func (c *userController) commitIfNeeded() {
	c.storage.Commit()
}
