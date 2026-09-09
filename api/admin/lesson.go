package admin

import (
	"mall/api"
	"mall/common"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) CategoryCreate(ctx *gin.Context) {
	req := &dto.AddCategoryReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.course.CategoryCreate(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) CategoryUpdate(ctx *gin.Context) {
	req := &dto.UpdateCategoryReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.CategoryUpdate(ctx.Request.Context(), req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) CategoryDelete(ctx *gin.Context) {
	req := &dto.DeleteCategoryReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.CategoryDelete(ctx.Request.Context(), req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) CategorySort(ctx *gin.Context) {
	req := make([]dto.UpdateSort, 0)
	if err := ctx.BindJSON(&req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.CategorySort(ctx.Request.Context(), req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) CategoryList(ctx *gin.Context) {
	req := &dto.ListCategoryReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.course.CategoryList(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) CreateLesson(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.CreateLessonReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	id, errno := c.course.CreateLesson(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, map[string]interface{}{
		"id": id,
	}, errno)
}

func (c *Ctrl) LessonList(ctx *gin.Context) {
	req := &dto.ListLessonReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.course.LessonList(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) LessonInfo(ctx *gin.Context) {
	req := &dto.LessonInfoReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.course.LessonInfo(ctx.Request.Context(), req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) MoveLesson(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.MoveLessonReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.MoveLesson(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) UpdateLessonStatus(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.UpdateLessonStatusReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.UpdateLessonStatus(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) UpdateLesson(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.UpdateLessonReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.UpdateLesson(ctx.Request.Context(), user, req)
	api.WriteResp(ctx, nil, errno)
}
