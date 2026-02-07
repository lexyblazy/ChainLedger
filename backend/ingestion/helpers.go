package ingestion

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func hexToUint64(s string) (uint64, error) {
	v := strings.TrimPrefix(s, "0x")
	return strconv.ParseUint(v, 16, 64)

}

func hexToBigInt(s string) (*big.Int, error) {
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

func normalizeAddress(addr string) string {
	a := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(addr), "0x"))

	if len(a) == 64 {
		// topic-padded address: take last 20 bytes (40 hex chars)
		a = a[24:]
	}

	return a
}

func formatAddress(address string) string {
	return "0x" + normalizeAddress(address)
}

func isAddressEqual(address1 string, address2 string) bool {
	return normalizeAddress(address1) == normalizeAddress(address2)
}

func hexToTimestamp(s string) (time.Time, error) {
	uint64, err := hexToUint64(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(uint64), 0).UTC(), nil
}
