package web3

import (
	"fmt"
	"math/big"
)

const (
	ERC721_TRANSFER_FROM_SELECTOR                = "23b872dd"
	ERC721_SAFE_TRANSFER_FROM_SELECTOR           = "42842e0e"
	ERC721_SAFE_TRANSFER_FROM_WITH_DATA_SELECTOR = "b88d4fde"
	ERC721_APPROVE_SELECTOR                      = "095ea7b3"
	ERC721_SET_APPROVAL_FOR_ALL_SELECTOR         = "a22cb465"
	ERC721_OWNER_OF_SELECTOR                     = "6352211e"
	ERC721_BALANCE_OF_SELECTOR                   = "70a08231"
	ERC721_GET_APPROVED_SELECTOR                 = "081812fc"
	ERC721_IS_APPROVED_FOR_ALL_SELECTOR          = "e985e9c5"
	ERC721_TOKEN_URI_SELECTOR                    = "c87b56dd"
	ERC721_NAME_SELECTOR                         = "06fdde03"
	ERC721_SYMBOL_SELECTOR                       = "95d89b41"
)

var (
	erc721TransferFromSelector             = mustSelector(ERC721_TRANSFER_FROM_SELECTOR)
	erc721SafeTransferFromSelector         = mustSelector(ERC721_SAFE_TRANSFER_FROM_SELECTOR)
	erc721SafeTransferFromWithDataSelector = mustSelector(ERC721_SAFE_TRANSFER_FROM_WITH_DATA_SELECTOR)
	erc721ApproveSelector                  = mustSelector(ERC721_APPROVE_SELECTOR)
	erc721SetApprovalForAllSelector        = mustSelector(ERC721_SET_APPROVAL_FOR_ALL_SELECTOR)
	erc721OwnerOfSelector                  = mustSelector(ERC721_OWNER_OF_SELECTOR)
	erc721BalanceOfSelector                = mustSelector(ERC721_BALANCE_OF_SELECTOR)
	erc721GetApprovedSelector              = mustSelector(ERC721_GET_APPROVED_SELECTOR)
	erc721IsApprovedForAllSelector         = mustSelector(ERC721_IS_APPROVED_FOR_ALL_SELECTOR)
	erc721TokenURISelector                 = mustSelector(ERC721_TOKEN_URI_SELECTOR)
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
	return encodeFromToTokenID(erc721TransferFromSelector, 3, from, to, tokenId)
}

// EncodeSafeTransferFrom encodes safeTransferFrom(address,address,uint256) when
// data is empty, and safeTransferFrom(address,address,uint256,bytes) otherwise.
func (nft *ERC721Token) EncodeSafeTransferFrom(from, to string, tokenId *big.Int, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return encodeFromToTokenID(erc721SafeTransferFromSelector, 3, from, to, tokenId)
	}

	// Head: from, to, tokenId, offset to bytes; tail: length + padded data
	callData, err := encodeFromToTokenID(erc721SafeTransferFromWithDataSelector, 5+paddedLen(len(data))/32, from, to, tokenId)
	if err != nil {
		return nil, err
	}
	putUint64(word(callData, 3), 4*32)
	putUint64(word(callData, 4), uint64(len(data)))
	copy(callData[4+5*32:], data)

	return callData, nil
}

// encodeFromToTokenID allocates calldata with the given number of words and
// fills the first three with from, to and tokenId.
func encodeFromToTokenID(selector [4]byte, words int, from, to string, tokenId *big.Int) ([]byte, error) {
	fromAddress, ok := parseAddress(from)
	if !ok {
		return nil, fmt.Errorf("invalid sender address")
	}
	toAddress, ok := parseAddress(to)
	if !ok {
		return nil, fmt.Errorf("invalid recipient address")
	}

	callData := newCallData(selector, words)
	putAddress(word(callData, 0), fromAddress)
	putAddress(word(callData, 1), toAddress)
	if err := putUint256(word(callData, 2), tokenId); err != nil {
		return nil, fmt.Errorf("invalid token ID: %w", err)
	}

	return callData, nil
}

func (nft *ERC721Token) EncodeApprove(to string, tokenId *big.Int) ([]byte, error) {
	toAddress, ok := parseAddress(to)
	if !ok {
		return nil, fmt.Errorf("invalid recipient address")
	}

	data := newCallData(erc721ApproveSelector, 2)
	putAddress(word(data, 0), toAddress)
	if err := putUint256(word(data, 1), tokenId); err != nil {
		return nil, fmt.Errorf("invalid token ID: %w", err)
	}

	return data, nil
}

func (nft *ERC721Token) EncodeSetApprovalForAll(operator string, approved bool) ([]byte, error) {
	operatorAddress, ok := parseAddress(operator)
	if !ok {
		return nil, fmt.Errorf("invalid operator address")
	}

	data := newCallData(erc721SetApprovalForAllSelector, 2)
	putAddress(word(data, 0), operatorAddress)
	if approved {
		word(data, 1)[31] = 1
	}

	return data, nil
}

func (nft *ERC721Token) EncodeOwnerOf(tokenId *big.Int) ([]byte, error) {
	return encodeTokenIDCall(erc721OwnerOfSelector, tokenId)
}

func (nft *ERC721Token) EncodeBalanceOf(owner string) ([]byte, error) {
	ownerAddress, ok := parseAddress(owner)
	if !ok {
		return nil, fmt.Errorf("invalid owner address")
	}

	data := newCallData(erc721BalanceOfSelector, 1)
	putAddress(word(data, 0), ownerAddress)

	return data, nil
}

func (nft *ERC721Token) EncodeGetApproved(tokenId *big.Int) ([]byte, error) {
	return encodeTokenIDCall(erc721GetApprovedSelector, tokenId)
}

func (nft *ERC721Token) EncodeIsApprovedForAll(owner, operator string) ([]byte, error) {
	ownerAddress, ok := parseAddress(owner)
	if !ok {
		return nil, fmt.Errorf("invalid owner address")
	}
	operatorAddress, ok := parseAddress(operator)
	if !ok {
		return nil, fmt.Errorf("invalid operator address")
	}

	data := newCallData(erc721IsApprovedForAllSelector, 2)
	putAddress(word(data, 0), ownerAddress)
	putAddress(word(data, 1), operatorAddress)

	return data, nil
}

func (nft *ERC721Token) EncodeTokenURI(tokenId *big.Int) ([]byte, error) {
	return encodeTokenIDCall(erc721TokenURISelector, tokenId)
}

func encodeTokenIDCall(selector [4]byte, tokenId *big.Int) ([]byte, error) {
	data := newCallData(selector, 1)
	if err := putUint256(word(data, 0), tokenId); err != nil {
		return nil, fmt.Errorf("invalid token ID: %w", err)
	}
	return data, nil
}

func (nft *ERC721Token) DecodeTransferEvent(logData string, topics []string) (*NFTTransferEvent, error) {
	return parseNFTTransferEvent(topics)
}

func (nft *ERC721Token) DecodeApprovalEvent(logData string, topics []string) (*NFTApprovalEvent, error) {
	if len(topics) < 4 {
		return nil, fmt.Errorf("insufficient topics for NFT approval event")
	}

	owner, err := topicToAddress(topics[1])
	if err != nil {
		return nil, fmt.Errorf("invalid owner topic: %w", err)
	}
	approved, err := topicToAddress(topics[2])
	if err != nil {
		return nil, fmt.Errorf("invalid approved topic: %w", err)
	}
	tokenId, err := topicToBigInt(topics[3])
	if err != nil {
		return nil, fmt.Errorf("invalid token ID topic: %w", err)
	}

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
