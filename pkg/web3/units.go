package web3

import (
	"fmt"
	"math/big"
	"strings"
)

const (
	WeiPerEther = 1e18
	WeiPerGwei  = 1e9
)

// maxCachedDecimals covers every decimals value that fits a uint256 amount.
const maxCachedDecimals = 77

var pow10Cache = func() [maxCachedDecimals + 1]*big.Int {
	var cache [maxCachedDecimals + 1]*big.Int
	ten := big.NewInt(10)
	cache[0] = big.NewInt(1)
	for i := 1; i <= maxCachedDecimals; i++ {
		cache[i] = new(big.Int).Mul(cache[i-1], ten)
	}
	return cache
}()

// pow10 returns 10^n. Cached values are shared and must not be modified.
func pow10(n int) *big.Int {
	if n <= maxCachedDecimals {
		return pow10Cache[n]
	}
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

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

// ParseEther parses a decimal ether amount (e.g. "1.5") into wei exactly.
func ParseEther(etherStr string) (*big.Int, error) {
	wei, err := ParseUnits(etherStr, 18)
	if err != nil {
		return nil, fmt.Errorf("invalid ether amount: %w", err)
	}
	return wei, nil
}

func FormatEther(wei *big.Int, decimals int) string {
	return WeiToEther(wei).Text('f', decimals)
}

func FormatGwei(wei *big.Int, decimals int) string {
	return WeiToGwei(wei).Text('f', decimals)
}

// ParseUnits parses a decimal string into an integer scaled by 10^decimals.
// Fractional digits beyond decimals are truncated.
func ParseUnits(amount string, decimals int) (*big.Int, error) {
	if decimals < 0 {
		return nil, fmt.Errorf("decimals must be non-negative")
	}

	integerPart, fractionalPart, _ := strings.Cut(amount, ".")
	if strings.Contains(fractionalPart, ".") {
		return nil, fmt.Errorf("invalid amount format")
	}

	if len(fractionalPart) > decimals {
		fractionalPart = fractionalPart[:decimals]
	}

	var sb strings.Builder
	sb.Grow(len(integerPart) + decimals)
	sb.WriteString(integerPart)
	sb.WriteString(fractionalPart)
	for i := len(fractionalPart); i < decimals; i++ {
		sb.WriteByte('0')
	}

	result, ok := new(big.Int).SetString(sb.String(), 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse amount")
	}

	return result, nil
}

// FormatUnits formats an integer amount scaled by 10^decimals as a decimal
// string, trimming trailing fractional zeros.
func FormatUnits(amount *big.Int, decimals int) string {
	if decimals <= 0 {
		return amount.String()
	}

	integerPart, remainder := new(big.Int).QuoRem(amount, pow10(decimals), new(big.Int))

	sign := ""
	if amount.Sign() < 0 {
		sign = "-"
		integerPart.Abs(integerPart)
		remainder.Abs(remainder)
	}

	if remainder.Sign() == 0 {
		return sign + integerPart.String()
	}

	fractionalStr := remainder.String()
	fractionalStr = strings.Repeat("0", decimals-len(fractionalStr)) + fractionalStr
	fractionalStr = strings.TrimRight(fractionalStr, "0")

	return sign + integerPart.String() + "." + fractionalStr
}
