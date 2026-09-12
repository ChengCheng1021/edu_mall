package customer

import (
	"mall/api"
	"mall/common"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) OrderCalcFee(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.OrderCalcFeeReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
	}
	resp, errno := c.order.OrderCalcFee(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) OrderPayNow(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.OrderPayNowReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
	}
	resp, errno := c.order.OrderPayNow(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) GetOrderInfo(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.GetOrderInfoReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.order.GetOrderInfo(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, resp, errno)
}
func (c *Ctrl) GetOrderList(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.GetOrderListReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.order.GetUserOrderList(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) CancelOrder(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.CancelOrderReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.order.CancelOrder(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}
