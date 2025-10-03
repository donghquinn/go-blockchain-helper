package web3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
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
	if !ValidateAddress(to) {
		return nil, fmt.Errorf("invalid recipient address")
	}

	selector, _ := hex.DecodeString(ERC20_TRANSFER_SELECTOR)

	toAddress := strings.TrimPrefix(to, "0x")
	toBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", toAddress))

	amountBytes := make([]byte, 32)
	amount.FillBytes(amountBytes)

	data := append(selector, toBytes...)
	data = append(data, amountBytes...)

	return data, nil
}

func (token *ERC20Token) EncodeTransferFrom(from, to string, amount *big.Int) ([]byte, error) {
	if !ValidateAddress(from) {
		return nil, fmt.Errorf("invalid sender address")
	}
	if !ValidateAddress(to) {
		return nil, fmt.Errorf("invalid recipient address")
	}

	selector, _ := hex.DecodeString(ERC20_TRANSFER_FROM_SELECTOR)

	fromAddress := strings.TrimPrefix(from, "0x")
	fromBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", fromAddress))

	toAddress := strings.TrimPrefix(to, "0x")
	toBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", toAddress))

	amountBytes := make([]byte, 32)
	amount.FillBytes(amountBytes)

	data := append(selector, fromBytes...)
	data = append(data, toBytes...)
	data = append(data, amountBytes...)

	return data, nil
}

func (token *ERC20Token) EncodeApprove(spender string, amount *big.Int) ([]byte, error) {
	if !ValidateAddress(spender) {
		return nil, fmt.Errorf("invalid spender address")
	}

	selector, _ := hex.DecodeString(ERC20_APPROVE_SELECTOR)

	spenderAddress := strings.TrimPrefix(spender, "0x")
	spenderBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", spenderAddress))

	amountBytes := make([]byte, 32)
	amount.FillBytes(amountBytes)

	data := append(selector, spenderBytes...)
	data = append(data, amountBytes...)

	return data, nil
}

func (token *ERC20Token) EncodeBalanceOf(owner string) ([]byte, error) {
	if !ValidateAddress(owner) {
		return nil, fmt.Errorf("invalid owner address")
	}

	selector, _ := hex.DecodeString(ERC20_BALANCE_OF_SELECTOR)

	ownerAddress := strings.TrimPrefix(owner, "0x")
	ownerBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", ownerAddress))

	data := append(selector, ownerBytes...)

	return data, nil
}

func (token *ERC20Token) EncodeAllowance(owner, spender string) ([]byte, error) {
	if !ValidateAddress(owner) {
		return nil, fmt.Errorf("invalid owner address")
	}
	if !ValidateAddress(spender) {
		return nil, fmt.Errorf("invalid spender address")
	}

	selector, _ := hex.DecodeString(ERC20_ALLOWANCE_SELECTOR)

	ownerAddress := strings.TrimPrefix(owner, "0x")
	ownerBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", ownerAddress))

	spenderAddress := strings.TrimPrefix(spender, "0x")
	spenderBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", spenderAddress))

	data := append(selector, ownerBytes...)
	data = append(data, spenderBytes...)

	return data, nil
}

func (token *ERC20Token) EncodeTotalSupply() ([]byte, error) {
	selector, _ := hex.DecodeString(ERC20_TOTAL_SUPPLY_SELECTOR)
	return selector, nil
}

func (token *ERC20Token) DecodeTransferEvent(logData string, topics []string) (*TransferEvent, error) {
	if len(topics) < 3 {
		return nil, fmt.Errorf("insufficient topics for transfer event")
	}

	from := "0x" + topics[1][26:]
	to := "0x" + topics[2][26:]

	amount := new(big.Int)
	if logData != "" && logData != "0x" {
		amountHex := strings.TrimPrefix(logData, "0x")
		amount.SetString(amountHex, 16)
	}

	return &TransferEvent{
		From:   from,
		To:     to,
		Amount: amount,
	}, nil
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
