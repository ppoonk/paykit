package test

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/ppoonk/paykit"
)

// clear && go test -v test/stripe_test.go
func TestStripe(t *testing.T) {
	ctx := gctx.New()
	stripeConfig := paykit.StripeConfig{
		PaymentKey:     "Stripe 1",
		StripeKey:      os.Getenv("StripeTestKey"),
		EndpointSecret: os.Getenv("StripeTestEndpointSecret"),
	}

	c, err := paykit.NewStripeClient(stripeConfig, func(s string) {
		t.Log("webhook handler, current out_trade_no: ", s)
	})
	if err != nil {
		t.Error("NewStripeClient error", err.Error())
		return
	}

	// 测试创建支付链接
	t.Log("test creat stripe checkout sessions:")
	cs, err := c.TradePrecreate(ctx, &paykit.TradePreCreateReq{
		ProductSubject: "Subject Subject",
		OutTradeNo:     "order_112233",
		TotalAmount:    1000,
		Currency:       paykit.CurrencyCNY,
	})
	if err != nil {
		t.Error("c.TradePrecreate error", err.Error())
		return
	}
	t.Logf("test create stripe checkout sessions success, url: %v", cs.PayURL)

	// 测试 webhook
	http.HandleFunc("/webhook", c.Notify)
	addr := ":4242"
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
