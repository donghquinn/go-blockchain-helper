# go-blockchain-helper

A comprehensive Go library for Ethereum and Web3 development without external dependencies. This library provides essential utilities for blockchain development including unit conversions, transaction handling, ERC token interactions, event processing, and ABI encoding/decoding.

## Features

- **🔄 Unit Conversions**: Wei, Gwei, Ether conversions and formatting
- **📝 Transaction Helpers**: Gas estimation, transaction creation, and validation
- **🪙 ERC-20 Support**: Complete ERC-20 token interaction utilities
- **🎨 ERC-721 Support**: NFT transfer, approval, and metadata functions
- **🎧 Event Listeners**: Event filtering, subscription management, and parsing
- **⚙️ ABI Encoding/Decoding**: Function call encoding and result decoding
- **🔐 Cryptographic Utilities**: Address validation, private key management
- **🚀 Zero Dependencies**: Built using only Go standard library

## Installation

```bash
go get github.com/kimdonghyun/go-blockchain-helper
```

## Quick Start

### Unit Conversions

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/kimdonghyun/go-blockchain-helper/pkg/web3"
)

func main() {
    // Convert Ether to Wei
    ethAmount := 1.5
    weiAmount := web3.EtherToWei(ethAmount)
    fmt.Printf("1.5 ETH = %s Wei\n", weiAmount.String())

    // Convert Wei to Ether
    wei := big.NewInt(1500000000000000000) // 1.5 ETH in Wei
    ether := web3.WeiToEther(wei)
    fmt.Printf("%s Wei = %s ETH\n", wei.String(), web3.FormatEther(wei, 4))

    // Parse Ether string
    weiFromStr, _ := web3.ParseEther("2.5")
    fmt.Printf("2.5 ETH = %s Wei\n", weiFromStr.String())
}
```

### Transaction Handling

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/kimdonghyun/go-blockchain-helper/pkg/web3"
)

func main() {
    // Create a transaction
    to := "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2"
    value := web3.EtherToWei(1.0) // 1 ETH
    data := []byte{}
    
    tx := web3.CreateTransaction(to, value, data)
    fmt.Printf("Transaction Gas Limit: %d\n", tx.Gas)
    fmt.Printf("Transaction Fee: %s Wei\n", tx.CalculateFee().String())

    // Validate addresses
    if web3.ValidateAddress(to) {
        fmt.Println("Address is valid")
    }

    // Generate private key and derive address
    privateKey := web3.GenerateRandomPrivateKey()
    address, _ := web3.PrivateKeyToAddress(privateKey)
    fmt.Printf("Private Key: %s\n", privateKey)
    fmt.Printf("Address: %s\n", address)
}
```

### ERC-20 Token Operations

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/kimdonghyun/go-blockchain-helper/pkg/web3"
)

func main() {
    // Create ERC-20 token instance
    token := web3.NewERC20Token(
        "0xA0b86a33E6440417C4eE5a3C2c6e9a8b8De2fD2A", // USDC
        "USD Coin",
        "USDC",
        6, // 6 decimals
    )

    // Encode transfer function call
    to := "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2"
    amount := big.NewInt(1000000) // 1 USDC (6 decimals)
    
    transferData, _ := token.EncodeTransfer(to, amount)
    fmt.Printf("Transfer function call data: %x\n", transferData)

    // Encode approve function call
    spender := "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2"
    approveAmount := big.NewInt(1000000000) // 1000 USDC
    
    approveData, _ := token.EncodeApprove(spender, approveAmount)
    fmt.Printf("Approve function call data: %x\n", approveData)

    // Format token amounts
    amount = big.NewInt(1500000) // 1.5 USDC
    formatted := token.FormatAmount(amount)
    fmt.Printf("Formatted amount: %s USDC\n", formatted)
}
```

### ERC-721 NFT Operations

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/kimdonghyun/go-blockchain-helper/pkg/web3"
)

func main() {
    // Create ERC-721 token instance
    nft := web3.NewERC721Token(
        "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D", // BAYC
        "Bored Ape Yacht Club",
        "BAYC",
    )

    // Encode transfer function call
    from := "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2"
    to := "0x8ba1f109551bD432803012645Hac136c63F5E5"
    tokenId := big.NewInt(1234)
    
    transferData, _ := nft.EncodeTransferFrom(from, to, tokenId)
    fmt.Printf("NFT transfer function call data: %x\n", transferData)

    // Encode approve function call
    approved := "0x8ba1f109551bD432803012645Hac136c63F5E5"
    approveData, _ := nft.EncodeApprove(approved, tokenId)
    fmt.Printf("NFT approve function call data: %x\n", approveData)

    // Encode balance check
    owner := "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2"
    balanceData, _ := nft.EncodeBalanceOf(owner)
    fmt.Printf("NFT balance check data: %x\n", balanceData)
}
```

### Event Processing

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/kimdonghyun/go-blockchain-helper/pkg/web3"
)

