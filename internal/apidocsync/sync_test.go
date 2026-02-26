package apidocsync

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestParseSectionLinks(t *testing.T) {
	content := `# 扣子

## 文档

### guides
- [忽略](https://docs.coze.cn/api/open/docs/guides/welcome)

### developer_guides
- [创建智能体](https://docs.coze.cn/api/open/docs/developer_guides/create_bot)
- [创建智能体](https://docs.coze.cn/api/open/docs/developer_guides/create_bot)
- [查看会话](https://docs.coze.cn/api/open/docs/developer_guides/retrieve_conversation)

### tutorial
- [忽略](https://docs.coze.cn/api/open/docs/tutorial/demo)
`

	links, err := parseSectionLinks(content, "developer_guides")
	if err != nil {
		t.Fatalf("parseSectionLinks() error = %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0].Slug != "create_bot" {
		t.Fatalf("unexpected first slug: %q", links[0].Slug)
	}
	if links[1].Slug != "retrieve_conversation" {
		t.Fatalf("unexpected second slug: %q", links[1].Slug)
	}
}

func TestParseSectionLinksMissingSection(t *testing.T) {
	_, err := parseSectionLinks("### guides\n", "developer_guides")
	if err == nil {
		t.Fatal("expected section missing error")
	}
}

func TestParseAPIDocHTTP(t *testing.T) {
	md := `# 创建智能体
创建一个新的智能体。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/v1/bot/create |
| **权限** | createBot |
| **接口说明** | 调用接口创建一个新的智能体。 |

## 请求参数
### Header
| 参数 | 取值 | 说明 |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 鉴权 |

### Body
| 参数 | 类型 | 是否必选 | 示例 | 说明 |
| --- | --- | --- | --- | --- |
| name | String | 必选 | demo | 名称 |
| prompt_info | Object of [PromptInfo](#promptinfo) | 可选 | {} | 人设 |

### PromptInfo
| 参数 | 类型 | 是否必选 | 示例 | 说明 |
| --- | --- | --- | --- | --- |
| prompt | String | 必选 | hi | 提示词 |

## 返回参数
| 参数 | 类型 | 示例 | 说明 |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码 |
| data | Object of [CreateData](#createdata) | {} | 数据 |

### CreateData
| 参数 | 类型 | 示例 | 说明 |
| --- | --- | --- | --- |
| id | String | 1 | id |
`

	doc, ok := parseAPIDoc(docLink{Title: "创建智能体", URL: "https://docs.coze.cn/api/open/docs/developer_guides/create_bot", Slug: "create_bot"}, md)
	if !ok {
		t.Fatal("expected api doc")
	}
	if doc.Path != "/v1/bot/create" {
		t.Fatalf("unexpected path: %q", doc.Path)
	}
	if doc.HTTPMethod != http.MethodPost {
		t.Fatalf("unexpected method: %q", doc.HTTPMethod)
	}
	if len(doc.HeaderParams) != 1 {
		t.Fatalf("expected 1 header field, got %d", len(doc.HeaderParams))
	}
	if len(doc.RequestBody) != 2 {
		t.Fatalf("expected 2 body fields, got %d", len(doc.RequestBody))
	}
	if len(doc.ResponseBody) != 2 {
		t.Fatalf("expected 2 response fields, got %d", len(doc.ResponseBody))
	}
	if _, ok := doc.ComponentSchemas["PromptInfo"]; !ok {
		t.Fatalf("expected PromptInfo schema, got %#v", doc.ComponentSchemas)
	}
	if _, ok := doc.ComponentSchemas["CreateData"]; !ok {
		t.Fatalf("expected CreateData schema, got %#v", doc.ComponentSchemas)
	}
}

func TestParseAPIDocWebsocket(t *testing.T) {
	md := `# 双向流式语音识别

## 接口信息
| **URL** | wss://ws.coze.cn/v1/audio/transcriptions |
| --- | --- |
| **Headers** | Authorization Bearer $AccessToken |
| **权限** | createTranscription |
| **接口说明** | 流式语音识别 |
`

	doc, ok := parseAPIDoc(docLink{Title: "双向流式语音识别", URL: "https://docs.coze.cn/api/open/docs/developer_guides/asr_api", Slug: "asr_api"}, md)
	if !ok {
		t.Fatal("expected websocket doc to be parsed")
	}
	if !doc.IsWebsocket {
		t.Fatal("expected websocket flag")
	}
	if doc.HTTPMethod != http.MethodGet {
		t.Fatalf("unexpected method for websocket: %q", doc.HTTPMethod)
	}
	if doc.Path != "/v1/audio/transcriptions" {
		t.Fatalf("unexpected websocket path: %q", doc.Path)
	}
}

func TestBuildSwaggerYAML(t *testing.T) {
	md := `# 创建智能体
创建一个新的智能体。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/v1/bot/create |
| **权限** | createBot |
| **接口说明** | 调用接口创建一个新的智能体。 |

## 请求参数
### Body
| 参数 | 类型 | 是否必选 | 示例 | 说明 |
| --- | --- | --- | --- | --- |
| prompt_info | Object of [PromptInfo](#promptinfo) | 必选 | {} | 人设 |

### PromptInfo
| 参数 | 类型 | 是否必选 | 示例 | 说明 |
| --- | --- | --- | --- | --- |
| prompt | String | 必选 | hi | 提示词 |

## 返回参数
| 参数 | 类型 | 示例 | 说明 |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码 |
`

	doc, ok := parseAPIDoc(docLink{Title: "创建智能体", URL: "https://docs.coze.cn/api/open/docs/developer_guides/create_bot", Slug: "create_bot"}, md)
	if !ok {
		t.Fatal("expected api doc")
	}

	encoded, err := buildSwaggerYAML(doc)
	if err != nil {
		t.Fatalf("buildSwaggerYAML() error = %v", err)
	}

	var parsed openapiDocument
	if err := yaml.Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("yaml unmarshal error = %v", err)
	}
	pathItem, ok := parsed.Paths["/v1/bot/create"]
	if !ok {
		t.Fatalf("expected /v1/bot/create in paths, got %#v", parsed.Paths)
	}
	op, ok := pathItem["post"]
	if !ok {
		t.Fatalf("expected post operation, got %#v", pathItem)
	}
	if op.RequestBody == nil {
		t.Fatal("expected request body")
	}
	if parsed.Components == nil || len(parsed.Components.Schemas) == 0 {
		t.Fatal("expected component schemas")
	}
}

