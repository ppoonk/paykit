package paykit

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	alipayv3 "github.com/go-pay/gopay/alipay/v3"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// TODO 适配或切换成 Alipay Plus https://www.alipayplus.com/cn/consumer

var _ Payment = (*AlipayClient)(nil)

type AlipayConfig struct {
	Name             string `json:"name"`
	AppId            string `json:"appId"`
	PrivateKey       string `json:"privateKey"`
	AppPublicCert    string `json:"appPublicCert"`
	AlipayRootCert   string `json:"alipayRootCert"`
	AlipayPublicCert string `json:"alipayPublicCert"`
}

// gopay docs https://github.com/go-pay/gopay/blob/main/doc/alipay_v3.md
type AlipayClient struct {
	client          *alipayv3.ClientV3
	logger          *glog.Logger
	fulfillCheckout func(string)
}

func NewAlipayClient(config AlipayConfig, logger *glog.Logger, fulfillCheckout func(string)) (*AlipayClient, error) {
	client, err := alipayv3.NewClientV3(config.AppId, config.PrivateKey, false)
	if err != nil {
		return nil, err
	}
	err = client.SetCert([]byte(config.AppPublicCert), []byte(config.AlipayRootCert), []byte(config.AlipayPublicCert))

	// 打开Debug开关，输出日志，默认关闭
	// client.DebugSwitch = gopay.DebugOn

	// 设置自定义配置（如需要）
	//client.
	//	SetAppAuthToken("app_auth_token").    // 设置授权token
	//	SetBodySize().                        // 自定义配置http请求接收返回结果body大小，默认 10MB，没有特殊需求，可忽略此配置
	//	SetRequestIdFunc().                   // 设置自定义RequestId生成方法
	//	SetAESKey("KvKUTqSVZX2fUgmxnFyMaQ==") // 设置biz_content加密KEY，设置此参数默认开启加密（目前不可用）
	if err != nil {
		return nil, err
	}
	ex, err := NewExchangeRate(logger)
	if err != nil {
		return nil, err
	}
	ex.StartCron()

	return &AlipayClient{
		client:          client,
		logger:          logger,
		fulfillCheckout: fulfillCheckout,
	}, nil

}

// TradePrecreate
//
//	alipay.trade.precreate(统一收单线下交易预创建)
//	https://opendocs.alipay.com/open/8ad49e4a_alipay.trade.precreate
func (a *AlipayClient) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (res *TradePreCreateRes, err error) {
	// 处理汇率
	amount, err := ExchangeRateInstance.ConvertExchangeRateToStandardUnit(ctx, req.TotalAmount, req.Currency, CurrencyCNY)
	if err != nil {
		return nil, err
	}
	// alipay 要求：订单总金额。单位为元，精确到小数点后两位，取值范围：[0.01,100000000] 。【示例值】88.88
	if amount < 0.01 {
		amount = 0.01
	}

	// 请求参数
	bm := make(gopay.BodyMap)
	bm.Set("subject", req.ProductSubject).
		Set("out_trade_no", req.OutTradeNo).
		Set("total_amount", fmt.Sprintf("%.2f", amount))

	aliRes, err := a.client.TradePrecreate(ctx, bm) // 生成订单二维码，用户使用支付宝 扫一扫 完成付款。预下单请求生成的二维码有效时间为2小时
	if err != nil {
		return
	}

	if aliRes.StatusCode != alipayv3.Success {
		err = errors.New(aliRes.ErrResponse.Message)
		return
	}

	return &TradePreCreateRes{
		OutTradeNo: aliRes.OutTradeNo,
		PayURL:     aliRes.QrCode,
	}, nil
}

func (a *AlipayClient) Notify(w http.ResponseWriter, req *http.Request) {
	ctx := gctx.New()
	// 解析参数
	notifyReq, err := alipay.ParseNotifyToBodyMap(req)
	if err != nil {
		a.logger.Error(ctx, "alipay ParseNotifyToBodyMap error: ", err.Error())
		return
	}

	// 支付宝异步通知验签（公钥证书模式）
	ok, err := alipay.VerifySignWithCert([]byte(a.client.AliPayPublicCertSN), notifyReq)
	if err != nil {
		a.logger.Error(ctx, "alipay ParseNotifyToBodyMap error: ", err.Error())
		return
	}
	if !ok {
		a.logger.Error(ctx, errors.New("alipay VerifySignWithCert failure"))
		return
	}

	a.fulfillCheckout(notifyReq.Get(OUT_TRADE_NO))
	// 文档：https://opendocs.alipay.com/open/203/105286
	// 程序执行完后必须打印输出“success”（不包含引号）。如果商户反馈给支付宝的字符不是success这7个字符，支付宝服务器会不断重发通知，直到超过24小时22分钟。一般情况下，25小时以内完成8次通知（通知的间隔频率一般是：4m,10m,10m,1h,2h,6h,15h）
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
	return
}
