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
