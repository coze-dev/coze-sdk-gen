package apidocs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultIndexURL    = "https://docs.coze.cn/llms.txt"
	DefaultSection     = "developer_guides"
	DefaultMarkdownDir = "docs/api_markdown"
	DefaultSwaggerDir  = "docs/swagger"
	DefaultTimeout     = 30 * time.Second
)

var markdownLinkPattern = regexp.MustCompile(`^- \[(.+)\]\((https?://[^)]+)\)\s*$`)

type Options struct {
	IndexURL          string
	Section           string
	MarkdownOutputDir string
	SwaggerOutputDir  string
	HTTPClient        *http.Client
}

type Result struct {
	TotalLinks     int
	FetchedDocs    int
	GeneratedFiles int
	SkippedDocs    int
	MarkdownDir    string
	SwaggerDir     string
}

type Link struct {
	Title string
	URL   string
}

func Sync(ctx context.Context, opt Options) (Result, error) {
	resolved := withDefaults(opt)
	if err := resolved.validate(); err != nil {
		return Result{}, err
	}

	indexContent, err := fetchText(ctx, resolved.HTTPClient, resolved.IndexURL)
	if err != nil {
		return Result{}, fmt.Errorf("fetch llms index: %w", err)
	}
	links := ParseSectionLinks(indexContent, resolved.Section)

	if err := resetOutputDir(resolved.MarkdownOutputDir); err != nil {
		return Result{}, err
	}
	if err := resetOutputDir(resolved.SwaggerOutputDir); err != nil {
		return Result{}, err
	}

	res := Result{
		TotalLinks:  len(links),
		MarkdownDir: resolved.MarkdownOutputDir,
		SwaggerDir:  resolved.SwaggerOutputDir,
	}

	filenameCounter := make(map[string]int)
	for _, link := range links {
		markdownBody, err := fetchText(ctx, resolved.HTTPClient, link.URL)
		if err != nil {
			return res, fmt.Errorf("fetch %s: %w", link.URL, err)
		}
		res.FetchedDocs++

		op, ok, err := BuildOpenAPI(link, markdownBody)
		if err != nil {
			return res, fmt.Errorf("parse %s: %w", link.URL, err)
		}
		if !ok {
			res.SkippedDocs++
			continue
		}

		name := uniqueBaseName(link.URL, filenameCounter)
		markdownPath := filepath.Join(resolved.MarkdownOutputDir, name+".md")
		swaggerPath := filepath.Join(resolved.SwaggerOutputDir, name+".yaml")

		if err := os.WriteFile(markdownPath, []byte(markdownBody), 0o644); err != nil {
			return res, fmt.Errorf("write markdown %s: %w", markdownPath, err)
		}
		yamlBody, err := yaml.Marshal(op)
		if err != nil {
			return res, fmt.Errorf("marshal swagger for %s: %w", link.URL, err)
		}
		if err := os.WriteFile(swaggerPath, yamlBody, 0o644); err != nil {
			return res, fmt.Errorf("write swagger %s: %w", swaggerPath, err)
		}
		res.GeneratedFiles++
	}

	return res, nil
}

func ParseSectionLinks(content string, section string) []Link {
	section = strings.TrimSpace(section)
	if section == "" {
		return nil
	}

	var links []Link
	inSection := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			heading := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			if inSection && heading != section {
				break
			}
			inSection = heading == section
			continue
		}
		if !inSection {
			continue
		}
		match := markdownLinkPattern.FindStringSubmatch(trimmed)
		if len(match) != 3 {
			continue
		}
		links = append(links, Link{Title: strings.TrimSpace(match[1]), URL: strings.TrimSpace(match[2])})
	}
	return links
}

func uniqueBaseName(rawURL string, counter map[string]int) string {
	base := baseNameFromURL(rawURL)
	counter[base]++
	if counter[base] == 1 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, counter[base])
}

func baseNameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return sanitizeFileName(rawURL)
	}
	name := path.Base(strings.TrimSuffix(u.Path, "/"))
	if name == "." || name == "/" || name == "" {
		return "unknown"
	}
	return sanitizeFileName(name)
}

func sanitizeFileName(in string) string {
	in = strings.ToLower(strings.TrimSpace(in))
	if in == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range in {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "unknown"
	}
	return out
}

func withDefaults(opt Options) Options {
	if strings.TrimSpace(opt.IndexURL) == "" {
		opt.IndexURL = DefaultIndexURL
	}
	if strings.TrimSpace(opt.Section) == "" {
		opt.Section = DefaultSection
	}
	if strings.TrimSpace(opt.MarkdownOutputDir) == "" {
		opt.MarkdownOutputDir = DefaultMarkdownDir
	}
	if strings.TrimSpace(opt.SwaggerOutputDir) == "" {
		opt.SwaggerOutputDir = DefaultSwaggerDir
	}
	if opt.HTTPClient == nil {
		opt.HTTPClient = &http.Client{Timeout: DefaultTimeout}
	}
	return opt
}

func (o Options) validate() error {
	if strings.TrimSpace(o.IndexURL) == "" {
		return fmt.Errorf("index url is required")
	}
	if strings.TrimSpace(o.Section) == "" {
		return fmt.Errorf("section is required")
	}
	if strings.TrimSpace(o.MarkdownOutputDir) == "" {
		return fmt.Errorf("markdown output dir is required")
	}
	if strings.TrimSpace(o.SwaggerOutputDir) == "" {
		return fmt.Errorf("swagger output dir is required")
	}
	if o.HTTPClient == nil {
		return fmt.Errorf("http client is required")
	}
	return nil
}

func fetchText(ctx context.Context, client *http.Client, uri string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func resetOutputDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("cleanup %s: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return nil
}
