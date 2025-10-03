package web3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const (
	ERC721_TRANSFER_FROM_SELECTOR        = "23b872dd"
	ERC721_SAFE_TRANSFER_FROM_SELECTOR   = "42842e0e"
	ERC721_APPROVE_SELECTOR              = "095ea7b3"
	ERC721_SET_APPROVAL_FOR_ALL_SELECTOR = "a22cb465"
	ERC721_OWNER_OF_SELECTOR             = "6352211e"
	ERC721_BALANCE_OF_SELECTOR           = "70a08231"
	ERC721_GET_APPROVED_SELECTOR         = "081812fc"
	ERC721_IS_APPROVED_FOR_ALL_SELECTOR  = "e985e9c5"
	ERC721_TOKEN_URI_SELECTOR            = "c87b56dd"
	ERC721_NAME_SELECTOR                 = "06fdde03"
	ERC721_SYMBOL_SELECTOR               = "95d89b41"
)

type ERC721Token struct {
	Address string
	Name    string
	Symbol  string
}

func NewERC721Token(address, name, symbol string) *ERC721Token {
	return &ERC721Token{
		Address: address,
		Name:    name,
		Symbol:  symbol,
	}
}

func (nft *ERC721Token) EncodeTransferFrom(from, to string, tokenId *big.Int) ([]byte, error) {
	if !ValidateAddress(from) {
		return nil, fmt.Errorf("invalid sender address")
	}
	if !ValidateAddress(to) {
		return nil, fmt.Errorf("invalid recipient address")
	}

	selector, _ := hex.DecodeString(ERC721_TRANSFER_FROM_SELECTOR)

	fromAddress := strings.TrimPrefix(from, "0x")
	fromBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", fromAddress))

	toAddress := strings.TrimPrefix(to, "0x")
	toBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", toAddress))

	tokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(tokenIdBytes)

	data := append(selector, fromBytes...)
	data = append(data, toBytes...)
	data = append(data, tokenIdBytes...)

	return data, nil
}

func (nft *ERC721Token) EncodeSafeTransferFrom(from, to string, tokenId *big.Int, data []byte) ([]byte, error) {
	if !ValidateAddress(from) {
		return nil, fmt.Errorf("invalid sender address")
	}
	if !ValidateAddress(to) {
		return nil, fmt.Errorf("invalid recipient address")
	}

	selector, _ := hex.DecodeString(ERC721_SAFE_TRANSFER_FROM_SELECTOR)

	fromAddress := strings.TrimPrefix(from, "0x")
	fromBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", fromAddress))

	toAddress := strings.TrimPrefix(to, "0x")
	toBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", toAddress))

	tokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(tokenIdBytes)

	callData := append(selector, fromBytes...)
	callData = append(callData, toBytes...)
	callData = append(callData, tokenIdBytes...)

	if len(data) > 0 {
		dataLengthBytes := make([]byte, 32)
		big.NewInt(int64(len(data))).FillBytes(dataLengthBytes)
		callData = append(callData, dataLengthBytes...)
		callData = append(callData, data...)
	}

	return callData, nil
}

func (nft *ERC721Token) EncodeApprove(to string, tokenId *big.Int) ([]byte, error) {
	if !ValidateAddress(to) {
		return nil, fmt.Errorf("invalid recipient address")
	}

	selector, _ := hex.DecodeString(ERC721_APPROVE_SELECTOR)

	toAddress := strings.TrimPrefix(to, "0x")
	toBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", toAddress))

	tokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(tokenIdBytes)

	data := append(selector, toBytes...)
	data = append(data, tokenIdBytes...)

	return data, nil
}

func (nft *ERC721Token) EncodeSetApprovalForAll(operator string, approved bool) ([]byte, error) {
	if !ValidateAddress(operator) {
		return nil, fmt.Errorf("invalid operator address")
	}

	selector, _ := hex.DecodeString(ERC721_SET_APPROVAL_FOR_ALL_SELECTOR)

	operatorAddress := strings.TrimPrefix(operator, "0x")
	operatorBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", operatorAddress))

	approvedBytes := make([]byte, 32)
	if approved {
		approvedBytes[31] = 1
	}

	data := append(selector, operatorBytes...)
	data = append(data, approvedBytes...)

	return data, nil
}

