package repository

import (
	"accesscontrol/internal/domain"
	"accesscontrol/internal/model"
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type gormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Insert(ctx context.Context, user *domain.User) error {
	modelUser := domainToModel(user)
	return r.db.WithContext(ctx).Create(modelUser).Error
}

func (r *gormUserRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE users SET password = ?, status = ?, validTime = ? WHERE id = ?",
		user.Password, user.Status, user.ValidTime.Format("2006-01-02 15:04:05"), user.ID,
	).Error
}

func (r *gormUserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *gormUserRepository) FindOne(ctx context.Context, id int64) (*domain.User, error) {
	var modelUser model.User
	err := r.db.WithContext(ctx).First(&modelUser, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return modelToDomain(&modelUser), nil
}

func (r *gormUserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var modelUser model.User
	err := r.db.WithContext(ctx).Where(`"phoneNumber" = ?`, phone).First(&modelUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return modelToDomain(&modelUser), nil
}

func (r *gormUserRepository) FindByStatus(ctx context.Context, status domain.UserStatus) (*domain.User, error) {
	var modelUser model.User
	err := r.db.WithContext(ctx).Where(`"status" = ?`, status).First(&modelUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return modelToDomain(&modelUser), nil
}

func (r *gormUserRepository) List(ctx context.Context, criteria domain.UserCriteria) ([]domain.User, int, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.User{})

	if criteria.PhoneNumber != nil {
		query = query.Where(`"phoneNumber" = ?`, *criteria.PhoneNumber)
	}
	if criteria.Status != nil {
		query = query.Where(`"status" = ?`, *criteria.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := criteria.PageSize
	offset := (criteria.Page - 1) * criteria.PageSize
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var modelUsers []model.User
	err := query.Order("id").Limit(limit).Offset(offset).Find(&modelUsers).Error
	if err != nil {
		return nil, 0, err
	}

	var users []domain.User
	for _, m := range modelUsers {
		users = append(users, *modelToDomain(&m))
	}
	return users, int(total), nil
}

func domainToModel(d *domain.User) *model.User {
	return &model.User{
		Id:          d.ID,
		PhoneNumber: d.PhoneNumber,
		Password:    d.Password,
		Status:      int(d.Status),
		ValidTime:   d.ValidTime.Format("2006-01-02 15:04:05"),
	}
}

func modelToDomain(m *model.User) *domain.User {
	validTime, _ := time.Parse("2006-01-02 15:04:05", m.ValidTime)
	return &domain.User{
		ID:          m.Id,
		PhoneNumber: m.PhoneNumber,
		Password:    m.Password,
		Status:      domain.UserStatus(m.Status),
		ValidTime:   validTime,
	}
}
