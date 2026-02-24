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

	logContent, err := renderLogPy()
	if err != nil {
		t.Fatalf("renderLogPy() error = %v", err)
	}
	if !strings.Contains(logContent, "def setup_logging") {
		t.Fatalf("expected setup_logging in log template, got: %q", logContent)
	}

	exceptionContent, err := renderExceptionPy()
	if err != nil {
		t.Fatalf("renderExceptionPy() error = %v", err)
	}
	if !strings.Contains(exceptionContent, "class CozeAPIError") {
		t.Fatalf("expected CozeAPIError in exception template, got: %q", exceptionContent)
	}

	versionContent, err := renderVersionPy()
	if err != nil {
		t.Fatalf("renderVersionPy() error = %v", err)
	}
	if !strings.Contains(versionContent, "VERSION = \"0.20.0\"") {
		t.Fatalf("expected VERSION in version template, got: %q", versionContent)
	}

	pyprojectContent, err := renderPyprojectToml()
	if err != nil {
		t.Fatalf("renderPyprojectToml() error = %v", err)
	}
	if !strings.Contains(pyprojectContent, "[tool.poetry]") {
		t.Fatalf("expected poetry config in pyproject template, got: %q", pyprojectContent)
	}
}

func TestRenderPythonTemplateMissing(t *testing.T) {
	if _, err := renderPythonTemplate("missing.tpl", nil); err == nil {
		t.Fatal("expected renderPythonTemplate to fail for missing template")
	}
}

func TestRenderPythonRawAsset(t *testing.T) {
	content, err := renderPythonRawAsset("special/cozepy/websockets/ws.py.raw")
	if err != nil {
		t.Fatalf("renderPythonRawAsset() error = %v", err)
	}
	if !strings.Contains(content, "class WebsocketsEventType(DynamicStrEnum):") {
		t.Fatalf("expected websocket event type class in raw asset, got: %q", content[:80])
	}

	initContent, err := renderPythonRawAsset("special/cozepy/__init__.py.raw")
	if err != nil {
		t.Fatalf("renderPythonRawAsset() error = %v", err)
	}
	if !strings.Contains(initContent, "from .version import VERSION") {
		t.Fatalf("expected VERSION import in raw package init, got: %q", initContent)
	}
}
