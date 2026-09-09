package admin

import (
	"mall/api"
	"mall/common"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) GetTempSecret(ctx *gin.Context) {
	req := &dto.GetTempSecretReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.storage.GetTempSecret(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}
