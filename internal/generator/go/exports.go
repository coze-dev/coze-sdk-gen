package gogen

import (
	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func FindGoOperationPath(cfg *config.Config, doc *openapi.Document, sdkMethod string, method string, fallback string) (string, error) {
	return findGoOperationPath(cfg, doc, sdkMethod, method, fallback)
}
