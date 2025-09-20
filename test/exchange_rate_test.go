package test

import (
	"testing"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/ppoonk/paykit"
)

func TestExchangeRate(t *testing.T) {
	t.Log("test exchange rate >>>>>>")
	ctx := gctx.New()
	ra, err := paykit.ERInstance.GetRate(ctx, paykit.CurrencyCNY, paykit.CurrencyJPY)
	if err != nil {
		t.Error("ex.GetRate error", err.Error())
		return
	}
	t.Log("current exchange rate [ CNY : JPY ]:", ra)
	targetUnitAmount, err := paykit.ERInstance.Convert(ctx, 10000, paykit.CurrencyCNY, paykit.CurrencyJPY)
	if err != nil {
		t.Error("ConvertExchangeRate error", err.Error())
		return
	}
	t.Logf("result: [ %v : 10000] [ %v: %d]", paykit.CurrencyCNY, paykit.CurrencyJPY, targetUnitAmount)
}