func TestRunEndToEnd(t *testing.T) {
	apiMarkdown := `# Demo API

## 基础信息
| 请求方式 | GET |
| --- | --- |
| 请求地址 | https://api.coze.cn/v1/demo |
| 接口说明 | demo |

## 请求参数
### Query
| 参数 | 类型 | 是否必选 | 示例 | 说明 |
| --- | --- | --- | --- | --- |
| id | String | 必选 | 1 | id |

## 返回参数
| 参数 | 类型 | 示例 | 说明 |
| --- | --- | --- | --- |
| code | Long | 0 | 状态 |
`

	nonAPIMarkdown := "# Not API\n普通文档"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	llmsContent := "## 文档\n\n### developer_guides\n" +
		"- [Demo API](" + server.URL + "/api/open/docs/developer_guides/demo_api)\n" +
		"- [Not API](" + server.URL + "/api/open/docs/developer_guides/not_api)\n" +
		"\n### tutorial\n- [Ignored](" + server.URL + "/api/open/docs/tutorial/demo)\n"

	mux.HandleFunc("/llms.txt", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, llmsContent)
	})
	mux.HandleFunc("/api/open/docs/developer_guides/demo_api", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, apiMarkdown)
	})
	mux.HandleFunc("/api/open/docs/developer_guides/not_api", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, nonAPIMarkdown)
	})

	outputRoot := t.TempDir()
	result, err := Run(context.Background(), io.Discard, Options{
		LLMSURL:     server.URL + "/llms.txt",
		Section:     "developer_guides",
		OutputRoot:  outputRoot,
		HTTPTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.TotalCandidates != 2 {
		t.Fatalf("expected 2 candidates, got %d", result.TotalCandidates)
	}
	if result.Generated != 1 {
		t.Fatalf("expected 1 generated file, got %d", result.Generated)
	}
	if result.Skipped != 1 {
		t.Fatalf("expected 1 skipped file, got %d", result.Skipped)
	}

	markdownPath := filepath.Join(outputRoot, "api-markdown", "demo_api.md")
	swaggerPath := filepath.Join(outputRoot, "api-swagger", "demo_api.yaml")
	if _, err := os.Stat(markdownPath); err != nil {
		t.Fatalf("expected markdown file %s: %v", markdownPath, err)
	}
	if _, err := os.Stat(swaggerPath); err != nil {
		t.Fatalf("expected swagger file %s: %v", swaggerPath, err)
	}
}
