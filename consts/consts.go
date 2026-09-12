package consts

import "time"

const (
	AdminUser    = 1 // 管理后台用户
	CustomerUser = 2 // 客户端用户
)

const (
	AdminTokenKey   = "token"
	UserTokenKey    = "token"
	CustomerUserKey = "user_key"
	AdminUserKey    = "admin_user_key"
)

const (
	AdminUserTokenExpire    = time.Hour * 24
	CustomerUserTokenExpire = time.Hour * 24 * 7
	PasswordErrExpire       = time.Minute * 10
	PasswordErrMaxCount     = 3
)

const (
	ExpireTokenDueDuration    = 200
	OrderCalcFeeExpire        = time.Minute * 10
	LessonLearnSessionExpire  = time.Hour * 24
	LessonLearnSessionTimeout = time.Minute
)

const (
	LessonLearnStatusLearning  = 1
	LessonLearnStatusCompleted = 2
)

const (
	LessonLearnTypePlay      = 1
	LessonLearnTypePause     = 2
	LessonLearnTypeExit      = 3
	LessonLearnTypeHeartbeat = 4
)

const (
	IsEnable  = 1
	IsDisable = -1
)

const (
	SexUnknown = 0
	SexMale    = 1
	SexFemale  = 2
)

const (
	WechatAppCode = 1000 // 公众号
	AppletAppCode = 1001 // 小程序
	LarkAppCode   = 2000
)

const (
	WechatScanTaskExpire = time.Minute * 2
	WechatAppletScanPage = "pages/scan-confirm/index"
)

const (
	WechatScanPurposeLogin = "login"
	WechatScanPurposeBind  = "bind"
)

const (
	WechatScanStateWaiting      = "waiting"
	WechatScanStateScanned      = "scanned"
	WechatScanStateLoginSuccess = "login_success"
	WechatScanStateUnbound      = "unbound"
	WechatScanStateBindSuccess  = "bind_success"
	WechatScanStateExpired      = "expired"
	WechatScanStateConflict     = "conflict"
	WechatScanStateInvalid      = "invalid"
)

// 验证码场景
const (
	AddAdminUserPasswordSmsCode   = "add_admin_user_password"
	AdminUserMobileLoginSmsCode   = "admin_user_mobile_login"
	AdminUserReSetPasswordSmsCode = "admin_user_reset_password"

	RegisterUserSmsCode      = "user_register"
	UserMobileLoginSmsCode   = "user_mobile_login"
	UserReSetPasswordSmsCode = "user_reset_password"
	UserMobileBindSmsCode    = "user_mobile_bind"
)

// 订单状态
// -1：已取消，1: 待支付 2：已支付（待发货） 3：已退款  4：已发货 5：已签收 6：已完成
const (
	OrderStatusCancel    = -1
	OrderStatusWaitPay   = 1
	OrderStatusPayed     = 2
	OrderStatusRefund    = 3 // 已退款
	OrderStatusShipped   = 4 // 已发货
	OrderStatusReceived  = 5 // 已签收
	OrderStatusCompleted = 6 // 已完成
)

const (
	RefundStatusProcessing = 1
	RefundStatusDone       = 2
	RefundStatusException  = 3
)

const (
	CustomerCancel = 1 // 用户取消
	AdminCancel    = 2 // 客服取消
	TimeoutCancel  = 3 // 超时取消
)

const (
	WechatRefundSuccess    = "REFUND_SUCCESS"
	WechatRefundClosed     = "CLOSED"
	WechatRefundProcessing = "PROCESSING"
	WechatRefundAbnormal   = "ABNORMAL"
)

func GetRefundStatus(wxStatus string) int32 {
	switch wxStatus {
	case WechatRefundSuccess:
		return RefundStatusDone
	case WechatRefundClosed:
		return RefundStatusException
	case WechatRefundAbnormal:
		return RefundStatusException
	default:
		return RefundStatusProcessing
	}
}

// 订单来源
// 1：用户下单  2：管理后台  3：系统赠送
const (
	OrderSourceUser  = 1
	OrderSourceAdmin = 2
	OrderSourceSys   = 3
)

// 商品类型
// 1: 课程商品
const (
	CourseGood = 1
)

// 时长， 1：一个月 2：三个月 3：半年 4：一年 5：二年 6：三年
const (
	LearnExpireTimeMonth      = 1
	LearnExpireTimeThreeMonth = 2
	LearnExpireTimeHalfYear   = 3
	LearnExpireTimeYear       = 4
	LearnExpireTimeTwoYear    = 5
	LearnExpireTimeThreeYear  = 6
)

func GetExpireTime(expireTime int32) int64 {
	switch expireTime {
	case LearnExpireTimeMonth:
		return time.Now().AddDate(0, 1, 1).UnixMilli()
	case LearnExpireTimeThreeMonth:
		return time.Now().AddDate(0, 3, 1).UnixMilli()
	case LearnExpireTimeHalfYear:
		return time.Now().AddDate(0, 6, 1).UnixMilli()
	case LearnExpireTimeYear:
		return time.Now().AddDate(1, 0, 1).UnixMilli()
	case LearnExpireTimeTwoYear:
		return time.Now().AddDate(2, 0, 1).UnixMilli()
	case LearnExpireTimeThreeYear:
		return time.Now().AddDate(3, 0, 1).UnixMilli()
	}
	return 0
}

const (
	DateTypeDay     = 1
	DateTypeMonth   = 2
	DateTypeQuarter = 3
	DateTypeYear    = 4
)
