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
    // 1. 拦截非法零值（手机号未传或为空字符串时拦截）
    if len(req.PhoneNumber) == 0 {
        return nil, errorx.NewCodeError(errorx.ErrCodeParamInvalid, "手机号码不能为空")
    }
    
    // 2. 动态拼接缓存 Key（使用传入的手机号字符串作为标识）
    cacheKey := fmt.Sprintf("user:phone:%s", req.PhoneNumber)
    
    // 3. 【防止缓存击穿】使用 SingleFlight 机制，确保同手机号的并发请求只会查一次数据库
    val, err := l.svcCtx.SingleGroup.Do(cacheKey, func() (interface{}, error) {
        var user model.User
        
        // 3.1 尝试从 Redis 获取缓存
        cacheVal, _ := l.svcCtx.Redis.Get(cacheKey)
        if cacheVal != "" {
            // 【防止缓存穿透】如果是空缓存标识，直接拦截
            if cacheVal == "empty" {
                return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在(缓存穿透保护)")
            }
            if err := json.Unmarshal([]byte(cacheVal), &user); err == nil {
                return &user, nil
            }
        }
        
        // 3.2 缓存未命中，调用 Repository 层根据手机号查询
        // 💡 提示：你需要确保你的 UserRepo 里有 FindOneByPhoneNumber 这个方法
        //  改成你接口里现有的方法名：
         u, err := l.svcCtx.UserRepo.FindOneByPhone(l.ctx, req.PhoneNumber)
        if err != nil {
            if err == sql.ErrNoRows || err.Error() == "sql: no rows in result set" {
                // 【防止缓存穿透】数据库没捞到，写入 60 秒的空值标识
                _ = l.svcCtx.Redis.Setex(cacheKey, "empty", 60)
                return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "用户不存在")
            }
            return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "数据库查询失败")
        }

        // 3.3 数据库查询成功，写入正常缓存 (10分钟)
        if data, err := json.Marshal(u); err == nil {
            _ = l.svcCtx.Redis.Setex(cacheKey, string(data), 600)
        }
        return u, nil
    })

    if err != nil {
        return nil, err
    }

    // 4. 安全获取 user 对象类型断言
    var user *model.User
    switch v := val.(type) {
    case *model.User:
        user = v
    case model.User:
        user = &v
    default:
        logx.Errorf("[QueryUser] 关键错误：数据存在，但类型不匹配! 实际类型为: %T", val)
        return nil, errorx.NewCodeError(errorx.ErrCodeServerInternal, "服务内部数据解析失败")
    }

    if user == nil {
        return nil, errorx.NewCodeError(errorx.ErrCodeUserNotFound, "未找到有效的用户信息")
    }

    // 5. 生成 JWT Token
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
