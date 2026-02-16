package hex

import (
	"math/big"
	"testing"
	"time"
)

func TestDecodeUint64(t *testing.T) {
	got, err := DecodeUint64("0x10")
	if err != nil {
		t.Fatalf("DecodeUint64 returned error: %v", err)
	}
	if got != 16 {
		t.Fatalf("DecodeUint64(0x10) = %d, want 16", got)
	}
}

func TestDecodeBigInt(t *testing.T) {
	got, err := DecodeBigInt("0xff")
	if err != nil {
		t.Fatalf("DecodeBigInt returned error: %v", err)
	}
	if got.Cmp(big.NewInt(255)) != 0 {
		t.Fatalf("DecodeBigInt(0xff) = %v, want 255", got)
	}

	zero, err := DecodeBigInt("0x")
	if err != nil {
		t.Fatalf("DecodeBigInt empty returned error: %v", err)
	}
	if zero.Sign() != 0 {
		t.Fatalf("DecodeBigInt empty = %v, want 0", zero)
	}

	if _, err := DecodeBigInt("0xzz"); err == nil {
		t.Fatal("expected invalid hex integer error")
	}
}

func TestDecodeTimestamp(t *testing.T) {
	got, err := DecodeTimestamp("0x1")
	if err != nil {
		t.Fatalf("DecodeTimestamp returned error: %v", err)
	}
	want := time.Unix(1, 0).UTC()
	if !got.Equal(want) {
		t.Fatalf("DecodeTimestamp(0x1) = %v, want %v", got, want)
	}
}

func TestDecodeUint8(t *testing.T) {
	valid := "000000000000000000000000000000000000000000000000000000000000002a"
	got, err := DecodeUint8(valid)
	if err != nil {
		t.Fatalf("DecodeUint8 returned error: %v", err)
	}
	if got != 42 {
		t.Fatalf("DecodeUint8(valid) = %d, want 42", got)
	}

	if _, err := DecodeUint8(""); err == nil {
		t.Fatal("expected empty hex response error")
	}
	if _, err := DecodeUint8("01"); err == nil {
		t.Fatal("expected invalid uint8 length error")
	}
}

func TestDecodeStringOrBytes32(t *testing.T) {
	bytes32 := "5553445400000000000000000000000000000000000000000000000000000000"
	got, err := DecodeStringOrBytes32(bytes32)
	if err != nil {
		t.Fatalf("DecodeStringOrBytes32(bytes32) returned error: %v", err)
	}
	if got != "USDT" {
		t.Fatalf("DecodeStringOrBytes32(bytes32) = %q, want USDT", got)
	}

	dynamic := "0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000005" +
		"68656c6c6f000000000000000000000000000000000000000000000000000000"
	got, err = DecodeStringOrBytes32(dynamic)
	if err != nil {
		t.Fatalf("DecodeStringOrBytes32(dynamic) returned error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("DecodeStringOrBytes32(dynamic) = %q, want hello", got)
	}

	if _, err := DecodeStringOrBytes32("abcd"); err == nil {
		t.Fatal("expected invalid ABI string error")
	}
}

func TestIntToHex(t *testing.T) {
	if got := IntToHex(255); got != "0xff" {
		t.Fatalf("IntToHex(255) = %q, want 0xff", got)
	}
}
