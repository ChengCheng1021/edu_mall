package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mall/adaptor"
	"mall/config"
	"mall/service/do"
	"mall/utils/http"
	"mall/utils/logger"

	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	HttpAccept         = "application/json"
	HttpContentType    = "application/json;charset=utf-8"
	larkHost           = "https://open.feishu.cn"
	LarkChatGroupType  = "chat_id"
	LarkChatPersonType = "open_id"
)

type ILark interface {
	GetLarkUserInfo(ctx context.Context, userAccessToken string) (*do.LarkUserInfo, error)

	GetLarkUserAccessToken(ctx context.Context, appCode int32, code, redirectUrl, scope string) (*do.LarkUserAccessToken, error)
	GetLarkTenantAccessToken(ctx context.Context, appCode int32) (*do.LarkTenantAccessToken, error)

	SendLarkMsg(ctx context.Context, getToken GetTokenFunc, req *do.SendLarkMsg) error
}

type GetTokenFunc func(ctx context.Context, force bool) (string, error)

type Lark struct {
	conf *config.Config
}

func NewLark(adaptor adaptor.IAdaptor) *Lark {
	return &Lark{
		conf: adaptor.GetConfig(),
	}
}

func (l *Lark) GetLarkUserInfo(ctx context.Context, userAccessToken string) (*do.LarkUserInfo, error) {
	url := fmt.Sprintf("%s/open-apis/authen/v1/user_info", larkHost)
	param := map[string]interface{}{}
	headers := map[string]string{
		"Content-Type":  "application/json; charset=utf-8",
		"Authorization": fmt.Sprintf("Bearer %s", userAccessToken),
	}

	type Response struct {
		Code int64           `json:"code"`
		Msg  string          `json:"msg"`
		Data do.LarkUserInfo `json:"data"`
	}

	resp := &Response{}
	err := http.Get(ctx, url, headers, param, resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("code:%d, msg:%s", resp.Code, resp.Msg)
	}

	return &resp.Data, nil
}

func (l *Lark) GetLarkUserAccessToken(
	ctx context.Context,
	appCode int32, code string,
	redirectUrl string, scope string) (*do.LarkUserAccessToken, error) {
	url := fmt.Sprintf("%s/open-apis/authen/v2/oauth/token", larkHost)

	body := map[string]interface{}{
		"grant_type":    "authorization_code",
		"client_id":     l.conf.AppConf[appCode].AppID,
		"client_secret": l.conf.AppConf[appCode].AppSecret,
		"code":          code,
		"redirect_uri":  redirectUrl,
		"scope":         scope,
	}
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json;charset=utf-8",
	}
	resp := &do.LarkUserAccessToken{}
	err := http.Post(ctx, url, headers, body, resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, err
	}

	return resp, nil
}

func (l *Lark) GetLarkTenantAccessToken(ctx context.Context, appCode int32) (*do.LarkTenantAccessToken, error) {
	url := fmt.Sprintf("%s/open-apis/auth/v3/tenant_access_token/internal", larkHost)

	body := map[string]interface{}{
		"app_id":     l.conf.AppConf[appCode].AppID,
		"app_secret": l.conf.AppConf[appCode].AppSecret,
	}
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json;charset=utf-8",
	}
	resp := &do.LarkTenantAccessToken{}
	err := http.Post(ctx, url, headers, body, resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("resp code != 0, %+v", resp)
	}
	return resp, nil
}

type LarkResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func (l LarkResp) IsOK() bool {
	return l.Code == 0
}

func (l LarkResp) IsTokenInvalid() bool {
	return l.Code == 4001
}

func (l LarkResp) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", l.Code, l.Msg)
}

// 发送消息
func (l *Lark) SendLarkMsg(ctx context.Context, getToken GetTokenFunc, req *do.SendLarkMsg) error {
	// https://open.feishu.cn/document/server-docs/im-v1/message/create
	path := fmt.Sprintf("%s/open-apis/im/v1/messages", larkHost)
	resp := &LarkResp{}
	force := false
	reqHeaders := map[string]string{
		"Content-Type": HttpContentType,
	}
	content, err := json.Marshal(map[string]string{
		"text": req.Content,
	})
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"content":    string(content),
		"msg_type":   "text",
		"receive_id": req.OpenID,
	}
	for _, v := range lo.Range(2) {
		token, err := getToken(ctx, force)
		if err != nil {
			return err
		}
		reqHeaders["Authorization"] = "Bearer " + token
		url := fmt.Sprintf("%s?receive_id_type=%s", path, req.IDType)
		err = http.Post(ctx, url, reqHeaders, body, resp)
		if err != nil {
			logger.Error("SendLarkMsg send fail", zap.Any("index", v), zap.String("err", err.Error()))
			continue
		}
		// 发送成功
		if resp.IsOK() {
			break
		}
		// token失效，触发强刷
		if resp.IsTokenInvalid() {
			force = true
			continue
		}
		logger.Error("SendLarkMsg send resp error", zap.Any("index", v), zap.Any("resp", resp.Error()))
	}

	if !resp.IsOK() {
		logger.Error("SendLarkMsg send fail", zap.Any("resp", resp))
		return errors.New(resp.Error())
	}

	return nil
}
