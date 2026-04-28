package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"ADRIFT-backend/internal/api/repository"
	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/entity"
	"ADRIFT-backend/internal/pkg/logger"
	"ADRIFT-backend/internal/pkg/mailer"
	"ADRIFT-backend/internal/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	UserService interface {
		RegisterUser(ctx context.Context, req dto.UserRegistrationRequest) (dto.UserResponse, error)
		Login(ctx context.Context, req dto.UserLoginRequest) (dto.UserLoginResponse, error)
		SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error
		VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error)
		ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error
		ResetPassword(ctx context.Context, token string, newPassword string) error
		GetUserByID(ctx context.Context, userId uuid.UUID) (dto.UserResponse, error)
		UpdateUser(ctx context.Context, userId uuid.UUID, req dto.UserUpdateRequest) (dto.UserResponse, error)
	}

	userService struct {
		userRepo   repository.UserRepository
		jwtService JWTService
		mailer     mailer.Mailer
		db         *gorm.DB
	}
)

func NewUserService(ur repository.UserRepository, jwt JWTService, mailer mailer.Mailer, db *gorm.DB) UserService {
	return &userService{
		userRepo:   ur,
		jwtService: jwt,
		mailer:     mailer,
		db:         db,
	}
}

var (
	VERIFY_EMAIL_TEMPLATE = "internal/pkg/mailer/template/verification_email.html"
	VERIFY_EMAIL_PATH     = "verify-email"
	FORGET_EMAIL_TEMPLATE = "internal/pkg/mailer/template/forgot_password_email.html"
	FORGET_EMAIL_PATH     = "reset-password"
)

func (s *userService) RegisterUser(ctx context.Context, req dto.UserRegistrationRequest) (dto.UserResponse, error) {
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	_, flag, _ := s.userRepo.GetUserByEmail(ctx, tx, req.Email)
	if flag {
		tx.Rollback()
		return dto.UserResponse{}, dto.ErrorEmailAlreadyExists
	}

	user := entity.User{
		NRP:            req.NRP,
		Name:           req.Name,
		Email:          req.Email,
		Password:       req.Password,
		EnrollmentYear: req.EnrollmentYear,
		Role:           entity.UserRoleStudent,
		IsVerified:     true, // Set to true for now to skip email verification
	}

	newUser, err := s.userRepo.RegisterUser(ctx, tx, user)
	if err != nil {
		tx.Rollback()
		return dto.UserResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return dto.UserResponse{}, err
	}

	// uncomment if email verification is needed, and comment out the IsVerified field in user entity struct
	// if err := tx.Commit().Error; err != nil {
	// 	return dto.UserResponse{}, err
	// }

	// expired := time.Now().Add(time.Minute * 2).Format("2006-01-02 15:04:05")
	// plainText := user.Email + "_" + expired
	// token, err := utils.AESEncrypt(plainText)
	// if err != nil {
	// 	return dto.UserResponse{}, err
	// }

	// verifyLink := os.Getenv("BE_URL") + "/" + VERIFY_EMAIL_PATH + "?token=" + token
	// data := map[string]any{
	// 	"Email":  user.Email,
	// 	"Verify": verifyLink,
	// }

	// mail := s.mailer.MakeMail(VERIFY_EMAIL_TEMPLATE, data)
	// if mail.Error != nil {
	// 	return dto.UserResponse{}, dto.ErrMakeMail
	// }

	// if err := mail.SendEmail(user.Email, "ADRIFT - Verification Email").Error; err != nil {
	// 	return dto.UserResponse{}, dto.ErrSendMail
	// }

	return dto.UserResponse{
		ID:             newUser.ID.String(),
		NRP:            newUser.NRP,
		Name:           newUser.Name,
		Email:          newUser.Email,
		EnrollmentYear: newUser.EnrollmentYear,
		Role:           string(newUser.Role),
		IsVerified:     newUser.IsVerified,
	}, nil
}

func (s *userService) Login(ctx context.Context, req dto.UserLoginRequest) (dto.UserLoginResponse, error) {
	user, flag, err := s.userRepo.GetUserByEmail(ctx, nil, req.Email)
	if err != nil || !flag {
		return dto.UserLoginResponse{}, dto.ErrInvalidCredentials
	}

	if !user.IsVerified {
		return dto.UserLoginResponse{}, dto.ErrInvalidCredentials
	}

	checkPassword, err := utils.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !checkPassword {
		return dto.UserLoginResponse{}, dto.ErrInvalidCredentials
	}

	token := s.jwtService.GenerateToken(user.ID.String(), string(user.Role))

	return dto.UserLoginResponse{
		Token: token,
		Role:  string(user.Role),
	}, nil
}

