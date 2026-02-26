package python

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type ClassMethodBlock struct {
	Name    string
	Content string
	IsChild bool
}

func mappingGeneratesSync(mapping *config.OperationMapping) bool {
	return true
}

func mappingGeneratesAsync(mapping *config.OperationMapping) bool {
	return true
}

func applyMethodDocstringOverrides(block string, classKey string, commentOverrides config.CommentOverrides) string {
	trimmedBlock := strings.TrimRight(block, "\n")
	if strings.TrimSpace(trimmedBlock) == "" {
		return block
	}
	lines := strings.Split(trimmedBlock, "\n")
	out := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		name, ok := ParseDefName(strings.TrimSpace(line))
		if !ok {
			out = append(out, line)
			continue
		}

		signatureEnd := i
		for signatureEnd+1 < len(lines) && !strings.HasSuffix(strings.TrimSpace(lines[signatureEnd]), ":") {
			signatureEnd++
		}
		for j := i; j <= signatureEnd && j < len(lines); j++ {
			out = append(out, lines[j])
		}

		methodKey := strings.TrimSpace(classKey) + "." + name
		rawDoc, exists := commentOverrides.MethodDocstrings[methodKey]
		if exists {
			docstring := strings.TrimSpace(rawDoc)
			if docstring != "" {
				nextNonEmpty := signatureEnd + 1
				for nextNonEmpty < len(lines) && strings.TrimSpace(lines[nextNonEmpty]) == "" {
					nextNonEmpty++
				}
				if nextNonEmpty >= len(lines) || !IsDocstringLine(strings.TrimSpace(lines[nextNonEmpty])) {
					indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
					docLines := RenderMethodDocstringLines(
						docstring,
						strings.TrimSpace(commentOverrides.MethodDocstringStyles[methodKey]),
						indent+"    ",
					)
					out = append(out, docLines...)
				}
			}
		}
		i = signatureEnd
	}
	return strings.Join(out, "\n")
}

func OrderClassMethodBlocks(blocks []ClassMethodBlock) []ClassMethodBlock {
	if len(blocks) == 0 {
		return blocks
	}

	prioritizedMethodNames := []string{"stream", "create", "clone", "retrieve", "update", "delete", "list"}
	prioritizedBuckets := make(map[string][]ClassMethodBlock, len(prioritizedMethodNames))
	prioritizedSet := make(map[string]struct{}, len(prioritizedMethodNames))
	for _, name := range prioritizedMethodNames {
		prioritizedSet[name] = struct{}{}
	}

	childMethods := make([]ClassMethodBlock, 0)
	otherMethods := make([]ClassMethodBlock, 0)
	for _, block := range blocks {
		if block.IsChild {
			childMethods = append(childMethods, block)
			continue
		}
		name := strings.TrimSpace(block.Name)
		if _, ok := prioritizedSet[name]; ok {
			prioritizedBuckets[name] = append(prioritizedBuckets[name], block)
			continue
		}
		otherMethods = append(otherMethods, block)
	}

	ordered := make([]ClassMethodBlock, 0, len(blocks))
	ordered = append(ordered, childMethods...)
	for _, name := range prioritizedMethodNames {
		ordered = append(ordered, prioritizedBuckets[name]...)
	}
	ordered = append(ordered, otherMethods...)
	return ordered
}

func renderChildClientProperty(
	meta PackageMeta,
	child childClient,
	async bool,
	classKey string,
	commentOverrides config.CommentOverrides,
) string {
	attribute := NormalizePythonIdentifier(child.Attribute)
	typeName := child.SyncClass
	if async {
		typeName = child.AsyncClass
	}
	module := strings.TrimSpace(child.Module)
	constructExpr := fmt.Sprintf("%s(base_url=self._base_url, requester=self._requester)", typeName)

	var buf bytes.Buffer
	buf.WriteString("    @property\n")
	buf.WriteString(fmt.Sprintf("    def %s(self) -> \"%s\":\n", attribute, typeName))
	methodKey := strings.TrimSpace(classKey) + "." + attribute
	if docstring, ok := commentOverrides.MethodDocstrings[methodKey]; ok {
		docstring = strings.TrimSpace(docstring)
		if docstring != "" {
			style := strings.TrimSpace(commentOverrides.MethodDocstringStyles[methodKey])
			WriteMethodDocstring(&buf, 2, docstring, style)
		}
	}
	buf.WriteString(fmt.Sprintf("        if not self._%s:\n", attribute))

	if module == "" {
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	} else if strings.HasPrefix(module, ".") {
		buf.WriteString(fmt.Sprintf("            from %s import %s\n\n", module, typeName))
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	} else {
		absModule := childImportModule(meta, module)
		buf.WriteString(fmt.Sprintf("            from %s import %s\n\n", absModule, typeName))
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	}
	buf.WriteString(fmt.Sprintf("        return self._%s\n", attribute))
	return buf.String()
}

