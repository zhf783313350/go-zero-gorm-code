package repository

import (
	"accesscontrol/internal/model"
	"context"

	"gorm.io/gorm"
)

type UserRepository interface {
	Insert(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, phoneNumber string) error
	FindOne(ctx context.Context, id int64) (*model.User, error)
	FindOneByPhone(ctx context.Context, phone string) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, int, error)
	FindOneByStatus(ctx context.Context, status int) (*model.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}
func (r *gormUserRepository) FindOneByStatus(ctx context.Context, status int) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where(`"status" = ?`, status).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Insert(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormUserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *gormUserRepository) Delete(ctx context.Context, phoneNumber string) error {
	return r.db.WithContext(ctx).Where(`"phoneNumber" = ?`, phoneNumber).Delete(&model.User{}).Error
}

func (r *gormUserRepository) FindOne(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindOneByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where(`"phoneNumber" = ?`, phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) List(ctx context.Context, limit, offset int) ([]model.User, int, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := r.db.WithContext(ctx).Order("id").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, int(total), nil
}

