package web3

import (
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

func EncodeFunctionCall(funcName string, params []ABIParam, values []interface{}) ([]byte, error) {
	signature := createFunctionSignature(funcName, params)
	selector := Keccak256([]byte(signature))[:8]

	selectorBytes, err := hex.DecodeString(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to decode selector: %w", err)
	}

	encodedParams, err := encodeParameters(params, values)
	if err != nil {
		return nil, fmt.Errorf("failed to encode parameters: %w", err)
	}

	return append(selectorBytes, encodedParams...), nil
}

func createFunctionSignature(funcName string, params []ABIParam) string {
	var paramTypes []string
	for _, param := range params {
		paramTypes = append(paramTypes, param.Type)
	}
	return funcName + "(" + strings.Join(paramTypes, ",") + ")"
}

func encodeParameters(params []ABIParam, values []interface{}) ([]byte, error) {
	if len(params) != len(values) {
		return nil, fmt.Errorf("parameter count mismatch: expected %d, got %d", len(params), len(values))
	}

	var encoded []byte
	var dynamicData []byte
	dynamicOffset := len(params) * 32

	for i, param := range params {
		value := values[i]

		if isDynamicType(param.Type) {
			offsetBytes := make([]byte, 32)
			big.NewInt(int64(dynamicOffset)).FillBytes(offsetBytes)
			encoded = append(encoded, offsetBytes...)

			dynamicEncoded, err := encodeValue(param.Type, value)
			if err != nil {
				return nil, fmt.Errorf("failed to encode dynamic parameter %d: %w", i, err)
			}
			dynamicData = append(dynamicData, dynamicEncoded...)
			dynamicOffset += len(dynamicEncoded)
		} else {
			staticEncoded, err := encodeValue(param.Type, value)
			if err != nil {
				return nil, fmt.Errorf("failed to encode static parameter %d: %w", i, err)
			}
			encoded = append(encoded, staticEncoded...)
		}
	}

	return append(encoded, dynamicData...), nil
}

func isDynamicType(abiType string) bool {
	return strings.Contains(abiType, "[]") || abiType == "string" || abiType == "bytes"
}

func encodeValue(abiType string, value interface{}) ([]byte, error) {
	switch {
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
	case strings.HasSuffix(abiType, "[]"):
		return encodeArray(abiType, value)
	default:
		return nil, fmt.Errorf("unsupported type: %s", abiType)
	}
}

func encodeAddress(value interface{}) ([]byte, error) {
	var addressStr string
	switch v := value.(type) {
	case string:
		addressStr = v
	default:
		return nil, fmt.Errorf("address must be string")
	}

	if !ValidateAddress(addressStr) {
		return nil, fmt.Errorf("invalid address format")
	}

	addressBytes, _ := hex.DecodeString(strings.TrimPrefix(addressStr, "0x"))
	result := make([]byte, 32)
	copy(result[12:], addressBytes)
	return result, nil
}

func encodeUint(abiType string, value interface{}) ([]byte, error) {
	var bigIntValue *big.Int

	switch v := value.(type) {
	case *big.Int:
		bigIntValue = v
	case string:
		var ok bool
		bigIntValue, ok = new(big.Int).SetString(v, 10)
		if !ok {
			return nil, fmt.Errorf("invalid uint string")
		}
	case int:
		bigIntValue = big.NewInt(int64(v))
	case int64:
		bigIntValue = big.NewInt(v)
	case uint64:
		bigIntValue = new(big.Int).SetUint64(v)
	default:
		return nil, fmt.Errorf("unsupported uint type")
	}

	result := make([]byte, 32)
	bigIntValue.FillBytes(result)
	return result, nil
}

func encodeInt(abiType string, value interface{}) ([]byte, error) {
	return encodeUint(abiType, value)
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
	var str string
	switch v := value.(type) {
	case string:
		str = v
	default:
		return nil, fmt.Errorf("string value must be string type")
	}

	strBytes := []byte(str)
	length := make([]byte, 32)
	big.NewInt(int64(len(strBytes))).FillBytes(length)

	paddedBytes := padBytes(strBytes, 32)

	return append(length, paddedBytes...), nil
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

	length := make([]byte, 32)
	big.NewInt(int64(len(bytes))).FillBytes(length)

	paddedBytes := padBytes(bytes, 32)

	return append(length, paddedBytes...), nil
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

	length := make([]byte, 32)
	big.NewInt(int64(len(elements))).FillBytes(length)

	var encodedElements []byte
	for _, element := range elements {
		encoded, err := encodeValue(elementType, element)
		if err != nil {
			return nil, fmt.Errorf("failed to encode array element: %w", err)
		}
		encodedElements = append(encodedElements, encoded...)
	}

	return append(length, encodedElements...), nil
}

func padBytes(data []byte, blockSize int) []byte {
	remainder := len(data) % blockSize
	if remainder == 0 {
		return data
	}

	padding := blockSize - remainder
	padded := make([]byte, len(data)+padding)
	copy(padded, data)

	return padded
}

func DecodeFunctionResult(abiTypes []string, data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	var results []interface{}
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

	addressBytes := data[offset+12 : offset+32]
	address := "0x" + hex.EncodeToString(addressBytes)

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
	if offset+32 > len(data) {
		return "", 0, fmt.Errorf("insufficient data for string offset")
	}

	stringOffset := new(big.Int).SetBytes(data[offset : offset+32]).Int64()
	if int(stringOffset)+32 > len(data) {
		return "", 0, fmt.Errorf("insufficient data for string length")
	}

	length := new(big.Int).SetBytes(data[stringOffset : stringOffset+32]).Int64()
	if int(stringOffset)+32+int(length) > len(data) {
		return "", 0, fmt.Errorf("insufficient data for string content")
	}

	stringBytes := data[stringOffset+32 : stringOffset+32+length]
	return string(stringBytes), offset + 32, nil
}

func ParseABISignature(signature string) (*ABIFunction, error) {
	re := regexp.MustCompile(`^(\w+)\((.*)\)$`)
	matches := re.FindStringSubmatch(signature)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid function signature format")
	}

	funcName := matches[1]
	paramStr := matches[2]

	var params []ABIParam
	if paramStr != "" {
		paramTypes := strings.Split(paramStr, ",")
		for i, paramType := range paramTypes {
			params = append(params, ABIParam{
				Name: fmt.Sprintf("param%d", i),
				Type: strings.TrimSpace(paramType),
			})
		}
	}

	return &ABIFunction{
		Name:   funcName,
		Inputs: params,
	}, nil
}
