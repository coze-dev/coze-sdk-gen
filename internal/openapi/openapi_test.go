package openapi

import (
	"path/filepath"
	"testing"
)

func TestLoadAndLookupOperations(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	op, ok := doc.Operation("post", "/v3/chat")
	if !ok {
		t.Fatalf("expected /v3/chat#post to exist")
	}
	if op.OperationID != "OpenApiChat" {
		t.Fatalf("unexpected operation id: %q", op.OperationID)
	}

	if !doc.HasOperation("GET", "/v3/chat/message/list") {
		t.Fatalf("expected GET /v3/chat/message/list to exist")
	}
	if doc.HasOperation("patch", "/v3/chat") {
		t.Fatalf("did not expect PATCH /v3/chat")
	}
	if _, ok := doc.Operation("invalid", "/v3/chat"); ok {
		t.Fatalf("did not expect invalid method to exist")
	}
}

func TestListOperationsSorted(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	ops := doc.ListOperations()
	if len(ops) != 5 {
		t.Fatalf("expected 5 operations, got %d", len(ops))
	}

	for i := 1; i < len(ops); i++ {
		prev := ops[i-1]
		curr := ops[i]
		if prev.Path > curr.Path || (prev.Path == curr.Path && prev.Method > curr.Method) {
			t.Fatalf("operations are not sorted at %d: %#v > %#v", i, prev, curr)
		}
	}
}

func TestPathsWithPrefix(t *testing.T) {
	doc, err := Load(filepath.Join("testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	chatPaths := doc.PathsWithPrefix("/v3/chat")
	if len(chatPaths) != 3 {
		t.Fatalf("expected 3 chat paths, got %d", len(chatPaths))
	}

	missing := doc.PathsWithPrefix("/v2/not-found")
	if len(missing) != 0 {
		t.Fatalf("expected no paths, got %v", missing)
	}
}

func TestParseInvalidYAML(t *testing.T) {
	if _, err := Parse([]byte("paths: [")); err == nil {
		t.Fatal("expected Parse() to fail for invalid yaml")
	}
}

func TestOperationSupportsAllMethodsAndNilDoc(t *testing.T) {
	doc, err := Parse([]byte(`
paths:
  /all:
    get: {operationId: getOp}
    put: {operationId: putOp}
    post: {operationId: postOp}
    delete: {operationId: deleteOp}
    patch: {operationId: patchOp}
    options: {operationId: optionsOp}
    head: {operationId: headOp}
    trace: {operationId: traceOp}
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	methods := []string{"get", "put", "post", "delete", "patch", "options", "head", "trace"}
	for _, method := range methods {
		if _, ok := doc.Operation(method, "/all"); !ok {
			t.Fatalf("expected method %s to exist", method)
		}
	}

	var nilDoc *Document
	if _, ok := nilDoc.Operation("get", "/all"); ok {
		t.Fatal("did not expect operation on nil doc")
	}
	if len(nilDoc.ListOperations()) != 0 {
		t.Fatal("expected nil list operations from nil doc")
	}
	if len(nilDoc.PathsWithPrefix("/")) != 0 {
		t.Fatal("expected nil paths from nil doc")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join("testdata", "missing.yaml")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
