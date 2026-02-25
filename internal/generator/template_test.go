package generator

import (
	pygen "github.com/coze-dev/coze-sdk-gen/internal/generator/python"
	"strings"
	"testing"
)

func TestRenderRequestTemplate(t *testing.T) {
	content, err := pygen.RenderRequestPy()
	if err != nil {
		t.Fatalf("pygen.RenderRequestPy() error = %v", err)
	}
	if !strings.Contains(content, "class Requester(object):") {
		t.Fatalf("expected Requester class in rendered request.py template, got: %q", content)
	}
}

func TestRenderStaticPythonTemplates(t *testing.T) {
	configContent, err := pygen.RenderConfigPy()
	if err != nil {
		t.Fatalf("pygen.RenderConfigPy() error = %v", err)
	}
	if !strings.Contains(configContent, "COZE_CN_BASE_URL") {
		t.Fatalf("expected coze config constants in config template, got: %q", configContent)
	}

	modelContent, err := pygen.RenderModelPy()
	if err != nil {
		t.Fatalf("pygen.RenderModelPy() error = %v", err)
	}
	if !strings.Contains(modelContent, "class CozeModel(BaseModel):") {
		t.Fatalf("expected CozeModel in model template, got: %q", modelContent)
	}

	utilContent, err := pygen.RenderUtilPy()
	if err != nil {
		t.Fatalf("pygen.RenderUtilPy() error = %v", err)
	}
	if !strings.Contains(utilContent, "def remove_url_trailing_slash") {
		t.Fatalf("expected remove_url_trailing_slash in util template, got: %q", utilContent)
	}

	logContent, err := pygen.RenderLogPy()
	if err != nil {
		t.Fatalf("pygen.RenderLogPy() error = %v", err)
	}
	if !strings.Contains(logContent, "def setup_logging") {
		t.Fatalf("expected setup_logging in log template, got: %q", logContent)
	}

	exceptionContent, err := pygen.RenderExceptionPy()
	if err != nil {
		t.Fatalf("pygen.RenderExceptionPy() error = %v", err)
	}
	if !strings.Contains(exceptionContent, "class CozeAPIError") {
		t.Fatalf("expected CozeAPIError in exception template, got: %q", exceptionContent)
	}

	versionContent, err := pygen.RenderVersionPy()
	if err != nil {
		t.Fatalf("pygen.RenderVersionPy() error = %v", err)
	}
	if !strings.Contains(versionContent, "VERSION = \"0.20.0\"") {
		t.Fatalf("expected VERSION in version template, got: %q", versionContent)
	}

	pyprojectContent, err := pygen.RenderPyprojectToml()
	if err != nil {
		t.Fatalf("pygen.RenderPyprojectToml() error = %v", err)
	}
	if !strings.Contains(pyprojectContent, "[tool.poetry]") {
		t.Fatalf("expected poetry config in pyproject template, got: %q", pyprojectContent)
	}
}

func TestRenderPythonTemplateMissing(t *testing.T) {
	if _, err := pygen.RenderPythonTemplate("missing.tpl", nil); err == nil {
		t.Fatal("expected pygen.RenderPythonTemplate to fail for missing template")
	}
}

func TestRenderPythonRawAsset(t *testing.T) {
	content, err := pygen.RenderPythonRawAsset("special/cozepy/websockets/ws.py.tpl")
	if err != nil {
		t.Fatalf("pygen.RenderPythonRawAsset() error = %v", err)
	}
	if !strings.Contains(content, "class WebsocketsEventType(DynamicStrEnum):") {
		t.Fatalf("expected websocket event type class in raw asset, got: %q", content[:80])
	}
}