func buildRenderQueryFields(
	doc *openapi.Document,
	details openapi.OperationDetails,
	mapping *config.OperationMapping,
	paramAliases map[string]string,
	argTypes map[string]string,
) []RenderQueryField {
	fields := make([]RenderQueryField, 0)
	if mapping != nil && len(mapping.QueryFields) > 0 {
		for _, field := range mapping.QueryFields {
			rawName := strings.TrimSpace(field.Name)
			if rawName == "" {
				continue
			}
			argName := OperationArgName(rawName, paramAliases)
			valueExpr := argName
			if field.UseValue {
				if field.Required {
					valueExpr = fmt.Sprintf("%s.value", argName)
				} else {
					valueExpr = fmt.Sprintf("%s.value if %s else None", argName, argName)
				}
			}
			if mapping != nil && len(mapping.QueryFieldValues) > 0 {
				if override, ok := mapping.QueryFieldValues[rawName]; ok && strings.TrimSpace(override) != "" {
					valueExpr = strings.TrimSpace(override)
				}
			}
			typeName := strings.TrimSpace(field.Type)
			if typeName == "" {
				typeName = "Any"
			}
			if !field.Required && !strings.HasPrefix(typeName, "Optional[") {
				typeName = "Optional[" + typeName + "]"
			}
			fields = append(fields, RenderQueryField{
				RawName:      rawName,
				ArgName:      argName,
				ValueExpr:    valueExpr,
				TypeName:     typeName,
				Required:     field.Required,
				DefaultValue: strings.TrimSpace(field.Default),
			})
		}
		return fields
	}

	for _, param := range details.QueryParameters {
		argName := OperationArgName(param.Name, paramAliases)
		typeName := TypeOverride(param.Name, param.Required, PythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		valueExpr := argName
		if mapping != nil && len(mapping.QueryFieldValues) > 0 {
			if override, ok := mapping.QueryFieldValues[param.Name]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
		}
		fields = append(fields, RenderQueryField{
			RawName:      param.Name,
			ArgName:      argName,
			ValueExpr:    valueExpr,
			TypeName:     typeName,
			Required:     param.Required,
			DefaultValue: "",
		})
	}
	return fields
}

func shouldSuppressSwaggerMethodDocstring(binding OperationBinding) bool {
	if binding.Mapping == nil {
		return false
	}
	pagination := strings.TrimSpace(binding.Mapping.Pagination)
	if pagination == "token" || pagination == "number" {
		return true
	}
	if methodOverride := strings.TrimSpace(binding.Mapping.HTTPMethodOverride); methodOverride != "" &&
		strings.ToLower(methodOverride) != strings.ToLower(binding.Details.Method) {
		return true
	}
	return false
}

func singleLineDescription(value string) string {
	lines := descriptionLines(value)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, " ")
}

func normalizeRichTextOverrideDocstring(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := make([]string, 0, 16)
	if strings.Contains(text, "\n") {
		for _, rawLine := range strings.Split(text, "\n") {
			line := strings.TrimSpace(rawLine)
			if line == "" {
				continue
			}
			lines = append(lines, splitAndNormalizeCollapsedRichTextLine(line)...)
		}
		return strings.Join(lines, "\n")
	}

	if !looksCollapsedRichTextDocstring(text) {
		return text
	}

	lines = append(lines, splitAndNormalizeCollapsedRichTextLine(text)...)
	if len(lines) == 0 {
		return text
	}
	return strings.Join(lines, "\n")
}

func looksCollapsedRichTextDocstring(text string) bool {
	if strings.Contains(text, "\n") {
		return false
	}
	if len([]rune(text)) < 80 {
		return false
	}
	return strings.Contains(text, "。 ") || strings.Contains(text, "！ ") || strings.Contains(text, "？ ") ||
		strings.Contains(text, " 接口限制 ")
}

func splitCollapsedRichTextSentences(text string) []string {
	lines := make([]string, 0, 16)
	var current strings.Builder
	flush := func() {
		line := strings.TrimSpace(current.String())
		if line != "" {
			lines = append(lines, line)
		}
		current.Reset()
	}
	for _, r := range text {
		current.WriteRune(r)
		switch r {
		case '。', '！', '？':
			flush()
		}
	}
	flush()
	return lines
}

func splitAndNormalizeCollapsedRichTextLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	chunks := []string{line}
	if looksCollapsedRichTextDocstring(line) {
		chunks = splitCollapsedRichTextSentences(line)
	}

	out := make([]string, 0, len(chunks)+2)
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		if strings.HasPrefix(chunk, "接口限制 ") {
			out = append(out, "接口限制")
			chunk = strings.TrimSpace(strings.TrimPrefix(chunk, "接口限制"))
		} else if strings.HasPrefix(chunk, "接口描述 ") {
			out = append(out, "接口描述")
			chunk = strings.TrimSpace(strings.TrimPrefix(chunk, "接口描述"))
		} else if strings.HasPrefix(chunk, "限制 说明 ") {
			out = append(out, "限制说明")
			chunk = strings.TrimSpace(strings.TrimPrefix(chunk, "限制 说明"))
		}
		if chunk != "" {
			out = append(out, chunk)
		}
	}
	return out
}

