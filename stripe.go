package paykit

import (
	"context"
	"encoding/json"

	"net/http"
	"strings"

	"io"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

const (
	OUT_TRADE_NO   = "OUT_TRADE_NO"
	stripeLogTag   = "[Stripe]"
	stripeLogPath  = "./.log/stripe"
	stripeLogLevel = "error"
)

type StripeConfig struct {
	Name           string
	Key            string
	EndpointSecret string
	SuccessURL     string // 成功回调URL
	CancelURL      string // 失败回调URL
}

var _ PaymentInterface = (*StripeClient)(nil)

type StripeClient struct {
	config          StripeConfig
	client          *stripe.Client
	endpointSecret  string
	logger          *glog.Logger
	fulfillCheckout func(string)
}

func NewStripeClient(config StripeConfig, fulfillCheckout func(string)) (*StripeClient, error) {
	// 设置日志
	l := glog.New()
	_ = l.SetPath(stripeLogPath)
	_ = l.SetLevelStr(stripeLogLevel)
	l.SetPrefix(stripeLogTag)
	l.SetStack(false)

	return &StripeClient{
		config:          config,
		client:          stripe.NewClient(config.Key),
		endpointSecret:  config.EndpointSecret,
		logger:          l,
		fulfillCheckout: fulfillCheckout,
	}, nil
}

// TradePrecreate 交易与创建
func (s *StripeClient) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (res *TradePreCreateRes, err error) {
	params := &stripe.CheckoutSessionCreateParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: stripe.StringSlice([]string{ // https://docs.stripe.com/payments/wallets
			"card",
			"wechat_pay",
			"alipay",
		}),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionCreateLineItemPriceDataParams{
					Currency: stripe.String(strings.ToLower(string(req.Currency))),
					ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
						Name: stripe.String(req.ProductSubject),
					},
					UnitAmount: stripe.Int64(req.TotalAmount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.config.SuccessURL),
		CancelURL:  stripe.String(s.config.CancelURL),
		Metadata: map[string]string{
			OUT_TRADE_NO: req.OutTradeNo,
		},
	}
	params.AddExtra("payment_method_options[wechat_pay][client]", "web")
	cs, err := s.client.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &TradePreCreateRes{
		OutTradeNo: cs.ID,
		PayURL:     cs.URL,
	}, nil

}

// Notify 异步回调，异步通知
//
//	docs:
//	https://docs.stripe.com/checkout/fulfillment?payment-ui=stripe-hosted
func (s *StripeClient) Notify(w http.ResponseWriter, req *http.Request) {
	ctx := gctx.New()
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// Pass the request body and Stripe-Signature header to ConstructEvent, along with the webhook signing key
	event, err := webhook.ConstructEvent(body, req.Header.Get("Stripe-Signature"), s.endpointSecret)

	if err != nil {
		s.logger.Errorf(ctx, "Error verifying webhook signature: %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	s.logger.Debug(ctx, "stripe event.Type: ", event.Type)

	if event.Type == stripe.EventTypeCheckoutSessionCompleted {
		var cs stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &cs)
		if err != nil {
			s.logger.Errorf(ctx, "Error parsing webhook JSON: %v\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.fulfillCheckout(cs.Metadata[OUT_TRADE_NO])
	}
	w.WriteHeader(http.StatusOK)
}
