package test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/ppoonk/paykit"
)

func TestTron(t *testing.T) {
	ctx := gctx.New()

	config := paykit.TronConfig{
		PaymentKey:   "Tron 1",
		APIKey:       "",
		Address:      os.Getenv("CollectionAddress"),
		OrderTimeout: 60,
		AcceptTokens: []paykit.TokenSymbol{paykit.USDC_TRC20, paykit.USDT_TRC20},
	}
	tronClient, err := paykit.NewTron(config, func(s string) {})
	if err != nil {
		t.Errorf("NewTron error: %v", err)
		return
	}

	t.Log("Test order concurrency")
	// 启动多个个协程并发调用TradePrecreate
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// 构造订单参数
			req := &paykit.TradePreCreateReq{
				OutTradeNo:  fmt.Sprintf("%d", 1001+index),
				TotalAmount: 1400,
				Currency:    "CNY",
				Extra: paykit.TronExtraForTradePreCreateReq{
					TokenSymbol: paykit.USDT_TRC20,
				},
			}

			// 调用TradePrecreate
			r, err := tronClient.TradePrecreate(ctx, req)

			if err != nil {
				t.Errorf("TradePrecreate error: %v", err)
				return
			}
			ta := r.Extra.(paykit.TronExtraForTradePreCreateRes).TotalAmountString
			t.Logf("TradePrecreate out_trade_no: %s, total_amount: %s", req.OutTradeNo, ta)
		}(i)
	}

	// 等待所有协程完成
	wg.Wait()

	t.Log("Test key reuse")
	// 创建两个相同金额的订单，验证key生成和重用
	req1 := &paykit.TradePreCreateReq{
		OutTradeNo:  "TEST_001",
		TotalAmount: 10000, // 100 CNY
		Currency:    "CNY",
		Extra: paykit.TronExtraForTradePreCreateReq{
			TokenSymbol: paykit.USDT_TRC20,
		},
	}

	req2 := &paykit.TradePreCreateReq{
		OutTradeNo:  "TEST_002",
		TotalAmount: 20000, // 200 CNY
		Currency:    "CNY",
		Extra: paykit.TronExtraForTradePreCreateReq{
			TokenSymbol: paykit.USDT_TRC20,
		},
	}

	// 第一个请求
	r1, err := tronClient.TradePrecreate(ctx, req1)
	if err != nil {
		t.Errorf("First TradePrecreate error: %v", err)
		return
	}
	ta1 := r1.Extra.(paykit.TronExtraForTradePreCreateRes).TotalAmountString
	t.Logf("TradePrecreate out_trade_no: %s, total_amount: %s", r1.OutTradeNo, ta1)

	// 删除 key
	tronClient.Cacha().Remove(ctx, ta1)

	// 第二个请求
	r2, err := tronClient.TradePrecreate(ctx, req2)
	if err != nil {
		t.Errorf("Second TradePrecreate error: %v", err)
		return
	}
	ta2 := r2.Extra.(paykit.TronExtraForTradePreCreateRes).TotalAmountString
	t.Logf("TradePrecreate out_trade_no: %s, total_amount: %s", r2.OutTradeNo, ta2)

}

func TestTronTransactions(t *testing.T) {
	ctx := gctx.New()

	config := paykit.TronConfig{
		APIKey:       "",
		Address:      os.Getenv("CollectionAddress"),
		OrderTimeout: 60,
		AcceptTokens: []paykit.TokenSymbol{paykit.USDC_TRC20, paykit.USDT_TRC20},
	}
	tronClient, err := paykit.NewTron(config, func(s string) {
		t.Logf("Fulfill checkout, outTradeNo: %s", s)
	})
	if err != nil {
		t.Errorf("NewTron error: %v", err)
		return
	}
	err = tronClient.Start()
	if err != nil {
		t.Errorf("Start error: %v", err)
		return
	}
	t.Log("Start Success")

	// 模拟数据, 可用测试钱包转入相同金额测试
	outTradeNo := "12345678"
	amount := 1.23 // 必须2位小数，和内部函数逻辑保持一致
	err = tronClient.Cacha().Set(ctx, amount, outTradeNo, 0)
	if err != nil {
		t.Errorf("Set cache error: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	for {
		select {
		case <-sigCtx.Done():
			t.Log("Received interrupt signal")
			return
		case <-ctx.Done():
			t.Log("Test timeout reached")
			return
		default:
			continue
		}

	}

}
