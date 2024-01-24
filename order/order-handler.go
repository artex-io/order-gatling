package order

import (
	"time"

	"github.com/google/uuid"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/fix50sp2/ordercancelreplacerequest"
	"github.com/quickfixgo/quickfix"
)

type OrderHandler struct {
	symbol      string
	refPrice    float64
	side        enum.Side
	lastClOrdId string
	account     string
	timestamp   time.Time
	messageType string
}

func NewOrderHandler(symbol string, price float64, side enum.Side, account string) *OrderHandler {
	return &OrderHandler{
		symbol:   symbol,
		refPrice: price,
		side:     side,
		account:  account,
	}
}

func (o *OrderHandler) CanCancelAllOrdersInOnRequest() bool {
	return false
}

func (o *OrderHandler) BuildAllCancelRequest(symbols []string) quickfix.Messagable {
	return nil
}

func (o *OrderHandler) BuildMassCancelRequest() quickfix.Messagable {
	return BuildMassCancelRequest(o.side, o.symbol, o.account)
}

func (o *OrderHandler) BuildOrderRequest() (quickfix.Messagable, string) {
	if len(o.lastClOrdId) > 0 {
		return o.buildOrderCancelReplaceRequest()
	} else {
		return o.buildNewOrderSingle()
	}
}

func (o *OrderHandler) buildNewOrderSingle() (quickfix.Messagable, string) {
	return buildNewOrderSingle(o.side, o.refPrice, o.symbol, o.account)
}

func (o *OrderHandler) buildOrderCancelReplaceRequest() (quickfix.Messagable, string) {
	clOrdId := uuid.New().String()
	order := ordercancelreplacerequest.New(
		field.NewClOrdID(clOrdId),
		field.NewSide(o.side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT),
	)
	order.Set(field.NewOrigClOrdID(o.lastClOrdId))
	order.Set(field.NewOrderQty(generateOrderQuantity(), 0))
	order.Set(field.NewPrice(generatePrice(o.refPrice), 2))
	order.Set(field.NewSymbol(o.symbol))
	order.Set(field.NewTimeInForce(enum.TimeInForce_DAY))
	partyIdsGroup := ordercancelreplacerequest.NewNoPartyIDsRepeatingGroup()
	partyIds := partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID(o.account))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_CUSTOMER_ACCOUNT))
	partyIds = partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID("ATH"))
	partyIds.Set(field.NewPartyIDSource(enum.PartyIDSource_CHINESE_INVESTOR_ID))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_INVESTMENT_DECISION_MAKER))
	order.SetNoPartyIDs(partyIdsGroup)
	return order, clOrdId
}

func (o *OrderHandler) GetSymbol() string {
	return o.symbol
}

func (o *OrderHandler) GetSide() enum.Side {
	return o.side
}

func (o *OrderHandler) GetLastOrderId() string {
	return o.lastClOrdId
}

func (o *OrderHandler) GetAccount() string {
	return o.account
}

func (o *OrderHandler) GetTimestamp() time.Time {
	return o.timestamp
}

func (o *OrderHandler) GetMessageType() string {
	return o.messageType
}

func (o *OrderHandler) UpdateClientOrderId(newId string) {
	if len(o.lastClOrdId) > 0 {
		o.messageType = "OrderCancelReplaceRequest"
	} else {
		o.messageType = "NewOrderSingle"
	}
	o.timestamp = time.Now()
	o.lastClOrdId = newId
}
