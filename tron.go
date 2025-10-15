package paykit

import (
	"context"
	"fmt"
	"net/http"
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

var _ PaymentInterface = (*TronClient)(nil)

const (
	_MAINNET           = "https://api.trongrid.io/v1/accounts/{address}/transactions/trc20"
	_NILE_TESTNET      = "https://nile.trongrid.io/v1/accounts/{address}/transactions/trc20"
	_TRONGRID_JOB_NAME = "Trongrid transactions"
	_TRON_LOG_TAG      = "[Tron]"
	_TRON_LOG_PATH     = "./.log/tron"
	_TRON_LOG_LEVEL    = "error"
)

type TokenSymbol string

const (
	USDT_TRC20 TokenSymbol = "USDT"
	USDC_TRC20 TokenSymbol = "USDC"
)

type TronConfig struct {
	PaymentKey   any
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

type TronClient struct {
	config          TronConfig
	logger          *glog.Logger
	httpClient      *gclient.Client
	cron            *gcron.Cron
	cache           *gcache.Cache
	fulfillCheckout func(string)
}

func NewTron(config TronConfig, fulfillCheckout func(string)) (*TronClient, error) {
	// 设置日志
	l := glog.New()
	_ = l.SetPath(_TRON_LOG_PATH)
	_ = l.SetLevelStr(_TRON_LOG_LEVEL)
	l.SetPrefix(_TRON_LOG_TAG)
	l.SetStack(false)

	return &TronClient{
		config:          config,
		httpClient:      gclient.New(),
		cron:            gcron.New(),
		cache:           gcache.New(),
		fulfillCheckout: fulfillCheckout,
		logger:          l,
	}, nil
}

func (t *TronClient) TradePrecreate(ctx context.Context, req *TradePreCreateReq) (res *TradePreCreateRes, err error) {
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
	// 缓存：key=价格 value=订单号
	for i := range 100 {
		if i == 100 {
			err = fmt.Errorf("amount %.2f failed to generate key", ta)
			return
		}
		if exist, _ := t.cache.Contains(ctx, ta); !exist {
			t.cache.Set(ctx, ta, req.OutTradeNo, time.Minute*30) // TODO 缓存时间：30分钟
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
func (t *TronClient) Cacha() *gcache.Cache {
	return t.cache
}

func (t *TronClient) Start() (err error) {
	ctx := gctx.New()
	_, err = t.cron.Add(ctx, "*/5 * * * * *", t.refresh, _TRONGRID_JOB_NAME) // 每 5 秒执行一次
	t.refresh(ctx)
	return
}
func (t *TronClient) Stop() {

}
func (t *TronClient) refresh(ctx context.Context) {
	var (
		url    string
		res    *trongridRes
		params = map[string]any{
			"only_confirmed": true,
			"only_to":        true,
			"min_timestamp":  time.Now().UnixMilli() - 10*1000, // 拉取 10秒 之前的数据
			"max_timestamp":  time.Now().UnixMilli(),
		}
	)
	if t.config.APIKey == "" {
		url = strings.ReplaceAll(_NILE_TESTNET, "{address}", t.config.Address)
	} else {
		url = strings.ReplaceAll(_MAINNET, "{address}", t.config.Address)
	}

	err := t.httpClient.GetVar(ctx, url, params).Scan(&res)
	if err != nil {
		t.logger.Error(ctx, "refresh tron transactions error:", err.Error())
		return
	}
	t.logger.Debug(ctx, "refresh tron transactions, data lenght: ", len(res.Data))
	for _, v := range res.Data {
		// api 接口返回的金额为 TRC20 代币的最小单位
		amount, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			t.logger.Error(ctx, err.Error())
			continue
		}
		// 目前只支持 USDT_TRC20，USDC_TRC20
		// TRC20 代币的最小单位值需要除以100万才能得到标准单位： 1 USDT = 1,000,000 最小单位
		// 精度为 0.01，和上面 TradePrecreate 时 最终的价格（key）保持一致
		key := amount / 1e6
		t.logger.Debugf(ctx, "refresh tron transactions, amount(key): %.2f", key)

		va, err := t.cache.Get(ctx, key)
		if err != nil {
			t.logger.Error(ctx, err.Error())
			continue
		}
		if va == nil {
			t.logger.Debugf(ctx, "refresh tron transactions, amount(key): %.2f, invalid order, continue", amount)
			continue
		}
		var outTradeNo string
		err = va.Scan(&outTradeNo)
		if err != nil {
			t.logger.Error(ctx, err.Error())
			continue
		}
		t.logger.Debugf(ctx, "refresh tron transactions, amount(key): %.2f, outTradeNo: %s", amount, outTradeNo)

		// 释放 key
		t.cache.Remove(ctx, v.Value)
		// 履约订单
		t.fulfillCheckout(outTradeNo)
	}
	return
}
func (t *TronClient) Notify(http.ResponseWriter, *http.Request) {

}
func (t *TronClient) PaymentKey() any {
	return t.config.PaymentKey
}
func (t *TronClient) PaymentType() PaymentType {
	return PAYMENT_TYPE_TRON
}
