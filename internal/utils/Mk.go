package utils

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// html 占位符前缀（HTML 注释形式，一般不会被其它规则改写）
func htmlPlaceholder(i int) string {
	return "<!--mk-p-" + strconv.Itoa(i) + "-->"
}

var voidHTMLTags = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isTagNameChar(b byte) bool {
	return isASCIILetter(b) || (b >= '0' && b <= '9') || b == '-' || b == ':'
}

// parseOpenTag 从 s[i]=='<' 处解析起始标签，返回小写标签名、是否 void、标签结束位置（不含）、是否成功。
func parseOpenTag(s string, i int) (name string, void bool, end int, ok bool) {
	if i >= len(s) || s[i] != '<' {
		return "", false, i, false
	}
	j := i + 1
	if j >= len(s) {
		return "", false, i, false
	}
	if s[j] == '/' || s[j] == '!' || s[j] == '?' {
		return "", false, i, false
	}
	if !isASCIILetter(s[j]) {
		return "", false, i, false
	}
	nameStart := j
	for j < len(s) && isTagNameChar(s[j]) {
		j++
	}
	name = strings.ToLower(s[nameStart:j])
	var quote byte
	selfClose := false
	for j < len(s) {
		c := s[j]
		if quote == 0 && (c == '"' || c == '\'') {
			quote = c
			j++
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			j++
			continue
		}
		if c == '/' && j+1 < len(s) && s[j+1] == '>' {
			selfClose = true
			j += 2
			end = j
			void = selfClose || voidHTMLTags[name]
			return name, void, end, true
		}
		if c == '>' {
			j++
			end = j
			void = voidHTMLTags[name]
			return name, void, end, true
		}
		j++
	}
	return "", false, i, false
}

// parseCloseTag 从 s[i]=='<' 且以 '</' 开头处解析闭合标签。
func parseCloseTag(s string, i int) (name string, end int, ok bool) {
	if i+2 >= len(s) || s[i] != '<' || s[i+1] != '/' {
		return "", i, false
	}
	j := i + 2
	for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
		j++
	}
	nameStart := j
	for j < len(s) && isTagNameChar(s[j]) {
		j++
	}
	if nameStart == j {
		return "", i, false
	}
	name = strings.ToLower(s[nameStart:j])
	var quote byte
	for j < len(s) {
		c := s[j]
		if quote == 0 && (c == '"' || c == '\'') {
			quote = c
			j++
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			j++
			continue
		}
		if c == '>' {
			j++
			return name, j, true
		}
		j++
	}
	return "", i, false
}

func skipHTMLComment(s string, i int) (end int, ok bool) {
	if !strings.HasPrefix(s[i:], "<!--") {
		return i, false
	}
	j := i + 4
	if idx := strings.Index(s[j:], "-->"); idx >= 0 {
		return j + idx + 3, true
	}
	return len(s), true
}

// extractBalancedHTMLFragment 从 s[start]（须为 '<'）起截取一个完整元素（含子节点），用于行内/块级原样保留。
func extractBalancedHTMLFragment(s string, start int) (frag string, end int, ok bool) {
	if start >= len(s) || s[start] != '<' {
		return "", start, false
	}
	if strings.HasPrefix(s[start:], "<!--") {
		e, okc := skipHTMLComment(s, start)
		if !okc {
			return "", start, false
		}
		return s[start:e], e, true
	}
	// <!DOCTYPE ...>、<?xml ...?> 等声明，引号内允许 '>'
	if start+1 < len(s) && (s[start+1] == '!' || s[start+1] == '?') {
		j := start + 2
		var quote byte
		for j < len(s) {
			c := s[j]
			if quote == 0 && (c == '"' || c == '\'') {
				quote = c
				j++
				continue
			}
			if quote != 0 {
				if c == quote {
					quote = 0
				}
				j++
				continue
			}
			if c == '>' {
				j++
				return s[start:j], j, true
			}
			j++
		}
		return "", start, false
	}
	name, void, pos, ok := parseOpenTag(s, start)
	if !ok {
		return "", start, false
	}
	if void {
		return s[start:pos], pos, true
	}
	var stack []string
	stack = append(stack, name)
	cur := pos
	for len(stack) > 0 && cur < len(s) {
		idx := strings.IndexByte(s[cur:], '<')
		if idx < 0 {
			return "", start, false
		}
		cur += idx
		if strings.HasPrefix(s[cur:], "<!--") {
			ne, _ := skipHTMLComment(s, cur)
			cur = ne
			continue
		}
		if cur+1 < len(s) && s[cur+1] == '/' {
			nm, e, okc := parseCloseTag(s, cur)
			if !okc {
				cur++
				continue
			}
			if len(stack) == 0 || stack[len(stack)-1] != nm {
				return "", start, false
			}
			stack = stack[:len(stack)-1]
			cur = e
			if len(stack) == 0 {
				return s[start:cur], cur, true
			}
			continue
		}
		nm2, void2, e2, ok2 := parseOpenTag(s, cur)
		if !ok2 {
			cur++
			continue
		}
		if !void2 {
			stack = append(stack, nm2)
		}
		cur = e2
	}
	return "", start, false
}

