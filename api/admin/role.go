package admin

import (
	"mall/api"
	"mall/common"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) AddRole(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.AddRoleReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr)
		return
	}
	id, errno := c.role.CreateRole(ctx, user, req)
	api.WriteResp(ctx, map[string]interface{}{
		"id": id,
	}, errno)
}

func (c *Ctrl) UpdateRole(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.UpdateRoleReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr)
		return
	}
	errno := c.role.UpdateRole(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) ListRole(ctx *gin.Context) {
	req := &dto.ListRoleReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr)
		return
	}
	resp, errno := c.role.ListRoles(ctx, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) MyRoles(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	roles, errno := c.role.GetMyRoles(ctx, user)
	api.WriteResp(ctx, roles, errno)
}

func (c *Ctrl) SetRolePerms(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.SetRolePermReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr)
		return
	}
	errno := c.role.SetRolePerms(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}
