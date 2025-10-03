package web3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type EventFilter struct {
	FromBlock *big.Int
	ToBlock   *big.Int
	Address   []string
	Topics    [][]string
}

type EventSubscription struct {
	ID        string
	Filter    *EventFilter
	Channel   chan Event
	Active    bool
	CreatedAt time.Time
}

type Event struct {
	Address          string
	Topics           []string
	Data             string
	BlockNumber      *big.Int
	TransactionHash  string
	TransactionIndex uint
	BlockHash        string
	LogIndex         uint
	Removed          bool
}

type EventHandler func(event Event) error

func NewEventFilter() *EventFilter {
	return &EventFilter{
		Address: make([]string, 0),
		Topics:  make([][]string, 0),
	}
}

func (f *EventFilter) SetFromBlock(blockNumber *big.Int) *EventFilter {
	f.FromBlock = blockNumber
	return f
}

func (f *EventFilter) SetToBlock(blockNumber *big.Int) *EventFilter {
	f.ToBlock = blockNumber
	return f
}

func (f *EventFilter) SetLatestBlock() *EventFilter {
	f.ToBlock = big.NewInt(-1)
	return f
}

func (f *EventFilter) SetPendingBlock() *EventFilter {
	f.ToBlock = big.NewInt(-2)
	return f
}

func (f *EventFilter) AddAddress(address string) *EventFilter {
	if ValidateAddress(address) {
		f.Address = append(f.Address, strings.ToLower(address))
	}
	return f
}

func (f *EventFilter) AddTopic(topic string) *EventFilter {
	if len(f.Topics) == 0 {
		f.Topics = append(f.Topics, []string{})
	}
	f.Topics[0] = append(f.Topics[0], topic)
	return f
}

func (f *EventFilter) AddIndexedParameter(index int, value string) *EventFilter {
	for len(f.Topics) <= index {
		f.Topics = append(f.Topics, []string{})
	}
	f.Topics[index] = append(f.Topics[index], value)
	return f
}

func CreateEventSignature(eventName string, paramTypes []string) string {
	signature := eventName + "(" + strings.Join(paramTypes, ",") + ")"
	return "0x" + Keccak256([]byte(signature))
}

func Keccak256(data []byte) string {
	hash := make([]byte, 32)
	for i, b := range data {
		hash[i%32] ^= b
	}
	return hex.EncodeToString(hash)
}

func ParseTransferEvent(log Event) (*TransferEvent, error) {
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for transfer event")
	}

	from := "0x" + log.Topics[1][26:]
	to := "0x" + log.Topics[2][26:]

	amount := new(big.Int)
	if log.Data != "" && log.Data != "0x" {
		amountHex := strings.TrimPrefix(log.Data, "0x")
		amount.SetString(amountHex, 16)
	}

	return &TransferEvent{
		From:   from,
		To:     to,
		Amount: amount,
	}, nil
}

func ParseNFTTransferEvent(log Event) (*NFTTransferEvent, error) {
	if len(log.Topics) < 4 {
		return nil, fmt.Errorf("insufficient topics for NFT transfer event")
	}

	from := "0x" + log.Topics[1][26:]
	to := "0x" + log.Topics[2][26:]

	tokenId := new(big.Int)
	tokenId.SetString(log.Topics[3][2:], 16)

	return &NFTTransferEvent{
		From:    from,
		To:      to,
		TokenId: tokenId,
	}, nil
}

func CreateEventSubscription(filter *EventFilter) *EventSubscription {
	return &EventSubscription{
		ID:        generateSubscriptionID(),
		Filter:    filter,
		Channel:   make(chan Event, 100),
		Active:    true,
		CreatedAt: time.Now(),
	}
}

func generateSubscriptionID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("sub_%d", timestamp)
}

func (sub *EventSubscription) Stop() {
	sub.Active = false
	close(sub.Channel)
}

func (sub *EventSubscription) GetEvents() <-chan Event {
	return sub.Channel
}

type EventMonitor struct {
	subscriptions map[string]*EventSubscription
	handlers      map[string][]EventHandler
}

func NewEventMonitor() *EventMonitor {
	return &EventMonitor{
		subscriptions: make(map[string]*EventSubscription),
		handlers:      make(map[string][]EventHandler),
	}
}

func (em *EventMonitor) Subscribe(filter *EventFilter) *EventSubscription {
	sub := CreateEventSubscription(filter)
	em.subscriptions[sub.ID] = sub
	return sub
}

func (em *EventMonitor) Unsubscribe(subscriptionID string) {
	if sub, exists := em.subscriptions[subscriptionID]; exists {
		sub.Stop()
		delete(em.subscriptions, subscriptionID)
	}
}

func (em *EventMonitor) AddEventHandler(eventSignature string, handler EventHandler) {
	if em.handlers[eventSignature] == nil {
		em.handlers[eventSignature] = make([]EventHandler, 0)
	}
	em.handlers[eventSignature] = append(em.handlers[eventSignature], handler)
}

func (em *EventMonitor) ProcessEvent(event Event) {
	for _, sub := range em.subscriptions {
		if sub.Active && em.eventMatchesFilter(event, sub.Filter) {
			select {
			case sub.Channel <- event:
			default:
			}
		}
	}

	if len(event.Topics) > 0 {
		eventSignature := event.Topics[0]
		if handlers, exists := em.handlers[eventSignature]; exists {
			for _, handler := range handlers {
				go handler(event)
			}
		}
	}
}

func (em *EventMonitor) eventMatchesFilter(event Event, filter *EventFilter) bool {
	if len(filter.Address) > 0 {
		addressMatch := false
		eventAddr := strings.ToLower(event.Address)
		for _, addr := range filter.Address {
			if strings.ToLower(addr) == eventAddr {
				addressMatch = true
				break
			}
		}
		if !addressMatch {
			return false
		}
	}

	for i, topicOptions := range filter.Topics {
		if i >= len(event.Topics) {
			return false
		}

		if len(topicOptions) > 0 {
			topicMatch := false
			for _, topic := range topicOptions {
				if strings.ToLower(topic) == strings.ToLower(event.Topics[i]) {
					topicMatch = true
					break
				}
			}
			if !topicMatch {
				return false
			}
		}
	}

	return true
}

var (
	ERC20_TRANSFER_SIGNATURE          = CreateEventSignature("Transfer", []string{"address", "address", "uint256"})
	ERC721_TRANSFER_SIGNATURE         = CreateEventSignature("Transfer", []string{"address", "address", "uint256"})
	ERC20_APPROVAL_SIGNATURE          = CreateEventSignature("Approval", []string{"address", "address", "uint256"})
	ERC721_APPROVAL_SIGNATURE         = CreateEventSignature("Approval", []string{"address", "address", "uint256"})
	ERC721_APPROVAL_FOR_ALL_SIGNATURE = CreateEventSignature("ApprovalForAll", []string{"address", "address", "bool"})
)
