package router

import (
	"context"
	"mall/adaptor"
	"mall/api/admin"
	"mall/api/customer"
	"mall/common"
	"mall/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type IRouter interface {
	Register(engine *gin.Engine)
	SpanFilter(r *gin.Context) bool
	AccessRecordFilter(r *gin.Context) bool
}

type Router struct {
	FullPPROF bool
	rootPath  string
	conf      *config.Config
	checkFunc func() error
	admin     *admin.Ctrl
	customer  *customer.Ctrl
}

func NewRouter(conf *config.Config, adaptor adaptor.IAdaptor, checkFunc func() error) *Router {
	return &Router{
		FullPPROF: conf.Server.EnablePprof,
		rootPath:  "/api/mall",
		conf:      conf,
		checkFunc: checkFunc,
		admin:     admin.NewCtrl(adaptor),
		customer:  customer.NewCtrl(adaptor),
	}
}

func (r *Router) checkServer() func(*gin.Context) {
	return func(ctx *gin.Context) {
		err := r.checkFunc()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{})
	}
}

func (r *Router) Register(app *gin.Engine) {
	if r.conf.Server.EnablePprof {
		SetupPprof(app, "/debug/pprof")
	}
	app.Any("/ping", r.checkServer())

	root := app.Group(r.rootPath)
	r.route(root)
}

func (r *Router) SpanFilter(ctx *gin.Context) bool {
	path := strings.Replace(ctx.Request.URL.Path, r.rootPath, "", 1)
	_, ok := AdminAuthWhiteList[path]
	if ok {
		return false
	}
	return true
}

func (r *Router) AccessRecordFilter(ctx *gin.Context) bool {
	return true
}

func (r *Router) route(root *gin.RouterGroup) {
	r.customerRoute(root)
	r.adminRoute(root)
}

func (r *Router) adminRoute(root *gin.RouterGroup) {
	adminRoot := root.Group("/admin", AdminAuthMiddleware(r.SpanFilter, func(ctx context.Context, token string) (*common.AdminUser, error) {
		return r.admin.GetAdminUserByToken(ctx, token)
	}))
	// 登录无鉴权：添加白名单
	adminRoot.GET("/v1/user/verify/captcha", r.admin.GetSmsCodeCaptcha)
	adminRoot.POST("/v1/user/verify/captcha/check", r.admin.CheckSmsCodeCaptcha)
	adminRoot.POST("/v1/user/verify/smscode", r.admin.GetSmsVerifyCode)
	adminRoot.POST("/v1/user/mobile/password_login", r.admin.MobilePasswordLogin)
	adminRoot.POST("/v1/user/mobile/verify_login", r.admin.MobileVerifyLogin)
	adminRoot.POST("/v1/user/lark/qrcode_login", r.admin.LarkQrCodeLogin)
	adminRoot.POST("/v1/user/mobile/reset_password", r.admin.MobilePasswordReset)

	// C端用户管理
	adminRoot.GET("/v1/customer/user/list", r.admin.CustomerUserList)    // 获取C端用户列表
	adminRoot.GET("/v1/customer/user/info", r.admin.GetCustomerUserInfo) // 获取C端用户信息

	// 管理员用户
	// 登出系统
	adminRoot.POST("/v1/user/logout", r.admin.AdminUserLogout)
	adminRoot.GET("/v1/user/list", r.admin.AdminUserList)
	adminRoot.GET("/v1/user/info", r.admin.GetUserInfo)
	adminRoot.POST("/v1/user/create", r.admin.CreateUser)
	adminRoot.POST("/v1/user/update", r.admin.UpdateUser)
	adminRoot.POST("/v1/user/delete", r.admin.DeleteUser)
	adminRoot.POST("/v1/user/lark_bind", r.admin.LarkBind)
	adminRoot.POST("/v1/user/lark_unbind", r.admin.LarkUnbind)

	// 权限菜单
	adminRoot.POST("/v1/perm/create", r.admin.CreatePermission)
	adminRoot.POST("/v1/perm/update", r.admin.UpdatePermission)
	adminRoot.POST("/v1/perm/delete", r.admin.DeletePermission)
	adminRoot.GET("/v1/perm/list", r.admin.PermissionList)
	adminRoot.GET("/v1/perm/my_perms", r.admin.MyPermissionList)
	// 角色管理
	adminRoot.POST("/v1/role/create", r.admin.AddRole)         // 添加角色
	adminRoot.POST("/v1/role/update", r.admin.UpdateRole)      // 更新角色
	adminRoot.GET("/v1/role/list", r.admin.ListRole)           // 角色列表
	adminRoot.GET("/v1/role/my_roles", r.admin.MyRoles)        // 获取自己的角色
	adminRoot.POST("/v1/role/perm/sets", r.admin.SetRolePerms) // 设置角色权限

	// 录播课时分类管理
	adminRoot.POST("/v1/lesson/category/create", r.admin.CategoryCreate)
	adminRoot.POST("/v1/lesson/category/update", r.admin.CategoryUpdate)
	adminRoot.POST("/v1/lesson/category/delete", r.admin.CategoryDelete)
	adminRoot.GET("/v1/lesson/category/list", r.admin.CategoryList)
	adminRoot.POST("/v1/lesson/category/update_sort", r.admin.CategorySort)

	// 对象存储密钥
	adminRoot.POST("/v1/storage/get_temp_secret", r.admin.GetTempSecret)

	// 录播课时管理
	adminRoot.POST("/v1/lesson/create", r.admin.CreateLesson)
	adminRoot.POST("/v1/lesson/update", r.admin.UpdateLesson)
	adminRoot.POST("/v1/lesson/move", r.admin.MoveLesson)
	adminRoot.POST("/v1/lesson/update_status", r.admin.UpdateLessonStatus)
	adminRoot.POST("/v1/lesson/list", r.admin.LessonList)
	adminRoot.GET("/v1/lesson/info", r.admin.LessonInfo)

	// 课程管理
	adminRoot.POST("/v1/course/create", r.admin.CreateCourse)
	adminRoot.GET("/v1/course/info", r.admin.GetCourseInfo)
	adminRoot.POST("/v1/course/update", r.admin.UpdateCourse)
	adminRoot.POST("/v1/course/update_status", r.admin.UpdateCourseStatus)
	adminRoot.GET("/v1/course/list", r.admin.ListCourse)

	// 课程目录
	adminRoot.POST("/v1/course/catalog/add", r.admin.AddCatalog)
	adminRoot.POST("/v1/course/catalog/update", r.admin.UpdateCatalog)
	adminRoot.POST("/v1/course/catalog/delete", r.admin.DeleteCatalog)
	adminRoot.POST("/v1/course/catalog/update_sort", r.admin.UpdateCatalogSort)
	adminRoot.GET("/v1/course/catalog/info", r.admin.CatalogInfo)

	// 课程下的课时管理
	adminRoot.POST("/v1/course/catalog/add_lesson", r.admin.AddCatalogLesson)
	adminRoot.POST("/v1/course/catalog/remove_lesson", r.admin.RemoveCatalogLesson)
	adminRoot.POST("/v1/course/catalog/update_lesson", r.admin.UpdateCatalogLesson)

	// 订单列表
	adminRoot.POST("v1/order/list", r.admin.OrderList)
	adminRoot.GET("v1/order/info", r.admin.OrderInfo)
	adminRoot.GET("v1/order/statistic", r.admin.OrderStatistic)

	// 订单退款
	adminRoot.POST("v1/order/refund", r.admin.OrderRefund)
	adminRoot.POST("v1/order/cancel", r.admin.OrderCancel)

}

