package python

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

func AppendIndentedCode(buf *bytes.Buffer, code string, indentLevel int) {
	block := strings.Trim(code, "\n")
	if strings.TrimSpace(block) == "" {
		return
	}
	lines := strings.Split(block, "\n")
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := 0
		for indent < len(line) && line[indent] == ' ' {
			indent++
		}
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent < 0 {
		return
	}
	prefix := strings.Repeat("    ", indentLevel)
	for _, line := range lines {
		cleanLine := strings.TrimRight(line, "\r")
		if minIndent > 0 && len(cleanLine) >= minIndent {
			cleanLine = cleanLine[minIndent:]
		}
		if strings.TrimSpace(cleanLine) == "" {
			buf.WriteString("\n")
			continue
		}
		buf.WriteString(prefix)
		buf.WriteString(cleanLine)
		buf.WriteString("\n")
	}
}

func EnsureTrailingNewlines(buf *bytes.Buffer, newlineCount int) {
	if newlineCount < 0 {
		newlineCount = 0
	}
	content := strings.TrimRight(buf.String(), "\n")
	buf.Reset()
	buf.WriteString(content)
	if newlineCount > 0 {
		buf.WriteString(strings.Repeat("\n", newlineCount))
	}
}

func LinesFromCommentOverride(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			line = strings.TrimPrefix(line, "#")
		}
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func CodeBlocksHaveLeadingDocstring(blocks []string) bool {
	for _, raw := range blocks {
		block := strings.TrimSpace(raw)
		if block == "" {
			continue
		}
		return strings.HasPrefix(block, "\"\"\"") || strings.HasPrefix(block, "'''")
	}
	return false
}

func WriteLineComments(buf *bytes.Buffer, indentLevel int, lines []string) {
	if len(lines) == 0 {
		return
	}
	indent := strings.Repeat("    ", indentLevel)
	for _, line := range lines {
		buf.WriteString(indent)
		buf.WriteString("# ")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
}

func NormalizedDocstringLines(docstring string) []string {
	text := strings.ReplaceAll(docstring, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	out := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		out = append(out, strings.TrimRight(line, "\r"))
	}
	return out
}

func WriteClassDocstring(buf *bytes.Buffer, indentLevel int, docstring string, style string) {
	lines := NormalizedDocstringLines(docstring)
	if len(lines) == 0 {
		return
	}
	indent := strings.Repeat("    ", indentLevel)
	if style == "inline" && len(lines) == 1 {
		buf.WriteString(indent)
		buf.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", lines[0]))
		buf.WriteString("\n")
		return
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			buf.WriteString("\n")
			continue
		}
		buf.WriteString(indent)
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"\n")
	buf.WriteString("\n")
}

func WriteMethodDocstring(buf *bytes.Buffer, indentLevel int, docstring string, style string) {
	lines := NormalizedDocstringLines(docstring)
	if len(lines) == 0 {
		return
	}
	indent := strings.Repeat("    ", indentLevel)
	if style == "block" {
		buf.WriteString(indent)
		buf.WriteString("\"\"\"\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				buf.WriteString("\n")
				continue
			}
			buf.WriteString(indent)
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		buf.WriteString(indent)
		buf.WriteString("\"\"\"\n")
		return
	}
	if len(lines) == 1 {
		buf.WriteString(indent)
		buf.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", EscapeDocstring(lines[0])))
		return
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"")
	buf.WriteString(lines[0])
	buf.WriteString("\n")
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			buf.WriteString("\n")
			continue
		}
		buf.WriteString(indent)
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"\n")
}

func RenderEnumValueLiteral(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "True"
		}
		return "False"
	default:
		return fmt.Sprintf("%q", fmt.Sprintf("%v", value))
	}
}

func EnumMemberName(value string) string {
	name := strings.TrimSpace(value)
	if name == "" {
		return "UNKNOWN"
	}
	name = strings.ToUpper(CollapseUnderscore(ToSnake(name)))
	name = strings.Trim(name, "_")
	if name == "" {
		name = "UNKNOWN"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "VALUE_" + name
	}
	return name
}

func DetectMethodBlockName(block string) string {
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if name, ok := ParseDefName(trimmed); ok {
			return name
		}
	}
	return ""
}

func ParseDefName(trimmedLine string) (string, bool) {
	defLine := strings.TrimSpace(trimmedLine)
	if strings.HasPrefix(defLine, "async def ") {
		defLine = strings.TrimSpace(strings.TrimPrefix(defLine, "async def "))
	} else if strings.HasPrefix(defLine, "def ") {
		defLine = strings.TrimSpace(strings.TrimPrefix(defLine, "def "))
	} else {
		return "", false
	}
	name := strings.TrimSpace(strings.SplitN(defLine, "(", 2)[0])
	if name == "" {
		return "", false
	}
	return name, true
}

func IsDocstringLine(trimmedLine string) bool {
	return strings.HasPrefix(trimmedLine, "\"\"\"") || strings.HasPrefix(trimmedLine, "'''")
}

func RenderMethodDocstringLines(docstring string, style string, indent string) []string {
	lines := NormalizedDocstringLines(docstring)
	if len(lines) == 0 {
		return nil
	}
	out := make([]string, 0, len(lines)+3)
	if style == "block" {
		out = append(out, indent+"\"\"\"")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				out = append(out, "")
				continue
			}
			out = append(out, indent+line)
		}
		out = append(out, indent+"\"\"\"")
		return out
	}
	if len(lines) == 1 {
		out = append(out, indent+fmt.Sprintf("\"\"\"%s\"\"\"", EscapeDocstring(lines[0])))
		return out
	}
	out = append(out, indent+"\"\"\""+lines[0])
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}
		out = append(out, indent+line)
	}
	out = append(out, indent+"\"\"\"")
	return out
}
