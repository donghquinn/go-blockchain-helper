package web3

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type ABIParam struct {
	Name string
	Type string
}

type ABIFunction struct {
	Name    string
	Inputs  []ABIParam
	Outputs []ABIParam
}

type ABIEvent struct {
	Name   string
	Inputs []ABIParam
}

var abiSignatureRegexp = regexp.MustCompile(`^(\w+)\((.*)\)$`)

var (
	two256    = new(big.Int).Lsh(big.NewInt(1), 256)
	maxInt256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))
	minInt256 = new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 255))
)

func EncodeFunctionCall(funcName string, params []ABIParam, values []interface{}) ([]byte, error) {
	signature := createFunctionSignature(funcName, params)

	types := make([]string, len(params))
	for i, param := range params {
		types[i] = param.Type
	}

	encodedParams, err := encodeTuple(types, values)
	if err != nil {
		return nil, fmt.Errorf("failed to encode parameters: %w", err)
	}

	data := make([]byte, 4+len(encodedParams))
	copy(data, keccak256([]byte(signature))[:4])
	copy(data[4:], encodedParams)
	return data, nil
}

func createFunctionSignature(funcName string, params []ABIParam) string {
	var sb strings.Builder
	sb.WriteString(funcName)
	sb.WriteByte('(')
	for i, param := range params {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(param.Type)
	}
	sb.WriteByte(')')
	return sb.String()
}

// encodeTuple ABI-encodes values as a tuple: a head of one 32-byte word per
// element (the value itself, or an offset for dynamic types) followed by the
// dynamic tails.
func encodeTuple(types []string, values []interface{}) ([]byte, error) {
	if len(types) != len(values) {
		return nil, fmt.Errorf("parameter count mismatch: expected %d, got %d", len(types), len(values))
	}

	headSize := len(types) * 32
	encoded := make([]byte, headSize, headSize+64*len(types))

	for i, abiType := range types {
		value, err := encodeValue(abiType, values[i])
		if err != nil {
			return nil, fmt.Errorf("failed to encode parameter %d: %w", i, err)
		}

		head := encoded[i*32 : (i+1)*32]
		if isDynamicType(abiType) {
			putUint64(head, uint64(len(encoded)))
			encoded = append(encoded, value...)
		} else {
			copy(head, value)
		}
	}

	return encoded, nil
}

func isDynamicType(abiType string) bool {
	return strings.HasSuffix(abiType, "[]") || abiType == "string" || abiType == "bytes"
}

func encodeValue(abiType string, value interface{}) ([]byte, error) {
	switch {
	case strings.HasSuffix(abiType, "[]"):
		return encodeArray(abiType, value)
	case abiType == "address":
		return encodeAddress(value)
	case strings.HasPrefix(abiType, "uint"):
		return encodeUint(abiType, value)
	case strings.HasPrefix(abiType, "int"):
		return encodeInt(abiType, value)
	case abiType == "bool":
		return encodeBool(value)
	case abiType == "string":
		return encodeString(value)
	case abiType == "bytes":
		return encodeBytes(value)
	default:
		return nil, fmt.Errorf("unsupported type: %s", abiType)
	}
}

func encodeAddress(value interface{}) ([]byte, error) {
	addressStr, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("address must be string")
	}

	address, ok := parseAddress(addressStr)
	if !ok {
		return nil, fmt.Errorf("invalid address format")
	}

	result := make([]byte, 32)
	putAddress(result, address)
	return result, nil
}

func toBigInt(value interface{}) (*big.Int, error) {
	switch v := value.(type) {
	case *big.Int:
		if v == nil {
			return nil, fmt.Errorf("nil integer value")
		}
		return v, nil
	case string:
		bigIntValue, ok := new(big.Int).SetString(v, 10)
		if !ok {
			return nil, fmt.Errorf("invalid integer string")
		}
		return bigIntValue, nil
	case int:
		return big.NewInt(int64(v)), nil
	case int64:
		return big.NewInt(v), nil
	case uint64:
		return new(big.Int).SetUint64(v), nil
	default:
		return nil, fmt.Errorf("unsupported integer type %T", value)
	}
}

