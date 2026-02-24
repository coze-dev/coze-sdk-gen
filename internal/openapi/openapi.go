package openapi

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Document struct {
	Paths map[string]PathItem `yaml:"paths"`
}

type PathItem struct {
	Get     *Operation `yaml:"get"`
	Put     *Operation `yaml:"put"`
	Post    *Operation `yaml:"post"`
	Delete  *Operation `yaml:"delete"`
	Patch   *Operation `yaml:"patch"`
	Options *Operation `yaml:"options"`
	Head    *Operation `yaml:"head"`
	Trace   *Operation `yaml:"trace"`
}

type Operation struct {
	OperationID string   `yaml:"operationId"`
	Summary     string   `yaml:"summary"`
	Tags        []string `yaml:"tags"`
}

type OperationRef struct {
	Path   string
	Method string
}

func Load(path string) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read openapi file %q: %w", path, err)
	}
	return Parse(content)
}

func Parse(content []byte) (*Document, error) {
	var doc Document
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi yaml: %w", err)
	}
	if doc.Paths == nil {
		doc.Paths = map[string]PathItem{}
	}
	return &doc, nil
}

func (d *Document) HasOperation(method string, path string) bool {
	_, ok := d.Operation(method, path)
	return ok
}

func (d *Document) Operation(method string, path string) (*Operation, bool) {
	if d == nil {
		return nil, false
	}

	item, ok := d.Paths[path]
	if !ok {
		return nil, false
	}

	switch normalizeMethod(method) {
	case "get":
		if item.Get == nil {
			return nil, false
		}
		return item.Get, true
	case "put":
		if item.Put == nil {
			return nil, false
		}
		return item.Put, true
	case "post":
		if item.Post == nil {
			return nil, false
		}
		return item.Post, true
	case "delete":
		if item.Delete == nil {
			return nil, false
		}
		return item.Delete, true
	case "patch":
		if item.Patch == nil {
			return nil, false
		}
		return item.Patch, true
	case "options":
		if item.Options == nil {
			return nil, false
		}
		return item.Options, true
	case "head":
		if item.Head == nil {
			return nil, false
		}
		return item.Head, true
	case "trace":
		if item.Trace == nil {
			return nil, false
		}
		return item.Trace, true
	default:
		return nil, false
	}
}

func (d *Document) ListOperations() []OperationRef {
	if d == nil {
		return nil
	}

	ops := make([]OperationRef, 0)
	for path, item := range d.Paths {
		if item.Get != nil {
			ops = append(ops, OperationRef{Path: path, Method: "get"})
		}
		if item.Put != nil {
			ops = append(ops, OperationRef{Path: path, Method: "put"})
		}
		if item.Post != nil {
			ops = append(ops, OperationRef{Path: path, Method: "post"})
		}
		if item.Delete != nil {
			ops = append(ops, OperationRef{Path: path, Method: "delete"})
		}
		if item.Patch != nil {
			ops = append(ops, OperationRef{Path: path, Method: "patch"})
		}
		if item.Options != nil {
			ops = append(ops, OperationRef{Path: path, Method: "options"})
		}
		if item.Head != nil {
			ops = append(ops, OperationRef{Path: path, Method: "head"})
		}
		if item.Trace != nil {
			ops = append(ops, OperationRef{Path: path, Method: "trace"})
		}
	}

	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path == ops[j].Path {
			return ops[i].Method < ops[j].Method
		}
		return ops[i].Path < ops[j].Path
	})
	return ops
}

func (d *Document) PathsWithPrefix(prefix string) []string {
	if d == nil {
		return nil
	}
	paths := make([]string, 0)
	for path := range d.Paths {
		if strings.HasPrefix(path, prefix) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func normalizeMethod(method string) string {
	return strings.ToLower(strings.TrimSpace(method))
}
