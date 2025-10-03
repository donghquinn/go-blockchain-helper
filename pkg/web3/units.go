package web3

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	WeiPerEther = 1e18
	WeiPerGwei  = 1e9
)

func WeiToEther(wei *big.Int) *big.Float {
	ether := new(big.Float).SetInt(wei)
	return ether.Quo(ether, big.NewFloat(WeiPerEther))
}

func EtherToWei(ether float64) *big.Int {
	etherBig := big.NewFloat(ether)
	wei := new(big.Float).Mul(etherBig, big.NewFloat(WeiPerEther))
	result, _ := wei.Int(nil)
	return result
}

func WeiToGwei(wei *big.Int) *big.Float {
	gwei := new(big.Float).SetInt(wei)
	return gwei.Quo(gwei, big.NewFloat(WeiPerGwei))
}

func GweiToWei(gwei float64) *big.Int {
	gweiFloat := big.NewFloat(gwei)
	wei := new(big.Float).Mul(gweiFloat, big.NewFloat(WeiPerGwei))
	result, _ := wei.Int(nil)
	return result
}

func ParseEther(etherStr string) (*big.Int, error) {
	ether, err := strconv.ParseFloat(etherStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ether amount: %w", err)
	}
	return EtherToWei(ether), nil
}

func FormatEther(wei *big.Int, decimals int) string {
	ether := WeiToEther(wei)
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, ether)
}

func FormatGwei(wei *big.Int, decimals int) string {
	gwei := WeiToGwei(wei)
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, gwei)
}

func ParseUnits(amount string, decimals int) (*big.Int, error) {
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid amount format")
	}

	integerPart := parts[0]
	fractionalPart := ""
	if len(parts) == 2 {
		fractionalPart = parts[1]
	}

	if len(fractionalPart) > decimals {
		fractionalPart = fractionalPart[:decimals]
	}

	for len(fractionalPart) < decimals {
		fractionalPart += "0"
	}

	fullAmount := integerPart + fractionalPart
	result, ok := new(big.Int).SetString(fullAmount, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse amount")
	}

	return result, nil
}

func FormatUnits(amount *big.Int, decimals int) string {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	integerPart := new(big.Int).Div(amount, divisor)
	remainder := new(big.Int).Mod(amount, divisor)

	if remainder.Cmp(big.NewInt(0)) == 0 {
		return integerPart.String()
	}

	fractionalStr := remainder.String()
	for len(fractionalStr) < decimals {
		fractionalStr = "0" + fractionalStr
	}

	fractionalStr = strings.TrimRight(fractionalStr, "0")
	if fractionalStr == "" {
		return integerPart.String()
	}

	return integerPart.String() + "." + fractionalStr
}
