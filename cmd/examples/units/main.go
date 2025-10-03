package main

// Run with: go run cmd/examples/units/main.go

import (
	"fmt"
	"math/big"

	"github.com/kimdonghyun/go-blockchain-helper/pkg/web3"
)

func main() {
	fmt.Println("=== Unit Conversion Examples ===")

	// Convert Ether to Wei
	ethAmount := 1.5
	weiAmount := web3.EtherToWei(ethAmount)
	fmt.Printf("1.5 ETH = %s Wei\n", weiAmount.String())

	// Convert Wei to Ether
	wei := big.NewInt(1500000000000000000) // 1.5 ETH in Wei
	fmt.Printf("%s Wei = %s ETH\n", wei.String(), web3.FormatEther(wei, 4))

	// Convert Gwei to Wei
	gweiAmount := 20.0
	weiFromGwei := web3.GweiToWei(gweiAmount)
	fmt.Printf("20 Gwei = %s Wei\n", weiFromGwei.String())

	// Convert Wei to Gwei
	weiForGas := big.NewInt(20000000000) // 20 Gwei
	fmt.Printf("%s Wei = %s Gwei\n", weiForGas.String(), web3.FormatGwei(weiForGas, 2))

	// Parse Ether string
	weiFromStr, err := web3.ParseEther("2.5")
	if err != nil {
		fmt.Printf("Error parsing ether: %v\n", err)
	} else {
		fmt.Printf("2.5 ETH = %s Wei\n", weiFromStr.String())
	}

	// Parse and format custom units (like token amounts)
	// Example: USDC has 6 decimals
	usdcAmount, err := web3.ParseUnits("100.50", 6)
	if err != nil {
		fmt.Printf("Error parsing USDC amount: %v\n", err)
	} else {
		fmt.Printf("100.50 USDC = %s (raw units)\n", usdcAmount.String())
		formatted := web3.FormatUnits(usdcAmount, 6)
		fmt.Printf("Formatted back: %s USDC\n", formatted)
	}

	// Example with different decimal places
	// DAI has 18 decimals
	daiAmount, err := web3.ParseUnits("50.123456789012345678", 18)
	if err != nil {
		fmt.Printf("Error parsing DAI amount: %v\n", err)
	} else {
		fmt.Printf("50.123456789012345678 DAI = %s (raw units)\n", daiAmount.String())
		formatted := web3.FormatUnits(daiAmount, 18)
		fmt.Printf("Formatted back: %s DAI\n", formatted)
	}
}