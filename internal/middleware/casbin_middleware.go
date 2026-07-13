package middleware

import (
	"fmt"
	"net/http"

	casbin "github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// CasbinMiddleware 基于 Casbin 的 RBAC 接口权限鉴权中间件。
//
// 工作原理：
//  1. go-zero 的 JWT 中间件会将 Token Claims 解析后写入 r.Context()。
//  2. 本中间件从 Context 中读取 userId（JWT Payload 中约定的字段）。
//  3. 以 userId 为 subject，请求路径为 resource，HTTP Method 为 action，
//     调用 Casbin Enforcer 做 RBAC 鉴权。
//  4. 鉴权不通过则直接返回 403，不调用下游 Handler。
type CasbinMiddleware struct {
	enforcer *casbin.Enforcer
}

func NewCasbinMiddleware(enforcer *casbin.Enforcer) *CasbinMiddleware {
	return &CasbinMiddleware{enforcer: enforcer}
}

// Handle 返回一个链式 http.HandlerFunc 包装器。
func (m *CasbinMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// go-zero JWT 中间件将 Claims 以 interface{} 类型存入 Context，
		// 数字类型（如 userId）在 JSON 解码后默认为 float64。
		userIdRaw := r.Context().Value("userId")
		if userIdRaw == nil {
			httpx.WriteJson(w, http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "Forbidden: 无法识别用户身份",
			})
			return
		}

		// 将 float64 转为字符串，作为 Casbin 的 subject
		userIdFloat, ok := userIdRaw.(float64)
		if !ok {
			httpx.WriteJson(w, http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "Forbidden: 用户身份格式错误",
			})
			return
		}
		subject := fmt.Sprintf("%d", int64(userIdFloat))

		// 请求的资源路径和 HTTP 方法
		resource := r.URL.Path
		action := r.Method

		// 调用 Casbin Enforcer 进行 RBAC 权限判定
		allowed, err := m.enforcer.Enforce(subject, resource, action)
		if err != nil {
			httpx.WriteJson(w, http.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "Internal Error: 权限系统异常",
			})
			return
		}
		if !allowed {
			httpx.WriteJson(w, http.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "Forbidden: 当前用户无此接口访问权限",
			})
			return
		}

		next(w, r)
	}
}
