package test

import (
	"testing"

	"github.com/smartwalle/alipay/v3"
)

func TestAlipay(t *testing.T) {

	// 获取私钥
	privateKey := "MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQC8qF748hTR7TpV37ddubV9R4XeHl748SIic2PhXfTb3w8eH7uWlcoJKkziGjDOLSIrUBo5S2ADcq+bXo+coW8RTRVzwxPzpmRWzLPEEqHbfy4umuRKAyog1ey9izO237gE1yYXiF1bCX22fhD3exW2eqeExVfKTlVcd3PiREASDI2zcLIYA5hAZQw4dgLzGTuInYbnHbm19Q4R7sdGULjOpTc2EWoyO3fj1KQ2ZveRf4bEyI3uKuDCAizVb+Qdjzg1e0FeNRED9+EPNZRzLAh+7nAA74pxkROu1xY5bHCJ8TdbhXph6YXn72FAmeMdi17G3igBGel9y1+At33tpdjRAgMBAAECggEAK+TAtAse7PjU6cXzU8sxfsR1UQif8CuqVXmjc3v4zG9JhSi87HxNVXSSDskpMc8udAVfFJWE8Uhtsyh9IWQuA0h8BUMOEVJVZhyadrQfFIKyrAU9uDqkQp+DRVZt4c6LchTct/zyO4wpw5vxNqNcmehPsYR4uIkhMzJXs/1NKuV6KCqH1WlQF5jpk6fdUUhucqRCU1YEWmM8lhgdlhuqpDvPsBJpm0by0XjTIQBIp1c91LApVXxPXn6c8c9BVzH6gfqFaoNgx6fZevZQYjwtsxIDGZyv5HlNVJiDzGxbDAefnqvOVFsGV1qHOqsaajkNHAO1oEouXWGzvUhgwmCpQQKBgQDwJD8uLeumiUyUeL51bfqCeufdqmkbZW6F7ZXNMf9e6CVVJ2+19R+91PvYT67SucQiQ7zObySYi4VfyqfinvsyxR02aSdmQ0mKq/JqhEgLy/cGoEpi5fYBriDZ51NTivUcibk2F29k0/kl4vcM2psA46wmt9fYfSMhTem71FHA2wKBgQDJHcE62Wtk8LLoXE0pK0133DyyPi1e6YO6fKV5K0OjgWNXyMD4giK+NvH4U/4X0he2gM+b8V08NkC0lxFJpXkhTz3zxVvqgBDnZfO6lCclAWjvw4C7AE2tCPDsdL+JHH0eNLqWfNJTzukX0elmbISGXWf4n310iR4mk5L2TOF2wwKBgQDnBIe/Xi/QM093mbzn7VhMg/5hUdnxoB+2Obyd/VZFsCCSDfE648iYb7ej/ewaDtnveKi/E07qbXZuk9/0dKsFyXjz6i8cAulRvV7lN8KzjpFjT3qgL8f9D83MsuyHdyucO6XwspTYM9AAsZqnQ/pP3ba0PLIqMyBDnteXeYb4dwKBgQCdMkWrku+PaVfdqO+iwzb8/cbvZwwdiJYu+Gh6aienMGYO4lp6o3U2ikndWQFdaxifzNT5RdIjUyCGRyH7F3yzXXXGCTgL9efAhn7YEh76nLyB06TWBamxGzD9EU/4gq0FJB/Hqm7XlP26YZd2OFHpmC7BNSBhKx/G9UhEfdko8wKBgQCUepsJzt0m1yWqE7bqNzrQEXERBzPkUqdPJIUPYnTdsaKSgvMjcE9evxXxMi29VUFHQ0mE5fF0MTDu4pj0xVI4tkR75StzqxYIQsW2P/AyxDFFQPtmqxWs+ldcCSoWxQi/h+ceom8IxhEvbiY9H1gRNpMNqrHVeLXCdav2ptDJdQ=="
	client, err := alipay.New(
		"9021000167693491",
		privateKey,
		false,
	)

	//err = client.LoadAppCertPublicKeyFromFile("你的应用公钥证书地址")
	//if err != nil {
	//	t.Fatal("加载应用公钥证书失败:", err)
	//}
	//
	//err = client.LoadAlipayCertPublicKeyFromFile("你的支付宝公钥证书地址")
	//if err != nil {
	//	t.Fatal("加载支付宝公钥证书失败:", err)
	//}
	//
	//err = client.LoadAliPayRootCertFromFile("你的支付宝根证书地址")
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
