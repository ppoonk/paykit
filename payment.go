package paykit

import (
	"context"
	"net/http"
)

type PaymentInterface interface {
	TradePrecreate(context.Context, *TradePreCreateReq) (*TradePreCreateRes, error)
	Notify(http.ResponseWriter, *http.Request)
}

type TradePreCreateReq struct {
	ProductSubject string   // 商品标题
	OutTradeNo     string   // 内部订单系统编号
	TotalAmount    int64    // 订单总价
	Currency       Currency // 订单总价货币
	Extra          any      // 扩展参数
}

type TradePreCreateRes struct {
	OutTradeNo string `json:"out_trade_no"` // 订单系统编号
	PayURL     string `json:"qr_code"`      // 支付链接
	Extra      any    `json:"extra"`        // 扩展参数
}
