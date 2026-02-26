package python

import (
	"fmt"
	"strings"
)

func normalizeAutoInferredImports(content string) string {
	sanitized := stripPythonStringsAndComments(content)
	ensureNamed := func(module string, symbol string, includeQuotedAnnotation bool) {
		if symbol == "" || module == "" {
			return
		}
		used := containsIdentifierOutsideImportSection(sanitized, symbol)
		if !used && includeQuotedAnnotation {
			used = containsQuotedAnnotationOutsideImportSection(content, symbol)
		}
		if !used {
			return
		}
		if declaresIdentifier(sanitized, symbol) {
			return
		}
		content = ensureNamedImport(content, module, symbol)
	}
	ensureModule := func(module string) {
		if module == "" {
			return
		}
		if !containsModuleReferenceOutsideImportSection(sanitized, module) {
			return
		}
		if declaresIdentifier(sanitized, module) {
			return
		}
		content = ensureModuleImport(content, module)
	}

	for _, module := range []string{"base64", "httpx", "json", "os", "time", "warnings"} {
		ensureModule(module)
	}

	knownFromImports := []struct {
		Module                  string
		Symbols                 []string
		IncludeQuotedAnnotation bool
	}{
		{Module: "typing", Symbols: []string{"Any", "AsyncIterator", "Dict", "IO", "List", "Optional", "TYPE_CHECKING", "Tuple", "Union", "overload"}},
		{Module: "typing_extensions", Symbols: []string{"Literal"}},
		{Module: "pathlib", Symbols: []string{"Path"}},
		{Module: "pydantic", Symbols: []string{"Field", "field_validator"}},
		{Module: "cozepy", Symbols: []string{"AudioFormat"}},
		{Module: "cozepy.audio.voiceprint_groups.features", Symbols: []string{"UserInfo"}},
		{Module: "cozepy.bots", Symbols: []string{"PublishStatus"}},
		{Module: "cozepy.chat", Symbols: []string{"ChatEvent", "ChatUsage", "Message", "MessageContentType", "MessageRole", "_chat_stream_handler"}},
		{Module: "cozepy.datasets.documents", Symbols: []string{"Document", "DocumentBase", "DocumentChunkStrategy", "DocumentUpdateRule"}},
		{Module: "cozepy.exception", Symbols: []string{"CozeAPIError"}},
		{Module: "cozepy.files", Symbols: []string{"FileTypes", "_try_fix_file"}},
		{Module: "cozepy.model", Symbols: []string{"AsyncIteratorHTTPResponse", "AsyncLastIDPaged", "AsyncNumberPaged", "AsyncStream", "AsyncTokenPaged", "CozeModel", "DynamicStrEnum", "FileHTTPResponse", "IteratorHTTPResponse", "LastIDPaged", "LastIDPagedResponse", "ListResponse", "NumberPaged", "NumberPagedResponse", "Stream", "TokenPaged", "TokenPagedResponse"}},
		{Module: "cozepy.request", Symbols: []string{"Requester"}},
		{Module: "cozepy.util", Symbols: []string{"base64_encode_string", "dump_exclude_none", "remove_none_values", "remove_url_trailing_slash"}},
		{Module: "cozepy.workspaces", Symbols: []string{"WorkspaceRoleType"}},
		{Module: ".versions", Symbols: []string{"WorkflowUserInfo"}},
	}
	for _, entry := range knownFromImports {
		for _, symbol := range entry.Symbols {
			ensureNamed(entry.Module, symbol, entry.IncludeQuotedAnnotation)
		}
	}
	feedbackSymbols := make([]string, 0, 2)
	if containsQuotedAnnotationOutsideImportSection(content, "AsyncMessagesFeedbackClient") {
		feedbackSymbols = append(feedbackSymbols, "AsyncMessagesFeedbackClient")
	}
	if containsQuotedAnnotationOutsideImportSection(content, "ConversationsMessagesFeedbackClient") {
		feedbackSymbols = append(feedbackSymbols, "ConversationsMessagesFeedbackClient")
	}
	if len(feedbackSymbols) > 0 {
		content = ensureTypeCheckingNamedImport(content, ".feedback", feedbackSymbols)
	}
	if containsIdentifierOutsideImportSection(sanitized, "HTTPRequest") &&
		!declaresIdentifier(sanitized, "HTTPRequest") &&
		!isSymbolImportedFromModule(content, "cozepy.model", "HTTPRequest") &&
		!isSymbolImportedFromModule(content, "cozepy.request", "HTTPRequest") {
		content = ensureNamedImport(content, "cozepy.model", "HTTPRequest")
	}
	return content
}

