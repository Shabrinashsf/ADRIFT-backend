package repository

import (
	"context"
	"errors"

	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	UserRepository interface {
		RegisterUser(ctx context.Context, tx *gorm.DB, user entity.User) (entity.User, error)
		UpdateUser(ctx context.Context, tx *gorm.DB, id uuid.UUID, updates map[string]interface{}) (entity.User, error)
		GetUserByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entity.User, error)
		GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entity.User, bool, error)
		ResetPassword(ctx context.Context, email, hashedPassword string) error
	}

	userRepository struct {
		db *gorm.DB
	}
)

func NewUserController(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) RegisterUser(ctx context.Context, tx *gorm.DB, user entity.User) (entity.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&user).Error; err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, tx *gorm.DB, id uuid.UUID, updates map[string]interface{}) (entity.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Updates(&updates).Error; err != nil {
		return entity.User{}, err
	}

	var user entity.User
	if err := tx.WithContext(ctx).Preload("Prodi").Where("id = ?", id).First(&user).Error; err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entity.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entity.User
	if err := tx.WithContext(ctx).Preload("Prodi").Where("id = ?", id).First(&user).Error; err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entity.User, bool, error) {
	if tx == nil {
		tx = r.db
	}

	var user entity.User
	if err := tx.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return entity.User{}, false, err
	}

	return user, true, nil
}

func (r *userRepository) ResetPassword(ctx context.Context, email, hashedPassword string) error {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ErrEmailNotFound
		}
		return err
	}

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Update("password", hashedPassword).Error; err != nil {
		return err
	}
	return nil
}
