package repository

import (
	"accesscontrol/internal/model"
	"context"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)
type UserRepository interface {
	Insert(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, phoneNumber string) error
	FindOne(ctx context.Context, id int64) (*model.User, error)
	FindOneByPhone(ctx context.Context, phone string) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, int, error)
}

type sqlUserRepository struct {
	db sqlx.SqlConn
}
func NewUserRepository(db sqlx.SqlConn) UserRepository {
	return &sqlUserRepository{db: db}
}
func (r *sqlUserRepository) Insert(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users ("phoneNumber", status, "validTime") VALUES ($1, $2, $3)`
	_, err := r.db.ExecCtx(ctx, query, user.PhoneNumber, user.Status, user.ValidTime)
	return err
}

func (r *sqlUserRepository) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET "phoneNumber" = $1, status = $2, "validTime" = $3 WHERE id = $4`
	_, err := r.db.ExecCtx(ctx, query, user.PhoneNumber, user.Status, user.ValidTime, user.Id)
	return err
}

func (r *sqlUserRepository) Delete(ctx context.Context, phoneNumber string) error {
	query := `DELETE FROM users WHERE "phoneNumber" = $1`
	_, err := r.db.ExecCtx(ctx, query, phoneNumber)
	return err
}

func (r *sqlUserRepository) FindOne(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	query := `SELECT id, "phoneNumber" AS phonenumber, status, "validTime" AS validtime FROM users WHERE id = $1`
	err := r.db.QueryRowCtx(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *sqlUserRepository) FindOneByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	query := `SELECT id, "phoneNumber" AS phonenumber, status, "validTime" AS validtime FROM users WHERE "phoneNumber" = $1`
	err := r.db.QueryRowCtx(ctx, &user, query, phone)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *sqlUserRepository) List(ctx context.Context, limit, offset int) ([]model.User, int, error) {
	var total int
	queryCount := "SELECT COUNT(*) FROM users"
	err := r.db.QueryRowCtx(ctx, &total, queryCount)
	if err != nil {
		return nil, 0, err
	}
	var users []model.User
	query := `SELECT id, "phoneNumber" AS phonenumber, status, "validTime" AS validtime FROM users ORDER BY id LIMIT $1 OFFSET $2`
	err = r.db.QueryRowsCtx(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