// protectRawHTML 将可解析的 HTML 片段（含块级、行内、注释、声明）替换为占位符，避免被 Markdown 规则破坏。
func protectRawHTML(s string, blocks *[]string, nextID *int) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if strings.HasPrefix(s[i:], "<!--mk-p-") {
			j := strings.Index(s[i:], "-->")
			if j < 0 {
				out.WriteString(s[i:])
				break
			}
			out.WriteString(s[i : i+j+3])
			i += j + 3
			continue
		}
		if s[i] != '<' {
			out.WriteByte(s[i])
			i++
			continue
		}
		frag, end, ok := extractBalancedHTMLFragment(s, i)
		if ok {
			*blocks = append(*blocks, frag)
			out.WriteString(htmlPlaceholder(*nextID))
			*nextID++
			i = end
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

func unprotectHTMLPlaceholders(s string, blocks []string) string {
	for i := len(blocks) - 1; i >= 0; i-- {
		s = strings.ReplaceAll(s, htmlPlaceholder(i), blocks[i])
	}
	return s
}

var cellRawHTMLRe = regexp.MustCompile(`<[a-zA-Z][\w:-]*`)

// cellMayContainRawHTML 表格单元格是否可能含 HTML（避免把已有标签转义成文本）。
func cellMayContainRawHTML(c string) bool {
	return cellRawHTMLRe.MatchString(c)
}

func tableCellInner(c string) string {
	if cellMayContainRawHTML(c) {
		return c
	}
	return html.EscapeString(c)
}

// TitletoHtml 将 ATX 标题（# …）转为 <h1>…</h1>
func TitletoHtml(s string) string {
	r := regexp.MustCompile(`(?m)^(#{1,6})\s+([^\r\n]+)`)
	sm := r.FindAllStringSubmatch(s, -1)
	index := 0
	return r.ReplaceAllStringFunc(s, func(_ string) string {
		n := len(sm[index][1])
		ret := fmt.Sprintf("<h%d>%s</h%d>", n, sm[index][2], n)
		index++
		return ret
	})
}

// HrtoHtml 将仅含 --- 的行转为 <hr>
func HrtoHtml(s string) string {
	// 允许行尾 \r，兼容 CRLF 源文件
	r := regexp.MustCompile(`(?m)^[ \t]*-{3,}[ \t]*\r?$`)
	return r.ReplaceAllString(s, "<hr>")
}

// FencedCodetoHtml 围栏代码 ``` … ``` → <pre><code>
func FencedCodetoHtml(s string) string {
	re := regexp.MustCompile("(?ms)^```([^`\\r\\n]*)\\r?\\n(.*?)\\r?\\n```")
	return re.ReplaceAllStringFunc(s, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if sub == nil {
			return m
		}
		lang := strings.TrimSpace(sub[1])
		body := sub[2]
		esc := html.EscapeString(body)
		if lang != "" {
			return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>", html.EscapeString(lang), esc)
		}
		return "<pre><code>" + esc + "</code></pre>"
	})
}

// isTableSeparatorRow 判断 GFM 表格分隔行 | --- | :--- |
func isTableSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		t := strings.TrimSpace(c)
		if t == "" {
			return false
		}
		ok := true
		for _, ch := range t {
			if ch != '-' && ch != ':' && ch != ' ' && ch != '\t' {
				ok = false
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") {
		return nil
	}
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(strings.TrimSpace(line), "|")
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// TabletoHtml 将连续的 | 表格行转为 <table>
func TabletoHtml(s string) string {
	lines := strings.Split(s, "\n")
	var out strings.Builder
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "|") {
			out.WriteString(line)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}
		var block []string
		for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
			block = append(block, strings.TrimSpace(lines[i]))
			i++
		}
		i--

		var rows [][]string
		for _, row := range block {
			if cells := splitTableRow(row); cells != nil {
				rows = append(rows, cells)
			}
		}
		if len(rows) == 0 {
			out.WriteString(line)
			if i < len(lines)-1 {
				out.WriteByte('\n')
			}
			continue
		}

		sepAt := -1
		if len(rows) >= 2 && isTableSeparatorRow(rows[1]) {
			sepAt = 1
		}

		out.WriteString("<table>\n")
		bodyStart := 0
		if sepAt == 1 {
			out.WriteString("<thead>\n<tr>")
			for _, c := range rows[0] {
				out.WriteString("<th>" + tableCellInner(c) + "</th>")
			}
			out.WriteString("</tr>\n</thead>\n<tbody>\n")
			bodyStart = 2
		}
		for r := bodyStart; r < len(rows); r++ {
			if sepAt == 1 && r == 1 {
				continue
			}
			out.WriteString("<tr>")
			for _, c := range rows[r] {
				tag := "td"
				if sepAt != 1 && r == 0 {
					tag = "th"
				}
				out.WriteString(fmt.Sprintf("<%s>%s</%s>", tag, tableCellInner(c), tag))
			}
			out.WriteString("</tr>\n")
		}
		if sepAt == 1 {
			out.WriteString("</tbody>\n")
		}
		out.WriteString("</table>")
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// TripleEmphasistoHtml 整行 ***文本***
func TripleEmphasistoHtml(s string) string {
	r := regexp.MustCompile(`(?m)^\*\*\*([^*\r\n]+)\*\*\*$`)
	return r.ReplaceAllString(s, "<strong><em>$1</em></strong>")
}

