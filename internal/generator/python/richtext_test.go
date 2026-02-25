package python

import "testing"

func TestExtractSwaggerRichTextDeterministicOrder(t *testing.T) {
	raw := `{"b":{"ops":[{"insert":"second"}]},"a":{"ops":[{"insert":"first"}]}}`

	for i := 0; i < 128; i++ {
		got := extractSwaggerRichText(raw)
		if got != "first second" {
			t.Fatalf("extractSwaggerRichText() = %q, want %q", got, "first second")
		}
	}
}
