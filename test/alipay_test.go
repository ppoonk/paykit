package test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/ppoonk/paykit"
)

// clear && go test ./alipay/test -v -run TestTradePrecreate
func TestTradePrecreate(t *testing.T) {
	var (
		ctx = gctx.New()

		req = &paykit.TradePreCreateReq{
			ProductSubject: "Test Product",
			OutTradeNo:     fmt.Sprintf("test_order_%d", time.Now().Unix()),
			TotalAmount:    101, // 101 分 == 1.01 元
			Currency:       "CNY",
		}
		config = paykit.AlipayConfig{
			Name:             "alipay sandbox account",
			AppId:            os.Getenv("AlipaySandbox_AppId"),
			AppPrivateKey:    os.Getenv("AlipaySandbox_AppPrivateKey"),
			AlipayPublicCert: os.Getenv("AlipaySandbox_AlipayPublicCert"),
			IsProd:           false,
			SuccessURL:       os.Getenv("NotifyURL"),
		}

		err = error(nil)
	)

	client, err := paykit.NewAlipayClient(config, func(s string) {
		fmt.Printf("Order completed: %s\n", s)
	})
	if err != nil {
		t.Fatalf("Failed to create Alipay client: %v", err)
	}
	client.SetDebug(true)

	res, err := client.TradePrecreate(ctx, req)
	if err != nil {
		t.Fatalf("TradePrecreate call failed: %v", err)
	}

	fmt.Printf("Pre-create order result:\n")
	fmt.Printf("Total amount: %d\n", req.TotalAmount)
	fmt.Printf("Order No: %s\n", res.OutTradeNo)
	fmt.Printf("Payment URL: %s\n", res.PayURL)

	http.HandleFunc("/", client.Notify)
	port := ":16666"
	fmt.Printf("Starting Alipay notification service, listening on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}

}
