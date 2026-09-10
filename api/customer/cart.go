package customer

import (
	"mall/api"
	"mall/common"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) AddGoods(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.AddGoodsReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
	}
	id, errno := c.cart.AddGoods(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, map[string]int64{"id": id}, errno)
}

func (c *Ctrl) RemoveGoods(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.RemoveGoodsReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.cart.RemoveGoods(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) ListGoods(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.ListGoodsReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.cart.ListGoods(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, resp, errno)
}
