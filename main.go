package main

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/errorx"
	"accesscontrol/internal/handler"
	"accesscontrol/internal/svc"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
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

	// 1. 启动轻量级后台事件总线消费者
	stopEventBus := make(chan struct{})
	ctx.EventBus.Start(stopEventBus)

	// 2. 启动 Web 服务
	go func() {
		server.Start()
	}()

	// 3. 监听系统终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 4. 优雅停机：先关闭异步事件总线以排空队列，再释放其它系统资源
	close(stopEventBus)
}