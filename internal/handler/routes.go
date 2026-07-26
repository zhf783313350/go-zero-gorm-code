package handler

import (
	"accesscontrol/internal/middleware"
	"accesscontrol/internal/svc"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

func SetupRoutes(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 中间件初始化
	rateLimitMatch := middleware.NewRateLimitMiddleware(serverCtx.RateLimiter).Handle
	casbinMatch := middleware.NewCasbinMiddleware(serverCtx.Enforcer).Handle

	// 1. 公共路由（无需 JWT，仅限流）
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/user/login",
				Handler: rateLimitMatch(LoginHandler(serverCtx)),
			},
		},
		rest.WithPrefix("/api"),
	)

	// 2. 受保护路由（JWT 认证 + Casbin RBAC 鉴权 + 限流）
	//    三层防护：限流 → JWT 身份识别 → Casbin 接口权限校验
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/user/add",
				Handler: rateLimitMatch(casbinMatch(AddUserHandler(serverCtx))),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/update",
				Handler: rateLimitMatch(casbinMatch(EditUserHandler(serverCtx))),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/delete",
				Handler: rateLimitMatch(casbinMatch(DeleteUserHandler(serverCtx))),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/list",
				Handler: rateLimitMatch(casbinMatch(ListUsersHandler(serverCtx))),
			},
		},
		rest.WithPrefix("/api"),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)

	// 3. 健康检查（供 Kubernetes Liveness/Readiness 探针调用）
	server.AddRoutes(
		[]rest.Route{
			{
				Method: http.MethodGet,
				Path:   "/health",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("OK"))
				},
			},
		},
	)
}
