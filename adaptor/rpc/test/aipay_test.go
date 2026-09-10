package test

import (
	"testing"

	"github.com/smartwalle/alipay/v3"
)

func TestAlipay(t *testing.T) {

	// 获取私钥
	privateKey := "你的应用私钥"
	client, err := alipay.New(
		"你的appId",
		privateKey,
		false,
	)

	err = client.LoadAppCertPublicKeyFromFile("你的应用公钥证书地址")
	if err != nil {
		t.Fatal("加载应用公钥证书失败:", err)
	}

	err = client.LoadAlipayCertPublicKeyFromFile("你的支付宝公钥证书地址")
	if err != nil {
		t.Fatal("加载支付宝公钥证书失败:", err)
	}

	err = client.LoadAliPayRootCertFromFile("你的支付宝根证书地址")
	if err != nil {
		t.Fatal("加载支付宝根证书失败:", err)
	}

	url, _ := client.TradePagePay(alipay.TradePagePay{
		Trade: alipay.Trade{
			Subject:     "test_product",
			OutTradeNo:  "TEST20260909223401",
			TotalAmount: "0.02",
			ProductCode: "FAST_INSTANT_TRADE_PAY",
		},
	})
	if err != nil {
		t.Fatal("创建支付宝支付链接失败:", err)
	}

	t.Logf("支付地址为: %s", url.String())
}
