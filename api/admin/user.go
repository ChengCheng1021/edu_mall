package admin

import (
	"context"
	"github.com/gin-gonic/gin"
	"mall/api"
	"mall/common"
	"mall/service/dto"
)

func (c *Ctrl) GetAdminUserByToken(ctx context.Context, token string) (*common.AdminUser, error) {
	adminUser, errno := c.user.GetAdminUserByToken(ctx, token)
	if errno.NotOk() {
		return nil, errno
	}
	return adminUser, nil
}

func (c *Ctrl) AdminUserLogout(ctx *gin.Context) {
	adminUser := api.GetAdminUserFromCtx(ctx)
	errno := c.user.AdminUserLogout(ctx, adminUser)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) AdminUserList(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.ListAdminUserReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithMsg(err.Error()))
		return
	}

	resp, errno := c.user.AdminUserList(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) GetUserInfo(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	if user == nil {
		api.WriteResp(ctx, nil, common.AuthErr)
		return
	}
	resp, errno := c.user.GetUserInfo(ctx.Request.Context(), user)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) CreateUser(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	if user == nil {
		api.WriteResp(ctx, nil, common.AuthErr)
		return
	}
	req := &dto.CreateUserReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithMsg(err.Error()))
		return
	}
	userId, errno := c.user.CreateUser(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, map[string]int64{
		"id": userId,
	}, errno)
}

func (c *Ctrl) UpdateUser(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	if user == nil {
		api.WriteResp(ctx, nil, common.AuthErr)
		return
	}
	req := &dto.UpdateUserReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithMsg(err.Error()))
		return
	}
	errno := c.user.UpdateUser(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) DeleteUser(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	if user == nil {
		api.WriteResp(ctx, nil, common.AuthErr)
		return
	}
	req := &dto.DeleteUserReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithMsg(err.Error()))
		return
	}
	errno := c.user.DeleteUser(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}
func (c *Ctrl) LarkBind(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	if user == nil {
		api.WriteResp(ctx, nil, common.AuthErr)
		return
	}
	req := &dto.LarkQrCodeBindReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithMsg(err.Error()))
		return
	}
	errno := c.user.LarkBind(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) LarkUnbind(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	if user == nil {
		api.WriteResp(ctx, nil, common.AuthErr)
		return
	}
	errno := c.user.LarkUnbind(ctx.Request.Context(), user)
	api.WriteResp(ctx, nil, errno)
}
