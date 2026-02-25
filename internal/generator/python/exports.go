package python

import (
	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type OperationBinding = operationBinding
type PackageMeta = packageMeta
type ClassMethodBlock = classMethodBlock

func RenderOperationMethod(doc *openapi.Document, binding OperationBinding, async bool) string {
	return renderOperationMethod(doc, binding, async)
}

func RenderOperationMethodWithComments(
	doc *openapi.Document,
	binding OperationBinding,
	async bool,
	modulePath string,
	className string,
	commentOverrides config.CommentOverrides,
) string {
	return renderOperationMethodWithComments(doc, binding, async, modulePath, className, commentOverrides)
}

func RenderPackageModule(
	doc *openapi.Document,
	meta PackageMeta,
	bindings []OperationBinding,
	commentOverrides config.CommentOverrides,
) string {
	return renderPackageModule(doc, meta, bindings, commentOverrides)
}

func OrderClassMethodBlocks(blocks []ClassMethodBlock, orderedNames []string) []ClassMethodBlock {
	return orderClassMethodBlocks(blocks, orderedNames)
}

func DeduplicateBindings(bindings []OperationBinding) []OperationBinding {
	return deduplicateBindings(bindings)
}

func RenderConfigPy() (string, error) {
	return renderConfigPy()
}

func RenderUtilPy() (string, error) {
	return renderUtilPy()
}

func RenderModelPy() (string, error) {
	return renderModelPy()
}

func RenderRequestPy() (string, error) {
	return renderRequestPy()
}

func RenderLogPy() (string, error) {
	return renderLogPy()
}

func RenderExceptionPy() (string, error) {
	return renderExceptionPy()
}

func RenderVersionPy() (string, error) {
	return renderVersionPy()
}

func RenderPyprojectToml() (string, error) {
	return renderPyprojectToml()
}

func RenderPythonTemplate(templateName string, data any) (string, error) {
	return renderPythonTemplate(templateName, data)
}

func RenderPythonRawAsset(assetName string) (string, error) {
	return renderPythonRawAsset(assetName)
}