// StrongtoHtml **文本**
func StrongtoHtml(s string) string {
	r := regexp.MustCompile(`\*\*([^*\r\n]+)\*\*`)
	return r.ReplaceAllString(s, "<strong>$1</strong>")
}

// ItalictoHtml *文本*（不含内部 *）
func ItalictoHtml(s string) string {
	r := regexp.MustCompile(`\*([^*\r\n]+)\*`)
	return r.ReplaceAllString(s, "<em>$1</em>")
}

// DeltoHtml ~~文本~~
func DeltoHtml(s string) string {
	r := regexp.MustCompile(`~~([^~\r\n]+)~~`)
	return r.ReplaceAllString(s, "<del>$1</del>")
}

// LinktoHtml 内联链接 [文本](url) 与 [文本](url "标题")；URL 中勿含空格或未转义 )
func LinktoHtml(s string) string {
	// 方括号内、URL 允许为空：![](x.png)、![]()、[描述]()
	re := regexp.MustCompile(`(!*)\[([^\]\r\n]*)\]\(([^\s)]*)(?:\s+"([^"]*)")?\)`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if sub == nil {
			return m
		}
		image := strings.TrimSpace(sub[1]) // 图片
		text := strings.TrimSpace(sub[2])  // 文本
		href := strings.TrimSpace(sub[3])  // 链接
		title := strings.TrimSpace(sub[4]) // 标题
		if image == "!" {
			if title != "" {
				return fmt.Sprintf(`<img src="%s" alt="%s" title="%s">`, html.EscapeString(href), html.EscapeString(text), html.EscapeString(title))
			}
			return fmt.Sprintf(`<img src="%s" alt="%s">`, html.EscapeString(href), html.EscapeString(text))
		}
		if title != "" {
			return fmt.Sprintf(`<a href="%s" title="%s">%s</a>`, html.EscapeString(href), html.EscapeString(title), html.EscapeString(text))
		}
		return fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(href), html.EscapeString(text))
	})
}

// CodetoHtml 行内 `代码`（忽略被反斜线转义的 \`）
func CodetoHtml(s string) string {
	var out strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch != '`' {
			out.WriteRune(ch)
			continue
		}
		// 忽略转义反引号
		if i > 0 && runes[i-1] == '\\' {
			out.WriteRune(ch)
			continue
		}
		// 查找配对的未转义反引号
		end := -1
		for j := i + 1; j < len(runes); j++ {
			if runes[j] == '`' && runes[j-1] != '\\' {
				end = j
				break
			}
			if runes[j] == '\n' || runes[j] == '\r' {
				break
			}
		}
		if end == -1 {
			out.WriteRune(ch)
			continue
		}
		code := string(runes[i+1 : end])
		out.WriteString("<code>")
		out.WriteString(html.EscapeString(code))
		out.WriteString("</code>")
		i = end
	}
	return out.String()
}

// ListtoHtml 无序列表：行首 `- ` → <li>
func ListtoHtml(s string) string {
	r := regexp.MustCompile(`(?m)^[ \t]*-\s+([^\r\n]+)`)
	return r.ReplaceAllString(s, "<li>$1</li>")
}

// WrapListBlocks 将连续多行 <li> 包进 <ul>
func WrapListBlocks(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for j := 0; j < len(lines); j++ {
		trim := strings.TrimSpace(lines[j])
		if strings.HasPrefix(trim, "<li>") {
			out = append(out, "<ul>")
			for j < len(lines) {
				t := strings.TrimSpace(lines[j])
				if !strings.HasPrefix(t, "<li>") {
					break
				}
				out = append(out, t)
				j++
			}
			j--
			out = append(out, "</ul>")
		} else {
			out = append(out, lines[j])
		}
	}
	return strings.Join(out, "\n")
}

