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

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/user/query",
				Handler: rateLimitMatch(QueryUserHandler(serverCtx)),
			},
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
}