func (s *userService) SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error {
	user, _, err := s.userRepo.GetUserByEmail(ctx, nil, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Infof("send verification requested for non-existing email: %s", req.Email)
			return nil
		}

		logger.Errorf("send verification lookup error for email %s: %v", req.Email, err)
		return nil
	}

	if user.IsVerified {
		logger.Infof("send verification requested for already verified email: %s", req.Email)
		return nil
	}

	expired := time.Now().Add(time.Minute * 2).Format("2006-01-02 15:04:05")
	plainText := user.Email + "_" + expired
	token, err := utils.AESEncrypt(plainText)
	if err != nil {
		logger.Errorf("failed to generate verification token for email %s: %v", req.Email, err)
		return nil
	}

	verifyLink := os.Getenv("BE_URL") + "/" + VERIFY_EMAIL_PATH + "?token=" + token
	data := map[string]any{
		"Email":  user.Email,
		"Verify": verifyLink,
	}

	mail := s.mailer.MakeMail(VERIFY_EMAIL_TEMPLATE, data)
	if mail.Error != nil {
		logger.Errorf("failed to build verification email for %s: %v", req.Email, mail.Error)
		return nil
	}

	if err := mail.SendEmail(user.Email, "ADRIFT - Verification Email").Error; err != nil {
		logger.Errorf("failed to send verification email to %s: %v", req.Email, err)
		return nil
	}

	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error) {
	decryptedToken, err := utils.AESDecrypt(req.Token)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	if !strings.Contains(decryptedToken, "_") {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	decryptedTokenSplit := strings.Split(decryptedToken, "_")
	email := decryptedTokenSplit[0]
	expired := decryptedTokenSplit[1]

	now := time.Now()
	expiredTime, err := time.Parse("2006-01-02 15:04:05", expired)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	if expiredTime.Sub(now) < 0 {
		return dto.VerifyEmailResponse{
			Email:      email,
			IsVerified: false,
		}, dto.ErrTokenExpired
	}

	user, _, err := s.userRepo.GetUserByEmail(ctx, nil, email)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrUserNotFound
	}

	if user.IsVerified {
		return dto.VerifyEmailResponse{}, dto.ErrAccountAlreadyVerified
	}

	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	updates := map[string]interface{}{}
	updates["is_verified"] = true

	updatedUser, err := s.userRepo.UpdateUser(ctx, tx, user.ID, updates)
	if err != nil {
		tx.Rollback()
		return dto.VerifyEmailResponse{}, dto.ErrUpdateUser
	}

	if err := tx.Commit().Error; err != nil {
		return dto.VerifyEmailResponse{}, err
	}

	return dto.VerifyEmailResponse{
		Email:      email,
		IsVerified: updatedUser.IsVerified,
	}, nil
}

func (s *userService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	user, _, err := s.userRepo.GetUserByEmail(ctx, nil, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Infof("forgot password requested for non-existing email: %s", req.Email)
			return nil
		}

		logger.Errorf("forgot password lookup error for email %s: %v", req.Email, err)
		return nil
	}

	expired := time.Now().Add(time.Minute * 2).Format("2006-01-02 15:04:05")
	plainText := user.Email + "_" + expired
	token, err := utils.AESEncrypt(plainText)
	if err != nil {
		logger.Errorf("failed to generate forgot-password token for email %s: %v", req.Email, err)
		return nil
	}

	verifyLink := os.Getenv("BE_URL") + "/" + FORGET_EMAIL_PATH + "?token=" + token
	data := map[string]any{
		"Email":  user.Email,
		"Verify": verifyLink,
	}

	mail := s.mailer.MakeMail(FORGET_EMAIL_TEMPLATE, data)
	if mail.Error != nil {
		logger.Errorf("failed to build forgot-password email for %s: %v", req.Email, mail.Error)
		return nil
	}

	if err := mail.SendEmail(user.Email, "ADRIFT - Reset Password").Error; err != nil {
		logger.Errorf("failed to send forgot-password email to %s: %v", req.Email, err)
		return nil
	}

	return nil
}

func (s *userService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	decryptedToken, err := utils.AESDecrypt(token)
	if err != nil {
		return dto.ErrTokenInvalid
	}

	tokenParts := strings.Split(decryptedToken, "_")
	if len(tokenParts) < 2 {
		return dto.ErrTokenInvalid
	}

	email := tokenParts[0]
	expirationDate := tokenParts[1]
	expirationTime, err := time.Parse("2006-01-02 15:04:05", expirationDate)

	if err != nil {
		return dto.ErrTokenInvalid
	}

	if time.Now().After(expirationTime) {
		return dto.ErrTokenExpired
	}

	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		tx.Rollback()
		return dto.ErrHashPasswordFailed
	}

	err = s.userRepo.ResetPassword(ctx, email, hashedPassword)
	if err != nil {
		tx.Rollback()
		return dto.ErrUpdateUser
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *userService) GetUserByID(ctx context.Context, userId uuid.UUID) (dto.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, nil, userId)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:             user.ID.String(),
		NRP:            user.NRP,
		Name:           user.Name,
		Email:          user.Email,
		EnrollmentYear: user.EnrollmentYear,
		Role:           string(user.Role),
		IsVerified:     user.IsVerified,
	}, nil
}

func (s *userService) UpdateUser(ctx context.Context, userId uuid.UUID, req dto.UserUpdateRequest) (dto.UserResponse, error) {
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	user, err := s.userRepo.GetUserByID(ctx, tx, userId)
	if err != nil {
		tx.Rollback()
		return dto.UserResponse{}, dto.ErrUserNotFound
	}

	updates := map[string]interface{}{}

	if req.Name != nil && *req.Name != user.Name {
		updates["name"] = *req.Name
	}
	if req.NRP != nil && *req.NRP != user.NRP {
		updates["nrp"] = *req.NRP
	}
	if req.EnrollmentYear != nil && *req.EnrollmentYear != user.EnrollmentYear {
		updates["enrollment_year"] = *req.EnrollmentYear
	}

	if len(updates) == 0 {
		tx.Rollback()
		return dto.UserResponse{}, dto.ErrNoChanges
	}

	userUpdate, err := s.userRepo.UpdateUser(ctx, tx, userId, updates)
	if err != nil {
		tx.Rollback()
		return dto.UserResponse{}, dto.ErrUpdateUser
	}

	if err := tx.Commit().Error; err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:             userUpdate.ID.String(),
		NRP:            userUpdate.NRP,
		Name:           userUpdate.Name,
		Email:          userUpdate.Email,
		EnrollmentYear: userUpdate.EnrollmentYear,
		Role:           string(user.Role),
		IsVerified:     user.IsVerified,
	}, nil
}
