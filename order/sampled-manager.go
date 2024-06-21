package order

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/fix50sp2/executionreport"
	"github.com/quickfixgo/quickfix"
)

type SampledManager struct {
	context            context.Context
	app                *SenderApp
	accounts           []string
	symbols            []string
	refPrices          []float64
	nbOrderPerSec      uint
	ordersTimestampMap map[string]time.Time
	orderLock          sync.Mutex
	Closed             chan bool
}

func NewSampledManager(
	context context.Context,
	app *SenderApp,
	accounts []string,
	symbols []string,
	refPrices []float64,
	nbOrderPerSec uint) *SampledManager {
	mgr := &SampledManager{
		context:            context,
		app:                app,
		accounts:           accounts,
		symbols:            symbols,
		refPrices:          refPrices,
		nbOrderPerSec:      nbOrderPerSec,
		ordersTimestampMap: make(map[string]time.Time),
		orderLock:          sync.Mutex{},
		Closed:             make(chan bool),
	}

	go mgr.processExecutionReports()
	return mgr
}

func (m *SampledManager) getOrderTimestamp(id string) (time.Time, bool) {
	m.orderLock.Lock()
	defer m.orderLock.Unlock()
	ts, found := m.ordersTimestampMap[id]
	if found {
		delete(m.ordersTimestampMap, id)
	}
	return ts, found
}

func (m *SampledManager) setOrderTimestamp(id string) {
	m.orderLock.Lock()
	defer m.orderLock.Unlock()
	m.ordersTimestampMap[id] = time.Now()
}

func (m *SampledManager) CancelAllOrders() {
	for _, account := range m.accounts {
		for _, symbol := range m.symbols {
			massCancel := BuildMassCancelRequest(enum.Side_BUY, symbol, account)
			err := quickfix.SendToTarget(massCancel, m.app.sessionId)
			if err != nil {
				m.app.Logger.Err(err).Str("account", account).Str("symbol", symbol).Any("side", "buy").Msg("Cannot send mass cancel request")
			}
			massCancel = BuildMassCancelRequest(enum.Side_SELL, symbol, account)
			err = quickfix.SendToTarget(massCancel, m.app.sessionId)
			if err != nil {
				m.app.Logger.Err(err).Str("account", account).Str("symbol", symbol).Any("side", "sell").Msg("Cannot send mass cancel request")
			}
		}
	}
}

func (m *SampledManager) Start() {
	interval := 1000000 / (m.nbOrderPerSec / 2) // /2 because we send 2 orders per iteration
	tick := time.After(time.Duration(interval) * time.Microsecond)
	go func() {

		for {
			select {
			case <-tick:
				err := m.sendTrade()
				if err != nil {
					m.app.Logger.Err(err).Msg("Stopping order sending routine")
					return
				}
				tick = time.After(time.Duration(interval) * time.Microsecond)

			case <-m.context.Done():
				m.app.Logger.Error().Err(m.context.Err()).Msg("Sampled order manager creation routine is stopping")
				m.Closed <- true
				return
			}
		}
	}()
}

func (m *SampledManager) sendTrade() error {
	id := rand.Intn(len(m.refPrices))
	refPrice := m.refPrices[id]
	symbol := m.symbols[id]
	err := m.sendOrderRequest(enum.Side_BUY, symbol, refPrice)
	if err != nil {
		return err
	}
	return m.sendOrderRequest(enum.Side_SELL, symbol, refPrice)
}

func (m *SampledManager) sendOrderRequest(side enum.Side, symbol string, refPrice float64) error {

	account := m.accounts[rand.Intn(len(m.accounts))]
	var order quickfix.Messagable
	var clOrdId string
	order, clOrdId = buildNewOrderSingle(side, refPrice, symbol, account)
	m.setOrderTimestamp(clOrdId)
	err := quickfix.SendToTarget(order, m.app.sessionId)
	if err != nil {
		m.app.Logger.Err(err).Str("account", account).Str("symbol", symbol).Msg("Cannot send new order single request")
		return errors.New("cannot send new order single request")
	}
	return nil
}

func (m *SampledManager) processExecutionReports() {
LOOP:
	for {
		select {
		case msg, ok := <-m.app.ExecReportNotification:
			if !ok {
				break LOOP
			}
			if err := m.processExecutionReport(msg); err != nil {
				m.app.Logger.Error().Err(err).Any("msg", strings.Replace(msg.Message.String(), "\001", "|", -1)).Msg("Cannot process fix message")
			}

		case <-m.context.Done():
			m.app.Logger.Error().Err(m.context.Err()).Msg("Sampled order manager is stopping")
			m.Closed <- true
			return
		}
	}
}

func (m *SampledManager) processExecutionReport(execReport executionreport.ExecutionReport) error {
	clOrdId, err := execReport.GetClOrdID()
	if err != nil {
		return errors.New("missing ClOrdID in ExecutionReport")
	}
	ts, found := m.getOrderTimestamp(clOrdId)
	if !found {
		m.app.Logger.Trace().Str("clOrdId", clOrdId).Msg("Order not found")
		return nil
	}
	metricOrderRoundtrip.WithLabelValues("NewOrderSingle").Observe(time.Since(ts).Seconds())
	status, err := execReport.GetOrdStatus()
	if err != nil {
		return errors.New("missing OrdStatus in ExecutionReport")
	}
	switch status {
	case enum.OrdStatus_NEW:
		return nil
	case enum.OrdStatus_PARTIALLY_FILLED:
		return nil
	case enum.OrdStatus_FILLED:
		return nil
	case enum.OrdStatus_REJECTED:
		return nil
	default:
		return fmt.Errorf("order status not handled: %v", status)
	}
}
