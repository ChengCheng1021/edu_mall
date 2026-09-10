package rpc

import (
	"context"
	"errors"
	"fmt"
	"mall/adaptor"
	"mall/adaptor/repo/query"
	"mall/config"
	"mall/utils/logger"
	"os"
	"path/filepath"
	"sync"

	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 未做抽象处理 抽象将出参,入参更改
type IPay interface {
	// 网页扫码下单支付
	PagePrePayOrder(ctx context.Context, req alipay.TradePagePay) (string, string, error)
	// 通过商户的订单号查询订单
	QueryOrderByOutTradeNo(ctx context.Context, req alipay.TradeQuery) (*alipay.TradeQueryRsp, error)
	// 通过支付平台的订单号查询订单
	QueryByTradeNo(ctx context.Context, req alipay.TradeQuery) (*alipay.TradeQueryRsp, error)
	//  关闭订单
	CloseOrder(ctx context.Context, req alipay.TradeClose) error
	//  订单申请退款
	RefundOrder(ctx context.Context, req alipay.TradeRefund) (*alipay.TradeRefundRsp, error)
	//  订单退款查询
	QueryRefund(ctx context.Context, req alipay.TradeFastPayRefundQuery) (*alipay.TradeFastPayRefundQueryRsp, error)
}

var (
	aliPayClient *alipay.Client
	once         sync.Once
)

type AliPay struct {
	conf         *config.Config
	db           *gorm.DB
	aliPayClient *alipay.Client
}

func NewAliPay(adaptor adaptor.IAdaptor) *AliPay {
	aliPay := &AliPay{
		conf: adaptor.GetConfig(),
		db:   adaptor.GetDB(),
	}
	once.Do(func() {
		payClient, err := aliPay.initAliPayClient()
		if err != nil {
			logger.Error("init ali pay client error", zap.Error(err))
			panic(err)
		}
		aliPayClient = payClient
	})
	aliPay.aliPayClient = aliPayClient
	return aliPay
}

func (a *AliPay) getPrivateKey() (string, error) {
	qs := query.Use(a.db).PaymentPrivateKey
	first, err := qs.WithContext(context.Background()).Where(qs.AppID.Eq(a.conf.AliPay.AppID)).First()
	if err != nil {
		return "", err
	}
	return first.PrivateKey, nil
}

func (a *AliPay) initAliPayClient() (*alipay.Client, error) {
	// 获取私钥
	privateKey, err := a.getPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("获取支付私钥失败: %w", err)
	}

	if privateKey == "" {
		return nil, errors.New("支付宝私钥为空")
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
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取工作目录失败: %w", err)
	}
	appPublicCert := filepath.Join(wd, a.conf.AliPay.AppPublicCert)
	alipayPublicCert := filepath.Join(wd, a.conf.AliPay.AlipayPublicCert)
	alipayRootCert := filepath.Join(wd, a.conf.AliPay.AlipayRootCert)

	err = client.LoadAppCertPublicKeyFromFile(appPublicCert)
	if err != nil {
		return nil, fmt.Errorf("加载支付宝应用公钥证书失败: %w", err)
	}

	err = client.LoadAlipayCertPublicKeyFromFile(alipayPublicCert)
	if err != nil {
		return nil, fmt.Errorf("加载支付宝公钥证书失败: %w", err)
	}

	err = client.LoadAliPayRootCertFromFile(alipayRootCert)
	if err != nil {
		return nil, fmt.Errorf("加载支付宝根证书失败: %w", err)
	}
	return client, nil
}

func (a *AliPay) PagePrePayOrder(ctx context.Context, req alipay.TradePagePay) (string, string, error) {
	req.ProductCode = "FAST_INSTANT_TRADE_PAY"
	payURL, err := a.aliPayClient.TradePagePay(req)
	if err != nil {
		return "", "", err
	}
	return payURL.String(), "ALI_PAGE", nil
}

func (a *AliPay) QueryOrderByOutTradeNo(ctx context.Context, req alipay.TradeQuery) (*alipay.TradeQueryRsp, error) {
	return a.aliPayClient.TradeQuery(ctx, req)
}

func (a *AliPay) QueryByTradeNo(ctx context.Context, req alipay.TradeQuery) (*alipay.TradeQueryRsp, error) {
	return a.aliPayClient.TradeQuery(ctx, req)
}

func (a *AliPay) CloseOrder(ctx context.Context, req alipay.TradeClose) error {
	_, err := a.aliPayClient.TradeClose(ctx, req)
	return err
}

func (a *AliPay) RefundOrder(ctx context.Context, req alipay.TradeRefund) (*alipay.TradeRefundRsp, error) {
	return a.aliPayClient.TradeRefund(ctx, req)
}

func (a *AliPay) QueryRefund(ctx context.Context, req alipay.TradeFastPayRefundQuery) (*alipay.TradeFastPayRefundQueryRsp, error) {
	return a.aliPayClient.TradeFastPayRefundQuery(ctx, req)
}
