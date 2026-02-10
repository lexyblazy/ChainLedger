package address

import "strings"

func Normalize(addr string) string {
	a := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(addr), "0x"))

	if len(a) == 64 {
		// topic-padded address: take last 20 bytes (40 hex chars)
		a = a[24:]
	}

	return a
}

func IsNativeAsset(assetAddress string) bool {
	// null, native, and empty string are special values for native assets
	return assetAddress == "native" || assetAddress == "null" || assetAddress == ""
}


func Format(address string) string {
	if IsNativeAsset(address) {
		return "native"
	}
	return "0x" + Normalize(address)
}

func IsEqual(address1 string, address2 string) bool {
	return Normalize(address1) == Normalize(address2)
}

func IsValidLength(address string) bool {
	return len(address) >= 40
}