func containsIdentifierOutsideImportSection(content string, ident string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if lineHasIdentifier(line, ident) {
			return true
		}
	}
	return false
}

func containsQuotedAnnotationOutsideImportSection(content string, ident string) bool {
	if ident == "" {
		return false
	}
	singleQuoted := "'" + ident + "'"
	doubleQuoted := `"` + ident + `"`
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if !strings.Contains(line, ":") && !strings.Contains(line, "->") {
			continue
		}
		if strings.Contains(line, singleQuoted) || strings.Contains(line, doubleQuoted) {
			return true
		}
	}
	return false
}

func containsModuleReferenceOutsideImportSection(content string, module string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if lineHasModuleReference(line, module) {
			return true
		}
	}
	return false
}

func lineHasModuleReference(line string, module string) bool {
	if module == "" {
		return false
	}
	for start := 0; start < len(line); {
		idx := strings.Index(line[start:], module)
		if idx < 0 {
			return false
		}
		idx += start
		beforeOK := idx == 0 || !isIdentifierRune(line[idx-1])
		afterIdx := idx + len(module)
		if beforeOK && afterIdx < len(line) && line[afterIdx] == '.' {
			return true
		}
		start = idx + len(module)
	}
	return false
}

func stripPythonStringsAndComments(content string) string {
	const (
		scanCode = iota
		scanSingleQuote
		scanDoubleQuote
		scanTripleSingleQuote
		scanTripleDoubleQuote
		scanComment
	)

	state := scanCode
	escaped := false
	var b strings.Builder
	b.Grow(len(content))

	for i := 0; i < len(content); i++ {
		ch := content[i]

		switch state {
		case scanComment:
			if ch == '\n' {
				state = scanCode
				b.WriteByte('\n')
			} else {
				b.WriteByte(' ')
			}
		case scanSingleQuote:
			if ch == '\n' {
				state = scanCode
				escaped = false
				b.WriteByte('\n')
				continue
			}
			if escaped {
				escaped = false
				b.WriteByte(' ')
				continue
			}
			if ch == '\\' {
				escaped = true
				b.WriteByte(' ')
				continue
			}
			if ch == '\'' {
				state = scanCode
			}
			b.WriteByte(' ')
		case scanDoubleQuote:
			if ch == '\n' {
				state = scanCode
				escaped = false
				b.WriteByte('\n')
				continue
			}
			if escaped {
				escaped = false
				b.WriteByte(' ')
				continue
			}
			if ch == '\\' {
				escaped = true
				b.WriteByte(' ')
				continue
			}
			if ch == '"' {
				state = scanCode
			}
			b.WriteByte(' ')
		case scanTripleSingleQuote:
			if ch == '\n' {
				b.WriteByte('\n')
				continue
			}
			if ch == '\'' && i+2 < len(content) && content[i+1] == '\'' && content[i+2] == '\'' {
				b.WriteString("   ")
				i += 2
				state = scanCode
				continue
			}
			b.WriteByte(' ')
		case scanTripleDoubleQuote:
			if ch == '\n' {
				b.WriteByte('\n')
				continue
			}
			if ch == '"' && i+2 < len(content) && content[i+1] == '"' && content[i+2] == '"' {
				b.WriteString("   ")
				i += 2
				state = scanCode
				continue
			}
			b.WriteByte(' ')
		default:
			if ch == '#' {
				state = scanComment
				b.WriteByte(' ')
				continue
			}
			if ch == '\'' {
				if i+2 < len(content) && content[i+1] == '\'' && content[i+2] == '\'' {
					state = scanTripleSingleQuote
					b.WriteString("   ")
					i += 2
					continue
				}
				state = scanSingleQuote
				b.WriteByte(' ')
				continue
			}
			if ch == '"' {
				if i+2 < len(content) && content[i+1] == '"' && content[i+2] == '"' {
					state = scanTripleDoubleQuote
					b.WriteString("   ")
					i += 2
					continue
				}
				state = scanDoubleQuote
				b.WriteByte(' ')
				continue
			}
			b.WriteByte(ch)
		}
	}

	return b.String()
}

