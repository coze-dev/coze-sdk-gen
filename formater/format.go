package formater

import (
	"context"
	"fmt"

	"github.com/coze-dev/coze-sdk-gen/consts"
	"github.com/coze-dev/coze-sdk-gen/formater/python"
)

func Format(ctx context.Context, lang string, path string) error {
	switch lang {
	case consts.Python:
		return python.Format(ctx, path)
	default:
		return fmt.Errorf("unsupported language %q", lang)
	}
}
