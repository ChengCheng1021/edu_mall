package common

type Errno struct {
	Code   int
	Msg    string
	ErrMsg string
}

func (err Errno) Error() string {
	return err.Msg
}

func (err Errno) WithMsg(msg string) Errno {
	err.Msg = err.Msg + "," + msg
	return err
}

func (err Errno) WithErr(rawErr error) Errno {
	var msg string
	if rawErr != nil {
		msg = rawErr.Error()
	}
	err.ErrMsg = err.Msg + "," + msg
	return err
}

func (err Errno) IsOk() bool {
	return err.Code == 200
}

func (err Errno) NotOk() bool {
	return !err.IsOk()
}

var (
	OK             = Errno{Code: 200, Msg: "OK"}
	ServerErr      = Errno{Code: 500, Msg: "Internal Server Error"}
	ParamErr       = Errno{Code: 400, Msg: "Param Error"}
	AuthErr        = Errno{Code: 401, Msg: "Auth Error"}
	PermissionErr  = Errno{Code: 403, Msg: "Permission Error"}
	LimitExceedErr = Errno{Code: 429, Msg: "达到频次限制，请稍后再试"}

	DatabaseErr = Errno{Code: 10000, Msg: "Database Error"}
	RedisErr    = Errno{Code: 10001, Msg: "Redis Error"}

	InvalidPasswordErr = Errno{Code: 11001, Msg: "用户不存在或密码错误"}
	InvalidCaptchaErr  = Errno{Code: 11002, Msg: "滑块校验失败，请重试"}
	PasswordErrLimit   = Errno{Code: 11003, Msg: "用户名或密码错误超过限定次数，请10分钟后再试"}
	AdminUserNotFound  = Errno{Code: 11004, Msg: "用户不存在或密码错误"}
	InvalidSmsCodeErr  = Errno{Code: 11005, Msg: "验证码错误"}
	InvalidLarkOpenID  = Errno{Code: 11006, Msg: "没有绑定飞书账号，请使用其他方式登录"}
	ConfirmPasswordErr = Errno{Code: 11007, Msg: "两次密码不一致"}

	UserNotFoundErr         = Errno{Code: 12001, Msg: "用户不存在或密码错误"}
	MobileRegisteredErr     = Errno{Code: 12002, Msg: "手机号已注册，请直接登录"}
	OrderCalcFeeErr         = Errno{Code: 12003, Msg: "商品信息已更新，请重新下单"}
	OrderLockedErr          = Errno{Code: 12004, Msg: "订单正在处理中，请刷新确认"}
	OrderNotFoundErr        = Errno{Code: 12005, Msg: "订单不存在"}
	OrderPayingErr          = Errno{Code: 12006, Msg: "订单支付确认中，请稍后刷新确认"}
	OrderCantCancelErr      = Errno{Code: 12007, Msg: "订单在该状态下，不支持取消"}
	OrderCantRefundErr      = Errno{Code: 12008, Msg: "订单在该状态下，不支持退款"}
	OrderRefundAmountErr    = Errno{Code: 12009, Msg: "退款金额错误"}
	OrderRefundNotEnoughErr = Errno{Code: 12010, Msg: "退款失败，支付平台账户余额不足"}
	WechatUnboundErr        = Errno{Code: 12011, Msg: "当前微信未绑定账号"}
	WechatBindConflictErr   = Errno{Code: 12012, Msg: "该微信已绑定其他账号"}
	WechatNeedBindErr       = Errno{Code: 12013, Msg: "请先使用手机号登录，再到个人中心绑定微信"}
	WechatScanExpiredErr    = Errno{Code: 12014, Msg: "二维码已过期，请刷新后重试"}
	WechatUnbindBlockedErr  = Errno{Code: 12015, Msg: "当前账号至少保留一种登录方式"}
	MobileBindRequiredErr   = Errno{Code: 12016, Msg: "当前账号未绑定手机号，无法设置密码"}
)
