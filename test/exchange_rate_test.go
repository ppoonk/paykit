package test

import (
	"testing"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/ppoonk/paykit"
)

// clear && go test -v ./test/exchange_rate_test.go
func TestExchangeRate(t *testing.T) {
	// 配置日志
	l := glog.New()
	l.SetPrefix("[Test]")
	l.SetLevelStr("info")
	l.SetPath("./.log/paykit")
	l.SetStack(false)

	// 创建客户端
	ex, err := paykit.NewExchangeRate(l)
	if err != nil {
		t.Error("NewExchangeRate error", err.Error())
		return
	}
	t.Log("start cron >>>>>>")
	err = ex.StartCron()
	if err != nil {
		t.Error("ex.StartCron() error", err.Error())
		return
	}

	// 汇率转换
	t.Log("test exchange rate >>>>>>")
	ctx := gctx.New()
	ra, err := ex.GetRate(ctx, paykit.CurrencyCNY, paykit.CurrencyJPY)
	if err != nil {
		t.Error("ex.GetRate error", err.Error())
		return
	}
	t.Log("current exchange rate [ CNY : JPY ]:", ra)
	targetUnitAmount, err := ex.ConvertExchangeRate(ctx, 10000, paykit.CurrencyCNY, paykit.CurrencyJPY)
	if err != nil {
		t.Error("ConvertExchangeRate error", err.Error())
		return
	}
	t.Logf("result: [ %v : 10000] [ %v: %d]", paykit.CurrencyCNY, paykit.CurrencyJPY, targetUnitAmount)

	// // 超时
	// t.Log("wait >>>>>>")
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// defer cancel()
	// <-ctx.Done()
	// t.Log("exit >>>>>>")

}
