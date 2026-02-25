package python

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectPublicTopLevelClassesFromSource(t *testing.T) {
	content := `
class PublicBase:
    pass

class _Private:
    pass

class PublicWithParent(PublicBase):
    pass

class AsyncBotsClient:
    pass

if True:
    class Nested:
        pass

class PublicMultiLine(
    PublicBase
):
    pass
`
	got := collectPublicTopLevelClassesFromSource(content)
	want := []string{"PublicBase", "PublicWithParent", "PublicMultiLine"}
	if len(got) != len(want) {
		t.Fatalf("class count mismatch got=%v want=%v", got, want)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("class[%d] mismatch got=%q want=%q", idx, got[idx], want[idx])
		}
	}
}

func TestRenderPythonRootInit(t *testing.T) {
	root := t.TempDir()

	writeTestPythonFile(t, filepath.Join(root, "__init__.py"), "class ShouldBeIgnored:\n    pass\n")
	writeTestPythonFile(t, filepath.Join(root, "coze.py"), "class Coze:\n    pass\n\nclass AsyncCoze:\n    pass\n")
	writeTestPythonFile(t, filepath.Join(root, "apps", "__init__.py"), "class AppsClient:\n    pass\n\nclass SimpleApp:\n    pass\n")
	writeTestPythonFile(t, filepath.Join(root, "apps", "collaborators", "__init__.py"), "class AppCollaborator:\n    pass\n")
	writeTestPythonFile(t, filepath.Join(root, "auth", "__init__.py"), "class Auth:\n    pass\n\ndef load_oauth_app_from_config():\n    pass\n")
	writeTestPythonFile(t, filepath.Join(root, "config.py"), "COZE_COM_BASE_URL = \"x\"\nCOZE_CN_BASE_URL = \"y\"\nDEFAULT_TIMEOUT = 10\nDEFAULT_CONNECTION_LIMITS = 5\n")
	writeTestPythonFile(t, filepath.Join(root, "log.py"), "def setup_logging():\n    pass\n")
	writeTestPythonFile(t, filepath.Join(root, "version.py"), "VERSION = \"0.0.0\"\n")

	content, err := renderPythonRootInit(root)
	if err != nil {
		t.Fatalf("renderPythonRootInit() error = %v", err)
	}

	containsAll := []string{
		"from .apps import SimpleApp",
		"from .apps.collaborators import AppCollaborator",
		"from .coze import (",
		"    AsyncCoze,",
		"    Coze,",
		"from .auth import (",
		"    Auth,",
		"    load_oauth_app_from_config,",
		"from .config import (",
		"    COZE_CN_BASE_URL,",
		"    COZE_COM_BASE_URL,",
		"    DEFAULT_CONNECTION_LIMITS,",
		"    DEFAULT_TIMEOUT,",
		"from .log import setup_logging",
		"from .version import VERSION",
		"\"SimpleApp\"",
		"\"AppCollaborator\"",
		"\"AsyncCoze\"",
		"\"Coze\"",
		"\"COZE_COM_BASE_URL\"",
		"\"setup_logging\"",
		"\"VERSION\"",
	}
	for _, expected := range containsAll {
		if !strings.Contains(content, expected) {
			t.Fatalf("rendered root init missing %q:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "ShouldBeIgnored") {
		t.Fatalf("root package __init__.py classes should not be scanned: %s", content)
	}
	if strings.Contains(content, "AppsClient") {
		t.Fatalf("client classes should not be exported: %s", content)
	}
}

func TestRenderPythonRootInitDuplicateExport(t *testing.T) {
	root := t.TempDir()

	writeTestPythonFile(t, filepath.Join(root, "a.py"), "class Duplicated:\n    pass\n")
	writeTestPythonFile(t, filepath.Join(root, "b.py"), "class Duplicated:\n    pass\n")

	_, err := renderPythonRootInit(root)
	if err == nil {
		t.Fatal("renderPythonRootInit() expected duplicate export error")
	}
	if !strings.Contains(err.Error(), "duplicate root export \"Duplicated\"") {
		t.Fatalf("unexpected duplicate error: %v", err)
	}
}

func writeTestPythonFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
