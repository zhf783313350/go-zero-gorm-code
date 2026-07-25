package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

type User struct {
	ID          int64
	PhoneNumber string
	Password    string
	Status      UserStatus
	ValidTime   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserStatus int

const (
	UserStatusNormal  UserStatus = 1
	UserStatusBlocked UserStatus = 2
)

func (u *User) Validate() error {
	if strings.TrimSpace(u.PhoneNumber) == "" {
		return errors.New("phone number cannot be empty")
	}
	return nil
}

func (u *User) IsExpired() bool {
	return u.ValidTime.Before(time.Now())
}

func (u *User) IsBlocked() bool {
	return u.Status == UserStatusBlocked
}

type UserCriteria struct {
	PhoneNumber *string
	Status      *UserStatus
	Page        int
	PageSize    int
}

type Repository[T any] interface {
	Insert(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id int64) error
	FindOne(ctx context.Context, id int64) (*T, error)
}

type UserQuery interface {
	FindByPhone(ctx context.Context, phone string) (*User, error)
	FindByStatus(ctx context.Context, status UserStatus) (*User, error)
	List(ctx context.Context, criteria UserCriteria) ([]User, int, error)
}

type UserRepository interface {
	Repository[User]
	UserQuery
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, phoneNumber string, validTime time.Time) (*User, error) {
	existing, _ := s.repo.FindByPhone(ctx, phoneNumber)
	if existing != nil {
		return nil, errors.New("user already exists")
	}

	user := &User{
		PhoneNumber: phoneNumber,
		Status:      UserStatusNormal,
		ValidTime:   validTime,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Insert(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, phoneNumber string) (*User, error) {
	user, err := s.repo.FindByPhone(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}

	if user.IsBlocked() {
		return nil, errors.New("user is blocked")
	}

	if user.IsExpired() {
		return nil, errors.New("user has expired")
	}

	return user, nil
}