package order

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/fix50sp2/newordersingle"
	"github.com/quickfixgo/fix50sp2/ordermasscancelrequest"
	"github.com/quickfixgo/quickfix"
	"github.com/shopspring/decimal"
)

type Handler interface {
	GetSymbol() string
	GetSide() enum.Side
	GetLastOrderId() string
	GetAccount() string
	GetTimestamp() time.Time
	GetMessageType() string

	UpdateClientOrderId(newId string)

	CanCancelAllOrdersInOnRequest() bool
	BuildAllCancelRequest(symbols []string) quickfix.Messagable
	BuildMassCancelRequest() quickfix.Messagable
	BuildOrderRequest() (quickfix.Messagable, string)
}

func generateOrderQuantity() decimal.Decimal {
	qty := 90 + rand.Intn(20)
	return decimal.NewFromInt(int64(qty))
}

func generatePrice(refPrice float64) decimal.Decimal {
	p := refPrice - 0.05 + rand.Float64()*0.10
	return decimal.NewFromFloat(p)
}

func buildNewOrderSingle(side enum.Side, price float64, symbol string, account string) (quickfix.Messagable, string) {
	clOrdId := uuid.New().String()
	order := newordersingle.New(
		field.NewClOrdID(clOrdId),
		field.NewSide(side),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT),
	)
	order.Set(field.NewOrderQty(generateOrderQuantity(), 0))
	order.Set(field.NewPrice(generatePrice(price), 2))
	order.Set(field.NewSymbol(symbol))
	order.Set(field.NewTimeInForce(enum.TimeInForce_DAY))
	partyIdsGroup := newordersingle.NewNoPartyIDsRepeatingGroup()
	partyIds := partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID(account))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_CUSTOMER_ACCOUNT))
	partyIds = partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID("ATH"))
	partyIds.Set(field.NewPartyIDSource(enum.PartyIDSource_CHINESE_INVESTOR_ID))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_INVESTMENT_DECISION_MAKER))
	order.SetNoPartyIDs(partyIdsGroup)
	return order, clOrdId
}

func BuildMassCancelRequest(side enum.Side, symbol string, account string) quickfix.Messagable {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%s%v%s", symbol, side, account)))
	massCancel := ordermasscancelrequest.New(
		field.NewClOrdID(hex.EncodeToString(h.Sum(nil))),
		field.NewMassCancelRequestType(enum.MassCancelRequestType_CANCEL_ORDERS_FOR_A_SECURITY),
		field.NewTransactTime(time.Now()),
	)
	massCancel.Set(field.NewSide(side))
	massCancel.Set(field.NewSymbol(symbol))
	partyIdsGroup := ordermasscancelrequest.NewNoPartyIDsRepeatingGroup()
	partyIds := partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID(account))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_CUSTOMER_ACCOUNT))
	partyIds = partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID("ATH"))
	partyIds.Set(field.NewPartyIDSource(enum.PartyIDSource_CHINESE_INVESTOR_ID))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_INVESTMENT_DECISION_MAKER))
	massCancel.SetNoPartyIDs(partyIdsGroup)
	return massCancel
}
