package generator

import (
	"strings"
	"testing"
)

func TestRenderRequestTemplate(t *testing.T) {
	content, err := renderRequestPy()
	if err != nil {
		t.Fatalf("renderRequestPy() error = %v", err)
	}
	if !strings.Contains(content, "class Requester(object):") {
		t.Fatalf("expected Requester class in rendered request.py template, got: %q", content)
	}
}

func TestRenderPythonTemplateMissing(t *testing.T) {
	if _, err := renderPythonTemplate("missing.tpl", nil); err == nil {
		t.Fatal("expected renderPythonTemplate to fail for missing template")
	}
}
