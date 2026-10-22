package web3

import (
	"math/big"
	"testing"
)

const (
	benchAddrA = "0x742d35cc6634c0532925a3b8d82c28d53e01bcf2"
	benchAddrB = "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23"
)

func BenchmarkERC20EncodeTransfer(b *testing.B) {
	token := NewERC20Token(benchAddrA, "USD Coin", "USDC", 6)
	amount := big.NewInt(1_000_000)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := token.EncodeTransfer(benchAddrB, amount); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkERC721EncodeTransferFrom(b *testing.B) {
	nft := NewERC721Token(benchAddrA, "NFT", "NFT")
	tokenID := big.NewInt(42)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := nft.EncodeTransferFrom(benchAddrA, benchAddrB, tokenID); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeFunctionCall(b *testing.B) {
	params := []ABIParam{{Type: "address"}, {Type: "uint256"}, {Type: "string"}}
	values := []interface{}{benchAddrB, big.NewInt(1_000_000), "hello world"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := EncodeFunctionCall("transfer", params, values); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseABISignature(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ParseABISignature("transferFrom(address,address,uint256)"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventMatchesFilter(b *testing.B) {
	em := NewEventMonitor()
	filter := NewEventFilter().AddAddress(benchAddrA).AddAddress(benchAddrB).AddTopic(ERC20_TRANSFER_SIGNATURE)
	event := Event{
		Address: "0x2C7536E3605D9C16A7A3D7B1898E529396A65C23",
		Topics:  []string{ERC20_TRANSFER_SIGNATURE},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !em.eventMatchesFilter(event, filter) {
			b.Fatal("expected match")
		}
	}
}

func BenchmarkFormatUnits(b *testing.B) {
	amount, _ := new(big.Int).SetString("123456789012345678901", 10)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FormatUnits(amount, 18)
	}
}

func BenchmarkParseUnits(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ParseUnits("50.1", 18); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateTransaction(b *testing.B) {
	data := make([]byte, 68)
	data[0] = 0xa9
	value := big.NewInt(0)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CreateTransaction(benchAddrB, value, data)
	}
}
