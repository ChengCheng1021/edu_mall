package admin

import (
	"mall/api"
	"mall/common"
	"mall/service/dto"

	"github.com/gin-gonic/gin"
)

func (c *Ctrl) CreateCourse(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.CreateCourseReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	courseID, err := c.course.CreateCourse(ctx, user, req)
	api.WriteResp(ctx, map[string]int64{
		"id": courseID,
	}, err)
}

func (c *Ctrl) GetCourseInfo(ctx *gin.Context) {
	req := &dto.CourseInfoReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	req.IsAdmin = true
	courseDto, err := c.course.GetCourseInfo(ctx, req)
	api.WriteResp(ctx, courseDto, err)
}

func (c *Ctrl) UpdateCourse(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.UpdateCourseReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	err := c.course.UpdateCourse(ctx, user, req)
	api.WriteResp(ctx, nil, err)
}

func (c *Ctrl) UpdateCourseStatus(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.UpdateCourseStatusReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.UpdateCourseStatus(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) ListCourse(ctx *gin.Context) {
	req := &dto.CourseListReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	req.IsAdmin = true
	resp, errno := c.course.ListCourse(ctx, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) AddCatalog(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.AddCatalogReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	catalogID, errno := c.course.AddCatalog(ctx, user, req)
	api.WriteResp(ctx, map[string]int64{"id": catalogID}, errno)
}

func (c *Ctrl) UpdateCatalog(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.UpdateCatalogReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.UpdateCatalog(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) DeleteCatalog(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.DeleteCatalogReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.DeleteCatalog(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) UpdateCatalogSort(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	sorts := make([]dto.UpdateCatalogSortDto, 0)
	if err := ctx.BindJSON(&sorts); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.UpdateCatalogSort(ctx, user, sorts)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) CatalogInfo(ctx *gin.Context) {
	req := &dto.CatalogInfoReq{}
	if err := ctx.BindQuery(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	resp, errno := c.course.CatalogInfo(ctx, req)
	api.WriteResp(ctx, resp, errno)
}

func (c *Ctrl) AddCatalogLesson(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.AddCatalogLessonReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.AddCatalogLesson(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) UpdateCatalogLesson(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.UpdateCatalogLessonReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.UpdateCatalogLesson(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}

func (c *Ctrl) RemoveCatalogLesson(ctx *gin.Context) {
	user := api.GetAdminUserFromCtx(ctx)
	req := &dto.RemoveCatalogLessonReq{}
	if err := ctx.BindJSON(req); err != nil {
		api.WriteResp(ctx, nil, common.ParamErr.WithErr(err))
		return
	}
	errno := c.course.RemoveCatalogLesson(ctx, user, req)
	api.WriteResp(ctx, nil, errno)
}
