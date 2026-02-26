package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunInvalidFlag(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--unknown"}, &out); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestRunSuccess(t *testing.T) {
	apiMarkdown := `# Demo API
## 基础信息
| 请求方式 | GET |
| --- | --- |
| 请求地址 | https://api.coze.cn/v1/demo |
## 返回参数
| 参数 | 类型 | 示例 | 说明 |
| --- | --- | --- | --- |
| code | Long | 0 | 状态 |
`

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	llms := "## 文档\n### developer_guides\n- [Demo](" + server.URL + "/api/open/docs/developer_guides/demo_api)\n### tutorial\n"
	mux.HandleFunc("/llms.txt", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, llms)
	})
	mux.HandleFunc("/api/open/docs/developer_guides/demo_api", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, apiMarkdown)
	})

	var out bytes.Buffer
	err := run([]string{
		"--llms-url", server.URL + "/llms.txt",
		"--output-root", t.TempDir(),
		"--http-timeout", "5s",
	}, &out)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected summary output")
	}
}
