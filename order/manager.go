package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
	"github.com/sylr/quickfixgo-fix50sp2/executionreport"
	"github.com/sylr/quickfixgo-fix50sp2/quotestatusreport"
)

var (
	metricOrderRoundtrip = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem: "order_gatling",
			Name:      "fix_roundtrip_duration_seconds_summary",
			Help:      "Fix requests roundtrip duration",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.05,
				0.95: 0.01,
				0.99: 0.005,
			},
		},
		[]string{"type"},
	)
)

type Manager struct {
	context     context.Context
	app         *SenderApp
	orders      []Handler
	updateTempo time.Duration
	ordersMap   map[string]Handler
	orderLock   sync.Mutex
	Closed      chan bool
}

func init() {
	prometheus.MustRegister(metricOrderRoundtrip)
}

func NewManager(
	context context.Context,
	app *SenderApp,
	accounts []string,
	symbols []string,
	refPrices []float64,
	useQuoteWorkflow bool,
	updateTempo time.Duration) *Manager {
	mgr := &Manager{
		context:     context,
		app:         app,
		orders:      make([]Handler, 0, len(accounts)*2*len(symbols)),
		updateTempo: updateTempo,
		ordersMap:   make(map[string]Handler, len(accounts)*2*len(symbols)),
		orderLock:   sync.Mutex{},
		Closed:      make(chan bool),
	}

	for idx, symbol := range symbols {
		for i, account := range accounts {
			offset := 0.10 + 0.01*float64(i)
			if useQuoteWorkflow {
				mgr.orders = append(
					mgr.orders,
					NewQuoteHandler(symbol, refPrices[idx]-offset, refPrices[idx]+offset, account),
				)
			} else {
				mgr.orders = append(
					mgr.orders,
					NewOrderHandler(symbol, refPrices[idx]-offset, enum.Side_BUY, account),
					NewOrderHandler(symbol, refPrices[idx]+offset, enum.Side_SELL, account),
				)
			}
		}
	}

	go mgr.processExecutionReports()
	return mgr
}
func (m *Manager) getOrder(id string) (Handler, bool) {
	m.orderLock.Lock()
	defer m.orderLock.Unlock()
	order, found := m.ordersMap[id]
	return order, found
}

func (m *Manager) updateClientOrderId(newId string, o Handler) {
	m.orderLock.Lock()
	defer m.orderLock.Unlock()
	if len(o.GetLastOrderId()) > 0 {
		delete(m.ordersMap, o.GetLastOrderId())
	}
	o.UpdateClientOrderId(newId)
	m.ordersMap[newId] = o
}

func (m *Manager) CancelAllOrders() {
	for _, order := range m.orders {
		massCancel := order.BuildMassCancelRequest()
		err := quickfix.SendToTarget(massCancel, m.app.sessionId)
		if err != nil {
			m.app.Logger.Err(err).Str("account", order.GetAccount()).Str("symbol", order.GetSymbol()).Any("side", order.GetSide()).Msg("Cannot send mass cancel request")
		}
	}
}

func (m *Manager) Start() {
	for _, order := range m.orders {
		_ = m.sendOrderRequest(order)
	}
}

func (m *Manager) sendOrderRequest(order Handler) error {
	nos, orderId := order.BuildOrderRequest()
	m.updateClientOrderId(orderId, order)
	err := quickfix.SendToTarget(nos, m.app.sessionId)
	m.app.Logger.Debug().Str("clordid", orderId).Msg("New order single sent")
	if err != nil {
		m.app.Logger.Err(err).Str("account", order.GetAccount()).Str("symbol", order.GetSymbol()).Any("side", order.GetSide()).Msg("Cannot send new order single")
		return err
	}
	return nil
}

func (m *Manager) sendMessage(order Handler, sendMessageFunc func(Handler) error) error {
	if m.updateTempo <= 0 {
		return sendMessageFunc(order)
	}
	go func() {
		time.Sleep(m.updateTempo)
		_ = sendMessageFunc(order)
	}()
	return nil
}

func (m *Manager) processExecutionReports() {
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

		case msg, ok := <-m.app.QuoteStatusReportNotification:
			if !ok {
				break LOOP
			}
			if err := m.processQuoteStatusReport(msg); err != nil {
				m.app.Logger.Error().Err(err).Any("msg", strings.Replace(msg.Message.String(), "\001", "|", -1)).Msg("Cannot process fix message")
			}

		case <-m.context.Done():
			m.app.Logger.Error().Err(m.context.Err()).Msg("Order manager is stopping")
			m.Closed <- true
			return
		}
	}
}

func (m *Manager) processExecutionReport(execReport executionreport.ExecutionReport) error {
	clOrdId, err := execReport.GetClOrdID()
	if err != nil {
		return errors.New("missing ClOrdID in ExecutionReport")
	}
	order, found := m.getOrder(clOrdId)
	if !found {
		m.app.Logger.Trace().Str("clOrdId", clOrdId).Msg("Order not found")
		return nil
	}
	metricOrderRoundtrip.WithLabelValues(order.GetMessageType()).Observe(time.Since(order.GetTimestamp()).Seconds())
	status, err := execReport.GetOrdStatus()
	if err != nil {
		return errors.New("missing OrdStatus in ExecutionReport")
	}
	switch status {
	case enum.OrdStatus_NEW:
		fallthrough
	case enum.OrdStatus_REPLACED:
		fallthrough
	case enum.OrdStatus_PARTIALLY_FILLED:
		return m.sendMessage(order, m.sendOrderRequest)
	case enum.OrdStatus_FILLED:
		m.updateClientOrderId("", order)
		return m.sendMessage(order, m.sendOrderRequest)
	default:
		return fmt.Errorf("order status not handled: %v", status)
	}
}

func (m *Manager) processQuoteStatusReport(qsReport quotestatusreport.QuoteStatusReport) error {
	quoteId, err := qsReport.GetQuoteID()
	if err != nil {
		return errors.New("missing QuoteID in QuoteStatusReport")
	}
	order, found := m.getOrder(quoteId)
	if !found {
		m.app.Logger.Trace().Str("quoteId", quoteId).Msg("Quote not found")
		return nil
	}
	metricOrderRoundtrip.WithLabelValues(order.GetMessageType()).Observe(time.Since(order.GetTimestamp()).Seconds())
	status, err := qsReport.GetQuoteStatus()
	if err != nil {
		return errors.New("missing QuoteStatus in QuoteStatusReport")
	}
	switch status {
	case enum.QuoteStatus_ACCEPTED:
		_, err := qsReport.GetBidQuoteID()
		if err != nil {
			return errors.New("missing BidQuoteID")
		}
		_, err = qsReport.GetOfferQuoteID()
		if err != nil {
			return errors.New("missing OfferQuoteID")
		}
		return m.sendMessage(order, m.sendOrderRequest)
	case enum.QuoteStatus_REJECTED:
		fallthrough
	case enum.QuoteStatus_CANCELED:
		return fmt.Errorf("quote %s is rejected(%v)", quoteId, status)
	default:
		return fmt.Errorf("quote status not handled: %v", status)
	}
}
