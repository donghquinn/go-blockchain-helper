package web3

import (
	"fmt"
	"math/big"
)

const (
	ERC20_TRANSFER_SELECTOR      = "a9059cbb"
	ERC20_TRANSFER_FROM_SELECTOR = "23b872dd"
	ERC20_APPROVE_SELECTOR       = "095ea7b3"
	ERC20_BALANCE_OF_SELECTOR    = "70a08231"
	ERC20_ALLOWANCE_SELECTOR     = "dd62ed3e"
	ERC20_TOTAL_SUPPLY_SELECTOR  = "18160ddd"
	ERC20_NAME_SELECTOR          = "06fdde03"
	ERC20_SYMBOL_SELECTOR        = "95d89b41"
	ERC20_DECIMALS_SELECTOR      = "313ce567"
)

var (
	erc20TransferSelector     = mustSelector(ERC20_TRANSFER_SELECTOR)
	erc20TransferFromSelector = mustSelector(ERC20_TRANSFER_FROM_SELECTOR)
	erc20ApproveSelector      = mustSelector(ERC20_APPROVE_SELECTOR)
	erc20BalanceOfSelector    = mustSelector(ERC20_BALANCE_OF_SELECTOR)
	erc20AllowanceSelector    = mustSelector(ERC20_ALLOWANCE_SELECTOR)
	erc20TotalSupplySelector  = mustSelector(ERC20_TOTAL_SUPPLY_SELECTOR)
)

type ERC20Token struct {
	Address  string
	Name     string
	Symbol   string
	Decimals uint8
}

func NewERC20Token(address, name, symbol string, decimals uint8) *ERC20Token {
	return &ERC20Token{
		Address:  address,
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}
}

func (token *ERC20Token) EncodeTransfer(to string, amount *big.Int) ([]byte, error) {
	toAddress, ok := parseAddress(to)
	if !ok {
		return nil, fmt.Errorf("invalid recipient address")
	}

	data := newCallData(erc20TransferSelector, 2)
	putAddress(word(data, 0), toAddress)
	if err := putUint256(word(data, 1), amount); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	return data, nil
}

func (token *ERC20Token) EncodeTransferFrom(from, to string, amount *big.Int) ([]byte, error) {
	fromAddress, ok := parseAddress(from)
	if !ok {
		return nil, fmt.Errorf("invalid sender address")
	}
	toAddress, ok := parseAddress(to)
	if !ok {
		return nil, fmt.Errorf("invalid recipient address")
	}

	data := newCallData(erc20TransferFromSelector, 3)
	putAddress(word(data, 0), fromAddress)
	putAddress(word(data, 1), toAddress)
	if err := putUint256(word(data, 2), amount); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	return data, nil
}

func (token *ERC20Token) EncodeApprove(spender string, amount *big.Int) ([]byte, error) {
	spenderAddress, ok := parseAddress(spender)
	if !ok {
		return nil, fmt.Errorf("invalid spender address")
	}

	data := newCallData(erc20ApproveSelector, 2)
	putAddress(word(data, 0), spenderAddress)
	if err := putUint256(word(data, 1), amount); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	return data, nil
}

func (token *ERC20Token) EncodeBalanceOf(owner string) ([]byte, error) {
	ownerAddress, ok := parseAddress(owner)
	if !ok {
		return nil, fmt.Errorf("invalid owner address")
	}

	data := newCallData(erc20BalanceOfSelector, 1)
	putAddress(word(data, 0), ownerAddress)

	return data, nil
}

func (token *ERC20Token) EncodeAllowance(owner, spender string) ([]byte, error) {
	ownerAddress, ok := parseAddress(owner)
	if !ok {
		return nil, fmt.Errorf("invalid owner address")
	}
	spenderAddress, ok := parseAddress(spender)
	if !ok {
		return nil, fmt.Errorf("invalid spender address")
	}

	data := newCallData(erc20AllowanceSelector, 2)
	putAddress(word(data, 0), ownerAddress)
	putAddress(word(data, 1), spenderAddress)

	return data, nil
}

func (token *ERC20Token) EncodeTotalSupply() ([]byte, error) {
	return newCallData(erc20TotalSupplySelector, 0), nil
}

func (token *ERC20Token) DecodeTransferEvent(logData string, topics []string) (*TransferEvent, error) {
	return parseTransferEvent(logData, topics)
}

func (token *ERC20Token) FormatAmount(amount *big.Int) string {
	return FormatUnits(amount, int(token.Decimals))
}

func (token *ERC20Token) ParseAmount(amountStr string) (*big.Int, error) {
	return ParseUnits(amountStr, int(token.Decimals))
}

type TransferEvent struct {
	From   string
	To     string
	Amount *big.Int
}

type ApprovalEvent struct {
	Owner   string
	Spender string
	Amount  *big.Int
}