func (nft *ERC721Token) EncodeOwnerOf(tokenId *big.Int) ([]byte, error) {
	selector, _ := hex.DecodeString(ERC721_OWNER_OF_SELECTOR)

	tokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(tokenIdBytes)

	data := append(selector, tokenIdBytes...)

	return data, nil
}

func (nft *ERC721Token) EncodeBalanceOf(owner string) ([]byte, error) {
	if !ValidateAddress(owner) {
		return nil, fmt.Errorf("invalid owner address")
	}

	selector, _ := hex.DecodeString(ERC721_BALANCE_OF_SELECTOR)

	ownerAddress := strings.TrimPrefix(owner, "0x")
	ownerBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", ownerAddress))

	data := append(selector, ownerBytes...)

	return data, nil
}

func (nft *ERC721Token) EncodeGetApproved(tokenId *big.Int) ([]byte, error) {
	selector, _ := hex.DecodeString(ERC721_GET_APPROVED_SELECTOR)

	tokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(tokenIdBytes)

	data := append(selector, tokenIdBytes...)

	return data, nil
}

func (nft *ERC721Token) EncodeIsApprovedForAll(owner, operator string) ([]byte, error) {
	if !ValidateAddress(owner) {
		return nil, fmt.Errorf("invalid owner address")
	}
	if !ValidateAddress(operator) {
		return nil, fmt.Errorf("invalid operator address")
	}

	selector, _ := hex.DecodeString(ERC721_IS_APPROVED_FOR_ALL_SELECTOR)

	ownerAddress := strings.TrimPrefix(owner, "0x")
	ownerBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", ownerAddress))

	operatorAddress := strings.TrimPrefix(operator, "0x")
	operatorBytes, _ := hex.DecodeString(fmt.Sprintf("%064s", operatorAddress))

	data := append(selector, ownerBytes...)
	data = append(data, operatorBytes...)

	return data, nil
}

func (nft *ERC721Token) EncodeTokenURI(tokenId *big.Int) ([]byte, error) {
	selector, _ := hex.DecodeString(ERC721_TOKEN_URI_SELECTOR)

	tokenIdBytes := make([]byte, 32)
	tokenId.FillBytes(tokenIdBytes)

	data := append(selector, tokenIdBytes...)

	return data, nil
}

func (nft *ERC721Token) DecodeTransferEvent(logData string, topics []string) (*NFTTransferEvent, error) {
	if len(topics) < 4 {
		return nil, fmt.Errorf("insufficient topics for NFT transfer event")
	}

	from := "0x" + topics[1][26:]
	to := "0x" + topics[2][26:]

	tokenId := new(big.Int)
	tokenId.SetString(topics[3][2:], 16)

	return &NFTTransferEvent{
		From:    from,
		To:      to,
		TokenId: tokenId,
	}, nil
}

func (nft *ERC721Token) DecodeApprovalEvent(logData string, topics []string) (*NFTApprovalEvent, error) {
	if len(topics) < 4 {
		return nil, fmt.Errorf("insufficient topics for NFT approval event")
	}

	owner := "0x" + topics[1][26:]
	approved := "0x" + topics[2][26:]

	tokenId := new(big.Int)
	tokenId.SetString(topics[3][2:], 16)

	return &NFTApprovalEvent{
		Owner:    owner,
		Approved: approved,
		TokenId:  tokenId,
	}, nil
}

type NFTTransferEvent struct {
	From    string
	To      string
	TokenId *big.Int
}

type NFTApprovalEvent struct {
	Owner    string
	Approved string
	TokenId  *big.Int
}

type NFTApprovalForAllEvent struct {
	Owner    string
	Operator string
	Approved bool
}
