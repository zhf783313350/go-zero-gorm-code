package handler

import (
	"accesscontrol/internal/logic"
	"accesscontrol/internal/svc"
	"accesscontrol/internal/types"
	"fmt"
	"net/http"
 
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 🔥 注入点 1：确认请求到底进没进这个 Handler
        fmt.Println("====== [FMT 调试] 请求已成功到达 QueryUserHandler ======")
        var req types.LoginRequest
        if err := httpx.Parse(r, &req); err != nil {
            // 🔥 注入点 2：用原生打印输出错误原因，防止被 logx 过滤
            fmt.Printf("====== [FMT 调试] httpx.Parse 失败!! 原因: %v\n", err)
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        l := logic.NewUserLogic(r.Context(), svcCtx)
        resp, err := l.QueryUser(&req)
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            httpx.OkJsonCtx(r.Context(), w, resp)
        }
    }
}
// func QueryUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
//     return func(w http.ResponseWriter, r *http.Request) {
//         fmt.Println("====== [FMT 调试] 请求已成功到达 QueryUserHandler ======")

//         var req types.LoginRequest
//         if err := httpx.Parse(r, &req); err != nil {
//             fmt.Printf("====== [FMT 调试] httpx.Parse 失败!! 原因: %v\n", err)
//             httpx.ErrorCtx(r.Context(), w, err)
//             return
//         }

//         // 🔥 严格按照要求：这里直接拦截，写死固定返回 "查询数据请求"
//         // 不再调用底层的 logic.NewUserLogic 和数据库
//         resp := &types.Response{
//             Code:    200,
//             Message: "查询数据请求", // 👈 按照你的要求，这里暂时固定返回该字符串
//             Data:    nil,        // 联调阶段暂时不需要具体 data 字段可以给 nil 
//         }

//         // 直接通过 go-zero 官方工具包响应给客户端
//         httpx.OkJsonCtx(r.Context(), w, resp)
//     }
// }
// AddUserHandler 添加用户
func AddUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUserLogic(r.Context(), svcCtx)
		resp, err := l.AddUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// EditUserHandler 编辑用户
func EditUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewUserLogic(r.Context(), svcCtx)
		resp, err := l.EditUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// DeleteUserHandler 删除用户
func DeleteUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUserLogic(r.Context(), svcCtx)
		resp, err := l.DeleteUser(req.PhoneNumber)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// ListUsersHandler 用户列表
func ListUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListUsersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// 默认值处理
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 10
		}
		l := logic.NewUserLogic(r.Context(), svcCtx)
		resp, err := l.ListUsers(req.Page, req.PageSize)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