func isSymbolImportedFromModule(content string, module string, symbol string) bool {
	if module == "" || symbol == "" {
		return false
	}
	lines := strings.Split(content, "\n")
	importPrefix := "from " + module + " import "
	insertIdx := topImportInsertIndex(lines)
	for i := 0; i < insertIdx; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line != trimmed {
			continue
		}
		if !strings.HasPrefix(trimmed, importPrefix) {
			continue
		}
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		for _, name := range strings.Split(namesPart, ",") {
			if strings.TrimSpace(name) == symbol {
				return true
			}
		}
	}
	return false
}

func ensureTypeCheckingNamedImport(content string, module string, symbols []string) string {
	if module == "" || len(symbols) == 0 {
		return content
	}
	ordered := make([]string, 0, len(symbols))
	seen := map[string]struct{}{}
	for _, symbol := range symbols {
		symbol = strings.TrimSpace(symbol)
		if symbol == "" {
			continue
		}
		if _, exists := seen[symbol]; exists {
			continue
		}
		seen[symbol] = struct{}{}
		ordered = append(ordered, symbol)
	}
	if len(ordered) == 0 {
		return content
	}

	content = ensureNamedImport(content, "typing", "TYPE_CHECKING")
	lines := strings.Split(content, "\n")
	insertIdx := topImportInsertIndex(lines)

	blockIdx := -1
	for i := insertIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line == trimmed && trimmed == "if TYPE_CHECKING:" {
			blockIdx = i
			break
		}
		if line == trimmed && trimmed != "" {
			break
		}
	}
	if blockIdx < 0 {
		importLine := fmt.Sprintf("    from %s import %s", module, strings.Join(ordered, ", "))
		prefix := append([]string{}, lines[:insertIdx]...)
		suffix := append([]string{}, lines[insertIdx:]...)
		if len(prefix) > 0 && strings.TrimSpace(prefix[len(prefix)-1]) != "" {
			prefix = append(prefix, "")
		}
		prefix = append(prefix, "if TYPE_CHECKING:", importLine, "")
		return strings.Join(append(prefix, suffix...), "\n")
	}

	blockEnd := blockIdx + 1
	for blockEnd < len(lines) {
		line := lines[blockEnd]
		trimmed := strings.TrimSpace(line)
		if line == trimmed && trimmed != "" {
			break
		}
		blockEnd++
	}

	importPrefix := "from " + module + " import "
	importLineIdx := -1
	for i := blockIdx + 1; i < blockEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, importPrefix) {
			continue
		}
		importLineIdx = i
		break
	}

	if importLineIdx >= 0 {
		trimmed := strings.TrimSpace(lines[importLineIdx])
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		updated := make([]string, 0, len(ordered)+4)
		localSeen := map[string]struct{}{}
		for _, name := range strings.Split(namesPart, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, exists := localSeen[name]; exists {
				continue
			}
			localSeen[name] = struct{}{}
			updated = append(updated, name)
		}
		for _, symbol := range ordered {
			if _, exists := localSeen[symbol]; exists {
				continue
			}
			localSeen[symbol] = struct{}{}
			updated = append(updated, symbol)
		}
		lines[importLineIdx] = "    " + importPrefix + strings.Join(updated, ", ")
		return strings.Join(lines, "\n")
	}

	importLine := "    " + importPrefix + strings.Join(ordered, ", ")
	prefix := append([]string{}, lines[:blockEnd]...)
	suffix := append([]string{}, lines[blockEnd:]...)
	prefix = append(prefix, importLine)
	return strings.Join(append(prefix, suffix...), "\n")
}

func lineHasIdentifier(line string, ident string) bool {
	if ident == "" {
		return false
	}
	for start := 0; start < len(line); {
		idx := strings.Index(line[start:], ident)
		if idx < 0 {
			return false
		}
		idx += start
		beforeOK := idx == 0 || !isIdentifierRune(line[idx-1])
		afterIdx := idx + len(ident)
		afterOK := afterIdx >= len(line) || !isIdentifierRune(line[afterIdx])
		if beforeOK && afterOK {
			return true
		}
		start = idx + len(ident)
	}
	return false
}

