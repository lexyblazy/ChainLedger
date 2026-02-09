package hex

import "strings"

func NormalizeAddress(addr string) string {
	a := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(addr), "0x"))

	if len(a) == 64 {
		// topic-padded address: take last 20 bytes (40 hex chars)
		a = a[24:]
	}

	return a
}

func FormatAddress(address string) string {
	return "0x" + NormalizeAddress(address)
}

func IsAddressEqual(address1 string, address2 string) bool {
	return NormalizeAddress(address1) == NormalizeAddress(address2)
}

func IsValidAddressLength(address string) bool {
	return len(address) >= 40
}
