package router

var AdminAuthWhiteList = map[string]bool{
	"/ping":                                true,
	"/metrics":                             true,
	"/admin/v1/user/verify/captcha/check":  true,
	"/admin/v1/user/verify/captcha":        true,
	"/admin/v1/user/verify/smscode":        true,
	"/admin/v1/user/mobile/verify_login":   true,
	"/admin/v1/user/mobile/password_login": true,
	"/admin/v1/user/password/reset":        true,

	"/customer/v1/user/verify/captcha/check":  true,
	"/customer/v1/user/verify/captcha":        true,
	"/customer/v1/user/verify/smscode":        true,
	"/customer/v1/user/mobile/verify_login":   true,
	"/customer/v1/user/mobile/password_login": true,
	"/customer/v1/user/wechat/qrcode_login":   true,
	"/customer/v1/user/wechat/qrcode_status":  true,
	"/customer/v1/user/wechat/scan_confirm":   true,
	"/customer/v1/user/applet/login":          true,
	"/customer/v1/user/mobile/reset_password": true,
	"/customer/v1/wechat/callback/payment":    true,
	"/customer/v1/wechat/callback/refund":     true,
	"/customer/v1/course/home/list":           true,
	"/customer/v1/course/home/info":           true,
}
