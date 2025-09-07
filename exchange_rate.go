package paykit

import (
	"context"
	"math"
	"slices"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	sourceApi      = "https://open.er-api.com/v6/latest/CNY"
	jobName        = "RefreshExchangeRateCronJob"
	logTag         = "[ExchangeRate]"
	defaultTimeout = 24 * time.Hour
)

var (
	ExchangeRateInstance *ExchangeRateClient // 全局单例
)

func init() {
	ExchangeRateInstance, err := NewExchangeRate(glog.New())
	if err != nil {
		panic(err)
	}
	err = ExchangeRateInstance.StartCron()
	if err != nil {
		panic(err)
	}
}

// exchangeRateResponse  汇率 api 响应结构体
type exchangeRateResponse struct {
	BaseCode string             `json:"base_code"`
	Rates    map[string]float64 `json:"rates"`
}

// ExchangeRateClient
type ExchangeRateClient struct {
	cron       *gcron.Cron
	logger     *glog.Logger
	httpClient *gclient.Client
	cache      *gcache.Cache
}

// NewExchangeRate
func NewExchangeRate(logger *glog.Logger) (client *ExchangeRateClient, err error) {
	// 设置日志
	logger.SetDebug(false)
	logger.SetPrefix(logger.GetConfig().Prefix + logTag)

	return &ExchangeRateClient{
		cron:       gcron.New(),
		logger:     logger,
		httpClient: gclient.New(),
		cache:      gcache.New(),
	}, nil

}
func (e *ExchangeRateClient) SetLogger(debug bool) {
	e.logger.SetDebug(debug)
}

// StartCron 启动定时任务，自动更新汇率数据并保存到数据库
func (e *ExchangeRateClient) StartCron() (err error) {
	ctx := gctx.New()
	_, err = e.cron.Add(ctx, "6 6 6 * * *", e.refresh, jobName) // 每天 06:06:06 执行
	e.refresh(ctx)                                              // 启动时更新一次
	return
}

// RoundType 取整类型
type RoundType int

const (
	RoundTypeDefault RoundType = iota // 默认四舍五入
	RoundTypeFloor                    // 向下取整
	RoundTypeCeil                     // 向上取整
)

// ConvertExchangeRate 执行货币最小单位金额的汇率转换
//
// 参数:
//
//	ctx          - 上下文对象，用于控制请求生命周期
//	unitAmount   - 原始货币最小单位金额（如CNY使用分，JPY使用元）
//	from         - 原始货币类型（Currency枚举值）
//	target       - 目标货币类型（Currency枚举值）
//	roundType    - 取整规则（RoundType枚举值）
//
// 返回:
//
//	targetUnitAmount - 转换后的目标货币最小单位金额
//	err             - 错误信息
func (e *ExchangeRateClient) ConvertExchangeRate(ctx context.Context, unitAmount int64, from, target Currency, roundType ...RoundType) (targetUnitAmount int64, err error) {
	if from == target {
		return unitAmount, nil
	}
	r, err := e.GetRate(ctx, from, target)
	if err != nil {
		return
	}
	if r == 0 {
		r = 1
	}
	var rt RoundType

	if len(roundType) > 0 {
		rt = roundType[0]
	}

	switch rt {
	case RoundTypeFloor:
		targetUnitAmount = int64(math.Floor(r * float64(unitAmount)))
	case RoundTypeCeil:
		targetUnitAmount = int64(math.Ceil(r * float64(unitAmount)))
	default:
		targetUnitAmount = int64(math.Round(r * float64(unitAmount)))
	}
	return
}

