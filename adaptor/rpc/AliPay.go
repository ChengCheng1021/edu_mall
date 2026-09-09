package rpc

import (
	"context"
	"mall/adaptor/repo/query"
	"mall/config"

	"github.com/smartwalle/alipay/v3"
	"gorm.io/gorm"
)

type AliPay struct {
	conf   *config.Config
	db     *gorm.DB
	aliPay *alipay.Client
}

func (a *AliPay) getPrivateKey() (string, error) {
	qs := query.Use(a.db).PaymentPrivateKey
	first, err := qs.WithContext(context.TODO()).Where(qs.AppID.Eq(a.conf.AliPay.AppID)).First()
	if err != nil {
		return "", err
	}
	return first.PrivateKey, nil
}

func (a *AliPay) initAliPayClient() (*AliPay, error) {
	// 获取私钥
	privateKey, err := a.getPrivateKey()
	if err != nil {
		return nil, err
	}
	// false = 沙箱
	client, err := alipay.New(
		a.conf.AliPay.AppID,
		privateKey,
		!a.conf.AliPay.Sandbox,
	)
	if err != nil {
		return nil, err
	}

	// 普通支付宝公钥模式
	if err := client.LoadAliPayPublicKey(a.conf.AliPay.PublicKey); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &AliPay{
		aliPay: client,
	}, nil
}