func (r *Router) customerRoute(root *gin.RouterGroup) {
	cstRoot := root.Group("/customer", AuthMiddleware(r.SpanFilter, func(ctx context.Context, token string) (*common.UserInfo, error) {
		return r.customer.GetUserByToken(ctx, token)
	}))
	// 开白接口
	cstRoot.GET("/v1/user/verify/captcha", r.customer.GetSmsCodeCaptcha)
	cstRoot.POST("/v1/user/verify/captcha/check", r.customer.CheckSmsCodeCaptcha)
	cstRoot.POST("/v1/user/verify/smscode", r.customer.GetSmsVerifyCode)
	cstRoot.POST("/v1/user/mobile/password_login", r.customer.MobilePasswordLogin)
	cstRoot.POST("/v1/user/mobile/verify_login", r.customer.MobileVerifyLogin)     // 登录即注册
	cstRoot.POST("/v1/user/mobile/reset_password", r.customer.MobilePasswordReset) // 手机号重置密码

	// 以下是需要鉴权的接口
	cstRoot.GET("/v1/user/info", r.customer.GetUserInfo)
	cstRoot.GET("/v1/course/list", r.customer.GetCourseList)                         // 课程列表
	cstRoot.GET("/v1/course/detail", r.customer.GetCourseDetail)                     // 课程信息
	cstRoot.GET("/v1/course/lesson/info", r.customer.GetCourseLessonInfo)            // 课时信息
	cstRoot.GET("/v1/course/purchased/list", r.customer.GetPurchasedCourseList)      // 已购买的课程列表
	cstRoot.GET("/v1/course/home/list", r.customer.GetHomeCourseList)                // 首页课程列表
	cstRoot.GET("/v1/course/home/info", r.customer.GetHomeCourseInfo)                // 首页课程信息
	cstRoot.GET("/v1/course/lesson/learn_info", r.customer.GetCourseLessonLearnInfo) // 课时学习进度

	// 用户购物车
	cstRoot.POST("/v1/cart/add_goods", r.customer.AddGoods)       // 添加商品到购物车
	cstRoot.POST("/v1/cart/remove_goods", r.customer.RemoveGoods) // 删除购物车商品
	cstRoot.GET("/v1/cart/list_goods", r.customer.ListGoods)      // 获取购物车商品列表

	// 订单相关
	cstRoot.POST("/v1/order/calc_fee", r.customer.OrderCalcFee) // 通过提交的课程商品ID，计算订单价格
	cstRoot.POST("/v1/order/pay_now", r.customer.OrderPayNow)   // 基于计算的价格进行订单创建
	//cstRoot.POST("/v1/order/pay_later", r.customer.OrderPayLater) // 从订单列表发起支付
	cstRoot.POST("/v1/order/cancel", r.customer.CancelOrder) // 取消订单，未支付前都可以取消
	cstRoot.GET("/v1/order/list", r.customer.GetOrderList)   // 我的订单列表
	cstRoot.GET("v1/order/info", r.customer.GetOrderInfo)    // 订单详情
}
