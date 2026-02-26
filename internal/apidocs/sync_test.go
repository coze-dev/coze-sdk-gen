package apidocs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSectionLinks(t *testing.T) {
	content := `### guides
- [Welcome](https://example.com/guides/welcome)
### developer_guides
- [创建智能体](https://example.com/developer_guides/create_bot)
- [更新日志](https://example.com/developer_guides/changelog)
### references
- [R](https://example.com/ref)
`

	links := ParseSectionLinks(content, "developer_guides")
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0].Title != "创建智能体" || links[0].URL != "https://example.com/developer_guides/create_bot" {
		t.Fatalf("unexpected first link: %#v", links[0])
	}
}

func TestBuildOpenAPI(t *testing.T) {
	link := Link{Title: "创建智能体", URL: "https://docs.coze.cn/api/open/docs/developer_guides/create_bot"}
	doc, ok, err := BuildOpenAPI(link, sampleAPIMarkdown)
	if err != nil {
		t.Fatalf("BuildOpenAPI() error = %v", err)
	}
	if !ok {
		t.Fatal("expected api doc to be detected")
	}

	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatalf("missing paths in generated doc: %#v", doc)
	}
	createPath, ok := paths["/v1/bot/create"].(map[string]any)
	if !ok {
		t.Fatalf("missing /v1/bot/create path: %#v", paths)
	}
	postOp, ok := createPath["post"].(map[string]any)
	if !ok {
		t.Fatalf("missing post operation: %#v", createPath)
	}
	if postOp["operationId"] != "post_create_bot" {
		t.Fatalf("unexpected operationId: %v", postOp["operationId"])
	}
	if _, ok := postOp["requestBody"]; !ok {
		t.Fatalf("requestBody missing: %#v", postOp)
	}

	components, ok := doc["components"].(map[string]any)
	if !ok {
		t.Fatalf("missing components: %#v", doc)
	}
	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		t.Fatalf("missing schemas: %#v", components)
	}
	if _, ok := schemas["PromptInfo"]; !ok {
		t.Fatalf("PromptInfo schema missing: %#v", schemas)
	}
	if _, ok := schemas["CreateDraftBotData"]; !ok {
		t.Fatalf("CreateDraftBotData schema missing: %#v", schemas)
	}
}

func TestBuildOpenAPINonAPI(t *testing.T) {
	_, ok, err := BuildOpenAPI(Link{Title: "更新日志", URL: "https://example.com/changelog"}, "# 更新日志\n\n这不是 API 文档。")
	if err != nil {
		t.Fatalf("BuildOpenAPI() error = %v", err)
	}
	if ok {
		t.Fatal("expected non-api markdown to be skipped")
	}
}

func TestSync(t *testing.T) {
	apiPath := "/api/open/docs/developer_guides/create_bot"
	nonAPIPath := "/api/open/docs/developer_guides/changelog"

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/llms.txt":
			_, _ = w.Write([]byte(fmt.Sprintf("### developer_guides\n- [创建智能体](%s%s)\n- [更新日志](%s%s)\n### other\n- [x](https://example.com)\n", server.URL, apiPath, server.URL, nonAPIPath)))
		case apiPath:
			_, _ = w.Write([]byte(sampleAPIMarkdown))
		case nonAPIPath:
			_, _ = w.Write([]byte("# 更新日志\n\n- 2026-01-01 发布"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tmp := t.TempDir()
	mdDir := filepath.Join(tmp, "docs", "api_markdown")
	yamlDir := filepath.Join(tmp, "docs", "swagger")

	res, err := Sync(context.Background(), Options{
		IndexURL:          server.URL + "/llms.txt",
		Section:           "developer_guides",
		MarkdownOutputDir: mdDir,
		SwaggerOutputDir:  yamlDir,
		HTTPClient:        server.Client(),
	})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if res.TotalLinks != 2 || res.GeneratedFiles != 1 || res.SkippedDocs != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}

	mdPath := filepath.Join(mdDir, "create_bot.md")
	yamlPath := filepath.Join(yamlDir, "create_bot.yaml")
	if _, err := os.Stat(mdPath); err != nil {
		t.Fatalf("markdown not generated: %v", err)
	}
	yamlBody, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("swagger not generated: %v", err)
	}
	text := string(yamlBody)
	if !strings.Contains(text, "openapi: 3.0.3") {
		t.Fatalf("unexpected yaml body: %s", text)
	}
	if !strings.Contains(text, "/v1/bot/create") {
		t.Fatalf("expected endpoint path in yaml: %s", text)
	}
}

const sampleAPIMarkdown = `# 创建智能体
创建一个新的智能体。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | https://api.coze.cn/v1/bot/create |
| **权限** | createBot |

## 请求参数
### Header
| 参数 | 取值 | 说明 |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于鉴权 |
| Content-Type | application/json | 请求类型 |

### Body
| 参数 | 类型 | 是否必选 | 说明 |
| --- | --- | --- | --- |
| space_id | String | 必选 | 空间 ID |
| name | String | 必选 | 名称 |
| prompt_info | Object of [PromptInfo](#promptinfo) | 可选 | 人设 |

### PromptInfo
| 参数 | 类型 | 是否必选 | 说明 |
| --- | --- | --- | --- |
| prompt | String | 可选 | 智能体提示词 |

## 返回参数
| 参数 | 类型 | 示例 | 说明 |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码 |
| msg | String | "" | 状态信息 |
| data | Object of [CreateDraftBotData](#createdraftbotdata) | {} | 返回数据 |

### CreateDraftBotData
| 参数 | 类型 | 示例 | 说明 |
| --- | --- | --- | --- |
| bot_id | String | 73428668***** | 创建的智能体 ID |
`
