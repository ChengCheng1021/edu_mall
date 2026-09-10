package test

import (
	"testing"

	"github.com/smartwalle/alipay/v3"
)

func TestAlipay(t *testing.T) {

	// 获取私钥
	//privateKey := "MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCL2J+C6eU9zGkYnN7T+oBxkOw9cJ7+Mrl17jTw+kPfKGi85Ysit6frcYZg+xaDaJOGSDaPyapQgAoL7hH+NlZGC6fIJG+y6qgAQaayoNlUg7ctprf3xtGyts5oOlXDm4vLINPeWF8q3sGzNlLwOjZiZQBbBquygyMhccsJsfQ0n5pEOysyYo8y5AUOc10XPwOy+wY7ywUu8ZdYPsrtw2M4t0KdKsqgUhOcXHXc2iCHTBVxdPzDnwsx1otnccXhShiSqVFO1Keg3Qpcg4vyQk6+6EwXaLl5TSuI6e72iBAAG5tWlQC7DaPrYdxQoJlo15vCCSseLMn1CReIAGbq4lYnAgMBAAECggEBAIqt0NDj3X8BHB9aQOZ5fbIhAwSSkDiIWL4H8Nwfcfr0eZkJEIbnFVA4DghSNqsto04Agoroc0rNDilydslfXQKtQD8LUvFcHinS8NonBB35WEefEsRVl1HgUqOtZatKrsBK14+glw9OQ0vSzUCImbHNcyLRZKbrwITD8ZK1s/QR65VyipaVJgoGdcr6pvN1iLeRtGRZVRjcIaSuGcjOlHgClO1SXb63R8OqNx172LYZIzqKaZFI3GZi3oZ5bN+XuIT9P31iad5RK4yBSmRL3vDx7ZguHt545dvHOpsns+V4qK8ld4bRLNC/3T5EnkUtXE6fHICyvRbdoBnuQauqU2ECgYEA5UXVfAVwM/FkPz/J3fzYbzabWdmQUnsKN8Iq08GNs8Kjbr+ldct7fQ6yKmZI+zppsVP18sUonM9EEhGjBHJ4IglyYgwnSRCq9TNfiGKZBPdZ0dOQUG6NoA7+S7L9/csJ8+yeRzAHDTw7pdzxKpk1oEppELiSn0TkgofpXrHFYVkCgYEAnCYKG9ON9oN19Q3+H6q98qYaQFvHNvPGXrFcaTiLoX2xBC6rHFdFwSudMLvb+LHNyNakR0+K7mxoLjhBx2iL/S6uEuIFUOJKIpgd3Y0BXGJr5/3DOM1EpwoH6S5GaNiuDNJxaIf1WEYdg+jBQkrjOyV1cEb1kdiODv65w7MLA38CgYBTMPe5vK9t6ZUabibtaaWPFR1hiNQZWZPnj4jCtWSZaXKr6NY828y/H+n+AIdSwWtAcNq5cFjALWThuYyRPIfisdLTSN2oYWfm+PEdJ8mmR6pLvJyM0tCI12fmR9hpkpbV73GvGvo0DzsFgBnx/w26T8W3z4FEUVcpFe/T8GVSYQKBgH2bv//4YzeNMqMpWWZB5EDAS1fAPHXBDa74v5zI5tHGmVIC9JR8w9kLa4xbYi0hYqePumC+5MS8oeWkTY3KVOoa1d7MwHf7QKWpdkTVe+XEKodZQ+R6gyJX2FtEZVFMFF6uHpp22+7hoDaPUn8wXLAkht8Fxd9Hs6buU6LQcSijAoGAJJg46ni430w5670c30ELCSw+8CI5q1Rc53IyTVqBs3engtavwzBfKeTAvJ+eA4hQqJKtQXhrE2iu73TtvzpXwAWQ7Tj5qkVbrkhCI6JTBMIHmd549c+k0oAvcJ2Hyc8VzCh71+fWowy4gUNwdtk9m73LeGPyhJcu1sZmwmsBj34="
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
