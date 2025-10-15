package paykit

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
)

type PaymentType string

const (
	PAYMENT_TYPE_ALIPAY PaymentType = "ALIPAY"
	PAYMENT_TYPE_EPAY   PaymentType = "EPAY"
	PAYMENT_TYPE_STRIPE PaymentType = "STRIPE"
	PAYMENT_TYPE_TRON   PaymentType = "TRON"
)

type PaymentInterface interface {
	TradePrecreate(context.Context, *TradePreCreateReq) (*TradePreCreateRes, error)
	Notify(http.ResponseWriter, *http.Request)
	// Start(context.Context) error
	Stop()
	PaymentKey() any
	PaymentType() PaymentType
}

type TradePreCreateReq struct {
	PaymentKey     any
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

type PayServer struct {
	payments  sync.Map
	cancel    context.CancelFunc
	isRunning atomic.Bool
}

func (ps *PayServer) AddOrUpdate(item ...PaymentInterface) {
	for _, pi := range item {
		if v, ok := ps.payments.Load(pi.PaymentKey()); ok {
			v.(PaymentInterface).Stop()
		}
		ps.payments.Store(pi.PaymentKey(), pi)
	}
}
func (ps *PayServer) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (*TradePreCreateRes, error) {
	p, ok := ps.payments.Load(req.PaymentKey)
	if !ok {
		return nil, fmt.Errorf("invalid payment of key: %v", req.PaymentKey)
	}
	return p.(PaymentInterface).TradePrecreate(ctx, req)
}
