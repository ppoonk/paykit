package paykit

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	mainnet         = "https://api.trongrid.io/v1/accounts/{address}/transactions/trc20"
	nileTestnet     = "https://nile.trongrid.io/v1/accounts/{address}/transactions/trc20"
	trongridJobName = "Trongrid transactions"
	tronLogTag      = "[Tron]"
	tronLogPath     = "./.log/tron"
	tronLogLevel    = "error"
)

type TronConfig struct {
	Address      string        // 收款钱包地址
	APIKey       string        // Trongrid API密钥，用于监听网络交易。如果为空，则使用测试网
	AcceptTokens []TokenSymbol // 可接受的代币类型，如 USDT, USDC 等
	OrderTimeout int           // 订单超时时间(秒)
}
type trongridRes struct {
	Success bool   `json:"success"`
	Data    []data `json:"data"`
}

type data struct {
	TransactionID string `json:"transaction_id"`
	TokenInfo     struct {
		Symbol   string `json:"symbol"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		Name     string `json:"name"`
	} `json:"token_info"`
	BlockTimestamp int64  `json:"block_timestamp"`
	From           string `json:"from"`
	To             string `json:"to"`
	Type           string `json:"type"`
	Value          string `json:"value"`
}

type TronExtraForTradePreCreateReq struct {
	TokenSymbol TokenSymbol `json:"tokenSymbol"` // 当前订单指定的代币符号(USDT, USDC等)
}
type TronExtraForTradePreCreateRes struct {
	TotalAmountString string      `json:"totalAmountString"` // 加密货币价格，直接返回 2 位小数的支付金额，并提示用户严格按照该价格支付，否则无法履约订单
	TokenSymbol       TokenSymbol `json:"tokenSymbol"`       // 当前订单指定的代币符号(USDT, USDC等)
}

type (
	// 区块链网络类型
	Blockchain string
	// 代币类型
	Token string
	// TRC20代币具体类型
	TokenSymbol string
)

const (
	TRON     Blockchain = "TRON"
	BSC      Blockchain = "BSC"
	ETHEREUM Blockchain = "ETHEREUM"

	TRX   Token = "TRX"
	TRC10 Token = "TRC10"
	TRC20 Token = "TRC20"

	USDT_TRC20 TokenSymbol = "USDT"
	USDC_TRC20 TokenSymbol = "USDC"
)

type Tron struct {
	config          TronConfig
	logger          *glog.Logger
	httpClient      *gclient.Client
	cron            *gcron.Cron
	cache           *gcache.Cache
	fulfillCheckout func(string)
}

func NewTron(config TronConfig, fulfillCheckout func(string)) (*Tron, error) {
	// 设置日志
	l := glog.New()
	_ = l.SetPath(tronLogPath)
	_ = l.SetLevelStr(tronLogLevel)
	l.SetPrefix(tronLogTag)
	l.SetStack(false)

	return &Tron{
		config:          config,
		httpClient:      gclient.New(),
		cron:            gcron.New(),
		cache:           gcache.New(),
		fulfillCheckout: fulfillCheckout,
		logger:          l,
	}, nil
}

func (t *Tron) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (res *TradePreCreateRes, err error) {
	// 处理 tron 扩展参数
	ex := req.Extra.(TronExtraForTradePreCreateReq)
	if !slices.Contains(t.config.AcceptTokens, ex.TokenSymbol) {
		return nil, fmt.Errorf("token %s not supported", ex.TokenSymbol)
	}

	// 汇率转换，暂时用法币汇率 USD 代替。USD 返回值已经保留2位小数
	ta, err := ERInstance.ConvertToStandard(ctx, req.TotalAmount, req.Currency, CurrencyUSD)
	if err != nil {
		return nil, err
	}
	if ta <= 0 {
		ta = 0.01
	}

	// 原理类似于：https://github.com/assimon/epusdt
	for i := range 100 {
		if i == 100 {
			err = fmt.Errorf("amount %.2f failed to generate key", ta)
			return
		}
		if exist, _ := t.cache.Contains(ctx, ta); !exist {
			t.cache.Set(ctx, ta, req.OutTradeNo, 0) // TODO 缓存时间
			break
		}
		ta += 0.01
	}

	return &TradePreCreateRes{
		OutTradeNo: req.OutTradeNo,
		PayURL:     t.config.Address,
		Extra: TronExtraForTradePreCreateRes{
			TotalAmountString: fmt.Sprintf("%.2f", ta),
			TokenSymbol:       ex.TokenSymbol,
		},
	}, nil

}
func (t *Tron) Cacha() *gcache.Cache {
	return t.cache
}

func (t *Tron) StartCron() (err error) {
	ctx := gctx.New()
	_, err = t.cron.Add(ctx, "*/5 * * * * *", t.refresh, trongridJobName)
	t.refresh(ctx)
	return
}
func (t *Tron) refresh(ctx context.Context) {
	var (
		url    string
		res    *trongridRes
		params = map[string]any{
			"only_confirmed": true,
			"only_to":        true,
			"min_timestamp":  time.Now().UnixMilli() - 10*1000, // 10秒
			"max_timestamp":  time.Now().UnixMilli(),
		}
	)
	if t.config.APIKey == "" {
		url = strings.ReplaceAll(nileTestnet, "{address}", t.config.Address)
	} else {
		url = strings.ReplaceAll(mainnet, "{address}", t.config.Address)
	}

	err := t.httpClient.GetVar(ctx, url, params).Scan(&res)
	if err != nil {
		t.logger.Error(ctx, "refresh tron transactions error:", err.Error())
		return
	}
	t.logger.Debug(ctx, "refresh tron transactions, data lenght: ", len(res.Data))
	for _, v := range res.Data {
		amount, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			t.logger.Error(ctx, err.Error())
			continue
		}
		amount = amount / 1e6 // 单位为 0.01，和上面 USD 保持一致
		t.logger.Debugf(ctx, "refresh tron transactions, amount: %.2f", amount)

		va, err := t.cache.Get(ctx, amount)
		if err != nil {
			t.logger.Error(ctx, err.Error())
			continue
		}
		if va == nil {
			t.logger.Debugf(ctx, "refresh tron transactions, amount: %.2f, invalid order, continue", amount)
			continue
		}
		var outTradeNo string
		err = va.Scan(&outTradeNo)
		if err != nil {
			t.logger.Error(ctx, err.Error())
			continue
		}
		t.logger.Debugf(ctx, "refresh tron transactions, amount: %.2f, outTradeNo: %s", amount, outTradeNo)

		// 释放 key
		t.cache.Remove(ctx, v.Value)
		// 履约订单
		t.fulfillCheckout(outTradeNo)
	}
	return
}