func buildSwaggerMethodDocstring(
	doc *openapi.Document,
	binding OperationBinding,
	details openapi.OperationDetails,
	pathParamNameMap map[string]string,
	queryFields []RenderQueryField,
	bodyFieldNames []string,
	requestBodyType string,
	returnType string,
	paramAliases map[string]string,
) string {
	if shouldSuppressSwaggerMethodDocstring(binding) {
		return ""
	}

	docLines := make([]string, 0, 16)
	summary := strings.TrimSpace(details.Summary)
	if summary != "" {
		docLines = append(docLines, summary)
	}
	description := strings.TrimSpace(details.Description)
	descriptionText := singleLineDescription(description)
	if descriptionText != "" && descriptionText != summary {
		if len(docLines) > 0 {
			docLines = append(docLines, "")
		}
		docLines = append(docLines, descriptionLines(description)...)
	}

	paramDocs := make([]string, 0, len(details.Parameters)+len(bodyFieldNames))
	for _, param := range details.PathParameters {
		description := singleLineDescription(param.Description)
		if description == "" {
			continue
		}
		argName := strings.TrimSpace(pathParamNameMap[param.Name])
		if argName == "" {
			argName = OperationArgName(param.Name, paramAliases)
		}
		paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
	}

	queryArgByRaw := make(map[string]string, len(queryFields))
	for _, field := range queryFields {
		queryArgByRaw[field.RawName] = field.ArgName
	}
	for _, param := range details.QueryParameters {
		description := singleLineDescription(param.Description)
		if description == "" {
			continue
		}
		argName := strings.TrimSpace(queryArgByRaw[param.Name])
		if argName == "" {
			argName = OperationArgName(param.Name, paramAliases)
		}
		paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
	}

	for _, param := range details.HeaderParameters {
		description := singleLineDescription(param.Description)
		if description == "" {
			continue
		}
		argName := OperationArgName(param.Name, paramAliases)
		paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
	}

	if len(bodyFieldNames) > 0 && details.RequestBodySchema != nil {
		seen := map[string]struct{}{}
		for _, bodyField := range bodyFieldNames {
			if _, exists := seen[bodyField]; exists {
				continue
			}
			seen[bodyField] = struct{}{}
			descriptionLines := schemaCommentLines(doc, BodyFieldSchema(doc, details.RequestBodySchema, bodyField))
			if len(descriptionLines) == 0 {
				continue
			}
			description := strings.Join(descriptionLines, " ")
			argName := OperationArgName(bodyField, paramAliases)
			paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
		}
	} else if requestBodyType != "" && details.RequestBody != nil {
		description := singleLineDescription(details.RequestBody.Description)
		if description != "" {
			paramDocs = append(paramDocs, fmt.Sprintf(":param body: %s", description))
		}
	}

	if len(paramDocs) > 0 {
		if len(docLines) > 0 {
			docLines = append(docLines, "")
		}
		docLines = append(docLines, paramDocs...)
	}

	if strings.TrimSpace(returnType) != "" && strings.TrimSpace(returnType) != "None" && details.Response != nil {
		description := singleLineDescription(details.Response.Description)
		if description != "" {
			if len(docLines) > 0 {
				docLines = append(docLines, "")
			}
			docLines = append(docLines, fmt.Sprintf(":return: %s", description))
		}
	}

	return strings.TrimSpace(strings.Join(docLines, "\n"))
}

func RenderOperationMethod(doc *openapi.Document, binding OperationBinding, async bool) string {
	return renderOperationMethodWithContext(doc, binding, async, "", "", config.CommentOverrides{}, nil)
}

func RenderOperationMethodWithComments(
	doc *openapi.Document,
	binding OperationBinding,
	async bool,
) string {
	return renderOperationMethodWithContext(doc, binding, async, "", "", config.CommentOverrides{}, nil)
}