func encodeUint(abiType string, value interface{}) ([]byte, error) {
	bigIntValue, err := toBigInt(value)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 32)
	if err := putUint256(result, bigIntValue); err != nil {
		return nil, err
	}
	return result, nil
}

func encodeInt(abiType string, value interface{}) ([]byte, error) {
	bigIntValue, err := toBigInt(value)
	if err != nil {
		return nil, err
	}

	if bigIntValue.Cmp(maxInt256) > 0 || bigIntValue.Cmp(minInt256) < 0 {
		return nil, fmt.Errorf("value out of int256 range")
	}

	result := make([]byte, 32)
	if bigIntValue.Sign() < 0 {
		// Two's complement: 2^256 + v
		new(big.Int).Add(two256, bigIntValue).FillBytes(result)
	} else {
		bigIntValue.FillBytes(result)
	}
	return result, nil
}

func encodeBool(value interface{}) ([]byte, error) {
	var boolValue bool
	switch v := value.(type) {
	case bool:
		boolValue = v
	case string:
		var err error
		boolValue, err = strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("invalid bool string")
		}
	default:
		return nil, fmt.Errorf("unsupported bool type")
	}

	result := make([]byte, 32)
	if boolValue {
		result[31] = 1
	}
	return result, nil
}

func encodeString(value interface{}) ([]byte, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("string value must be string type")
	}
	return encodeDynamicBytes([]byte(str)), nil
}

func encodeBytes(value interface{}) ([]byte, error) {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		var err error
		bytes, err = hex.DecodeString(strings.TrimPrefix(v, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid hex string")
		}
	default:
		return nil, fmt.Errorf("bytes value must be []byte or hex string")
	}

	return encodeDynamicBytes(bytes), nil
}

// encodeDynamicBytes returns the length word followed by data right-padded to
// a multiple of 32 bytes.
func encodeDynamicBytes(data []byte) []byte {
	result := make([]byte, 32+paddedLen(len(data)))
	putUint64(result[:32], uint64(len(data)))
	copy(result[32:], data)
	return result
}

func paddedLen(n int) int {
	return (n + 31) / 32 * 32
}

func encodeArray(abiType string, value interface{}) ([]byte, error) {
	elementType := strings.TrimSuffix(abiType, "[]")

	var elements []interface{}
	switch v := value.(type) {
	case []interface{}:
		elements = v
	case []string:
		elements = make([]interface{}, len(v))
		for i, s := range v {
			elements[i] = s
		}
	default:
		return nil, fmt.Errorf("array value must be slice")
	}

	types := make([]string, len(elements))
	for i := range types {
		types[i] = elementType
	}

	encodedElements, err := encodeTuple(types, elements)
	if err != nil {
		return nil, fmt.Errorf("failed to encode array element: %w", err)
	}

	result := make([]byte, 32+len(encodedElements))
	putUint64(result[:32], uint64(len(elements)))
	copy(result[32:], encodedElements)
	return result, nil
}

func DecodeFunctionResult(abiTypes []string, data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	results := make([]interface{}, 0, len(abiTypes))
	offset := 0

	for _, abiType := range abiTypes {
		if offset+32 > len(data) {
			return nil, fmt.Errorf("insufficient data for type %s", abiType)
		}

		value, newOffset, err := decodeValue(abiType, data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to decode type %s: %w", abiType, err)
		}

		results = append(results, value)
		offset = newOffset
	}

	return results, nil
}

func decodeValue(abiType string, data []byte, offset int) (interface{}, int, error) {
	switch {
	case strings.HasSuffix(abiType, "[]"):
		return nil, 0, fmt.Errorf("unsupported decode type: %s", abiType)
	case abiType == "address":
		return decodeAddress(data, offset)
	case strings.HasPrefix(abiType, "uint"):
		return decodeUint(data, offset)
	case abiType == "bool":
		return decodeBool(data, offset)
	case abiType == "string":
		return decodeString(data, offset)
	default:
		return nil, 0, fmt.Errorf("unsupported decode type: %s", abiType)
	}
}

