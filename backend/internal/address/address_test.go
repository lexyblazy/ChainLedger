package address

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trims lowercase and strips prefix", input: " 0xAbC123 ", want: "abc123"},
		{name: "handles topic padded address", input: "0x00000000000000000000000090f8bf6a479f320ead074411a4b0e7944ea8c9c1", want: "90f8bf6a479f320ead074411a4b0e7944ea8c9c1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Normalize(tc.input); got != tc.want {
				t.Fatalf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestIsNativeAsset(t *testing.T) {
	if !IsNativeAsset("native") || !IsNativeAsset("null") || !IsNativeAsset("") {
		t.Fatal("expected native/null/empty to be treated as native assets")
	}
	if IsNativeAsset("0xabc") {
		t.Fatal("did not expect regular address to be treated as native asset")
	}
}

func TestFormat(t *testing.T) {
	if got := Format("null"); got != "native" {
		t.Fatalf("Format(null) = %q, want native", got)
	}
	if got := Format("0xABC"); got != "0xabc" {
		t.Fatalf("Format(0xABC) = %q, want 0xabc", got)
	}
}

func TestIsEqual(t *testing.T) {
	if !IsEqual("0xABC", "abc") {
		t.Fatal("expected addresses to be equal after normalization")
	}
	if IsEqual("0xabc", "0xdef") {
		t.Fatal("expected different addresses to not be equal")
	}
}

func TestIsValidLength(t *testing.T) {
	if !IsValidLength("1234567890123456789012345678901234567890") {
		t.Fatal("expected 40-char address to be valid")
	}
	if IsValidLength("short") {
		t.Fatal("expected short address to be invalid")
	}
}
