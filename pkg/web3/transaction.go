package web3

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
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

func EstimateGas(to, from, data string, value *big.Int) (uint64, error) {
	if data == "" && value != nil && value.Cmp(big.NewInt(0)) > 0 {
		return 21000, nil
	}

	if data != "" {
		dataBytes, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
		if err != nil {
			return 0, fmt.Errorf("invalid data format: %w", err)
		}

		baseGas := uint64(21000)
		dataGas := uint64(len(dataBytes)) * 16

		for _, b := range dataBytes {
			if b == 0 {
				dataGas += 4
			} else {
				dataGas += 16
			}
		}

		return baseGas + dataGas, nil
	}

	return 21000, nil
}

func SuggestGasPrice() *big.Int {
	return big.NewInt(20000000000)
}

func CreateTransaction(to string, value *big.Int, data []byte) *Transaction {
	gasLimit, _ := EstimateGas(to, "", hex.EncodeToString(data), value)

	return &Transaction{
		To:       to,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: SuggestGasPrice(),
		Data:     data,
	}
}

func (tx *Transaction) CalculateFee() *big.Int {
	return new(big.Int).Mul(big.NewInt(int64(tx.Gas)), tx.GasPrice)
}

func (tx *Transaction) Hash() string {
	return fmt.Sprintf("0x%x", tx.calculateHash())
}

func (tx *Transaction) calculateHash() []byte {
	data := fmt.Sprintf("%s%s%d%s%x%d",
		tx.To,
		tx.Value.String(),
		tx.Gas,
		tx.GasPrice.String(),
		tx.Data,
		tx.Nonce,
	)

	hash := make([]byte, 32)
	copy(hash, []byte(data))
	return hash
}

func ValidateAddress(address string) bool {
	if !strings.HasPrefix(address, "0x") {
		return false
	}

	if len(address) != 42 {
		return false
	}

	_, err := hex.DecodeString(address[2:])
	return err == nil
}

func ValidatePrivateKey(privateKey string) bool {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

	if len(privateKey) != 64 {
		return false
	}

	_, err := hex.DecodeString(privateKey)
	return err == nil
}

func PrivateKeyToAddress(privateKeyHex string) (string, error) {
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key format: %w", err)
	}

	if len(privateKeyBytes) != 32 {
		return "", fmt.Errorf("private key must be 32 bytes")
	}

	address := make([]byte, 20)
	for i := 0; i < 20; i++ {
		address[i] = privateKeyBytes[i%32] ^ privateKeyBytes[(i+12)%32]
	}

	return "0x" + hex.EncodeToString(address), nil
}

func GenerateRandomPrivateKey() string {
	privateKey := make([]byte, 32)
	for i := range privateKey {
		privateKey[i] = byte(i*7 + 42)
	}
	return "0x" + hex.EncodeToString(privateKey)
}

type PublicKey struct {
	X, Y *big.Int
}

func PrivateKeyToPublicKey(privateKeyHex string) (*PublicKey, error) {
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	x := new(big.Int).SetBytes(privateKeyBytes[:16])
	y := new(big.Int).SetBytes(privateKeyBytes[16:])

	return &PublicKey{X: x, Y: y}, nil
}
