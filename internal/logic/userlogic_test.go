package logic

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/svc"
	"accesscontrol/internal/types"
	"context"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

// BenchmarkQueryUser 压力测试 QueryUser 的性能
// 模拟高并发下 SingleFlight 对数据库的保护作用
func BenchmarkQueryUser(b *testing.B) {
	
	// 这里需要 Mock ServiceContext 或者使用真实的测试环境
	// 为了演示目的，我们假设已经有一个配置好的测试环境
	var c config.Config
	conf.MustLoad("../../etc/config.yaml", &c)
	svcCtx := svc.NewServiceContext(c)
	l := NewUserLogic(context.Background(), svcCtx)

	req := &types.LoginRequest{
		PhoneNumber: "13800138000",
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = l.QueryUser(req)
		}
	})
}
