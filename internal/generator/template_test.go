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

func TestRenderStaticPythonTemplates(t *testing.T) {
	configContent, err := renderConfigPy()
	if err != nil {
		t.Fatalf("renderConfigPy() error = %v", err)
	}
	if !strings.Contains(configContent, "COZE_CN_BASE_URL") {
		t.Fatalf("expected coze config constants in config template, got: %q", configContent)
	}

	modelContent, err := renderModelPy()
	if err != nil {
		t.Fatalf("renderModelPy() error = %v", err)
	}
	if !strings.Contains(modelContent, "class CozeModel(BaseModel):") {
		t.Fatalf("expected CozeModel in model template, got: %q", modelContent)
	}

	utilContent, err := renderUtilPy()
	if err != nil {
		t.Fatalf("renderUtilPy() error = %v", err)
	}
	if !strings.Contains(utilContent, "def remove_url_trailing_slash") {
		t.Fatalf("expected remove_url_trailing_slash in util template, got: %q", utilContent)
	}
}

func TestRenderPythonTemplateMissing(t *testing.T) {
	if _, err := renderPythonTemplate("missing.tpl", nil); err == nil {
		t.Fatal("expected renderPythonTemplate to fail for missing template")
	}
}

func TestPythonSupportFiles(t *testing.T) {
	files, err := listPythonSupportFiles()
	if err != nil {
		t.Fatalf("listPythonSupportFiles() error = %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected embedded support files")
	}

	content, err := readPythonSupportFile("README.md")
	if err != nil {
		t.Fatalf("readPythonSupportFile() error = %v", err)
	}
	if !strings.Contains(string(content), "# Coze Python API SDK") {
		t.Fatalf("unexpected support README content: %q", string(content))
	}
}
