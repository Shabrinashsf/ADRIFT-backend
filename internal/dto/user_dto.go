package dto

import (
	"errors"
)

const (
	// Failed
	MESSAGE_FAILED_REGISTER_USER   = "gagal melakukan registrasi user"
	MESSAGE_FAILED_LOGIN_USER      = "gagal melakukan login user"
	MESSAGE_FAILED_VERIFY_EMAIL    = "gagal memverifikasi email"
	MESSAGE_FAILED_FORGET_PASSWORD = "gagal memproses permintaan lupa password"
	MESSAGE_FAILED_RESET_PASSWORD  = "gagal mereset password"
	MESSAGE_FAILED_GET_USER        = "gagal mendapatkan data user"
	MESSAGE_FAILED_UPDATE_USER     = "gagal memperbarui data user"
	MESSAGE_FAILED_UPLOAD_PROFILE  = "gagal mengupload foto profil"

	// Success
	MESSAGE_SUCCESS_REGISTER_USER           = "berhasil melakukan registrasi user"
	MESSAGE_SUCCESS_LOGIN_USER              = "berhasil melakukan login user"
	MESSAGE_SEND_VERIFICATION_EMAIL_SUCCESS = "jika email terdaftar dan belum terverifikasi, email verifikasi akan dikirim"
	MESSAGE_SUCCESS_VERIFY_EMAIL            = "berhasil memverifikasi email"
	MESSAGE_SUCCESS_FORGET_PASSWORD         = "jika email terdaftar, instruksi reset password akan dikirim"
	MESSAGE_SUCCESS_RESET_PASSWORD          = "berhasil mereset password"
	MESSAGE_SUCCESS_GET_USER                = "berhasil mendapatkan data user"
	MESSAGE_SUCCESS_UPDATE_USER             = "berhasil memperbarui data user"
)

var (
	ErrorEmailAlreadyExists   = errors.New("email sudah terdaftar")
	ErrMakeMail               = errors.New("gagal membuat email")
	ErrSendMail               = errors.New("gagal mengirim email")
	ErrTokenInvalid           = errors.New("token tidak valid atau kadaluarsa")
	ErrTokenExpired           = errors.New("token telah kadaluarsa")
	ErrUserNotFound           = errors.New("user tidak ditemukan")
	ErrAccountAlreadyVerified = errors.New("akun sudah terverifikasi")
	ErrUpdateUser             = errors.New("gagal memperbarui data user")
	ErrEmailNotFound          = errors.New("email tidak ditemukan")
	ErrHashPasswordFailed     = errors.New("gagal melakukan hash password")
	ErrNoChanges              = errors.New("tidak ada perubahan pada data user")
	ErrInvalidCredentials     = errors.New("kredensial tidak valid")
)

type (
	UserRegistrationRequest struct {
		Name           string `json:"name" form:"name" binding:"required"`
		NRP            string `json:"nrp" form:"nrp" binding:"required"`
		Email          string `json:"email" form:"email" binding:"required,email"`
		Password       string `json:"password" form:"password" binding:"required"`
		EnrollmentYear int    `json:"enrollment_year" form:"enrollment_year" binding:"required"`
	}

	UserResponse struct {
		ID             string `json:"id"`
		NRP            string `json:"nrp"`
		Name           string `json:"name"`
		Email          string `json:"email"`
		EnrollmentYear int    `json:"enrollment_year"`
		Role           string `json:"role"`
		IsVerified     bool   `json:"is_verified"`
	}

	UserLoginRequest struct {
		Email    string `json:"email" form:"email" binding:"required,email"`
		Password string `json:"password" form:"password" binding:"required"`
	}

	UserLoginResponse struct {
		Token string `json:"token"`
		Role  string `json:"role"`
	}

	SendVerificationEmailRequest struct {
		Email string `json:"email" form:"email" binding:"required,email"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" form:"token" binding:"required"`
	}

	VerifyEmailResponse struct {
		Email      string `json:"email"`
		IsVerified bool   `json:"is_verified"`
	}

	ForgotPasswordRequest struct {
		Email string `json:"email" form:"email" binding:"required,email"`
	}

	ResetPasswordRequest struct {
		Password string `json:"password" form:"password" binding:"required"`
	}

	ResetPasswordResponse struct {
		Email string `json:"email"`
	}

	UserUpdateRequest struct {
		Name           *string `json:"name" form:"name" binding:"omitempty"`
		NRP            *string `json:"nrp" form:"nrp" binding:"omitempty"`
		EnrollmentYear *int    `json:"enrollment_year" form:"enrollment_year" binding:"omitempty"`
	}
)
