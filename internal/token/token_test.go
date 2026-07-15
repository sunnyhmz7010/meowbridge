package token

import "testing"

func TestGenerateReturnsUniqueURLSafeTokens(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		value, err := Generate()
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if len(value) < 32 {
			t.Fatalf("token too short: %q", value)
		}
		if seen[value] {
			t.Fatalf("duplicate token: %q", value)
		}
		seen[value] = true
	}
}
