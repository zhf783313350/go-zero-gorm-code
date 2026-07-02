package main

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/errorx"
	"accesscontrol/internal/handler"
	"accesscontrol/internal/svc"
	"flag"
	"fmt"
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
     httpx.SetErrorHandler(func(err error) (int, interface{}) {
        switch e := err.(type) {
        case *errorx.CodeError:
            return http.StatusOK, e.Data()
        default:

            // 🔥 注入点 3：看看全局捕获到了什么底层错误
            fmt.Printf("====== [FMT 调试] 全局异常处理器捕获到未知错误: %v\n", err)
            return http.StatusInternalServerError, nil
        }
    })

     server.Use(func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token")
            w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
            w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
            w.Header().Set("Access-Control-Allow-Credentials", "true")

            // 如果是浏览器的 OPTIONS 探路请求，直接返回 204
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next(w, r)
        }
    })

	server.Start()
}
