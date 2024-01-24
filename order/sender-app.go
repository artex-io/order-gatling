package order

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/fix50sp2/executionreport"
	"github.com/quickfixgo/fix50sp2/ordercancelreject"
	"github.com/quickfixgo/fix50sp2/ordermasscancelreport"
	"github.com/quickfixgo/fix50sp2/quotestatusreport"
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/tag"
	"github.com/rs/zerolog"
	"sylr.dev/fix/config"
	fixerrors "sylr.dev/fix/pkg/errors"
	"sylr.dev/fix/pkg/initiator"
	fixutils "sylr.dev/fix/pkg/utils"
)

type SenderApp struct {
	// Logger.
	fixutils.QuickFixAppMessageLogger

	// Quickfix settings for connection API Key.
	settings *quickfix.Settings

	// Quickfix initiator.
	initiator *quickfix.Initiator

	// Quickfix settings for trading API Key.
	sessionConfig *config.Session

	// Message router.
	*quickfix.MessageRouter

	// logonStatusChan is a chan raising connection status event.
	logonStatusChan chan bool

	// ExecReportNotification forwards received fix message to subscriber.
	ExecReportNotification chan executionreport.ExecutionReport

	// QuoteStatusReportNotification forwards received fix message to subscriber.
	QuoteStatusReportNotification chan quotestatusreport.QuoteStatusReport

	// sessionId is the session connected to the market to send orders.
	sessionId quickfix.SessionID

	isConnectionUp bool

	// Closed is a chan to notify when application is closed properly.
	Closed chan bool

	isStopping atomic.Bool
}

var (
	_ quickfix.Application = (*SenderApp)(nil)
)

// NewOrderSender creates an Application which implements quickfix.Application.
func NewOrderSender(
	ctx context.Context,
	quickFixAppMessageLogger fixutils.QuickFixAppMessageLogger,
	settings *quickfix.Settings,
	sessionConfig *config.Session) (*SenderApp, error) {
	app := SenderApp{
		QuickFixAppMessageLogger:      quickFixAppMessageLogger,
		MessageRouter:                 quickfix.NewMessageRouter(),
		settings:                      settings,
		sessionConfig:                 sessionConfig,
		logonStatusChan:               make(chan bool),
		ExecReportNotification:        make(chan executionreport.ExecutionReport, 10),
		QuoteStatusReportNotification: make(chan quotestatusreport.QuoteStatusReport, 10),
		isConnectionUp:                false,
		Closed:                        make(chan bool),
		isStopping:                    atomic.Bool{},
	}

	app.MessageRouter.AddRoute(executionreport.Route(app.onExecutionReport))
	app.MessageRouter.AddRoute(quotestatusreport.Route(app.OnQuoteStatusReport))
	app.MessageRouter.AddRoute(ordercancelreject.Route(app.onOrderCancelReject))
	app.MessageRouter.AddRoute(ordermasscancelreport.Route(app.onOrderMassCancelReport))

	go app.handleContextDone(ctx)

	return &app, nil
}

func (a *SenderApp) handleContextDone(ctx context.Context) {
	<-ctx.Done()
	a.isStopping.Store(true)
	if a.initiator != nil {
		a.initiator.Stop()
	}
	close(a.logonStatusChan)
	close(a.ExecReportNotification)
	close(a.QuoteStatusReportNotification)
	a.Closed <- true
}

// OnCreate is called when a session is created. Note that sessions are created
// upon initiator/acceptor start and not when a connection is established.
func (a *SenderApp) OnCreate(sessionID quickfix.SessionID) {
	a.Logger.Debug().Str("session", sessionID.String()).Msg("Created")
}

// OnLogon is called when a FIX logon occurs.
func (a *SenderApp) OnLogon(sessionID quickfix.SessionID) {
	a.Logger.Info().Str("session", sessionID.String()).Msg("Successful Logon")
	a.sessionId = sessionID
	a.logonStatusChan <- true
}

// OnLogout is called when a FIX logout occurs.
func (a *SenderApp) OnLogout(sessionID quickfix.SessionID) {
	a.Logger.Info().Str("session", sessionID.String()).Msg("Successful Logout")
	if !a.isStopping.Load() {
		a.logonStatusChan <- false
	}
	a.Logger.Debug().Str("session", sessionID.String()).Msg("End of OnLogout")
}

