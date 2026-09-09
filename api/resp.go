package api

import (
	"mall/common"
	"mall/consts"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	ErrMsg string `json:"err_msg"`
	IsAes  bool   `json:"is_aes"`
	Data   any    `json:"data"`
}

func WriteResp(ctx *gin.Context, data any, errno common.Errno) {
	ctx.JSON(http.StatusOK, Resp{
		Code:   errno.Code,
		Msg:    errno.Msg,
		ErrMsg: errno.ErrMsg,
		Data:   data,
	})
}

func WriteRespAes(ctx *gin.Context, data any, errno common.Errno) {
	ctx.JSON(http.StatusOK, Resp{
		Code:   errno.Code,
		Msg:    errno.Msg,
		ErrMsg: errno.ErrMsg,
		IsAes:  true,
		Data:   data,
	})
}

func GetUserFromCtx(ctx *gin.Context) *common.UserInfo {
	user, exist := ctx.Get(consts.CustomerUserKey)
	if !exist {
		return nil
	}
	return user.(*common.UserInfo)
}

func GetAdminUserFromCtx(ctx *gin.Context) *common.AdminUser {
	user, exist := ctx.Get(consts.AdminUserKey)
	if !exist {
		return nil
	}
	return user.(*common.AdminUser)
}
