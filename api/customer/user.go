package customer

import (
	"context"
	"mall/api"
	"mall/common"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) GetUserByToken(ctx context.Context, token string) (*common.UserInfo, error) {
	return c.user.GetUserByToken(ctx, token)
}

func (c *Ctrl) GetUserInfo(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	if user == nil {
		api.WriteResp(ctx, nil, common.AuthErr)
		return
	}
	resp, errno := c.user.GetUserInfo(ctx.Request.Context(), user.User.ID)
	api.WriteResp(ctx, resp, errno)
}
