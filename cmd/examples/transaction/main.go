package main

// Run with: go run cmd/examples/transaction/main.go

import (
	"fmt"
	"math/big"

	"github.com/donghquinn/go-blockchain-helper/pkg/web3"
)

func main() {
	fmt.Println("=== Transaction Examples ===")

	// Example addresses
	to := "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2"
	from := "0x8ba1f109551bD432803012645Hac136c63F5E5"

	// Validate addresses
	fmt.Println("\n--- Address Validation ---")
	fmt.Printf("Address %s is valid: %t\n", to, web3.ValidateAddress(to))
	fmt.Printf("Address %s is valid: %t\n", "invalid-address", web3.ValidateAddress("invalid-address"))
	fmt.Printf("Address %s is valid: %t\n", "0x123", web3.ValidateAddress("0x123"))

	// Create simple ETH transfer transaction
	fmt.Println("\n--- Simple ETH Transfer ---")
	value := web3.EtherToWei(1.5) // 1.5 ETH
	data := []byte{}

	tx := web3.CreateTransaction(to, value, data)
	fmt.Printf("Transaction Details:\n")
	fmt.Printf("  To: %s\n", tx.To)
	fmt.Printf("  Value: %s Wei (%s ETH)\n", tx.Value.String(), web3.FormatEther(tx.Value, 4))
	fmt.Printf("  Gas Limit: %d\n", tx.Gas)
	fmt.Printf("  Gas Price: %s Wei (%s Gwei)\n", tx.GasPrice.String(), web3.FormatGwei(tx.GasPrice, 2))
	fmt.Printf("  Transaction Fee: %s Wei (%s ETH)\n", tx.CalculateFee().String(), web3.FormatEther(tx.CalculateFee(), 6))
	fmt.Printf("  Hash: %s\n", tx.Hash())

	// Contract interaction transaction
	fmt.Println("\n--- Contract Interaction Transaction ---")

	// Create some sample contract call data (ERC-20 transfer)
	contractData := []byte{
		0xa9, 0x05, 0x9c, 0xbb, // transfer(address,uint256) selector
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // padding
		0x74, 0x2d, 0x35, 0xcc, 0x66, 0x34, 0xc0, 0x53, 0x29, 0x25, 0xa3, 0xb8, 0xd8, 0x2c, 0x28, 0xd5, 0x3e, 0x01, 0xbc, 0xf2, // to address
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0f, 0x42, 0x40, // amount (1000000 = 1 USDC)
	}

	contractTx := web3.CreateTransaction(to, big.NewInt(0), contractData)
	fmt.Printf("Contract Transaction Details:\n")
	fmt.Printf("  To: %s\n", contractTx.To)
	fmt.Printf("  Value: %s Wei\n", contractTx.Value.String())
	fmt.Printf("  Data Length: %d bytes\n", len(contractTx.Data))
	fmt.Printf("  Gas Limit: %d\n", contractTx.Gas)
	fmt.Printf("  Gas Price: %s Wei (%s Gwei)\n", contractTx.GasPrice.String(), web3.FormatGwei(contractTx.GasPrice, 2))
	fmt.Printf("  Transaction Fee: %s Wei (%s ETH)\n", contractTx.CalculateFee().String(), web3.FormatEther(contractTx.CalculateFee(), 6))

	// Gas estimation examples
	fmt.Println("\n--- Gas Estimation ---")

	// Simple ETH transfer
	gasETH, err := web3.EstimateGas(to, from, "", value)
	if err != nil {
		fmt.Printf("Error estimating gas for ETH transfer: %v\n", err)
	} else {
		fmt.Printf("Estimated gas for ETH transfer: %d\n", gasETH)
	}

	// Contract call
	gasContract, err := web3.EstimateGas(to, from, "0xa9059cbb000000000000000000000000742d35cc6634c0532925a3b8d82c28d53e01bcf200000000000000000000000000000000000000000000000000000000000f4240", big.NewInt(0))
	if err != nil {
		fmt.Printf("Error estimating gas for contract call: %v\n", err)
	} else {
		fmt.Printf("Estimated gas for contract call: %d\n", gasContract)
	}

	// Private key and address generation
	fmt.Println("\n--- Private Key & Address Generation ---")

	privateKey := web3.GenerateRandomPrivateKey()
	fmt.Printf("Generated Private Key: %s\n", privateKey)

	if web3.ValidatePrivateKey(privateKey) {
		fmt.Println("Private key is valid")

		address, err := web3.PrivateKeyToAddress(privateKey)
		if err != nil {
			fmt.Printf("Error deriving address: %v\n", err)
		} else {
			fmt.Printf("Derived Address: %s\n", address)
		}

		// Derive public key
		pubKey, err := web3.PrivateKeyToPublicKey(privateKey)
		if err != nil {
			fmt.Printf("Error deriving public key: %v\n", err)
		} else {
			fmt.Printf("Public Key X: %s\n", pubKey.X.String())
			fmt.Printf("Public Key Y: %s\n", pubKey.Y.String())
		}
	}

	// Gas price suggestion
	fmt.Println("\n--- Gas Price Suggestion ---")
	suggestedGasPrice := web3.SuggestGasPrice()
	fmt.Printf("Suggested Gas Price: %s Wei (%s Gwei)\n", suggestedGasPrice.String(), web3.FormatGwei(suggestedGasPrice, 2))

	// Transaction with custom gas settings
	fmt.Println("\n--- Custom Gas Transaction ---")
	customTx := &web3.Transaction{
		To:       to,
		Value:    web3.EtherToWei(0.1),
		Gas:      25000,
		GasPrice: web3.GweiToWei(25.0), // 25 Gwei
		Data:     []byte{},
		Nonce:    42,
	}

	fmt.Printf("Custom Transaction:\n")
	fmt.Printf("  Gas Limit: %d\n", customTx.Gas)
	fmt.Printf("  Gas Price: %s Gwei\n", web3.FormatGwei(customTx.GasPrice, 2))
	fmt.Printf("  Total Fee: %s ETH\n", web3.FormatEther(customTx.CalculateFee(), 6))
	fmt.Printf("  Nonce: %d\n", customTx.Nonce)
}
