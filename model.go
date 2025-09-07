package paykit

import (
	"context"
	"net/http"
)

type Payment interface {
	TradePrecreate(context.Context, *TradePreCreateReq) (*TradePreCreateRes, error)
	Notify(http.ResponseWriter, *http.Request)
}

type TradePreCreateReq struct {
	ProductSubject string   `json:"productSubject"` // 商品标题
	OutTradeNo     string   `json:"outTradeNo"`     // 内部订单系统编号
	TotalAmount    int64    `json:"totalAmount"`    // 订单总价
	Currency       Currency `json:"currency"`       // 订单总价货币
	SuccessURL     string   `json:"successURL"`
	CancelURL      string   `json:"cancelURL"`
}

type TradePreCreateRes struct {
	OutTradeNo string `json:"out_trade_no"`
	PayURL     string `json:"qr_code"`
}

const (
	OUT_TRADE_NO = "OUT_TRADE_NO"
	// OUT_PRODUCT_ID = "OUT_PRODUCT_ID"
	// OUT_PRICE_ID   = "OUT_PRICE_ID"
)
