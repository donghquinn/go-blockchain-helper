package web3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/sha3"
)

// DefaultMaxConcurrentHandlers bounds how many event handlers an EventMonitor
// runs at once; ProcessEvent blocks when the limit is reached.
const DefaultMaxConcurrentHandlers = 64

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

	mu sync.Mutex
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
	f.Topics[0] = append(f.Topics[0], strings.ToLower(topic))
	return f
}

func (f *EventFilter) AddIndexedParameter(index int, value string) *EventFilter {
	for len(f.Topics) <= index {
		f.Topics = append(f.Topics, []string{})
	}
	f.Topics[index] = append(f.Topics[index], strings.ToLower(value))
	return f
}

func CreateEventSignature(eventName string, paramTypes []string) string {
	signature := eventName + "(" + strings.Join(paramTypes, ",") + ")"
	return "0x" + Keccak256([]byte(signature))
}

// Keccak256 returns the hex-encoded (no 0x prefix) Ethereum Keccak-256 hash.
func Keccak256(data []byte) string {
	return hex.EncodeToString(keccak256(data))
}

func keccak256(data []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func ParseTransferEvent(log Event) (*TransferEvent, error) {
	return parseTransferEvent(log.Data, log.Topics)
}

func ParseNFTTransferEvent(log Event) (*NFTTransferEvent, error) {
	return parseNFTTransferEvent(log.Topics)
}

func parseTransferEvent(logData string, topics []string) (*TransferEvent, error) {
	if len(topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for transfer event")
	}

	from, err := topicToAddress(topics[1])
	if err != nil {
		return nil, fmt.Errorf("invalid from topic: %w", err)
	}
	to, err := topicToAddress(topics[2])
	if err != nil {
		return nil, fmt.Errorf("invalid to topic: %w", err)
	}

	amount := new(big.Int)
	if logData != "" && logData != "0x" {
		if _, ok := amount.SetString(strings.TrimPrefix(logData, "0x"), 16); !ok {
			return nil, fmt.Errorf("invalid transfer amount data")
		}
	}

	return &TransferEvent{
		From:   from,
		To:     to,
		Amount: amount,
	}, nil
}

func parseNFTTransferEvent(topics []string) (*NFTTransferEvent, error) {
	if len(topics) < 4 {
		return nil, fmt.Errorf("insufficient topics for NFT transfer event")
	}

	from, err := topicToAddress(topics[1])
	if err != nil {
		return nil, fmt.Errorf("invalid from topic: %w", err)
	}
	to, err := topicToAddress(topics[2])
	if err != nil {
		return nil, fmt.Errorf("invalid to topic: %w", err)
	}
	tokenId, err := topicToBigInt(topics[3])
	if err != nil {
		return nil, fmt.Errorf("invalid token ID topic: %w", err)
	}

	return &NFTTransferEvent{
		From:    from,
		To:      to,
		TokenId: tokenId,
	}, nil
}

// topicToAddress extracts the address from a 32-byte (0x + 64 hex) topic.
func topicToAddress(topic string) (string, error) {
	if len(topic) != 66 || !strings.HasPrefix(topic, "0x") {
		return "", fmt.Errorf("topic must be 0x-prefixed 32 bytes")
	}
	return "0x" + topic[26:], nil
}

func topicToBigInt(topic string) (*big.Int, error) {
	if len(topic) < 3 || !strings.HasPrefix(topic, "0x") {
		return nil, fmt.Errorf("topic must be 0x-prefixed hex")
	}
	value, ok := new(big.Int).SetString(topic[2:], 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex value")
	}
	return value, nil
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

var subscriptionCounter atomic.Uint64

func generateSubscriptionID() string {
	return fmt.Sprintf("sub_%d_%d", time.Now().UnixNano(), subscriptionCounter.Add(1))
}

// Stop deactivates the subscription and closes its channel. It is safe to call
// more than once and concurrently with event delivery.
func (sub *EventSubscription) Stop() {
	sub.mu.Lock()
	defer sub.mu.Unlock()

	if !sub.Active {
		return
	}
	sub.Active = false
	close(sub.Channel)
}

// IsActive reports whether the subscription is still receiving events.
func (sub *EventSubscription) IsActive() bool {
	sub.mu.Lock()
	defer sub.mu.Unlock()
	return sub.Active
}

func (sub *EventSubscription) GetEvents() <-chan Event {
	return sub.Channel
}

// deliver sends the event without blocking, dropping it if the buffer is full.
func (sub *EventSubscription) deliver(event Event) {
	sub.mu.Lock()
	defer sub.mu.Unlock()

	if !sub.Active {
		return
	}
	select {
	case sub.Channel <- event:
	default:
	}
}

type EventMonitor struct {
	mu            sync.RWMutex
	subscriptions map[string]*EventSubscription
	handlers      map[string][]EventHandler
	handlerSlots  chan struct{}
}

func NewEventMonitor() *EventMonitor {
	return NewEventMonitorWithConcurrency(DefaultMaxConcurrentHandlers)
}

// NewEventMonitorWithConcurrency creates a monitor that runs at most
// maxConcurrentHandlers event handlers at once.
func NewEventMonitorWithConcurrency(maxConcurrentHandlers int) *EventMonitor {
	if maxConcurrentHandlers < 1 {
		maxConcurrentHandlers = 1
	}
	return &EventMonitor{
		subscriptions: make(map[string]*EventSubscription),
		handlers:      make(map[string][]EventHandler),
		handlerSlots:  make(chan struct{}, maxConcurrentHandlers),
	}
}

func (em *EventMonitor) Subscribe(filter *EventFilter) *EventSubscription {
	sub := CreateEventSubscription(filter)

	em.mu.Lock()
	em.subscriptions[sub.ID] = sub
	em.mu.Unlock()

	return sub
}

func (em *EventMonitor) Unsubscribe(subscriptionID string) {
	em.mu.Lock()
	sub, exists := em.subscriptions[subscriptionID]
	delete(em.subscriptions, subscriptionID)
	em.mu.Unlock()

	if exists {
		sub.Stop()
	}
}

func (em *EventMonitor) AddEventHandler(eventSignature string, handler EventHandler) {
	key := strings.ToLower(eventSignature)

	em.mu.Lock()
	em.handlers[key] = append(em.handlers[key], handler)
	em.mu.Unlock()
}

func (em *EventMonitor) ProcessEvent(event Event) {
	var handlers []EventHandler

	em.mu.RLock()
	for _, sub := range em.subscriptions {
		if em.eventMatchesFilter(event, sub.Filter) {
			sub.deliver(event)
		}
	}
	if len(event.Topics) > 0 {
		handlers = em.handlers[strings.ToLower(event.Topics[0])]
	}
	em.mu.RUnlock()

	for _, handler := range handlers {
		em.handlerSlots <- struct{}{}
		go func(handler EventHandler) {
			defer func() { <-em.handlerSlots }()
			handler(event)
		}(handler)
	}
}

func (em *EventMonitor) eventMatchesFilter(event Event, filter *EventFilter) bool {
	if filter == nil {
		return true
	}

	if len(filter.Address) > 0 && !containsFold(filter.Address, event.Address) {
		return false
	}

	for i, topicOptions := range filter.Topics {
		if i >= len(event.Topics) {
			return false
		}
		if len(topicOptions) > 0 && !containsFold(topicOptions, event.Topics[i]) {
			return false
		}
	}

	return true
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

var (
	ERC20_TRANSFER_SIGNATURE          = CreateEventSignature("Transfer", []string{"address", "address", "uint256"})
	ERC721_TRANSFER_SIGNATURE         = CreateEventSignature("Transfer", []string{"address", "address", "uint256"})
	ERC20_APPROVAL_SIGNATURE          = CreateEventSignature("Approval", []string{"address", "address", "uint256"})
	ERC721_APPROVAL_SIGNATURE         = CreateEventSignature("Approval", []string{"address", "address", "uint256"})
	ERC721_APPROVAL_FOR_ALL_SIGNATURE = CreateEventSignature("ApprovalForAll", []string{"address", "address", "bool"})
)
