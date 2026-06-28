package main

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/errorx"
	"accesscontrol/internal/handler"
	"accesscontrol/internal/svc"
	"context"
	"flag"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)
var configFile = flag.String("f", "etc/config.yaml", "the config file")
func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop() 
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		switch e := err.(type) {
		case *errorx.CodeError:
			return http.StatusOK, e.Data()
		default:
			return http.StatusInternalServerError, nil
		}
	})
	ctx := svc.NewServiceContext(c)
	handler.SetupRoutes(server, ctx) 
	
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, interface{}) {
        // 尝试将普通 error 断言为你自定义的 CodeError 结构体/接口
        switch e := err.(type) {
        case *errorx.CodeError: // 注意这里的类型要跟你的实际名字对上
            // 返回给前端 HTTP 200 状态码，并将错误格式化为 JSON
            return http.StatusOK, map[string]interface{}{
                "code": e.Data().Code, // 请根据你 errorx 具体的结构体字段（如 Code / Msg / Data 等）进行修改
                "msg":  "errorx.CodeError", 
            }
        default:
            // 真正的系统级未知错误，兜底返回 500
            return http.StatusInternalServerError, map[string]interface{}{
                "code": 500,
                "msg":  "服务器开小差了，请稍后再试",
            }
        }
    })
	server.Start()
}