func isBlockHTMLLine(trim string) bool {
	if trim == "" {
		return true
	}
	blockPrefixes := []string{
		"<h1>", "<h2>", "<h3>", "<h4>", "<h5>", "<h6>",
		"<hr", "<table", "</table>", "<thead>", "</thead>", "<tbody>", "</tbody>",
		"<tr>", "</tr>", "<ul>", "</ul>", "<li>", "<pre>", "</pre>",
		"<strong><em>", "<img ", "<p>", "</p>",
		"<!--mk-p-",
	}
	for _, p := range blockPrefixes {
		if strings.HasPrefix(trim, p) {
			return true
		}
	}
	return false
}

// ParagraphToHtml 将普通文本段落包裹为 <p>，提升页面可读性。
func ParagraphToHtml(s string) string {
	lines := strings.Split(s, "\n") // 将文本按行分割
	var out []string                // 输出结果
	var para []string               // 当前段落

	flush := func() {
		if len(para) == 0 { // 如果当前段落为空，则返回
			return
		}
		text := strings.Join(para, " ")      // 将当前段落按空格分割
		out = append(out, "<p>"+text+"</p>") // 将当前段落包裹为 <p>
		para = para[:0]                      // 清空当前段落
	}

	for _, line := range lines { // 遍历每一行
		trim := strings.TrimSpace(line) // 去除行首尾空格
		if trim == "" {
			flush()               // 如果当前行为空，则刷新段落
			out = append(out, "") // 添加空行
			continue
		}
		if isBlockHTMLLine(trim) {
			flush()                 // 如果当前行为块级 HTML 标签，则刷新段落
			out = append(out, line) // 添加当前行
			continue
		}
		para = append(para, trim) // 将当前行添加到当前段落
	}
	flush()                        // 刷新段落
	return strings.Join(out, "\n") // 将输出结果按行拼接
}

// WrapStyledHtml 包装为可直接在浏览器展示的完整 HTML 页面。
func WrapStyledHtml(body string) string {
	const css = `<style>
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,"Noto Sans",sans-serif;line-height:1.75;color:#1f2328;background:#f6f8fa;margin:0}
.mk-container{max-width:900px;margin:32px auto;padding:28px;background:#fff;border:1px solid #d0d7de;border-radius:12px;box-shadow:0 1px 2px rgba(31,35,40,.08)}
h1,h2,h3,h4,h5,h6{line-height:1.3;color:#24292f}
hr{border:none;border-top:1px solid #d8dee4;margin:20px 0}
p{margin:12px 0}
code{background:#f6f8fa;padding:2px 6px;border-radius:6px;border:1px solid #d8dee4}
pre{background:#0d1117;color:#c9d1d9;padding:14px;border-radius:8px;overflow:auto}
pre code{background:transparent;border:none;padding:0;color:inherit}
table{width:100%;border-collapse:collapse;margin:14px 0}
th,td{border:1px solid #d0d7de;padding:8px 10px;text-align:left}
thead{background:#f6f8fa}
ul{padding-left:24px}
a{color:#0969da;text-decoration:none}
a:hover{text-decoration:underline}
img{max-width:100%;height:auto;border-radius:8px}
div,section{margin:12px 0}
details{margin:12px 0;padding:8px 12px;border:1px solid #d0d7de;border-radius:8px;background:#f6f8fa}
summary{cursor:pointer;font-weight:600}
kbd{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;border:1px solid #d0d7de;border-bottom-width:2px;border-radius:4px;padding:2px 6px;background:#fff}
blockquote{margin:12px 0;padding:8px 16px;border-left:4px solid #0969da;background:#f6f8fa;color:#57606a}
mark{background:#fff8c5;padding:0 2px}
</style>`
	return "<!doctype html><html><head><meta charset=\"utf-8\">" + css + "</head><body><main class=\"mk-container\">" + body + "</main></body></html>"
}

func MarkdownToHtml(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = FencedCodetoHtml(s)
	var htmlBlocks []string
	n := 0
	s = protectRawHTML(s, &htmlBlocks, &n)
	s = TitletoHtml(s)
	s = HrtoHtml(s)
	s = TabletoHtml(s)
	s = ListtoHtml(s)
	s = WrapListBlocks(s)
	s = TripleEmphasistoHtml(s)
	s = LinktoHtml(s)
	s = CodetoHtml(s)
	s = StrongtoHtml(s)
	s = ItalictoHtml(s)
	s = DeltoHtml(s)
	s = ParagraphToHtml(s)
	s = unprotectHTMLPlaceholders(s, htmlBlocks)
	return s
}
