package order

import (
	"time"

	"github.com/google/uuid"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/fix50sp2/quote"
	"github.com/quickfixgo/fix50sp2/quotecancel"
	"github.com/quickfixgo/quickfix"
)

type QuoteHandler struct {
	symbol        string
	bidRefPrice   float64
	offerRefPrice float64
	side          enum.Side
	lastClOrdId   string
	account       string
	timestamp     time.Time
	messageType   string
}

func NewQuoteHandler(symbol string, bidPrice, offerPrice float64, account string) *QuoteHandler {
	return &QuoteHandler{
		symbol:        symbol,
		bidRefPrice:   bidPrice,
		offerRefPrice: offerPrice,
		side:          enum.Side_AS_DEFINED,
		account:       account,
	}
}

func (q *QuoteHandler) CanCancelAllOrdersInOnRequest() bool {
	return false
}

func (q *QuoteHandler) BuildAllCancelRequest(symbols []string) quickfix.Messagable {
	var quoteCancel quotecancel.QuoteCancel
	if len(symbols) == 0 {
		quoteCancel = quotecancel.New(field.NewQuoteCancelType(enum.QuoteCancelType_CANCEL_ALL_QUOTES))
	} else {
		quoteCancel = quotecancel.New(field.NewQuoteCancelType(enum.QuoteCancelType_CANCEL_FOR_ONE_OR_MORE_SECURITIES))
		quoteEntriesGroup := quotecancel.NewNoQuoteEntriesRepeatingGroup()
		for _, s := range symbols {
			quoteEntry := quoteEntriesGroup.Add()
			quoteEntry.Set(field.NewSymbol(s))
		}
		quoteCancel.SetNoQuoteEntries(quoteEntriesGroup)
	}
	quoteCancel.Set(field.NewQuoteID("cancel_all"))
	partyIdsGroup := quotecancel.NewNoPartyIDsRepeatingGroup()
	partyIds := partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID(q.account))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_CUSTOMER_ACCOUNT))
	partyIds = partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID("ATH"))
	partyIds.Set(field.NewPartyIDSource(enum.PartyIDSource_CHINESE_INVESTOR_ID))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_INVESTMENT_DECISION_MAKER))
	quoteCancel.SetNoPartyIDs(partyIdsGroup)
	return quoteCancel
}

func (q *QuoteHandler) BuildMassCancelRequest() quickfix.Messagable {
	return q.BuildAllCancelRequest(nil)
}

func (q *QuoteHandler) BuildOrderRequest() (quickfix.Messagable, string) {
	clOrdId := uuid.New().String()
	quoteMsg := quote.New(
		field.NewQuoteID(clOrdId),
	)
	quoteMsg.Set(field.NewSymbol(q.symbol))
	quoteMsg.Set(field.NewBidPx(generatePrice(q.bidRefPrice), 2))
	quoteMsg.Set(field.NewBidSize(generateOrderQuantity(), 0))
	quoteMsg.Set(field.NewOfferPx(generatePrice(q.offerRefPrice), 2))
	quoteMsg.Set(field.NewOfferSize(generateOrderQuantity(), 0))
	partyIdsGroup := quote.NewNoPartyIDsRepeatingGroup()
	partyIds := partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID(q.account))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_CUSTOMER_ACCOUNT))
	partyIds = partyIdsGroup.Add()
	partyIds.Set(field.NewPartyID("ATH"))
	partyIds.Set(field.NewPartyIDSource(enum.PartyIDSource_CHINESE_INVESTOR_ID))
	partyIds.Set(field.NewPartyRole(enum.PartyRole_INVESTMENT_DECISION_MAKER))
	quoteMsg.SetNoPartyIDs(partyIdsGroup)
	return quoteMsg, clOrdId
}

func (q *QuoteHandler) GetSymbol() string {
	return q.symbol
}

func (q *QuoteHandler) GetSide() enum.Side {
	return q.side
}

func (q *QuoteHandler) GetLastOrderId() string {
	return q.lastClOrdId
}

func (q *QuoteHandler) GetAccount() string {
	return q.account
}

func (q *QuoteHandler) GetTimestamp() time.Time {
	return q.timestamp
}

func (q *QuoteHandler) GetMessageType() string {
	return "Quote"
}

func (q *QuoteHandler) UpdateClientOrderId(newId string) {
	q.timestamp = time.Now()
	q.lastClOrdId = newId
}
