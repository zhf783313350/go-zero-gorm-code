package types

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string   `json:"accessToken"`
	AccessExpire int64    `json:"accessExpire"`
	UserInfo     UserInfo `json:"userInfo"`
}

type UserInfo struct {
	ID          int64  `json:"id"`
	PhoneNumber string `json:"phoneNumber"`
	Status      int    `json:"status"`
	ValidTime   string `json:"validTime"`
}

type RegisterRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
	ValidTime   string `json:"validTime"`
	Status      int    `json:"status"`
}

type UpdateUserRequest struct {
	ID          int64  `json:"id"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
	Status      int    `json:"status"`
	ValidTime   string `json:"validTime"`
}

type DeleteUserRequest struct {
	ID          int64  `json:"id"`
	PhoneNumber string `json:"phoneNumber,optional"`
}

type ListUsersRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type ListUsersResponse struct {
	Total int        `json:"total"`
	List  []UserInfo `json:"list"`
}
