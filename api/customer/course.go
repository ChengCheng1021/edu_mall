package customer

import (
	"mall/api"
	"mall/common"
	"mall/consts"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) GetCourseList(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.CourseListReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	req.UserID = user.User.ID
	req.Status = consts.IsEnable
	resp, errno := c.course.ListCourse(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) GetCourseDetail(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.CourseInfoReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	req.UserID = user.User.ID
	resp, errno := c.course.GetCourseDetail(ctx.Request.Context(), req)

	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) GetCourseLessonInfo(ctx *gin.Context) {
	_ = api.GetUserFromCtx(ctx)
	req := &dto.LessonInfoReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, isAes, errno := c.course.LessonDetail(ctx.Request.Context(), req)
	if isAes {
		api.WriteRespAes(ctx, resp, errno)
		return
	}
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) GetPurchasedCourseList(ctx *gin.Context) {
	user := api.GetUserFromCtx(ctx)
	req := &dto.GetPurchasedCourseReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.course.GetPurchasedCourseList(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, resp, errno)
}
