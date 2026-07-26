package handler

import (
	"accesscontrol/internal/logic"
	"accesscontrol/internal/svc"
	"accesscontrol/internal/types"
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUserLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
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

func EditUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 400, Message: "请求参数解析失败"})
			return
		}

		if req.ID <= 0 {
			httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 400, Message: "用户ID不能为空"})
			return
		}

		if req.Password != "" {
			logx.Infof("Processing password update for user %d", req.ID)
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				logx.Errorf("bcrypt error: %v", err)
				httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 500, Message: "密码处理失败"})
				return
			}
			logx.Infof("Executing SQL: UPDATE users SET password = ? WHERE id = ?")
			err = svcCtx.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), req.ID).Error
			if err != nil {
				logx.Errorf("SQL exec error: %v", err)
				httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 500, Message: "更新失败: " + err.Error()})
				return
			}
			logx.Infof("Password update success for user %d", req.ID)
		}

		if req.PhoneNumber != "" {
			err := svcCtx.DB.Exec(`UPDATE users SET "phoneNumber" = ? WHERE id = ?`, req.PhoneNumber, req.ID).Error
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 500, Message: "更新失败: " + err.Error()})
				return
			}
		}

		if req.Status != 0 {
			err := svcCtx.DB.Exec("UPDATE users SET status = ? WHERE id = ?", req.Status, req.ID).Error
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 500, Message: "更新失败: " + err.Error()})
				return
			}
		}

		if req.ValidTime != "" {
			err := svcCtx.DB.Exec(`UPDATE users SET "validTime" = ? WHERE id = ?`, req.ValidTime, req.ID).Error
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 500, Message: "更新失败: " + err.Error()})
				return
			}
		}

		httpx.OkJsonCtx(r.Context(), w, types.Response{Code: 200, Message: "更新成功"})
	}
}

func DeleteUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUserLogic(r.Context(), svcCtx)
		resp, err := l.DeleteUser(req.ID, req.PhoneNumber)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func ListUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListUsersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
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
