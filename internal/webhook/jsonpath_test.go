package webhook

import "testing"

func TestParseJSONPathExtractsObjectAndArrayValues(t *testing.T) {
	payload := map[string]any{
		"hook": map[string]any{
			"url": "https://github.com/sunnyhmz7010/meowbridge",
		},
		"commits": []any{
			map[string]any{"message": "first"},
			map[string]any{"message": "second"},
		},
		"count": float64(2),
	}

	cases := []struct {
		name string
		path string
		want string
	}{
		{name: "object path", path: "$.hook.url", want: "https://github.com/sunnyhmz7010/meowbridge"},
		{name: "array index", path: "$.commits[1].message", want: "second"},
		{name: "number", path: "$.count", want: "2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseJSONPath(payload, tc.path)
			if !ok {
				t.Fatalf("ParseJSONPath(%q) did not match", tc.path)
			}
			if got != tc.want {
				t.Fatalf("ParseJSONPath(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestParseJSONPathReturnsFalseForMissingOrUnsupportedPath(t *testing.T) {
	payload := map[string]any{"items": []any{map[string]any{"name": "one"}}}

	for _, path := range []string{
		"",
		"items[0].name",
		"$.items[2].name",
		"$.items[*].name",
		"$.items[0].missing",
		"$.items[bad].name",
	} {
		if got, ok := ParseJSONPath(payload, path); ok {
			t.Fatalf("ParseJSONPath(%q) = %q, want no match", path, got)
		}
	}
}
