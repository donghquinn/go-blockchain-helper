package main

// Run with: go run cmd/examples/erc20/main.go

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/donghquinn/go-blockchain-helper/pkg/web3"
)

func main() {
	fmt.Println("=== ERC-20 Token Examples ===")

	// Create ERC-20 token instances
	usdc := web3.NewERC20Token(
		"0xA0b86a33E6440417C4eE5a3C2c6e9a8b8De2fD2A",
		"USD Coin",
		"USDC",
		6,
	)

	dai := web3.NewERC20Token(
		"0x6B175474E89094C44Da98b954EedeAC495271d0F",
		"Dai Stablecoin",
		"DAI",
		18,
	)

	// Example addresses
	from := "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2"
	to := "0x8ba1f109551bD432803012645Hac136c63F5E5"
	spender := "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D" // Uniswap V2 Router

	fmt.Println("\n--- USDC Operations ---")

	// Transfer 100 USDC
	transferAmount, _ := usdc.ParseAmount("100.0")
	transferData, err := usdc.EncodeTransfer(to, transferAmount)
	if err != nil {
		fmt.Printf("Error encoding transfer: %v\n", err)
	} else {
		fmt.Printf("Transfer 100 USDC to %s\n", to)
		fmt.Printf("Call data: 0x%s\n", hex.EncodeToString(transferData))
	}

	// Approve 1000 USDC for Uniswap
	approveAmount, _ := usdc.ParseAmount("1000.0")
	approveData, err := usdc.EncodeApprove(spender, approveAmount)
	if err != nil {
		fmt.Printf("Error encoding approve: %v\n", err)
	} else {
		fmt.Printf("Approve 1000 USDC for %s\n", spender)
		fmt.Printf("Call data: 0x%s\n", hex.EncodeToString(approveData))
	}

	// Check balance
	balanceData, err := usdc.EncodeBalanceOf(from)
	if err != nil {
		fmt.Printf("Error encoding balanceOf: %v\n", err)
	} else {
		fmt.Printf("Check balance of %s\n", from)
		fmt.Printf("Call data: 0x%s\n", hex.EncodeToString(balanceData))
	}

	// Check allowance
	allowanceData, err := usdc.EncodeAllowance(from, spender)
	if err != nil {
		fmt.Printf("Error encoding allowance: %v\n", err)
	} else {
		fmt.Printf("Check allowance from %s to %s\n", from, spender)
		fmt.Printf("Call data: 0x%s\n", hex.EncodeToString(allowanceData))
	}

	fmt.Println("\n--- DAI Operations ---")

	// Transfer 50.5 DAI
	daiTransferAmount, _ := dai.ParseAmount("50.5")
	daiTransferData, err := dai.EncodeTransfer(to, daiTransferAmount)
	if err != nil {
		fmt.Printf("Error encoding DAI transfer: %v\n", err)
	} else {
		fmt.Printf("Transfer 50.5 DAI to %s\n", to)
		fmt.Printf("Call data: 0x%s\n", hex.EncodeToString(daiTransferData))
	}

	// TransferFrom example
	transferFromAmount, _ := usdc.ParseAmount("25.0")
	transferFromData, err := usdc.EncodeTransferFrom(from, to, transferFromAmount)
	if err != nil {
		fmt.Printf("Error encoding transferFrom: %v\n", err)
	} else {
		fmt.Printf("TransferFrom 25 USDC from %s to %s\n", from, to)
		fmt.Printf("Call data: 0x%s\n", hex.EncodeToString(transferFromData))
	}

	fmt.Println("\n--- Amount Formatting ---")

	// Format different amounts
	amount1 := big.NewInt(1000000)    // 1 USDC
	amount2 := big.NewInt(1500000)    // 1.5 USDC
	amount3 := big.NewInt(1000000000) // 1000 USDC

	fmt.Printf("1000000 raw = %s USDC\n", usdc.FormatAmount(amount1))
	fmt.Printf("1500000 raw = %s USDC\n", usdc.FormatAmount(amount2))
	fmt.Printf("1000000000 raw = %s USDC\n", usdc.FormatAmount(amount3))

	// DAI formatting (18 decimals)
	daiAmount1 := big.NewInt(1000000000000000000) // 1 DAI
	daiAmount2, _ := dai.ParseAmount("123.456789012345678901")
	fmt.Printf("1000000000000000000 raw = %s DAI\n", dai.FormatAmount(daiAmount1))
	fmt.Printf("Large precision DAI = %s DAI\n", dai.FormatAmount(daiAmount2))

	fmt.Println("\n--- Event Decoding Example ---")

	// Simulate a transfer event log
	topics := []string{
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", // Transfer event signature
		"0x000000000000000000000000742d35cc6634c0532925a3b8d82c28d53e01bcf2", // from
		"0x0000000000000000000000008ba1f109551bd432803012645hac136c63f5e5",   // to
	}
	logData := "0x00000000000000000000000000000000000000000000000000000000000f4240" // 1 USDC

	transferEvent, err := usdc.DecodeTransferEvent(logData, topics)
	if err != nil {
		fmt.Printf("Error decoding transfer event: %v\n", err)
	} else {
		fmt.Printf("Transfer Event Decoded:\n")
		fmt.Printf("  From: %s\n", transferEvent.From)
		fmt.Printf("  To: %s\n", transferEvent.To)
		fmt.Printf("  Amount: %s USDC\n", usdc.FormatAmount(transferEvent.Amount))
	}
}
