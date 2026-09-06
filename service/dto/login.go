package dto

import (
	"fmt"
	"mall/utils/tools"
)

type GetVerifyCaptchaReq struct {
	Once string `form:"once"`
	Time int64  `form:"ts"`
	Sign string `form:"sign"` // 秘钥固定加密： md5(once+daqing2025+ts) 转小写
}

func (r *GetVerifyCaptchaReq) CheckSign() bool {
	return r.Sign == tools.Sha256Hash(fmt.Sprintf("%s%s%d", r.Once, "daqing2025", r.Time))
}

type GetVerifyCaptchaResp struct {
	Key            string `json:"key"`
	ImageBs64      string `json:"image_base64"`       // 包含“data:image/jpeg;base64
	TitleImageBs64 string `json:"title_image_base64"` // 滑块图片，包含“data:image/jpeg;base64
	TitleHeight    int    `json:"title_height"`       // 滑块图片高
	TitleWidth     int    `json:"title_width"`        // 滑块图片宽
	TitleX         int    `json:"title_x"`            // 滑块图的x坐标
	TitleY         int    `json:"title_y"`            // 滑块图的y坐标
	Expire         int64  `json:"expire"`             // 过期时间
}

type CheckCaptchaReq struct {
	Key    string `json:"key"`
	SlideX int    `json:"slide_x"`
	SlideY int    `json:"slide_y"`
}

type CheckCaptchaDtoResp struct {
	Ticket string `json:"ticket"`
	Expire int64  `json:"expire"`
}

type MobilePasswordLoginReq struct {
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
	Ticket   string `json:"ticket"`
}

type MobilePasswordResetReq struct {
	Mobile          string `json:"mobile"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_pwd"`
	VerifyCode      string `json:"verify_code"`
}

type AdminLoginResp struct {
	Token string       `json:"token"`
	User  AdminUserDto `json:"user"`
}

type LarkQrCodeLoginReq struct {
	AppCode     int32  `json:"app_code"`
	Code        string `json:"code"`
	RedirectUri string `json:"redirect_uri"`
}

type GetSmsVerifyCodeReq struct {
	Scene    string `json:"scene"` // login, register, reset_password
	Mobile   string `json:"mobile"`
	Ticket   string `json:"ticket"`
	ClientIP string `json:"client_ip"`
}

type MobileVerifyCodeLoginReq struct {
	Mobile     string `json:"mobile"`
	VerifyCode string `json:"verify_code"`
}

type AppletLoginReq struct {
	AppCode  int32  `json:"app_code"`
	Code     string `json:"code"`
	Platform string `json:"platform"`
}

type ChangePasswordReq struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
	VerifyCode      string `json:"verify_code"`
}

type ChangePasswordSmsCodeReq struct {
	Ticket string `json:"ticket"`
}

type SmsVerifyCodeResp struct {
	DebugVerifyCode string `json:"debug_verify_code,omitempty"`
}

type ChangePasswordResp struct {
	ReloginRequired bool `json:"relogin_required"`
}

type WechatQrCodeReq struct {
	Purpose string `json:"purpose"`
}

type WechatQrCodeResp struct {
	ExpireIn   int64  `json:"expire_in"`
	SceneToken string `json:"scene_token"`
	QrcodeURL  string `json:"qrcode_url"`
}

type WechatQrCodeStatusReq struct {
	SceneToken string `form:"scene_token"`
}

type WechatScanConfirmReq struct {
	SceneToken string `json:"scene_token"`
	Code       string `json:"code"`
}

type WechatQrCodeStatusResp struct {
	State     string       `json:"state"`
	Purpose   string       `json:"purpose,omitempty"`
	Message   string       `json:"message,omitempty"`
	Token     string       `json:"token,omitempty"`
	UserInfo  *UserInfoDto `json:"user_info,omitempty"`
	BindExist bool         `json:"bind_exist,omitempty"`
}

type LoginResp struct {
	Token    string       `json:"token"`
	UserInfo *UserInfoDto `json:"user_info"`
}
