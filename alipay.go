package paykit

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/smartwalle/alipay/v3"
)

var (
	_ PaymentInterface = (*AlipayClient)(nil)
)

const (
	alipayLogTag   = "[Alipay]"
	alipayLogPath  = "./log/alipay"
	alipayLogLevel = "error"
)

type (
	AlipayConfig struct {
		Name             string
		AppId            string
		AppPrivateKey    string // 应用私钥
		AlipayPublicCert string // 支付宝公钥
		IsProd           bool   // 是否生产环境
		SuccessURL       string // 成功回调URL
	}
	AlipayClient struct {
		config          AlipayConfig
		client          *alipay.Client
		logger          *glog.Logger
		fulfillCheckout func(string)
	}
)

func NewAlipayClient(config AlipayConfig, fulfillCheckout func(string)) (*AlipayClient, error) {
	// 设置日志
	l := glog.New()
	_ = l.SetPath(alipayLogPath)
	_ = l.SetLevelStr(alipayLogLevel)
	l.SetPrefix(alipayLogTag)
	l.SetStack(false)

	client, err := alipay.New(config.AppId, config.AppPrivateKey, config.IsProd)
	if err != nil {
		return nil, err
	}
	// 使用支付宝公钥模式
	_ = client.LoadAliPayPublicKey(config.AlipayPublicCert)

	return &AlipayClient{
		config:          config,
		client:          client,
		logger:          l,
		fulfillCheckout: fulfillCheckout,
	}, nil

}

// TradePrecreate
func (a *AlipayClient) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (res *TradePreCreateRes, err error) {
	// 处理汇率
	amount, err := ERInstance.ConvertToStandard(ctx, req.TotalAmount, req.Currency, CurrencyCNY)
	if err != nil {
		return nil, err
	}
	if amount < 0.01 {
		amount = 0.01
	}

	//创建支付宝订单
	var order alipay.TradePreCreate
	order.NotifyURL = a.config.SuccessURL
	order.Subject = req.ProductSubject
	order.OutTradeNo = req.OutTradeNo
	order.TotalAmount = fmt.Sprintf("%.2f", amount) // alipay 要求：订单总金额。单位为元，精确到小数点后两位，取值范围：[0.01,100000000] 。【示例值】88.88

	//	alipay.trade.precreate(统一收单线下交易预创建)
	//	https://opendocs.alipay.com/open/8ad49e4a_alipay.trade.precreate
	aliRes, err := a.client.TradePreCreate(ctx, order)
	if err != nil {
		return nil, err
	}
	return &TradePreCreateRes{
		OutTradeNo: aliRes.OutTradeNo,
		PayURL:     aliRes.QRCode,
	}, nil
}

func (a *AlipayClient) Notify(w http.ResponseWriter, req *http.Request) {
	ctx := gctx.New()
	err := req.ParseForm()
	if err != nil {
		a.logger.Error(ctx, "alipay ParseForm error: ", err.Error())
	}
	noti, err := a.client.DecodeNotification(req.Form)
	if err != nil {
		a.logger.Error(ctx, "alipay DecodeNotification error: ", err.Error())
		return
	}
	alipay.ACKNotification(w)
	// a.logger.Debug(ctx, "alipay notifyReq: ", noti)
	a.logger.Debug(ctx, "alipay notifyReq OutTradeNo: ", noti.OutTradeNo)
	a.fulfillCheckout(noti.OutTradeNo)
	return
}

func (a *AlipayClient) SetDebug(debug bool) {
	a.logger.SetDebug(debug)
}
