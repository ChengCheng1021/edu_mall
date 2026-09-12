package admin

import (
	"mall/api"
	"mall/common"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) OrderList(ctx *gin.Context) {
	_ = api.GetAdminUserFromCtx(ctx)
	req := &dto.GetOrderListReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.order.GetOrderList(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) OrderInfo(ctx *gin.Context) {
	_ = api.GetAdminUserFromCtx(ctx)
	req := &dto.GetOrderInfoReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.order.GetOrderInfo(ctx.Request.Context(), nil, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) OrderStatistic(ctx *gin.Context) {
	_ = api.GetAdminUserFromCtx(ctx)
	req := &dto.OrderStatReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.order.OrderStatistic(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) OrderRefund(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.OrderRefundReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.order.OrderRefund(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) OrderCancel(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.CancelOrderReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.order.AdminCancelOrder(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}