// ToAdmin is called when sending a FIX message regarding the FIX protocol, e.g.:
// LOGIN, LOGOUT, HEARTBEAT, TEST ... etc.
func (a *SenderApp) ToAdmin(message *quickfix.Message, sessionID quickfix.SessionID) {
	a.LogMessage(zerolog.TraceLevel, message, sessionID, true)
	a.Logger.Debug().Msg("-> Sending message to admin")

	typ, err := message.MsgType()
	if err != nil {
		a.Logger.Error().Msgf("Message type error: %s", err)
	}

	// Logon
	if err == nil && typ == string(enum.MsgType_LOGON) {
		sets := a.settings.SessionSettings()
		if session, ok := sets[sessionID]; ok {
			if session.HasSetting("Username") {
				username, err := session.Setting("Username")
				if err == nil && len(username) > 0 {
					a.Logger.Debug().Msg("Username injected in logon message")
					message.Header.SetField(tag.Username, quickfix.FIXString(username))
				}
			}
			if session.HasSetting("Password") {
				password, err := session.Setting("Password")
				if err == nil && len(password) > 0 {
					a.Logger.Debug().Msg("Password injected in logon message")
					message.Header.SetField(tag.Password, quickfix.FIXString(password))
				}
			}
		}
	}
}

// FromAdmin is called when receiving a FIX message regarding the FIX protocol, e.g.:
// LOGIN, LOGOUT, HEARTBEAT, TEST ... etc.
func (a *SenderApp) FromAdmin(message *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	a.LogMessage(zerolog.TraceLevel, message, sessionID, false)

	return nil
}

// ToApp is called when sending a FIX message that is not considered "Admin".
func (a *SenderApp) ToApp(message *quickfix.Message, sessionID quickfix.SessionID) error {
	a.LogMessage(zerolog.TraceLevel, message, sessionID, true)

	return nil
}

// FromApp is called when receiving a FIX message that is not considered "Admin".
func (a *SenderApp) FromApp(message *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	a.LogMessage(zerolog.TraceLevel, message, sessionID, false)
	return a.MessageRouter.Route(message, sessionID)
}

func (a *SenderApp) onExecutionReport(msg executionreport.ExecutionReport, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	a.ExecReportNotification <- msg
	return nil
}

func (a *SenderApp) OnQuoteStatusReport(msg quotestatusreport.QuoteStatusReport, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	a.QuoteStatusReportNotification <- msg
	return nil
}

func (a *SenderApp) onOrderMassCancelReport(msg ordermasscancelreport.OrderMassCancelReport, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	clOrdId, err := msg.GetClOrdID()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Field ClOrdID not set")
		return err
	}
	rsp, err := msg.GetMassCancelResponse()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Field MassCancelResponse not set")
		return err
	}
	txt, err := msg.GetText()
	if err != nil {
		txt = "no reason"
	}
	switch rsp {
	case enum.MassCancelResponse_CANCEL_ORDERS_FOR_A_SECURITY:
		a.Logger.Info().Str("clOrdId", clOrdId).Msg("OrderMassCancelRequest accepted")
	case enum.MassCancelResponse_CANCEL_REQUEST_REJECTED:
		a.Logger.Error().Str("clOrdId", clOrdId).Str("reason", txt).Msg("OrderMassCancelRequest rejected")
	default:
		a.Logger.Error().Any("value", rsp).Str("clOrdId", clOrdId).Str("reason", txt).Msg("OrderMassCancelResponse invalid")
	}
	return nil
}

func (a *SenderApp) onOrderCancelReject(msg ordercancelreject.OrderCancelReject, id quickfix.SessionID) quickfix.MessageRejectError {
	clOrdId, err := msg.GetClOrdID()
	if err != nil {
		a.Logger.Error().Err(err).Msg("Field ClOrdID not set")
		return err
	}
	reason, err := msg.GetText()
	if err != nil {
		reason = "No exchange reason"
	}
	a.Logger.Warn().Str("clOrdId", clOrdId).Str("text", reason).Msg("OrderCancelReject received")
	return nil
}

func (a *SenderApp) Send(message quickfix.Messagable) error {
	if !a.isConnectionUp {
		return errors.New("order fix session is logged out")
	}
	return quickfix.SendToTarget(message, a.sessionId)
}

func (a *SenderApp) Connect() error {
	opt := config.GetOptions()
	var quickfixLogger *zerolog.Logger
	if opt.QuickFixLogging {
		quickfixLogger = a.Logger
	}
	var err error
	a.initiator, err = initiator.Initiate(a, a.settings, quickfixLogger)
	if err != nil {
		return fmt.Errorf("unable to create order initiator: %s", err)
	}

	err = a.initiator.Start()
	if err != nil {
		return fmt.Errorf("unable to start order initiator: %s", err)
	}

	// Wait for session connection
	select {
	case <-time.After(30 * time.Second):
		return errors.New("cannot connect to FIX order acceptor")
	case status, ok := <-a.logonStatusChan:
		if !ok {
			return fixerrors.FixLogout
		}
		a.isConnectionUp = status
	}
	go func() {
		for {
			a.isConnectionUp = <-a.logonStatusChan
		}
	}()

	return nil
}
