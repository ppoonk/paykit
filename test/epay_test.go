package test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ppoonk/paykit"
)

// clear && go test ./epay/test -v -run TestEpayTradePrecreate
func TestEpayTradePrecreate(t *testing.T) {
	var (
		req = &paykit.TradePreCreateReq{
			ProductSubject: "Test Product",
			OutTradeNo:     fmt.Sprintf("test_order_%d", time.Now().Unix()),
			TotalAmount:    1, // 1 分 == 0.01 元
			Currency:       "CNY",
		}

		config = paykit.EpayConfig{
			PaymentKey: "Epay 1",
			Url:        os.Getenv("EpayUrl"),
			Pid:        os.Getenv("EpayPid"),
			Key:        os.Getenv("EpayKey"),
			NotifyURL:  os.Getenv("NotifyUrl"),
			ReturnURL:  os.Getenv("ReturnUrl"),
		}

		err = error(nil)
	)

	client, err := paykit.NewEpayClient(config, func(s string) {
		fmt.Printf("Order completed: %s\n", s)
	})
	if err != nil {
		t.Fatalf("Failed to create Epay client: %v", err)
	}

	client.SetDebug(true)

	res, err := client.TradePrecreate(nil, req)
	if err != nil {
		t.Fatalf("TradePrecreate call failed: %v", err)
	}

	fmt.Printf("Pre-create order result:\n")
	fmt.Printf("Total amount: %d\n", req.TotalAmount)
	fmt.Printf("Order No: %s\n", res.OutTradeNo)
	fmt.Printf("Payment URL: %s\n", res.PayURL)

	http.HandleFunc("/", client.Notify)
	port := ":16666"
	fmt.Printf("Starting Epay notification service, listening on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
}