func isIdentifierRune(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

func declaresIdentifier(content string, ident string) bool {
	if ident == "" {
		return false
	}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "class "+ident+"(") {
			return true
		}
		if strings.HasPrefix(trimmed, "def "+ident+"(") {
			return true
		}
		if strings.HasPrefix(trimmed, "async def "+ident+"(") {
			return true
		}
		if strings.HasPrefix(trimmed, ident+" =") || strings.HasPrefix(trimmed, ident+":") {
			return true
		}
	}
	return false
}

func ensureNamedImport(content string, module string, symbol string) string {
	if module == "" || symbol == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	insertIdx := topImportInsertIndex(lines)
	importPrefix := "from " + module + " import "
	moduleImportIndex := -1
	for i := 0; i < insertIdx; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line != trimmed {
			continue
		}
		if !strings.HasPrefix(trimmed, importPrefix) {
			continue
		}
		moduleImportIndex = i
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		for _, name := range strings.Split(namesPart, ",") {
			if strings.TrimSpace(name) == symbol {
				return content
			}
		}
	}

	if moduleImportIndex >= 0 {
		trimmed := strings.TrimSpace(lines[moduleImportIndex])
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		names := strings.Split(namesPart, ",")
		seen := map[string]struct{}{}
		updated := make([]string, 0, len(names)+1)
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			updated = append(updated, name)
		}
		if _, exists := seen[symbol]; !exists {
			updated = append(updated, symbol)
		}
		formatted := formatNamedImportLines(module, updated, "")
		if len(formatted) == 1 {
			lines[moduleImportIndex] = formatted[0]
			return strings.Join(lines, "\n")
		}
		prefix := append([]string{}, lines[:moduleImportIndex]...)
		suffix := append([]string{}, lines[moduleImportIndex+1:]...)
		lines = append(prefix, append(formatted, suffix...)...)
		return strings.Join(lines, "\n")
	}

	line := importPrefix + symbol
	prefix := append([]string{}, lines[:insertIdx]...)
	suffix := append([]string{}, lines[insertIdx:]...)
	formatted := formatNamedImportLines(module, []string{symbol}, "")
	if len(formatted) == 0 {
		formatted = []string{line}
	}
	prefix = append(prefix, formatted...)
	return strings.Join(append(prefix, suffix...), "\n")
}

func ensureModuleImport(content string, module string) string {
	if module == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	insertIdx := topImportInsertIndex(lines)
	for i := 0; i < insertIdx; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line != trimmed {
			continue
		}
		if !strings.HasPrefix(trimmed, "import ") {
			continue
		}
		items := strings.Split(strings.TrimSpace(strings.TrimPrefix(trimmed, "import ")), ",")
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == module || strings.HasPrefix(item, module+" as ") {
				return content
			}
		}
	}
	prefix := append([]string{}, lines[:insertIdx]...)
	suffix := append([]string{}, lines[insertIdx:]...)
	prefix = append(prefix, "import "+module)
	return strings.Join(append(prefix, suffix...), "\n")
}

func shouldWrapNamedImport(module string, names []string) bool {
	if module == "cozepy.chat" {
		for _, name := range names {
			if strings.TrimSpace(name) == "_chat_stream_handler" {
				return true
			}
		}
	}
	if module == "cozepy.datasets.documents" && len(names) >= 4 {
		return true
	}
	return false
}

func formatNamedImportLines(module string, names []string, indent string) []string {
	trimmed := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		trimmed = append(trimmed, name)
	}
	if len(trimmed) == 0 {
		return nil
	}
	if !shouldWrapNamedImport(module, trimmed) {
		return []string{fmt.Sprintf("%sfrom %s import %s", indent, module, strings.Join(trimmed, ", "))}
	}
	lines := make([]string, 0, len(trimmed)+2)
	lines = append(lines, fmt.Sprintf("%sfrom %s import (", indent, module))
	for _, name := range trimmed {
		lines = append(lines, fmt.Sprintf("%s    %s,", indent, name))
	}
	lines = append(lines, fmt.Sprintf("%s)", indent))
	return lines
}

