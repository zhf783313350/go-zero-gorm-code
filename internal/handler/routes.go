package handler

import (
	"accesscontrol/internal/middleware"
	"accesscontrol/internal/svc"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

func SetupRoutes(server *rest.Server, serverCtx *svc.ServiceContext) {
	rateLimitMatch := middleware.NewRateLimitMiddleware(serverCtx.RateLimiter).Handle

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