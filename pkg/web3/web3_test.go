package web3

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestKeccak256(t *testing.T) {
	if got := Keccak256(nil); got != "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" {
		t.Errorf("Keccak256(empty) = %s", got)
	}
}

func TestEventSignatures(t *testing.T) {
	tests := map[string]string{
		"Transfer":       ERC20_TRANSFER_SIGNATURE,
		"Approval":       ERC20_APPROVAL_SIGNATURE,
		"ApprovalForAll": ERC721_APPROVAL_FOR_ALL_SIGNATURE,
	}
	want := map[string]string{
		"Transfer":       "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"Approval":       "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925",
		"ApprovalForAll": "0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31",
	}
	for name, got := range tests {
		if got != want[name] {
			t.Errorf("%s signature = %s, want %s", name, got, want[name])
		}
	}
}

func TestSelectorConstantsMatchKeccak(t *testing.T) {
	tests := map[string]string{
		"transfer(address,uint256)":                       ERC20_TRANSFER_SELECTOR,
		"transferFrom(address,address,uint256)":           ERC20_TRANSFER_FROM_SELECTOR,
		"approve(address,uint256)":                        ERC20_APPROVE_SELECTOR,
		"balanceOf(address)":                              ERC20_BALANCE_OF_SELECTOR,
		"allowance(address,address)":                      ERC20_ALLOWANCE_SELECTOR,
		"totalSupply()":                                   ERC20_TOTAL_SUPPLY_SELECTOR,
		"safeTransferFrom(address,address,uint256)":       ERC721_SAFE_TRANSFER_FROM_SELECTOR,
		"safeTransferFrom(address,address,uint256,bytes)": ERC721_SAFE_TRANSFER_FROM_WITH_DATA_SELECTOR,
		"setApprovalForAll(address,bool)":                 ERC721_SET_APPROVAL_FOR_ALL_SELECTOR,
		"ownerOf(uint256)":                                ERC721_OWNER_OF_SELECTOR,
		"getApproved(uint256)":                            ERC721_GET_APPROVED_SELECTOR,
		"isApprovedForAll(address,address)":               ERC721_IS_APPROVED_FOR_ALL_SELECTOR,
		"tokenURI(uint256)":                               ERC721_TOKEN_URI_SELECTOR,
	}
	for signature, selector := range tests {
		if got := Keccak256([]byte(signature))[:8]; got != selector {
			t.Errorf("%s selector = %s, constant is %s", signature, got, selector)
		}
	}
}

func TestEncodeTransferMatchesGenericEncoder(t *testing.T) {
	token := NewERC20Token(benchAddrA, "USD Coin", "USDC", 6)
	amount := big.NewInt(1_000_000)

	specific, err := token.EncodeTransfer(benchAddrA, amount)
	if err != nil {
		t.Fatal(err)
	}
	generic, err := EncodeFunctionCall("transfer",
		[]ABIParam{{Type: "address"}, {Type: "uint256"}},
		[]interface{}{benchAddrA, amount})
	if err != nil {
		t.Fatal(err)
	}

	want := mustHex(t, "a9059cbb000000000000000000000000742d35cc6634c0532925a3b8d82c28d53e01bcf200000000000000000000000000000000000000000000000000000000000f4240")
	if !bytes.Equal(specific, want) {
		t.Errorf("EncodeTransfer = %x", specific)
	}
	if !bytes.Equal(generic, want) {
		t.Errorf("EncodeFunctionCall = %x", generic)
	}
}

func TestEncodeRejectsBadInput(t *testing.T) {
	token := NewERC20Token(benchAddrA, "", "", 18)
	if _, err := token.EncodeTransfer("0x123", big.NewInt(1)); err == nil {
		t.Error("expected error for short address")
	}
	if _, err := token.EncodeTransfer(benchAddrA, big.NewInt(-1)); err == nil {
		t.Error("expected error for negative amount")
	}
	if _, err := token.EncodeTransfer(benchAddrA, new(big.Int).Lsh(big.NewInt(1), 256)); err == nil {
		t.Error("expected error for amount over uint256")
	}
}

func TestEncodeIntNegative(t *testing.T) {
	got, err := encodeInt("int256", -1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, bytes.Repeat([]byte{0xff}, 32)) {
		t.Errorf("encodeInt(-1) = %x", got)
	}
}

