package handler

import (
	"accesscontrol/internal/middleware"
	"accesscontrol/internal/svc"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

func SetupRoutes(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 初始化中间件
	rateLimitMatch := middleware.NewRateLimitMiddleware(serverCtx.RateLimiter).Handle

	// 1. 无需 JWT 认证的公共路由 (登录/查询)
   server.AddRoutes(
    []rest.Route{
        {
            Method:  http.MethodPost,
            Path:    "/user/query",
            Handler: rateLimitMatch(QueryUserHandler(serverCtx)), // 👈 先把这一行注释掉
         
        },
    },
    rest.WithPrefix("/api"),
)

	// 2. 需要 JWT 认证的保护路由
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/user/add",
				Handler: rateLimitMatch(AddUserHandler(serverCtx)),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/edit",
				Handler: rateLimitMatch(EditUserHandler(serverCtx)),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/delete",
				Handler: rateLimitMatch(DeleteUserHandler(serverCtx)),
			},
			{
				Method:  http.MethodPost,
				Path:    "/user/list",
				Handler: rateLimitMatch(ListUsersHandler(serverCtx)),
			},
		},
		rest.WithPrefix("/api"),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)

	// 3. 基础健康检查 (供 Kubernetes Liveness/Readiness 探针调用)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/health",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("OK"))
				},
			},
		},
	)
}
