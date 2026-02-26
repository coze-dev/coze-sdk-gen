package python

import (
	"strings"
	"testing"
)

func TestExtractSwaggerRichTextDeterministicOrder(t *testing.T) {
	raw := `{"b":{"ops":[{"insert":"second"}]},"a":{"ops":[{"insert":"first"}]}}`

	for i := 0; i < 128; i++ {
		got := extractSwaggerRichText(raw)
		flat := strings.ReplaceAll(strings.ReplaceAll(got, "\n", ""), " ", "")
		if flat != "firstsecond" {
			t.Fatalf("extractSwaggerRichText() = %q, normalized=%q, want %q", got, flat, "firstsecond")
		}
	}
}

func TestExtractSwaggerRichTextPreservesLineBreaks(t *testing.T) {
	raw := `{"0":{"ops":[{"insert":"删除扣子应用的协作者。\n"},{"attributes":{"lmkr":"1"},"insert":"*"},{"insert":"删除协作者时，扣子会将该协作者创建的工作流、插件等资源转移给应用的所有者。\n"},{"attributes":{"anchor":"3dc926e4","heading":"h2","lmkr":"1"},"insert":"*"},{"insert":"接口限制\n"},{"attributes":{"list":"bullet1","lmkr":"1"},"insert":"*"},{"insert":"每次请求只能删除一位协作者。如需删除多位，请依次发送请求。\n"}]}}`
	got := extractSwaggerRichText(raw)
	want := strings.Join([]string{
		"删除扣子应用的协作者。",
		"删除协作者时，扣子会将该协作者创建的工作流、插件等资源转移给应用的所有者。",
		"接口限制",
		"每次请求只能删除一位协作者。如需删除多位，请依次发送请求。",
	}, "\n")
	if got != want {
		t.Fatalf("extractSwaggerRichText() = %q, want %q", got, want)
	}
}

func TestExtractSwaggerRichTextKeepsInlineStyledText(t *testing.T) {
	raw := `{"0":{"ops":[{"insert":"使用"},{"attributes":{"bold":"true"},"insert":"OAuth"},{"insert":" 应用和服务访问令牌时，只需要有对应权限点即可。"}]}}`
	got := extractSwaggerRichText(raw)
	want := "使用OAuth 应用和服务访问令牌时，只需要有对应权限点即可。"
	if got != want {
		t.Fatalf("extractSwaggerRichText() = %q, want %q", got, want)
	}
}