func TestEncodeDynamicArrays(t *testing.T) {
	got, err := encodeTuple([]string{"string[]"}, []interface{}{[]string{"a", "b"}})
	if err != nil {
		t.Fatal(err)
	}
	want := mustHex(t, ""+
		"0000000000000000000000000000000000000000000000000000000000000020"+ // offset to array
		"0000000000000000000000000000000000000000000000000000000000000002"+ // length
		"0000000000000000000000000000000000000000000000000000000000000040"+ // offset to "a"
		"0000000000000000000000000000000000000000000000000000000000000080"+ // offset to "b"
		"0000000000000000000000000000000000000000000000000000000000000001"+
		"6100000000000000000000000000000000000000000000000000000000000000"+
		"0000000000000000000000000000000000000000000000000000000000000001"+
		"6200000000000000000000000000000000000000000000000000000000000000")
	if !bytes.Equal(got, want) {
		t.Errorf("string[] encoding =\n%x\nwant\n%x", got, want)
	}

	uints, err := encodeTuple([]string{"uint256[]"}, []interface{}{[]interface{}{1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(uints) != 4*32 || uints[3*32-1] != 1 || uints[4*32-1] != 2 {
		t.Errorf("uint256[] encoding = %x", uints)
	}
}

func TestDecodeFunctionResult(t *testing.T) {
	encoded, err := encodeTuple(
		[]string{"address", "uint256", "bool", "string"},
		[]interface{}{benchAddrB, 42, true, "hello"})
	if err != nil {
		t.Fatal(err)
	}

	results, err := DecodeFunctionResult([]string{"address", "uint256", "bool", "string"}, encoded)
	if err != nil {
		t.Fatal(err)
	}
	if results[0] != benchAddrB || results[1].(*big.Int).Int64() != 42 || results[2] != true || results[3] != "hello" {
		t.Errorf("decoded = %v", results)
	}

	// A huge string offset must return an error rather than panic.
	malformed := make([]byte, 64)
	malformed[0] = 0xff
	if _, err := DecodeFunctionResult([]string{"string"}, malformed); err == nil {
		t.Error("expected error for malformed string offset")
	}
}

func TestEncodeSafeTransferFromWithData(t *testing.T) {
	nft := NewERC721Token(benchAddrA, "", "")
	data, err := nft.EncodeSafeTransferFrom(benchAddrA, benchAddrB, big.NewInt(7), []byte{0xde, 0xad})
	if err != nil {
		t.Fatal(err)
	}

	generic, err := EncodeFunctionCall("safeTransferFrom",
		[]ABIParam{{Type: "address"}, {Type: "address"}, {Type: "uint256"}, {Type: "bytes"}},
		[]interface{}{benchAddrA, benchAddrB, 7, []byte{0xde, 0xad}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, generic) {
		t.Errorf("EncodeSafeTransferFrom =\n%x\nwant\n%x", data, generic)
	}

	noData, err := nft.EncodeSafeTransferFrom(benchAddrA, benchAddrB, big.NewInt(7), nil)
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(noData[:4]) != ERC721_SAFE_TRANSFER_FROM_SELECTOR || len(noData) != 4+3*32 {
		t.Errorf("EncodeSafeTransferFrom without data = %x", noData)
	}
}

func TestPrivateKeyToAddress(t *testing.T) {
	// Well-known test vector from the web3.js documentation.
	address, err := PrivateKeyToAddress("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	if err != nil {
		t.Fatal(err)
	}
	if address != "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23" {
		t.Errorf("PrivateKeyToAddress = %s", address)
	}

	if ValidatePrivateKey("0x" + strings.Repeat("0", 64)) {
		t.Error("zero private key must be invalid")
	}
	if ValidatePrivateKey("0x" + strings.Repeat("f", 64)) {
		t.Error("private key above curve order must be invalid")
	}
}

func TestGenerateRandomPrivateKey(t *testing.T) {
	a, b := GenerateRandomPrivateKey(), GenerateRandomPrivateKey()
	if a == b {
		t.Error("generated identical private keys")
	}
	if !ValidatePrivateKey(a) {
		t.Errorf("generated invalid private key %s", a)
	}
}

func TestTransactionRLP(t *testing.T) {
	// EIP-155 example transaction, without the chain ID fields.
	value, _ := new(big.Int).SetString("1000000000000000000", 10)
	tx := &Transaction{
		Nonce:    9,
		GasPrice: big.NewInt(20000000000),
		Gas:      21000,
		To:       "0x3535353535353535353535353535353535353535",
		Value:    value,
	}
	want := "e9098504a817c800825208943535353535353535353535353535353535353535880de0b6b3a764000080"
	if got := hex.EncodeToString(tx.encodeRLP()); got != want {
		t.Errorf("RLP = %s, want %s", got, want)
	}
	if got := tx.Hash(); got != "0x"+Keccak256(mustHex(t, want)) {
		t.Errorf("Hash = %s", got)
	}
}

func TestRLPLongPayload(t *testing.T) {
	data := bytes.Repeat([]byte{1}, 60)
	got := rlpAppendBytes(nil, data)
	if got[0] != 0xb8 || got[1] != 60 || len(got) != 62 {
		t.Errorf("long string header = %x", got[:2])
	}
}

func TestEstimateGas(t *testing.T) {
	gas, err := EstimateGas("", "", "0x0001", nil)
	if err != nil {
		t.Fatal(err)
	}
	if gas != 21000+4+16 {
		t.Errorf("EstimateGas = %d", gas)
	}
	if gas, _ := EstimateGas("", "", "", big.NewInt(1)); gas != 21000 {
		t.Errorf("EstimateGas(plain transfer) = %d", gas)
	}
}

func TestUnits(t *testing.T) {
	tests := []struct {
		in       string
		decimals int
		wei      string
		out      string
	}{
		{"100.50", 6, "100500000", "100.5"},
		{"50.123456789012345678", 18, "50123456789012345678", "50.123456789012345678"},
		{"0.000001", 18, "1000000000000", "0.000001"},
		{"-1.5", 18, "-1500000000000000000", "-1.5"},
		{"-0.5", 18, "-500000000000000000", "-0.5"},
		{"42", 0, "42", "42"},
	}
	for _, tt := range tests {
		wei, err := ParseUnits(tt.in, tt.decimals)
		if err != nil {
			t.Fatalf("ParseUnits(%q): %v", tt.in, err)
		}
		if wei.String() != tt.wei {
			t.Errorf("ParseUnits(%q) = %s, want %s", tt.in, wei, tt.wei)
		}
		if out := FormatUnits(wei, tt.decimals); out != tt.out {
			t.Errorf("FormatUnits(%s) = %s, want %s", wei, out, tt.out)
		}
	}

	if _, err := ParseUnits("1.2.3", 18); err == nil {
		t.Error("expected error for 1.2.3")
	}

	wei, err := ParseEther("0.1")
	if err != nil || wei.String() != "100000000000000000" {
		t.Errorf("ParseEther(0.1) = %v, %v", wei, err)
	}

	if got := FormatEther(wei, 4); got != "0.1000" {
		t.Errorf("FormatEther = %s", got)
	}
}

func TestParseTransferEventRejectsShortTopics(t *testing.T) {
	_, err := ParseTransferEvent(Event{Topics: []string{ERC20_TRANSFER_SIGNATURE, "0x12", "0x34"}})
	if err == nil {
		t.Error("expected error for short topics")
	}

	padded := "0x000000000000000000000000" + benchAddrA[2:]
	event, err := ParseTransferEvent(Event{
		Topics: []string{ERC20_TRANSFER_SIGNATURE, padded, padded},
		Data:   "0x0f4240",
	})
	if err != nil {
		t.Fatal(err)
	}
	if event.From != benchAddrA || event.Amount.Int64() != 1_000_000 {
		t.Errorf("event = %+v", event)
	}
}

func TestEventMonitorConcurrent(t *testing.T) {
	em := NewEventMonitorWithConcurrency(4)
	var handled atomic.Int64
	em.AddEventHandler(ERC20_TRANSFER_SIGNATURE, func(Event) error {
		handled.Add(1)
		return nil
	})

	event := Event{Address: strings.ToUpper(benchAddrA), Topics: []string{strings.ToUpper(ERC20_TRANSFER_SIGNATURE)}}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub := em.Subscribe(NewEventFilter().AddAddress(benchAddrA))
			for j := 0; j < 50; j++ {
				em.ProcessEvent(event)
			}
			sub.Stop()
			sub.Stop() // must not panic
			em.Unsubscribe(sub.ID)
		}()
	}
	wg.Wait()

	deadline := time.Now().Add(2 * time.Second)
	for handled.Load() < 400 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if handled.Load() != 400 {
		t.Errorf("handled %d events, want 400", handled.Load())
	}
}