func decodeAddress(data []byte, offset int) (string, int, error) {
	if offset+32 > len(data) {
		return "", 0, fmt.Errorf("insufficient data for address")
	}

	address := "0x" + hex.EncodeToString(data[offset+12:offset+32])
	return address, offset + 32, nil
}

func decodeUint(data []byte, offset int) (*big.Int, int, error) {
	if offset+32 > len(data) {
		return nil, 0, fmt.Errorf("insufficient data for uint")
	}

	value := new(big.Int).SetBytes(data[offset : offset+32])
	return value, offset + 32, nil
}

func decodeBool(data []byte, offset int) (bool, int, error) {
	if offset+32 > len(data) {
		return false, 0, fmt.Errorf("insufficient data for bool")
	}

	value := data[offset+31] != 0
	return value, offset + 32, nil
}

func decodeString(data []byte, offset int) (string, int, error) {
	stringOffset, err := readWordAsLength(data, offset, len(data))
	if err != nil {
		return "", 0, fmt.Errorf("invalid string offset: %w", err)
	}

	length, err := readWordAsLength(data, stringOffset, len(data)-stringOffset-32)
	if err != nil {
		return "", 0, fmt.Errorf("invalid string length: %w", err)
	}

	start := stringOffset + 32
	return string(data[start : start+length]), offset + 32, nil
}

// readWordAsLength reads the 32-byte word at offset as a non-negative integer
// no greater than limit, guarding against huge or malformed values.
func readWordAsLength(data []byte, offset, limit int) (int, error) {
	if offset < 0 || offset+32 > len(data) {
		return 0, fmt.Errorf("insufficient data")
	}

	word := data[offset : offset+32]
	for _, b := range word[:24] {
		if b != 0 {
			return 0, fmt.Errorf("value too large")
		}
	}

	value := binary.BigEndian.Uint64(word[24:])
	if limit < 0 || value > uint64(limit) {
		return 0, fmt.Errorf("value %d out of bounds", value)
	}
	return int(value), nil
}

func ParseABISignature(signature string) (*ABIFunction, error) {
	matches := abiSignatureRegexp.FindStringSubmatch(signature)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid function signature format")
	}

	funcName := matches[1]
	paramStr := matches[2]

	var params []ABIParam
	if paramStr != "" {
		paramTypes := strings.Split(paramStr, ",")
		params = make([]ABIParam, len(paramTypes))
		for i, paramType := range paramTypes {
			params[i] = ABIParam{
				Name: "param" + strconv.Itoa(i),
				Type: strings.TrimSpace(paramType),
			}
		}
	}

	return &ABIFunction{
		Name:   funcName,
		Inputs: params,
	}, nil
}

// parseAddress validates a 0x-prefixed 20-byte hex address and decodes it.
func parseAddress(address string) ([20]byte, bool) {
	var result [20]byte
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return result, false
	}
	if _, err := hex.Decode(result[:], []byte(address[2:])); err != nil {
		return result, false
	}
	return result, true
}

func mustSelector(selectorHex string) [4]byte {
	var selector [4]byte
	if n, err := hex.Decode(selector[:], []byte(selectorHex)); err != nil || n != 4 {
		panic("invalid selector: " + selectorHex)
	}
	return selector
}

// newCallData allocates calldata for a selector followed by the given number
// of 32-byte words.
func newCallData(selector [4]byte, words int) []byte {
	data := make([]byte, 4+32*words)
	copy(data, selector[:])
	return data
}

// word returns the i-th 32-byte argument word of calldata.
func word(callData []byte, i int) []byte {
	return callData[4+32*i : 4+32*(i+1)]
}

func putAddress(dst []byte, address [20]byte) {
	copy(dst[12:32], address[:])
}

func putUint64(dst []byte, n uint64) {
	binary.BigEndian.PutUint64(dst[24:32], n)
}

func putUint256(dst []byte, n *big.Int) error {
	if n == nil {
		return fmt.Errorf("nil integer value")
	}
	if n.Sign() < 0 {
		return fmt.Errorf("negative value for unsigned integer")
	}
	if n.BitLen() > 256 {
		return fmt.Errorf("value exceeds uint256")
	}
	n.FillBytes(dst[:32])
	return nil
}
