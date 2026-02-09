package hex

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func DecodeUint64(s string) (uint64, error) {
	v := strings.TrimPrefix(s, "0x")
	return strconv.ParseUint(v, 16, 64)

}

func DecodeBigInt(s string) (*big.Int, error) {
	v := strings.TrimPrefix(s, "0x")
	if v == "" {
		return big.NewInt(0), nil
	}

	n := new(big.Int)
	_, ok := n.SetString(v, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex integer: %s", s)
	}

	return n, nil
}



func DecodeTimestamp(s string) (time.Time, error) {
	uint64, err := DecodeUint64(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(uint64), 0).UTC(), nil
}

func DecodeUint8(hexData string) (uint8, error) {
	hexData = strings.TrimSpace(hexData)
	hexData = strings.TrimPrefix(hexData, "0x")
	hexData = strings.TrimPrefix(hexData, "0X")

	if len(hexData) == 0 {
		return 0, errors.New("empty hex response")
	}

	b, err := hex.DecodeString(hexData)
	if err != nil {
		return 0, fmt.Errorf("hex decode failed: %w", err)
	}

	if len(b) != 32 {
		return 0, fmt.Errorf("invalid uint8 length: %d", len(b))
	}

	return uint8(b[31]), nil
}

func DecodeStringOrBytes32(hexData string) (string, error) {
	hexData = strings.TrimSpace(hexData)
	hexData = strings.TrimPrefix(hexData, "0x")

	b, err := hex.DecodeString(hexData)
	if err != nil {
		return "", err
	}

	// bytes32 case (old tokens)
	if len(b) == 32 {
		n := bytes.IndexByte(b, 0)
		if n == -1 {
			n = 32
		}
		return string(b[:n]), nil
	}

	// ABI dynamic string
	if len(b) < 64 {
		return "", errors.New("invalid ABI string")
	}

	offset := int(new(big.Int).SetBytes(b[:32]).Int64())
	if offset+32 > len(b) {
		return "", errors.New("invalid ABI offset")
	}

	length := int(new(big.Int).SetBytes(b[offset : offset+32]).Int64())
	start := offset + 32
	end := start + length

	if end > len(b) {
		return "", errors.New("invalid ABI length")
	}

	return string(b[start:end]), nil
}
