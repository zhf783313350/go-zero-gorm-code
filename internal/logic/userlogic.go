package logic

import (
	"accesscontrol/internal/domain"
	"accesscontrol/internal/errorx"
	"accesscontrol/internal/svc"
	"accesscontrol/internal/types"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLogic {
	return &UserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogic) Login(req *types.LoginRequest) (*types.Response, error) {
	if len(req.PhoneNumber) == 0 {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "手机号码不能为空")
	}

	cacheKey := fmt.Sprintf("user:phone:%s", req.PhoneNumber)

	val, err := l.svcCtx.SingleGroup.Do(cacheKey, func() (interface{}, error) {
		cacheVal, _ := l.svcCtx.Redis.Get(cacheKey)
		if cacheVal != "" {
			if cacheVal == "empty" {
				return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在(缓存穿透保护)")
			}
			var cachedUser domain.User
			if err := json.Unmarshal([]byte(cacheVal), &cachedUser); err == nil {
				return &cachedUser, nil
			}
		}

		u, err := l.svcCtx.UserRepo.FindByPhone(l.ctx, req.PhoneNumber)
		if err != nil {
			if err == sql.ErrNoRows || err.Error() == "sql: no rows in result set" {
				_ = l.svcCtx.Redis.Setex(cacheKey, "empty", 60)
				return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在")
			}
			return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "数据库查询失败")
		}

		if data, err := json.Marshal(u); err == nil {
			_ = l.svcCtx.Redis.Setex(cacheKey, string(data), 600)
		}
		return u, nil
	})

	if err != nil {
		return nil, err
	}

	var user *domain.User
	switch v := val.(type) {
	case *domain.User:
		user = v
	case domain.User:
		user = &v
	default:
		logx.Errorf("[Login] data type mismatch: %T", val)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "服务内部数据解析失败")
	}

	if user == nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "未找到有效的用户信息")
	}

	if user.IsBlocked() {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户已被封禁")
	}

	if user.IsExpired() {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户已过期")
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	token, err := l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire, user.ID)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "生成Token失败")
	}

	return &types.Response{
		Code:    200,
		Message: "登录成功",
		Data: types.LoginResponse{
			AccessToken:  token,
			AccessExpire: now + accessExpire,
			UserInfo: types.UserInfo{
				ID:          user.ID,
				PhoneNumber: user.PhoneNumber,
				Status:      int(user.Status),
				ValidTime:   user.ValidTime.Format("2006-01-02 15:04:05"),
			},
		},
	}, nil
}

func (l *UserLogic) getJwtToken(secretKey string, iat, seconds int64, userId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["userId"] = userId
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func (l *UserLogic) AddUser(req *types.RegisterRequest) (*types.Response, error) {
	if req.PhoneNumber == "" || req.ValidTime == "" {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "手机号或有效时间不能为空")
	}

	validTime, err := time.Parse("2006-01-02 15:04:05", req.ValidTime)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "有效时间格式错误，应为: 2006-01-02 15:04:05")
	}

	user := &domain.User{
		PhoneNumber: req.PhoneNumber,
		Status:      domain.UserStatus(req.Status),
		ValidTime:   validTime,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, err.Error())
	}

	existing, _ := l.svcCtx.UserRepo.FindByPhone(l.ctx, req.PhoneNumber)
	if existing != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserAlreadyExist, "用户已存在")
	}

	err = l.svcCtx.UserRepo.Insert(l.ctx, user)
	if err != nil {
		logx.Errorf("添加用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "添加用户失败")
	}

	return &types.Response{
		Code:    http.StatusOK,
		Message: "用户创建成功",
	}, nil
}

func (l *UserLogic) EditUser(req *types.UpdateUserRequest) (*types.Response, error) {
	if req.PhoneNumber == "" || req.ValidTime == "" {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "手机号或有效时间不能为空")
	}

	validTime, err := time.Parse("2006-01-02 15:04:05", req.ValidTime)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "有效时间格式错误，应为: 2006-01-02 15:04:05")
	}

	user, err := l.svcCtx.UserRepo.FindByPhone(l.ctx, req.PhoneNumber)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在")
	}

	user.Status = domain.UserStatus(req.Status)
	user.ValidTime = validTime
	user.UpdatedAt = time.Now()

	err = l.svcCtx.UserRepo.Update(l.ctx, user)
	if err != nil {
		logx.Errorf("更新用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "更新用户失败")
	}

	cacheKey := "user:phone:" + user.PhoneNumber
	_, _ = l.svcCtx.Redis.Del(cacheKey)

	return &types.Response{
		Code:    http.StatusOK,
		Message: "用户信息更新成功",
	}, nil
}

func (l *UserLogic) DeleteUser(phoneNumber string) (*types.Response, error) {
	if phoneNumber == "" {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "手机号不能为空")
	}

	user, err := l.svcCtx.UserRepo.FindByPhone(l.ctx, phoneNumber)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在")
	}

	err = l.svcCtx.UserRepo.Delete(l.ctx, user.ID)
	if err != nil {
		logx.Errorf("删除用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "删除用户失败")
	}

	cacheKey := "user:phone:" + phoneNumber
	_, _ = l.svcCtx.Redis.Del(cacheKey)

	return &types.Response{
		Code:    http.StatusOK,
		Message: "用户删除成功",
	}, nil
}

func (l *UserLogic) ListUsers(page, pageSize int) (*types.Response, error) {
	criteria := domain.UserCriteria{
		Page:     page,
		PageSize: pageSize,
	}
	users, total, err := l.svcCtx.UserRepo.List(l.ctx, criteria)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "查询用户列表失败")
	}

	var userInfos []types.UserInfo
	for _, u := range users {
		userInfos = append(userInfos, types.UserInfo{
			ID:          u.ID,
			PhoneNumber: u.PhoneNumber,
			Status:      int(u.Status),
			ValidTime:   u.ValidTime.Format("2006-01-02 15:04:05"),
		})
	}


	return &types.Response{
		Code:    http.StatusOK,
		Message: "用户列表查询成功",
		Data: types.ListUsersResponse{
			Total: total,
			List:  userInfos,
		},
	}, nil
}