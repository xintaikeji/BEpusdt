package task

import (
	"context"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/v03413/bepusdt/app/conf"
	"github.com/v03413/bepusdt/app/model"
)

type task struct {
	duration time.Duration
	callback func(ctx context.Context)
}

var (
	tasks []task
	mu    sync.Mutex
)

func Init() error {
	bscInit()
	ethInit()
	polygonInit()
	arbitrumInit()
	plasmaInit()
	xlayerInit()
	baseInit()

	return nil
}

func register(t task) {
	mu.Lock()
	defer mu.Unlock()

	if t.callback == nil {

		panic("task callback cannot be nil")
	}

	tasks = append(tasks, t)
}

// 按 tradeType 参数读取限额范围
func inAmountRange(payAmount decimal.Decimal, tradeType string) bool {
	var min, max decimal.Decimal
	
	switch tradeType {
	case model.OrderTradeTypeEthErc20:
		min = conf.GetPaymentAmountEthMin()
		max = conf.GetPaymentAmountEthMax()
	case model.OrderTradeTypeBnbBep20:
		min = conf.GetPaymentAmountBnbMin()
		max = conf.GetPaymentAmountBnbMax()
	default:
		min = conf.GetPaymentAmountMin()
		max = conf.GetPaymentAmountMax()
	}

	if payAmount.GreaterThan(max) {
		return false
	}

	if payAmount.LessThan(min) {
		return false
	}

	return true
}

func Start(ctx context.Context) {
	mu.Lock()
	defer mu.Unlock()

	for _, t := range tasks {
		go func(t task) {
			if t.duration <= 0 {
				t.callback(ctx)

				return
			}

			t.callback(ctx)

			ticker := time.NewTicker(t.duration)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					t.callback(ctx)
				}
			}
		}(t)
	}
}