func main() {
    // Create event filter
    filter := web3.NewEventFilter()
    filter.AddAddress("0xA0b86a33E6440417C4eE5a3C2c6e9a8b8De2fD2A") // USDC
    filter.AddTopic(web3.ERC20_TRANSFER_SIGNATURE)

    // Create event monitor
    monitor := web3.NewEventMonitor()
    
    // Add event handler
    monitor.AddEventHandler(web3.ERC20_TRANSFER_SIGNATURE, func(event web3.Event) error {
        transferEvent, err := web3.ParseTransferEvent(event)
        if err != nil {
            return err
        }
        
        fmt.Printf("Transfer: %s -> %s, Amount: %s\n", 
            transferEvent.From, 
            transferEvent.To, 
            transferEvent.Amount.String())
        return nil
    })

    // Subscribe to events
    subscription := monitor.Subscribe(filter)
    fmt.Printf("Subscribed with ID: %s\n", subscription.ID)

    // Simulate processing an event
    sampleEvent := web3.Event{
        Address: "0xA0b86a33E6440417C4eE5a3C2c6e9a8b8De2fD2A",
        Topics: []string{
            web3.ERC20_TRANSFER_SIGNATURE,
            "0x000000000000000000000000742d35cc6634c0532925a3b8d82c28d53e01bcf2",
            "0x0000000000000000000000008ba1f109551bd432803012645hac136c63f5e5",
        },
        Data:        "0x00000000000000000000000000000000000000000000000000000000000f4240",
        BlockNumber: big.NewInt(18500000),
    }
    
    monitor.ProcessEvent(sampleEvent)
}
```

### ABI Encoding and Decoding

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/kimdonghyun/go-blockchain-helper/pkg/web3"
)

func main() {
    // Parse function signature
    funcSig, _ := web3.ParseABISignature("transfer(address,uint256)")
    fmt.Printf("Function: %s\n", funcSig.Name)

    // Encode function call
    params := []web3.ABIParam{
        {Name: "to", Type: "address"},
        {Name: "amount", Type: "uint256"},
    }
    values := []interface{}{
        "0x742d35Cc6634C0532925a3b8D82C28d53e01BCf2",
        big.NewInt(1000000),
    }
    
    encodedData, _ := web3.EncodeFunctionCall("transfer", params, values)
    fmt.Printf("Encoded function call: %x\n", encodedData)

    // Decode function result
    resultTypes := []string{"uint256"}
    resultData := []byte{
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0f, 0x42, 0x40,
    }
    
    results, _ := web3.DecodeFunctionResult(resultTypes, resultData)
    if len(results) > 0 {
        balance := results[0].(*big.Int)
        fmt.Printf("Decoded balance: %s\n", balance.String())
    }
}
```

## API Reference

### Unit Conversion Functions

- `EtherToWei(ether float64) *big.Int`
- `WeiToEther(wei *big.Int) *big.Float`
- `GweiToWei(gwei float64) *big.Int`
- `WeiToGwei(wei *big.Int) *big.Float`
- `ParseEther(etherStr string) (*big.Int, error)`
- `FormatEther(wei *big.Int, decimals int) string`
- `ParseUnits(amount string, decimals int) (*big.Int, error)`
- `FormatUnits(amount *big.Int, decimals int) string`

### Transaction Functions

- `EstimateGas(to, from, data string, value *big.Int) (uint64, error)`
- `SuggestGasPrice() *big.Int`
- `CreateTransaction(to string, value *big.Int, data []byte) *Transaction`
- `ValidateAddress(address string) bool`
- `ValidatePrivateKey(privateKey string) bool`
- `PrivateKeyToAddress(privateKeyHex string) (string, error)`
- `GenerateRandomPrivateKey() string`

### ERC-20 Token Methods

- `EncodeTransfer(to string, amount *big.Int) ([]byte, error)`
- `EncodeTransferFrom(from, to string, amount *big.Int) ([]byte, error)`
- `EncodeApprove(spender string, amount *big.Int) ([]byte, error)`
- `EncodeBalanceOf(owner string) ([]byte, error)`
- `EncodeAllowance(owner, spender string) ([]byte, error)`
- `DecodeTransferEvent(logData string, topics []string) (*TransferEvent, error)`

### ERC-721 NFT Methods

- `EncodeTransferFrom(from, to string, tokenId *big.Int) ([]byte, error)`
- `EncodeSafeTransferFrom(from, to string, tokenId *big.Int, data []byte) ([]byte, error)`
- `EncodeApprove(to string, tokenId *big.Int) ([]byte, error)`
- `EncodeSetApprovalForAll(operator string, approved bool) ([]byte, error)`
- `EncodeOwnerOf(tokenId *big.Int) ([]byte, error)`
- `EncodeBalanceOf(owner string) ([]byte, error)`

### Event Processing

- `NewEventFilter() *EventFilter`
- `NewEventMonitor() *EventMonitor`
- `CreateEventSignature(eventName string, paramTypes []string) string`
- `ParseTransferEvent(log Event) (*TransferEvent, error)`
- `ParseNFTTransferEvent(log Event) (*NFTTransferEvent, error)`

### ABI Encoding/Decoding

- `EncodeFunctionCall(funcName string, params []ABIParam, values []interface{}) ([]byte, error)`
- `DecodeFunctionResult(abiTypes []string, data []byte) ([]interface{}, error)`
- `ParseABISignature(signature string) (*ABIFunction, error)`

## Testing

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Ethereum Foundation for the EIP standards
- Go community for excellent tooling and libraries

## Support

If you find this library useful, please consider starring the repository and sharing it with others!

For questions, issues, or contributions, please visit our [GitHub repository](https://github.com/kimdonghyun/go-blockchain-helper).
