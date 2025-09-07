package paykit

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// TODO
type EpayClient struct {
	config          EpayConfig
	logger          *glog.Logger
	fulfillCheckout func(string)
}
type EpayConfig struct {
	Url string // 页面跳转支付 url, 格式: https://zpayz.cn/submit.php
	Key string
	Pid string
}

func NewEpayClient(config EpayConfig, logger *glog.Logger, fulfillCheckout func(string)) (*EpayClient, error) {
	ex, err := NewExchangeRate(logger)
	if err != nil {
		return nil, err
	}
	ex.StartCron()
	return &EpayClient{
		config:          config,
		logger:          logger,
		fulfillCheckout: fulfillCheckout,
	}, nil
}

func (e *EpayClient) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (res *TradePreCreateRes, err error) {
	// 处理汇率
	amount, err := ExchangeRateInstance.ConvertExchangeRateToStandardUnit(ctx, req.TotalAmount, req.Currency, CurrencyCNY)
	if err != nil {
		return nil, err
	}
	if amount < 0.01 {
		amount = 0.01
	}
	text := fmt.Sprintf(
		"money=%s&name=%s&out_trade_no=%s&notify_url=%s&pid=%s&return_url=%s",
		fmt.Sprintf("%.2f", amount),
		req.ProductSubject,
		req.OutTradeNo,
		req.SuccessURL,
		e.config.Pid,
		req.CancelURL,
	)
	sign := gmd5.MustEncryptString(text + e.config.Key)
	text = fmt.Sprintf("%s?%s&sign=%s&sign_type=MD5", e.config.Url, text, sign)

	e.logger.Info(ctx, "text: ", text)

	return &TradePreCreateRes{
		OutTradeNo: req.OutTradeNo,
		PayURL:     text,
	}, nil

}
func (e *EpayClient) Notify(w http.ResponseWriter, req *http.Request) {
	ctx := gctx.New()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	result := make(map[string]any)
	err = gjson.DecodeTo(body, &result)
	if err != nil {
		return
	}
	e.logger.Infof(ctx, "out_trade_no: %v, trade_status: %s", result["out_trade_no"], result["trade_status"])
	if result["trade_status"] == "TRADE_SUCCESS" {
		e.fulfillCheckout(result["out_trade_no"].(string))
	}
	return
}
