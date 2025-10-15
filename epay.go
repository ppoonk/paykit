package paykit

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

var _ PaymentInterface = (*EpayClient)(nil)

const (
	_EPAY_LOG_TAG       = "[Epay]"
	_EPAY_LOG_PATH      = "./.log/epay"
	_EPAY_LOG_LEVEL     = "error"
	_EPAY_TRADE_SUCCESS = "TRADE_SUCCESS"
	_EPAY_SUCCESS_RESP  = "success"
)

type EpayConfig struct {
	PaymentKey any
	Url        string // 页面跳转支付 url, 一般格式为: https://xxxxxxxx/submit.php
	Key        string
	Pid        string
	NotifyURL  string
	ReturnURL  string
}

type EpayClient struct {
	config          EpayConfig
	logger          *glog.Logger
	fulfillCheckout func(string)
}

func NewEpayClient(config EpayConfig, fulfillCheckout func(string)) (*EpayClient, error) {
	// 设置日志
	l := glog.New()
	_ = l.SetPath(_EPAY_LOG_PATH)
	_ = l.SetLevelStr(_EPAY_LOG_LEVEL)
	l.SetPrefix(_EPAY_LOG_TAG)
	l.SetStack(false)

	return &EpayClient{
		config:          config,
		logger:          l,
		fulfillCheckout: fulfillCheckout,
	}, nil
}

// TradePrecreate 创建 epay 订单，页面跳转支付，方法：GET
func (e *EpayClient) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (res *TradePreCreateRes, err error) {
	// 处理汇率
	amount, err := ERInstance.ConvertToStandard(ctx, req.TotalAmount, req.Currency, CurrencyCNY)
	if err != nil {
		return nil, err
	}
	if amount < 0.01 {
		amount = 0.01
	}
	payURL := fmt.Sprintf(
		"money=%s&name=%s&notify_url=%s&out_trade_no=%s&pid=%s&return_url=%s",
		fmt.Sprintf("%.2f", amount),
		req.ProductSubject,
		e.config.NotifyURL,
		req.OutTradeNo,
		e.config.Pid,
		e.config.ReturnURL,
	)
	sign := gmd5.MustEncryptString(payURL + e.config.Key)
	payURL = fmt.Sprintf("%s?%s&sign=%s&sign_type=MD5", e.config.Url, payURL, sign)

	e.logger.Debugf(ctx, "epay pay url: %s", payURL)

	return &TradePreCreateRes{
		OutTradeNo: req.OutTradeNo,
		PayURL:     payURL,
	}, nil

}

// Notify 接收epay的异步通知，方法：GET，收到异步通知后，需返回success以表示服务器接收到了订单通知
func (e *EpayClient) Notify(w http.ResponseWriter, req *http.Request) {
	ctx := gctx.New()
	err := req.ParseForm()
	if err != nil {
		e.logger.Error(ctx, "epay ParseForm error: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	no := req.Form.Get("out_trade_no")
	status := req.Form.Get("trade_status")

	e.logger.Debugf(ctx, "out_trade_no: %v, trade_status: %s", no, status)

	if status == _EPAY_TRADE_SUCCESS {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(_EPAY_SUCCESS_RESP))
		e.fulfillCheckout(no)
		return
	}
	return
}
func (e *EpayClient) SetDebug(debug bool) {
	e.logger.SetDebug(debug)
}

func (s *EpayClient) Start() error {
	return nil
}
func (s *EpayClient) Stop() {

}
func (s *EpayClient) PaymentKey() any {
	return s.config.PaymentKey
}
func (s *EpayClient) PaymentType() PaymentType {
	return PAYMENT_TYPE_EPAY
}