func topImportInsertIndex(lines []string) int {
	insertIdx := 0
	seenImport := false
	parenDepth := 0
	for insertIdx < len(lines) {
		trimmed := strings.TrimSpace(lines[insertIdx])
		if !seenImport {
			if trimmed == "" {
				insertIdx++
				continue
			}
			if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
				seenImport = true
				parenDepth += strings.Count(trimmed, "(")
				parenDepth -= strings.Count(trimmed, ")")
				insertIdx++
				continue
			}
			return insertIdx
		}

		if parenDepth > 0 {
			parenDepth += strings.Count(trimmed, "(")
			parenDepth -= strings.Count(trimmed, ")")
			insertIdx++
			continue
		}
		if trimmed == "" {
			insertIdx++
			continue
		}
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			parenDepth += strings.Count(trimmed, "(")
			parenDepth -= strings.Count(trimmed, ")")
			insertIdx++
			continue
		}
		return insertIdx
	}
	return insertIdx
}

func normalizeTypingOptionalImport(content string) string {
	lines := strings.Split(content, "\n")
	usesOptional := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from typing import ") {
			continue
		}
		if strings.Contains(line, "Optional[") {
			usesOptional = true
			break
		}
	}

	isTypingNameUsed := func(name string) bool {
		switch name {
		case "Optional":
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "from typing import ") {
					continue
				}
				if strings.Contains(line, "Optional[") {
					return true
				}
			}
			return false
		case "List":
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "from typing import ") {
					continue
				}
				if strings.Contains(line, "List[") {
					return true
				}
			}
			return false
		case "Dict":
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "from typing import ") {
					continue
				}
				if strings.Contains(line, "Dict[") {
					return true
				}
			}
			return false
		case "TYPE_CHECKING":
			return containsIdentifierOutsideImportSection(content, "TYPE_CHECKING")
		default:
			return containsIdentifierOutsideImportSection(content, name)
		}
	}

	typingImportIndexes := make([]int, 0)
	optionalImported := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "from typing import ") {
			continue
		}
		typingImportIndexes = append(typingImportIndexes, i)
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, "from typing import "))
		if namesPart == "" {
			continue
		}
		for _, name := range strings.Split(namesPart, ",") {
			if strings.TrimSpace(name) == "Optional" {
				optionalImported = true
				break
			}
		}
		if optionalImported {
			break
		}
	}

	normalizeTypingLine := func(line string, includeOptional bool) (string, bool) {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "from typing import ") {
			return line, true
		}
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, "from typing import "))
		names := strings.Split(namesPart, ",")
		seen := map[string]struct{}{}
		kept := make([]string, 0, len(names)+1)
		for _, name := range names {
			typedName := strings.TrimSpace(name)
			if typedName == "" {
				continue
			}
			if typedName == "Optional" && !includeOptional {
				continue
			}
			if typedName != "Optional" && !isTypingNameUsed(typedName) {
				continue
			}
			if _, exists := seen[typedName]; exists {
				continue
			}
			seen[typedName] = struct{}{}
			kept = append(kept, typedName)
		}
		if includeOptional {
			if _, exists := seen["Optional"]; !exists {
				kept = append(kept, "Optional")
			}
		}
		if len(kept) == 0 {
			return "", false
		}
		return "from typing import " + strings.Join(kept, ", "), true
	}

	includeOptional := usesOptional
	pruned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "from typing import ") {
			pruned = append(pruned, line)
			continue
		}
		updated, keep := normalizeTypingLine(line, includeOptional)
		if !keep {
			continue
		}
		pruned = append(pruned, updated)
	}
	if includeOptional && !optionalImported {
		if len(typingImportIndexes) > 0 {
			idx := typingImportIndexes[0]
			lines = pruned
			if idx < len(lines) {
				updated, keep := normalizeTypingLine(lines[idx], true)
				if keep {
					lines[idx] = updated
					return strings.Join(lines, "\n")
				}
			}
		}
		insertIdx := 0
		for insertIdx < len(pruned) {
			trimmed := strings.TrimSpace(pruned[insertIdx])
			if trimmed == "" {
				insertIdx++
				continue
			}
			if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
				insertIdx++
				continue
			}
			break
		}
		prefix := append([]string{}, pruned[:insertIdx]...)
		suffix := append([]string{}, pruned[insertIdx:]...)
		prefix = append(prefix, "from typing import Optional")
		if insertIdx < len(pruned) && strings.TrimSpace(pruned[insertIdx]) != "" {
			prefix = append(prefix, "")
		}
		return strings.Join(append(prefix, suffix...), "\n")
	}
	return strings.Join(pruned, "\n")
}
