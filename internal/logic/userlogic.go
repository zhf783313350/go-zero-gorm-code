package logic
import (
	"accesscontrol/internal/errorx"
	"accesscontrol/internal/model"
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
func (l *UserLogic) QueryUser(req *types.LoginRequest) (*types.Response, error) {
    // 1. 拦截非法零值（int 类型未传或为 0 时拦截）
    if req.Status == 0 {
        return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "状态码不能为空")
    }
    
    // 2. 依然保持你原有的缓存 Key 拼接方式（只把变量换成请求过来的 int 状态值）
    cacheKey := fmt.Sprintf("user:phone:%d", req.Status)
    
    // 3. SingleFlight + Cache protection 应对大规模并发
    val, err := l.svcCtx.SingleGroup.Do(cacheKey, func() (interface{}, error) {
        // 3.1 尝试从 Redis 获取缓存
        var user model.User
        cacheVal, _ := l.svcCtx.Redis.Get(cacheKey)
        if cacheVal != "" {
            if cacheVal == "empty" {
                return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在(缓存穿透保护)")
            }
            if err := json.Unmarshal([]byte(cacheVal), &user); err == nil {
                return &user, nil
            }
        }
        
        // 3.2 缓存未命中，调用你刚刚写好的 FindOneByStatus 方法
        u, err := l.svcCtx.UserRepo.FindOneByStatus(l.ctx, req.Status)
        if err != nil {
            if err == sql.ErrNoRows || err.Error() == "sql: no rows in result set" {
                // 写入空缓存，防止缓存穿透
                _ = l.svcCtx.Redis.Setex(cacheKey, "empty", 60)
                return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在")
            }
            return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "数据库查询失败")
        }

        // 3.3 写入正常缓存 (10分钟)
        if data, err := json.Marshal(u); err == nil {
            _ = l.svcCtx.Redis.Setex(cacheKey, string(data), 600)
        }
        return u, nil
    })

    if err != nil {
        return nil, err
    }

    user, ok := val.(*model.User)
    if !ok {
        if uVal, ok := val.(model.User); ok {
            user = &uVal
        }
    }

    // 生成 JWT Token
    now := time.Now().Unix()
    accessExpire := l.svcCtx.Config.Auth.AccessExpire
    token, err := l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire, user.Id)
    if err != nil {
        return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "生成Token失败")
    }

    return &types.Response{
        Code:    200,
        Message: "登录成功",
        Data: types.LoginResponse{
            AccessToken:  token,
            AccessExpire: now + accessExpire,
            UserInfo:     *user,
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

	// 检查用户是否已存在
	_, err := l.svcCtx.UserRepo.FindOneByPhone(l.ctx, req.PhoneNumber)
	if err == nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserAlreadyExist, "用户已存在")
	}

	user := &model.User{
		PhoneNumber: req.PhoneNumber,
		Status:      req.Status,
		ValidTime:   req.ValidTime,
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

// 编辑用户
func (l *UserLogic) EditUser(req *types.UpdateUserRequest) (*types.Response, error) {
	if req.PhoneNumber == "" || req.ValidTime == "" {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "手机号或有效时间不能为空")
	}
	// 检查用户是否存在 (根据手机号)
	user, err := l.svcCtx.UserRepo.FindOneByPhone(l.ctx, req.PhoneNumber)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在")
	}
	// 更新用户信息
	user.Status = req.Status
	user.ValidTime = req.ValidTime
	err = l.svcCtx.UserRepo.Update(l.ctx, user)
	if err != nil {
		logx.Errorf("更新用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "更新用户失败")
	}

	// 3. 清除 Redis 缓存
	cacheKey := "user:phone:" + user.PhoneNumber
	_, _ = l.svcCtx.Redis.Del(cacheKey)

	return &types.Response{
		Code:    http.StatusOK,
		Message: "用户信息更新成功",
	}, nil
}

// 删除用户 根据手机号码
func (l *UserLogic) DeleteUser(phoneNumber string) (*types.Response, error) {
	if phoneNumber == "" {
		return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "手机号不能为空")
	}

	err := l.svcCtx.UserRepo.Delete(l.ctx, phoneNumber)
	if err != nil {
		logx.Errorf("删除用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "删除用户失败")
	}

	// 清除 Redis 缓存
	cacheKey := "user:phone:" + phoneNumber
	_, _ = l.svcCtx.Redis.Del(cacheKey)

	return &types.Response{
		Code:    http.StatusOK,
		Message: "用户删除成功",
	}, nil
}

// 用户列表 分页加载
func (l *UserLogic) ListUsers(page, pageSize int) (*types.Response, error) {
	users, total, err := l.svcCtx.UserRepo.List(l.ctx, pageSize, (page-1)*pageSize)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "查询用户列表失败")
	}

	return &types.Response{
		Code:    http.StatusOK,
		Message: "用户列表查询成功",
		Data: map[string]interface{}{
			"total": total,
			"list":  users,
		},
	}, nil
}