// ConvertExchangeRateToStandardUnit 执行货币最小单位金额的汇率转换, 返回标准货币单位
//
// 参数:
//
//	ctx          - 上下文对象，用于控制请求生命周期
//	unitAmount   - 原始货币最小单位金额（如CNY使用分，JPY使用元）
//	from         - 原始货币类型（Currency枚举值）
//	target       - 目标货币类型（Currency枚举值）
//	roundType    - 取整规则（RoundType枚举值）
//
// 返回:
//
//	amount - 标准货币单位
//	err    - 错误信息
func (e *ExchangeRateClient) ConvertExchangeRateToStandardUnit(ctx context.Context, unitAmount int64, from, target Currency, roundType ...RoundType) (amount float64, err error) {
	targetUnitAmount, err := e.ConvertExchangeRate(ctx, unitAmount, from, target, roundType...)
	if err != nil {
		return
	}
	return ToAmount(targetUnitAmount, target), nil

}
func (e *ExchangeRateClient) GetRate(ctx context.Context, from, target Currency) (rate float64, err error) {
	r1, err := e.getCache(ctx, string(from))
	if err != nil {
		return
	}
	r2, err := e.getCache(ctx, string(target))
	if err != nil {
		return
	}
	if r1 == 0 || r2 == 0 {
		return
	}
	return r2 / r1, nil
}
func (e *ExchangeRateClient) getCache(ctx context.Context, key string) (rate float64, err error) {
	ca, err := e.cache.Get(ctx, key)
	if err != nil {
		return
	}
	if ca == nil {
		return
	}
	var res *float64
	err = ca.Scan(&res)
	if err != nil || res == nil {
		return
	}
	return *res, nil
}
func (e *ExchangeRateClient) setCache(ctx context.Context, key string, rates float64) {
	e.cache.Set(ctx, key, rates, defaultTimeout)
}

func (e *ExchangeRateClient) refresh(ctx context.Context) {
	var res *exchangeRateResponse
	err := e.httpClient.Retry(2, 2*time.Second).GetVar(ctx, sourceApi).Scan(&res)
	if err != nil {
		e.logger.Error(ctx, "[get remote data error]:", err.Error())
		return
	}
	if res == nil {
		e.logger.Error(ctx, "[get remote data null]")
		return
	}
	// e.logger.Info(ctx, "remote exchange rate, CurrencyJPY:", res.Rates[string(CurrencyJPY)])
	// 存入缓存
	for k, v := range res.Rates {
		e.setCache(ctx, k, v)
	}

}

// ToUnitAmount 将标准货币单位转为最小货币单位
//
// 参数：
//
//	amount     - 标准货币单位金额（如10.99表示10元99分）
//	currency   - 三字母ISO货币代码（如USD/JPY）
//	roundType  - 舍入方式（Floor向下取整/Ceil向上取整/默认四舍五入）
//
// 返回值：
//
//	unitAmount - 最小货币单位金额（如USD返回1099表示10.99美元，JPY返回500表示500日元）
//
// 货币处理规则：
//
//	1.零小数货币（如JPY）直接返回整数金额
//	2.常规货币（如USD/EUR）需乘以100转换为分单位
//	3.特殊货币处理参考Stripe官方标准：
//	   - 零小数货币列表：BIF,CLP,DJF,GNF,JPY,KMF,KRW,MGA,PYG,RWF,UGX,VND,VUV,XAF,XOF,XPF
//	   - 其他货币默认视为两位小数货币
//
// 参考文档：
//
//	Stripe货币处理规范：https://docs.stripe.com/currencies#zero-decimal
func ToUnitAmount(amount float64, c Currency, roundType ...RoundType) (unitAmount int) {
	if !slices.Contains(ZeroDecimalCurrency, c) {
		amount = amount * 100
	}

	var rt RoundType
	if len(roundType) > 0 {
		rt = roundType[0]
	}

	switch rt {
	case RoundTypeFloor:
		unitAmount = int(math.Floor(amount))
	case RoundTypeCeil:
		unitAmount = int(math.Ceil(amount))
	default:
		unitAmount = int(math.Round(amount))
	}
	return
}

// ToAmount 将最小货币单位转为标准货币单位
//
// 参数：
//
//	unitAmount - 最小货币单位金额（如USD 1099 表示 10.99 美元，JPY 500 表示 500 日元）
//	currency   - 三字母ISO货币代码（如USD/JPY）
//	roundType  - 舍入方式（Floor向下取整/Ceil向上取整/默认四舍五入）
//
// 返回值：
//
//	amount     - 标准货币单位金额（如USD输入1099返回10.99，JPY输入500返回500）
func ToAmount(unitAmount int64, c Currency, roundType ...RoundType) (amount float64) {
	amount = float64(unitAmount)
	if !slices.Contains(ZeroDecimalCurrency, c) {
		amount = amount / 100
	}
	var rt RoundType
	if len(roundType) > 0 {
		rt = roundType[0]
	}
	switch rt {
	case RoundTypeFloor:
		amount = math.Floor(amount)
	case RoundTypeCeil:
		amount = math.Ceil(amount)
	default:
		amount = math.Round(amount)
	}
	return
}
