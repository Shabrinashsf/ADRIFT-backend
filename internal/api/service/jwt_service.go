package service

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	// Track latest reset token issue time per email
	resetTokenTimes = make(map[string]int64)
	resetTokenMutex sync.RWMutex
)

const minJWTSecretLength = 32

type JWTService interface {
	GenerateToken(userId string, role string) string
	GenerateResetPasswordToken(email string) string
	ValidateToken(token string) (*jwt.Token, error)
	GetUserIDByToken(token string) (string, error)
	GetEmailByToken(token string) (string, error)
	ValidateResetToken(token string) (string, error)
}

type jwtCustomClaim struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey string
	issuer    string
}

func NewJWTService() JWTService {
	secretKey := mustGetSecretKey()

	return &jwtService{
		secretKey: secretKey,
		issuer:    "ADRIFT",
	}
}

func mustGetSecretKey() string {
	secretKey := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if secretKey == "" {
		log.Panic("JWT_SECRET must be set and non-empty")
	}

	if strings.EqualFold(secretKey, "your-jwt-secret-key-here") || secretKey == "Template" {
		log.Panic("JWT_SECRET must not use default/placeholder values")
	}

	if len(secretKey) < minJWTSecretLength {
		log.Panicf("JWT_SECRET must be at least %d characters", minJWTSecretLength)
	}

	return secretKey
}

func (j *jwtService) GenerateToken(userId string, role string) string {
	claims := jwtCustomClaim{
		userId,
		role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 120)),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tx, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		log.Println(err)
	}
	return tx
}

func (j *jwtService) GenerateResetPasswordToken(email string) string {
	// Store the current timestamp for this email - invalidates all previous tokens
	resetTokenMutex.Lock()
	issuedAt := time.Now().Unix()
	resetTokenTimes[email] = issuedAt
	resetTokenMutex.Unlock()

	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 1).Unix(), // Token expires in 1 hour
		"iat":   issuedAt,
		"iss":   j.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tx, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		log.Println(err)
	}
	return tx
}

func (j *jwtService) parseToken(t_ *jwt.Token) (any, error) {
	if _, ok := t_.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method %v", t_.Header["alg"])
	}
	return []byte(j.secretKey), nil
}

func (j *jwtService) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, j.parseToken)
}

func (j *jwtService) GetUserIDByToken(token string) (string, error) {
	t_Token, err := j.ValidateToken(token)
	if err != nil {
		return "", err
	}

	claims := t_Token.Claims.(jwt.MapClaims)
	id := fmt.Sprintf("%v", claims["user_id"])
	return id, nil
}

func (j *jwtService) GetEmailByToken(token string) (string, error) {
	t_Token, err := j.ValidateToken(token)
	if err != nil {
		return "", err
	}

	claims := t_Token.Claims.(jwt.MapClaims)
	email := fmt.Sprintf("%v", claims["email"])
	return email, nil
}

func (j *jwtService) ValidateResetToken(token string) (string, error) {
	t_Token, err := j.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid token")
	}

	claims := t_Token.Claims.(jwt.MapClaims)
	email := fmt.Sprintf("%v", claims["email"])

	iat, ok := claims["iat"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid token: missing iat")
	}

	resetTokenMutex.RLock()
	latestIat, exists := resetTokenTimes[email]
	resetTokenMutex.RUnlock()

	if exists && int64(iat) < latestIat {
		return "", fmt.Errorf("token has been superseded by a newer reset request")
	}

	return email, nil
}