func renderOperationMethodWithContext(
	doc *openapi.Document,
	binding OperationBinding,
	async bool,
	modulePath string,
	className string,
	commentOverrides config.CommentOverrides,
	classMethodNames map[string]struct{},
) string {
	details := binding.Details
	requestMethod := strings.ToLower(strings.TrimSpace(details.Method))
	paginationMode := ""
	returnType, returnCast := ReturnTypeInfo(doc, details.ResponseSchema)
	requestBodyType, bodyRequired := RequestBodyTypeInfo(doc, details.RequestBodySchema, details.RequestBody)
	ignoreHeaderParams := binding.Mapping != nil && binding.Mapping.IgnoreHeaderParams
	streamWrap := binding.Mapping != nil && binding.Mapping.StreamWrap
	streamWrapHandler := ""
	streamWrapFields := []string{}
	streamWrapAsyncYield := false
	// Keep a stable sync stream response variable name after removing mapping override.
	streamWrapSyncResponseVar := "response"
	bodyFieldValues := map[string]string{}
	paramAliases := map[string]string{}
	argTypes := map[string]string{}
	headersExpr := ""
	paginationRequestArg := "params"
	if binding.Mapping != nil && len(binding.Mapping.ParamAliases) > 0 {
		paramAliases = binding.Mapping.ParamAliases
	}
	if binding.Mapping != nil && len(binding.Mapping.ArgTypes) > 0 {
		argTypes = binding.Mapping.ArgTypes
	}
	if binding.Mapping != nil {
		if methodOverride := strings.TrimSpace(binding.Mapping.HTTPMethodOverride); methodOverride != "" {
			requestMethod = methodOverride
		}
		paginationMode = strings.TrimSpace(binding.Mapping.Pagination)
		if isTokenPagination(paginationMode) || isNumberPagination(paginationMode) {
			itemType := strings.TrimSpace(binding.Mapping.PaginationItemType)
			if itemType != "" {
				if isTokenPagination(paginationMode) {
					if async {
						returnType = fmt.Sprintf("AsyncTokenPaged[%s]", itemType)
					} else {
						returnType = fmt.Sprintf("TokenPaged[%s]", itemType)
					}
				} else {
					if async {
						returnType = fmt.Sprintf("AsyncNumberPaged[%s]", itemType)
					} else {
						returnType = fmt.Sprintf("NumberPaged[%s]", itemType)
					}
				}
			}
		}
		if mappedReturnType := strings.TrimSpace(binding.Mapping.ResponseType); mappedReturnType != "" {
			returnType = mappedReturnType
		}
		if async {
			if mappedAsyncReturnType := strings.TrimSpace(binding.Mapping.AsyncResponseType); mappedAsyncReturnType != "" {
				returnType = mappedAsyncReturnType
			}
		}
		shouldInferResponseModel := strings.TrimSpace(binding.Mapping.ResponseType) == "" &&
			strings.TrimSpace(binding.Mapping.AsyncResponseType) == "" &&
			(strings.TrimSpace(returnType) == "" || strings.TrimSpace(returnType) == "Dict[str, Any]")
		if shouldInferResponseModel {
			if inferredModelName, ok := inferBindingResponseModelName(doc, &config.Package{Name: binding.PackageName}, binding); ok {
				returnType = inferredModelName
				returnCast = inferredModelName
			}
		}
		streamWrapHandler = strings.TrimSpace(binding.Mapping.StreamWrapHandler)
		if len(binding.Mapping.StreamWrapFields) > 0 {
			streamWrapFields = append(streamWrapFields, binding.Mapping.StreamWrapFields...)
		}
		streamWrapAsyncYield = binding.Mapping.StreamWrapAsyncYield
		headersExpr = strings.TrimSpace(binding.Mapping.HeadersExpr)
		if len(binding.Mapping.BodyFieldValues) > 0 {
			for k, v := range binding.Mapping.BodyFieldValues {
				bodyFieldValues[k] = v
			}
		}
	}
	paginationRequestArg = inferPaginationRequestArg(requestMethod)
	returnCast = inferResponseCast(binding.Mapping, returnType, returnCast)
	if ignoreHeaderParams {
		details.HeaderParameters = nil
	}
	dataField := ""
	requestStream := false
	queryBuilder := "dump_exclude_none"
	bodyBuilder := "dump_exclude_none"
	if binding.Mapping != nil {
		dataField = strings.TrimSpace(binding.Mapping.DataField)
		requestStream = binding.Mapping.RequestStream
		queryBuilder = normalizeMapBuilder(binding.Mapping.QueryBuilder)
		bodyBuilder = normalizeMapBuilder(binding.Mapping.BodyBuilder)
	}
	autoDelegateTo := autoDelegateTarget(binding.MethodName, classMethodNames)
	autoDelegateExtraArgs := make([]string, 0, 1)
	if autoDelegateTo != "" {
		if streamArg, ok := mappingFixedStreamArg(binding.Mapping); ok {
			autoDelegateExtraArgs = append(autoDelegateExtraArgs, "stream="+streamArg)
		}
	}
	if async && requestStream && streamWrap && !streamWrapAsyncYield && binding.MethodName == "stream" {
		streamWrapAsyncYield = true
	}
	bodyFieldNames := make([]string, 0)
	bodyFixedValues := map[string]string{}
	filesFieldNames := make([]string, 0)
	filesFieldValues := map[string]string{}
	filesExpr := ""
	filesBeforeBody := false
	if binding.Mapping != nil && len(binding.Mapping.BodyFields) > 0 {
		bodyFieldNames = append(bodyFieldNames, binding.Mapping.BodyFields...)
	}
	if binding.Mapping != nil && len(binding.Mapping.BodyFixedValues) > 0 {
		for k, v := range binding.Mapping.BodyFixedValues {
			bodyFixedValues[k] = v
		}
	}
	if binding.Mapping != nil && len(binding.Mapping.FilesFields) > 0 {
		filesFieldNames = append(filesFieldNames, binding.Mapping.FilesFields...)
	}
	if binding.Mapping != nil && len(binding.Mapping.FilesFieldValues) > 0 {
		for k, v := range binding.Mapping.FilesFieldValues {
			filesFieldValues[k] = v
		}
	}
	if binding.Mapping != nil {
		filesExpr = strings.TrimSpace(binding.Mapping.FilesExpr)
		filesBeforeBody = binding.Mapping.FilesBeforeBody
	}
	if !shouldGenerateImplicitRequestBody(details.Method, binding.Mapping, details.RequestBodySchema) {
		requestBodyType = ""
		bodyRequired = false
	}
	queryFields := buildRenderQueryFields(doc, details, binding.Mapping, paramAliases, argTypes)
	signatureQueryFields := OrderSignatureQueryFields(queryFields, binding.Mapping, async)
	bodyRequiredSet := map[string]bool{}
	if details.RequestBodySchema != nil {
		for _, requiredName := range details.RequestBodySchema.Required {
			bodyRequiredSet[requiredName] = true
		}
	}
	if binding.Mapping != nil && len(binding.Mapping.BodyRequiredFields) > 0 {
		bodyRequiredSet = map[string]bool{}
		for _, requiredName := range binding.Mapping.BodyRequiredFields {
			bodyRequiredSet[requiredName] = true
		}
	}

	pathParamNameMap := map[string]string{}
	signatureArgs := make([]string, 0)
	signatureArgNames := map[string]struct{}{}
	for _, param := range details.PathParameters {
		name := OperationArgName(param.Name, paramAliases)
		pathParamNameMap[param.Name] = name
		typeName := TypeOverride(param.Name, true, PythonTypeForSchema(doc, param.Schema, true), argTypes)
		signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		signatureArgNames[name] = struct{}{}
	}
	for _, field := range signatureQueryFields {
		defaultValue := strings.TrimSpace(field.DefaultValue)
		if override, ok := OperationArgDefault(binding.Mapping, field.RawName, field.ArgName, async); ok {
			defaultValue = override
		}
		if field.Required && defaultValue == "" {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", field.ArgName, field.TypeName))
		} else if defaultValue != "" {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", field.ArgName, field.TypeName, defaultValue))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", field.ArgName, field.TypeName))
		}
		signatureArgNames[field.ArgName] = struct{}{}
	}
	for _, param := range details.HeaderParameters {
		name := OperationArgName(param.Name, paramAliases)
		typeName := TypeOverride(param.Name, param.Required, PythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		if defaultValue, ok := OperationArgDefault(binding.Mapping, param.Name, name, async); ok {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", name, typeName, defaultValue))
		} else if param.Required {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", name, typeName))
		}
		signatureArgNames[name] = struct{}{}
	}

	if len(bodyFieldNames) > 0 {
		for _, bodyField := range bodyFieldNames {
			argName := OperationArgName(bodyField, paramAliases)
			if _, exists := signatureArgNames[argName]; exists {
				continue
			}
			fieldSchema := BodyFieldSchema(doc, details.RequestBodySchema, bodyField)
			required := bodyRequiredSet[bodyField]
			typeName := TypeOverride(bodyField, required, PythonTypeForSchema(doc, fieldSchema, required), argTypes)
			if defaultValue, ok := OperationArgDefault(binding.Mapping, bodyField, argName, async); ok {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", argName, typeName, defaultValue))
			} else if required {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", argName, typeName))
			} else {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", argName, typeName))
			}
			signatureArgNames[argName] = struct{}{}
		}
	} else if requestBodyType != "" {
		if defaultValue, ok := OperationArgDefault(binding.Mapping, "body", "body", async); ok {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = %s", requestBodyType, defaultValue))
		} else if bodyRequired {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: %s", requestBodyType))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = None", requestBodyType))
		}
	}
	if len(filesFieldNames) > 0 {
		for _, filesField := range filesFieldNames {
			argName := OperationArgName(filesField, paramAliases)
			if _, exists := signatureArgNames[argName]; exists {
				continue
			}
			fieldSchema := BodyFieldSchema(doc, details.RequestBodySchema, filesField)
			required := bodyRequiredSet[filesField]
			typeName := TypeOverride(filesField, required, PythonTypeForSchema(doc, fieldSchema, required), argTypes)
			if defaultValue, ok := OperationArgDefault(binding.Mapping, filesField, argName, async); ok {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", argName, typeName, defaultValue))
			} else if required {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", argName, typeName))
			} else {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", argName, typeName))
			}
			signatureArgNames[argName] = struct{}{}
		}
	}
	includeKwargsHeaders := true
	if includeKwargsHeaders {
		signatureArgs = append(signatureArgs, "**kwargs")
	}
	signatureArgs = NormalizeSignatureArgs(signatureArgs)

	methodKeyword := "def"
	requestCall := "self._requester.request"
	if async {
		methodKeyword = "async def"
		requestCall = "await self._requester.arequest"
	}
	headersAssigned := false

	var buf bytes.Buffer
	returnAnnotation := fmt.Sprintf(" -> %s", returnType)
	nonKwargsSignatureArgCount := 0
	for _, argDecl := range signatureArgs {
		if IsKwargsSignatureArg(argDecl) {
			continue
		}
		nonKwargsSignatureArgCount++
	}
	compactSignature := nonKwargsSignatureArgCount <= 2
	if binding.Mapping != nil {
		if !async {
			if binding.Mapping.ForceMultilineSignatureSync {
				compactSignature = false
			}
		}
	}
	if compactSignature {
		if len(signatureArgs) == 0 {
			buf.WriteString(fmt.Sprintf("    %s %s(self)%s:\n", methodKeyword, binding.MethodName, returnAnnotation))
		} else if len(signatureArgs) == 1 && IsKwargsSignatureArg(signatureArgs[0]) {
			buf.WriteString(fmt.Sprintf("    %s %s(self, %s)%s:\n", methodKeyword, binding.MethodName, signatureArgs[0], returnAnnotation))
		} else {
			buf.WriteString(fmt.Sprintf("    %s %s(self, *, %s)%s:\n", methodKeyword, binding.MethodName, strings.Join(signatureArgs, ", "), returnAnnotation))
		}
	} else {
		buf.WriteString(fmt.Sprintf("    %s %s(\n", methodKeyword, binding.MethodName))
		buf.WriteString("        self,\n")
		buf.WriteString("        *,\n")
		for _, arg := range signatureArgs {
			buf.WriteString(fmt.Sprintf("        %s,\n", arg))
		}
		buf.WriteString(fmt.Sprintf("    )%s:\n", returnAnnotation))
	}
	if binding.Mapping != nil && len(binding.Mapping.PreDocstringCode) > 0 {
		for _, block := range binding.Mapping.PreDocstringCode {
			AppendIndentedCode(&buf, block, 2)
		}
	}
	methodDocstring := buildSwaggerMethodDocstring(doc, binding, details, pathParamNameMap, queryFields, bodyFieldNames, requestBodyType, returnType, paramAliases)
	docstringStyle := "block"
	docstringFromOverride := false
	methodKey := ""
	if modulePath != "" && className != "" {
		methodKey = strings.TrimSpace(modulePath) + "." + strings.TrimSpace(className) + "." + binding.MethodName
	}
	if methodDocstring == "" && methodKey != "" {
		if raw, ok := commentOverrides.RichTextMethodDocstrings[methodKey]; ok {
			methodDocstring = normalizeRichTextOverrideDocstring(raw)
			docstringFromOverride = true
		}
	}
	if methodDocstring == "" && methodKey != "" {
		if raw, ok := commentOverrides.MethodDocstrings[methodKey]; ok {
			methodDocstring = strings.TrimSpace(raw)
			docstringFromOverride = true
		}
	}
	if methodDocstring != "" && methodKey != "" && swaggerDescriptionLooksLikeRichText(details.Description) {
		if raw, ok := commentOverrides.RichTextMethodDocstrings[methodKey]; ok {
			methodDocstring = normalizeRichTextOverrideDocstring(raw)
			docstringFromOverride = true
		}
	}
	if methodKey != "" && docstringFromOverride {
		if style := strings.TrimSpace(commentOverrides.MethodDocstringStyles[methodKey]); style != "" {
			docstringStyle = style
		}
	}
	if methodDocstring != "" {
		WriteMethodDocstring(&buf, 2, methodDocstring, docstringStyle)
	}
	if autoDelegateTo != "" {
		callArgs := BuildAutoDelegateCallArgs(signatureArgs, autoDelegateExtraArgs)
		delegateAsyncYield := async && binding.MethodName == "stream"
		RenderDelegatedCall(&buf, autoDelegateTo, callArgs, async, delegateAsyncYield)
		return buf.String()
	}

	urlPath := details.Path
	for rawName, pyName := range pathParamNameMap {
		if rawName == pyName {
			continue
		}
		urlPath = strings.ReplaceAll(urlPath, "{"+rawName+"}", "{"+pyName+"}")
	}
	buf.WriteString(fmt.Sprintf("        url = f\"{self._base_url}%s\"\n", urlPath))

	if len(queryFields) > 0 && !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) {
		if queryBuilder == "raw" {
			buf.WriteString("        params = {\n")
			itemIndent := "            "
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			buf.WriteString("        }\n")
		} else {
			buf.WriteString(fmt.Sprintf("        params = %s(\n", queryBuilder))
			buf.WriteString("            {\n")
			itemIndent := "                "
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			buf.WriteString("            }\n")
			buf.WriteString("        )\n")
		}
	}

	if headersExpr == "" && includeKwargsHeaders && (isTokenPagination(paginationMode) || isNumberPagination(paginationMode) || len(details.HeaderParameters) > 0) {
		buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n\n")
	}

	if len(details.HeaderParameters) > 0 {
		buf.WriteString("        header_values = dict(headers or {})\n")
		for _, param := range details.HeaderParameters {
			name := OperationArgName(param.Name, paramAliases)
			if param.Required {
				buf.WriteString(fmt.Sprintf("        header_values[%q] = str(%s)\n", param.Name, name))
			} else {
				buf.WriteString(fmt.Sprintf("        if %s is not None:\n", name))
				buf.WriteString(fmt.Sprintf("            header_values[%q] = str(%s)\n", param.Name, name))
			}
		}
		buf.WriteString("        headers = header_values\n")
	}
	if headersExpr != "" {
		buf.WriteString(fmt.Sprintf("        headers = %s\n", headersExpr))
		headersAssigned = true
		if isTokenPagination(paginationMode) || isNumberPagination(paginationMode) {
			buf.WriteString("\n")
		}
	}
	if (isTokenPagination(paginationMode) || isNumberPagination(paginationMode)) && binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0 {
		for _, block := range binding.Mapping.PreBodyCode {
			AppendIndentedCode(&buf, block, 2)
		}
		if headersExpr == "" {
			buf.WriteString("\n")
		}
	}
	includePaginationHeaders := true
	paginationRequestMethod := strings.ToUpper(requestMethod)

	if isTokenPagination(paginationMode) && binding.Mapping != nil {
		dataClass := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		pageTokenField := strings.TrimSpace(binding.Mapping.PaginationPageTokenField)
		if pageTokenField == "" {
			pageTokenField = "page_token"
		}
		pageSizeField := strings.TrimSpace(binding.Mapping.PaginationPageSizeField)
		if pageSizeField == "" {
			pageSizeField = "page_size"
		}
		tokenExpr := "\"\""
		sizeExpr := "20"
		for _, field := range queryFields {
			if field.RawName == pageTokenField {
				tokenExpr = field.ArgName
				if !field.Required {
					tokenExpr = fmt.Sprintf("%s or \"\"", field.ArgName)
				}
			}
			if field.RawName == pageSizeField {
				sizeExpr = field.ArgName
			}
		}
		EnsureTrailingNewlines(&buf, 2)
		if async {
			buf.WriteString("        async def request_maker(i_page_token: str, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return await self._requester.amake_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if queryBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
			} else {
				buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
				buf.WriteString("                    {\n")
			}
			itemIndent := "                        "
			if queryBuilder == "raw" {
				itemIndent = "                    "
			}
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				if field.RawName == pageTokenField {
					valueExpr = "i_page_token"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			if queryBuilder == "raw" {
				buf.WriteString("                },\n")
			} else {
				buf.WriteString("                    }\n")
				buf.WriteString("                ),\n")
			}
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return await AsyncTokenPaged.build(\n")
			buf.WriteString(fmt.Sprintf("            page_token=%s,\n", tokenExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		} else {
			buf.WriteString("        def request_maker(i_page_token: str, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return self._requester.make_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if queryBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
			} else {
				buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
				buf.WriteString("                    {\n")
			}
			itemIndent := "                        "
			if queryBuilder == "raw" {
				itemIndent = "                    "
			}
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				if field.RawName == pageTokenField {
					valueExpr = "i_page_token"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			if queryBuilder == "raw" {
				buf.WriteString("                },\n")
			} else {
				buf.WriteString("                    }\n")
				buf.WriteString("                ),\n")
			}
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return TokenPaged(\n")
			buf.WriteString(fmt.Sprintf("            page_token=%s,\n", tokenExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		}
		return buf.String()
	}
	if isNumberPagination(paginationMode) && binding.Mapping != nil {
		dataClass := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		pageNumField := strings.TrimSpace(binding.Mapping.PaginationPageNumField)
		if pageNumField == "" {
			pageNumField = "page_num"
		}
		pageSizeField := strings.TrimSpace(binding.Mapping.PaginationPageSizeField)
		if pageSizeField == "" {
			pageSizeField = "page_size"
		}
		pageNumExpr := "1"
		sizeExpr := "20"
		for _, field := range queryFields {
			if field.RawName == pageNumField {
				pageNumExpr = field.ArgName
			}
			if field.RawName == pageSizeField {
				sizeExpr = field.ArgName
			}
		}
		EnsureTrailingNewlines(&buf, 2)
		if async {
			buf.WriteString("        async def request_maker(i_page_num: int, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return await self._requester.amake_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if queryBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
			} else {
				buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
				buf.WriteString("                    {\n")
			}
			itemIndent := "                        "
			if queryBuilder == "raw" {
				itemIndent = "                    "
			}
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				if field.RawName == pageNumField {
					valueExpr = "i_page_num"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			if queryBuilder == "raw" {
				buf.WriteString("                },\n")
			} else {
				buf.WriteString("                    }\n")
				buf.WriteString("                ),\n")
			}
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return await AsyncNumberPaged.build(\n")
			buf.WriteString(fmt.Sprintf("            page_num=%s,\n", pageNumExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		} else {
			buf.WriteString("        def request_maker(i_page_num: int, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return self._requester.make_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if queryBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
			} else {
				buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
				buf.WriteString("                    {\n")
			}
			itemIndent := "                        "
			if queryBuilder == "raw" {
				itemIndent = "                    "
			}
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				if field.RawName == pageNumField {
					valueExpr = "i_page_num"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			if queryBuilder == "raw" {
				buf.WriteString("                },\n")
			} else {
				buf.WriteString("                    }\n")
				buf.WriteString("                ),\n")
			}
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return NumberPaged(\n")
			buf.WriteString(fmt.Sprintf("            page_num=%s,\n", pageNumExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		}
		return buf.String()
	}

	if headersExpr == "" && includeKwargsHeaders && !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) && len(details.HeaderParameters) == 0 {
		buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n")
		headersAssigned = true
	}
	if binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0 {
		for _, block := range binding.Mapping.PreBodyCode {
			AppendIndentedCode(&buf, block, 2)
		}
	}
	bodyVarAssign := "body"
	renderFilesAssignment := func() bool {
		if filesExpr != "" {
			buf.WriteString(fmt.Sprintf("        files = %s\n", filesExpr))
			return true
		}
		if len(filesFieldNames) == 0 {
			return false
		}
		if len(filesFieldNames) == 1 {
			fieldName := filesFieldNames[0]
			argName := OperationArgName(fieldName, paramAliases)
			valueExpr := argName
			if override, ok := filesFieldValues[fieldName]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
			buf.WriteString(fmt.Sprintf("        files = {%q: %s}\n", fieldName, valueExpr))
			return true
		}

		buf.WriteString("        files = {\n")
		for _, filesField := range filesFieldNames {
			argName := OperationArgName(filesField, paramAliases)
			valueExpr := argName
			if override, ok := filesFieldValues[filesField]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
			buf.WriteString(fmt.Sprintf("            %q: %s,\n", filesField, valueExpr))
		}
		buf.WriteString("        }\n")
		return true
	}

	if filesBeforeBody {
		renderFilesAssignment()
	}

	if len(bodyFieldNames) > 0 {
		if bodyBuilder == "raw" {
			buf.WriteString(fmt.Sprintf("        %s = {\n", bodyVarAssign))
			itemIndent := "            "
			for _, bodyField := range bodyFieldNames {
				argName := OperationArgName(bodyField, paramAliases)
				valueExpr := argName
				if override, ok := bodyFieldValues[bodyField]; ok && strings.TrimSpace(override) != "" {
					valueExpr = strings.TrimSpace(override)
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, bodyField, valueExpr))
			}
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("        }\n")
		} else {
			buf.WriteString(fmt.Sprintf("        %s = %s(\n", bodyVarAssign, bodyBuilder))
			buf.WriteString("            {\n")
			itemIndent := "                "
			for _, bodyField := range bodyFieldNames {
				argName := OperationArgName(bodyField, paramAliases)
				valueExpr := argName
				if override, ok := bodyFieldValues[bodyField]; ok && strings.TrimSpace(override) != "" {
					valueExpr = strings.TrimSpace(override)
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, bodyField, valueExpr))
			}
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("            }\n")
			buf.WriteString("        )\n")
		}
	} else if len(bodyFixedValues) > 0 {
		if bodyBuilder == "raw" {
			buf.WriteString(fmt.Sprintf("        %s = {\n", bodyVarAssign))
			itemIndent := "            "
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("        }\n")
		} else {
			buf.WriteString(fmt.Sprintf("        %s = %s(\n", bodyVarAssign, bodyBuilder))
			buf.WriteString("            {\n")
			itemIndent := "                "
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("            }\n")
			buf.WriteString("        )\n")
		}
	} else if requestBodyType != "" {
		buf.WriteString("        request_body: Any = None\n")
		if bodyRequired {
			buf.WriteString("        request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		} else {
			buf.WriteString("        if body is not None:\n")
			buf.WriteString("            request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		}
	}
	if !filesBeforeBody {
		renderFilesAssignment()
	}
	if headersExpr == "" && !headersAssigned && includeKwargsHeaders && !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) && len(details.HeaderParameters) == 0 {
		needsBlankLine := len(queryFields) > 0 ||
			len(bodyFieldNames) > 0 ||
			len(bodyFixedValues) > 0 ||
			filesExpr != "" ||
			len(filesFieldNames) > 0 ||
			requestBodyType != "" ||
			strings.EqualFold(requestMethod, "delete") ||
			(binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0)
		if needsBlankLine {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n\n")
		} else {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n")
		}
	}

	castExpr := "None"
	if returnCast != "" {
		castExpr = returnCast
	}
	bodyArgExpr := ""
	if len(bodyFieldNames) > 0 {
		bodyArgExpr = "body"
	} else if len(bodyFixedValues) > 0 {
		bodyArgExpr = "body"
	} else if requestBodyType != "" {
		bodyArgExpr = "request_body"
	}
	callArgs := []string{
		fmt.Sprintf("%q", requestMethod),
		"url",
	}
	type requestCallArg struct {
		Expr string
	}
	optionalArgs := make([]requestCallArg, 0, 8)
	streamLiteral := "False"
	if requestStream {
		streamLiteral = "True"
	}
	optionalArgs = append(optionalArgs, requestCallArg{Expr: streamLiteral})
	castExprValue := fmt.Sprintf("cast=%s", castExpr)
	optionalArgs = append(optionalArgs, requestCallArg{Expr: castExprValue})
	if len(queryFields) > 0 {
		optionalArgs = append(optionalArgs, requestCallArg{Expr: "params=params"})
	}
	hasHeadersArg := true
	if hasHeadersArg {
		optionalArgs = append(optionalArgs, requestCallArg{Expr: "headers=headers"})
	}
	if bodyArgExpr != "" {
		optionalArgs = append(optionalArgs, requestCallArg{Expr: fmt.Sprintf("body=%s", bodyArgExpr)})
	}
	if filesExpr != "" || len(filesFieldNames) > 0 {
		optionalArgs = append(optionalArgs, requestCallArg{Expr: "files=files"})
	}
	if dataField != "" {
		optionalArgs = append(optionalArgs, requestCallArg{Expr: fmt.Sprintf("data_field=%q", dataField)})
	}
	for _, item := range optionalArgs {
		callArgs = append(callArgs, item.Expr)
	}
	requestExpr := fmt.Sprintf("%s(%s)", requestCall, strings.Join(callArgs, ", "))
	forceMultilineRequestCall := async && binding.Mapping != nil && binding.Mapping.ForceMultilineRequestCallAsync
	if binding.Mapping != nil && binding.Mapping.ResponseUnwrapListFirst {
		buf.WriteString(fmt.Sprintf("        res = %s\n", requestExpr))
		buf.WriteString("        data = res.data[0]\n")
		buf.WriteString("        data._raw_response = res._raw_response\n")
		buf.WriteString("        return data\n")
		return buf.String()
	}
	if requestStream && streamWrap {
		fieldLiterals := make([]string, 0, len(streamWrapFields))
		for _, field := range streamWrapFields {
			trimmed := strings.TrimSpace(field)
			if trimmed == "" {
				continue
			}
			fieldLiterals = append(fieldLiterals, fmt.Sprintf("%q", trimmed))
		}
		if async {
			if forceMultilineRequestCall {
				buf.WriteString(fmt.Sprintf("        resp: AsyncIteratorHTTPResponse[str] = %s(\n", requestCall))
				for _, arg := range callArgs {
					buf.WriteString(fmt.Sprintf("            %s,\n", arg))
				}
				buf.WriteString("        )\n")
			} else {
				buf.WriteString(fmt.Sprintf("        resp: AsyncIteratorHTTPResponse[str] = %s\n", requestExpr))
			}
			if streamWrapAsyncYield {
				buf.WriteString("        async for item in AsyncStream(\n")
				buf.WriteString("            resp.data,\n")
				if len(fieldLiterals) > 0 {
					buf.WriteString(fmt.Sprintf("            fields=[%s],\n", strings.Join(fieldLiterals, ", ")))
				}
				if streamWrapHandler != "" {
					buf.WriteString(fmt.Sprintf("            handler=%s,\n", streamWrapHandler))
				}
				buf.WriteString("            raw_response=resp._raw_response,\n")
				buf.WriteString("        ):\n")
				buf.WriteString("            yield item\n")
			} else {
				buf.WriteString("        return AsyncStream(\n")
				buf.WriteString("            resp.data,\n")
				if len(fieldLiterals) > 0 {
					buf.WriteString(fmt.Sprintf("            fields=[%s],\n", strings.Join(fieldLiterals, ", ")))
				}
				if streamWrapHandler != "" {
					buf.WriteString(fmt.Sprintf("            handler=%s,\n", streamWrapHandler))
				}
				buf.WriteString("            raw_response=resp._raw_response,\n")
				buf.WriteString("        )\n")
			}
		} else {
			if forceMultilineRequestCall {
				buf.WriteString(fmt.Sprintf("        %s: IteratorHTTPResponse[str] = %s(\n", streamWrapSyncResponseVar, requestCall))
				for _, arg := range callArgs {
					buf.WriteString(fmt.Sprintf("            %s,\n", arg))
				}
				buf.WriteString("        )\n")
			} else {
				buf.WriteString(fmt.Sprintf("        %s: IteratorHTTPResponse[str] = %s\n", streamWrapSyncResponseVar, requestExpr))
			}
			buf.WriteString("        return Stream(\n")
			buf.WriteString(fmt.Sprintf("            %s._raw_response,\n", streamWrapSyncResponseVar))
			buf.WriteString(fmt.Sprintf("            %s.data,\n", streamWrapSyncResponseVar))
			if len(fieldLiterals) > 0 {
				buf.WriteString(fmt.Sprintf("            fields=[%s],\n", strings.Join(fieldLiterals, ", ")))
			}
			if streamWrapHandler != "" {
				buf.WriteString(fmt.Sprintf("            handler=%s,\n", streamWrapHandler))
			}
			buf.WriteString("        )\n")
		}
	} else {
		if forceMultilineRequestCall {
			buf.WriteString(fmt.Sprintf("        return %s(\n", requestCall))
			for _, arg := range callArgs {
				buf.WriteString(fmt.Sprintf("            %s,\n", arg))
			}
			buf.WriteString("        )\n")
		} else {
			buf.WriteString(fmt.Sprintf("        return %s\n", requestExpr))
		}
	}

	return buf.String()
}

func autoDelegateTarget(methodName string, classMethodNames map[string]struct{}) string {
	if classMethodNames == nil {
		return ""
	}
	name := strings.TrimSpace(methodName)
	if name != "create" && name != "stream" {
		return ""
	}
	if _, ok := classMethodNames["_create"]; !ok {
		return ""
	}
	return "_create"
}

func mappingFixedStreamArg(mapping *config.OperationMapping) (string, bool) {
	if mapping == nil || len(mapping.BodyFixedValues) == 0 {
		return "", false
	}
	value, ok := mapping.BodyFixedValues["stream"]
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
}
