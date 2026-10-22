package web3

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const (
	txBaseGas        = 21000
	txDataZeroGas    = 4
	txDataNonZeroGas = 16
)

type Transaction struct {
	To       string
	Value    *big.Int
	Gas      uint64
	GasPrice *big.Int
	Data     []byte
	Nonce    uint64
}

type TransactionReceipt struct {
	Hash              string
	BlockNumber       *big.Int
	BlockHash         string
	TransactionIndex  uint
	From              string
	To                string
	GasUsed           uint64
	Status            uint64
	ContractAddress   string
	Logs              []Log
	CumulativeGasUsed uint64
}

type Log struct {
	Address     string
	Topics      []string
	Data        string
	BlockNumber *big.Int
	BlockHash   string
	TxHash      string
	TxIndex     uint
	LogIndex    uint
}

// EstimateGas returns the intrinsic gas of a transaction: the 21000 base cost
// plus 4 gas per zero byte and 16 gas per non-zero byte of calldata.
func EstimateGas(to, from, data string, value *big.Int) (uint64, error) {
	dataBytes, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
	if err != nil {
		return 0, fmt.Errorf("invalid data format: %w", err)
	}
	return intrinsicGas(dataBytes), nil
}

func intrinsicGas(data []byte) uint64 {
	gas := uint64(txBaseGas)
	for _, b := range data {
		if b == 0 {
			gas += txDataZeroGas
		} else {
			gas += txDataNonZeroGas
		}
	}
	return gas
}

func SuggestGasPrice() *big.Int {
	return big.NewInt(20000000000)
}

func CreateTransaction(to string, value *big.Int, data []byte) *Transaction {
	return &Transaction{
		To:       to,
		Value:    value,
		Gas:      intrinsicGas(data),
		GasPrice: SuggestGasPrice(),
		Data:     data,
	}
}

func (tx *Transaction) CalculateFee() *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(tx.Gas), tx.GasPrice)
}

// Hash returns the Keccak-256 hash of the RLP-encoded unsigned legacy
// transaction [nonce, gasPrice, gas, to, value, data] (the pre-EIP-155
// signing hash). An empty or invalid To is encoded as contract creation.
func (tx *Transaction) Hash() string {
	return "0x" + hex.EncodeToString(tx.calculateHash())
}

func (tx *Transaction) calculateHash() []byte {
	return keccak256(tx.encodeRLP())
}

func (tx *Transaction) encodeRLP() []byte {
	var to []byte
	if address, ok := parseAddress(tx.To); ok {
		to = address[:]
	}

	var payload []byte
	payload = rlpAppendUint64(payload, tx.Nonce)
	payload = rlpAppendBigInt(payload, tx.GasPrice)
	payload = rlpAppendUint64(payload, tx.Gas)
	payload = rlpAppendBytes(payload, to)
	payload = rlpAppendBigInt(payload, tx.Value)
	payload = rlpAppendBytes(payload, tx.Data)

	encoded := rlpAppendHeader(make([]byte, 0, len(payload)+9), 0xc0, len(payload))
	return append(encoded, payload...)
}

func rlpAppendHeader(dst []byte, offset byte, length int) []byte {
	if length <= 55 {
		return append(dst, offset+byte(length))
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(length))
	lenBytes := trimLeadingZeros(buf[:])
	dst = append(dst, offset+55+byte(len(lenBytes)))
	return append(dst, lenBytes...)
}

func rlpAppendBytes(dst, data []byte) []byte {
	if len(data) == 1 && data[0] < 0x80 {
		return append(dst, data[0])
	}
	dst = rlpAppendHeader(dst, 0x80, len(data))
	return append(dst, data...)
}

func rlpAppendUint64(dst []byte, n uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], n)
	return rlpAppendBytes(dst, trimLeadingZeros(buf[:]))
}

func rlpAppendBigInt(dst []byte, n *big.Int) []byte {
	if n == nil {
		return rlpAppendBytes(dst, nil)
	}
	return rlpAppendBytes(dst, n.Bytes())
}

func trimLeadingZeros(b []byte) []byte {
	for len(b) > 0 && b[0] == 0 {
		b = b[1:]
	}
	return b
}

func ValidateAddress(address string) bool {
	_, ok := parseAddress(address)
	return ok
}

func ValidatePrivateKey(privateKey string) bool {
	_, err := parsePrivateKey(privateKey)
	return err == nil
}

// parsePrivateKey decodes a hex secp256k1 private key, rejecting keys that are
// zero or not below the curve order.
func parsePrivateKey(privateKeyHex string) (*secp256k1.PrivateKey, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	if len(privateKeyHex) != 64 {
		return nil, fmt.Errorf("private key must be 32 bytes")
	}

	var privateKeyBytes [32]byte
	if _, err := hex.Decode(privateKeyBytes[:], []byte(privateKeyHex)); err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	var scalar secp256k1.ModNScalar
	if overflow := scalar.SetBytes(&privateKeyBytes); overflow != 0 || scalar.IsZero() {
		return nil, fmt.Errorf("private key out of range")
	}

	return secp256k1.NewPrivateKey(&scalar), nil
}

func PrivateKeyToAddress(privateKeyHex string) (string, error) {
	privateKey, err := parsePrivateKey(privateKeyHex)
	if err != nil {
		return "", err
	}

	// Address = last 20 bytes of keccak256(uncompressed pubkey without 0x04 prefix)
	publicKey := privateKey.PubKey().SerializeUncompressed()
	return "0x" + hex.EncodeToString(keccak256(publicKey[1:])[12:]), nil
}

// GenerateRandomPrivateKey returns a new private key from a cryptographically
// secure random source. It panics if the system random source fails.
func GenerateRandomPrivateKey() string {
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		panic(fmt.Sprintf("failed to generate private key: %v", err))
	}
	return "0x" + hex.EncodeToString(privateKey.Serialize())
}

type PublicKey struct {
	X, Y *big.Int
}

func PrivateKeyToPublicKey(privateKeyHex string) (*PublicKey, error) {
	privateKey, err := parsePrivateKey(privateKeyHex)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.PubKey()
	return &PublicKey{X: publicKey.X(), Y: publicKey.Y()}, nil
}
