package types

import "accesscontrol/internal/model"
// 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
// 登录请求
// type LoginRequest struct {
// 	PhoneNumber string `json:"phoneNumber"`
// }
// 查询请求
type LoginRequest struct {
    Status int `json:"status"` // 明确使用 int 类型的 status
}
// 登录响应
type LoginResponse struct {
	AccessToken  string     `json:"accessToken"`
	AccessExpire int64      `json:"accessExpire"`
	UserInfo     model.User `json:"userInfo"`
}
// 注册/创建用户请求
type RegisterRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	ValidTime   string `json:"validTime"`
	Status      int    `json:"status"`
}

// 更新用户请求
type UpdateUserRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Status      int    `json:"status"`
	ValidTime   string `json:"validTime"`
}
// 删除用户请求
type DeleteUserRequest struct {
	PhoneNumber string `json:"phoneNumber"`
}

// 用户列表请求
type ListUsersRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}
