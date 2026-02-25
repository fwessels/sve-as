package preprocessor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ---------------- Preprocessor ----------------

type Preprocessor struct {
	IncludeDirs       []string
	KeepLineComments  bool
	EntryFile         string
	obj               map[string]string
	fn                map[string]FnMacro
	includeStackGuard map[string]bool
}

type FnMacro struct {
	Params []string
	Body   string
}

func NewPreprocessor() *Preprocessor {
	return &Preprocessor{
		obj:               map[string]string{},
		fn:                map[string]FnMacro{},
		includeStackGuard: map[string]bool{},
	}
}

func (p *Preprocessor) DefineObject(name, value string) {
	p.obj[name] = value
}

func (p *Preprocessor) DefineFunc(name string, params []string, body string) {
	p.fn[name] = FnMacro{Params: params, Body: body}
}

// Process preprocesses file content and writes expanded output.
func (p *Preprocessor) Process(filename string, r io.Reader, w io.Writer) error {
	abs, err := p.resolveAsFile(filename, "")
	if err == nil {
		filename = abs
	}

	if p.includeStackGuard[filename] {
		// Prevent include cycles from exploding; you can change this behavior if you want
		return fmt.Errorf("include cycle detected at %q", filename)
	}
	p.includeStackGuard[filename] = true
	defer delete(p.includeStackGuard, filename)

	var out bytes.Buffer
	lr := newLineReader(r)

	cond := newCondStack()

	lineNo := 0
	var pendingLine string
	pendingLineNo := 0
	hasPending := false
	for {
		var line string
		if hasPending {
			line = pendingLine
			lineNo = pendingLineNo
			hasPending = false
		} else {
			var ok bool
			var err error
			line, _, ok, err = lr.next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if !ok {
				break
			}
			lineNo++
		}

		trim := strings.TrimSpace(line)
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			first := firstNonSpaceIndex(line)
			if first >= 0 && idx > first {
				after := strings.TrimSpace(line[idx:])
				if isDirectivePrefix(after) {
					return fmt.Errorf("%s:%d: '#' must be first item on line", shortPath(filename), lineNo)
				}
			}
		}
		if strings.HasPrefix(trim, "#") {
			startLineNo := lineNo
			fullLine, endLineNo, nextLine, nextLineNo, ok, endedAtEOF, err := readDirectiveLine(line, lineNo, lr)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			lineNo = endLineNo
			if nextLine != "" || nextLineNo != 0 {
				pendingLine = nextLine
				pendingLineNo = nextLineNo
				hasPending = true
			}
			fullTrim := strings.TrimSpace(fullLine)
			fields := splitDirective(fullTrim)
			if fields.cmd == "define" && endedAtEOF && !lr.lastHasNL {
				_, _, body, ok := parseDefineDirective(fields.arg)
				if ok && strings.TrimSpace(body) != "" {
					return fmt.Errorf("%s:%d: no newline after macro definition", shortPath(filename), startLineNo)
				}
			}
			if err := p.handleDirective(&out, filename, startLineNo, fullTrim, cond); err != nil {
				if err.Error() == "redefinition of macro" {
					return fmt.Errorf("%s:%d: %s", shortPath(filename), startLineNo, err.Error())
				}
				return err
			}
			if preserveDirectiveLines(fields.cmd) {
				for i := 0; i < lineNo-startLineNo+1; i++ {
					out.WriteByte('\n')
				}
			}
			continue
		}

		if !cond.Active() {
			continue
		}

		expanded, err := p.expandLineForProcess(line)
		if err != nil {
			if err.Error() == "recursive macro invocation" {
				return fmt.Errorf("%s:%d: %s", shortPath(filename), lineNo, err.Error())
			}
			return err
		}
		if p.KeepLineComments {
			// comment marker is safe for Go asm; adjust if you prefer
			out.WriteString(fmt.Sprintf("// %s:%d\n", shortPath(filename), lineNo))
		}
		out.WriteString(expanded)
		if !strings.HasSuffix(expanded, "\n") {
			out.WriteByte('\n')
		}
	}

	if cond.Depth() != 0 {
		if line := cond.UnclosedLine(); line > 0 {
			return fmt.Errorf("%s:%d: unclosed #ifdef or #ifndef", shortPath(filename), line)
		}
		return fmt.Errorf("%s: unclosed #ifdef or #ifndef", shortPath(filename))
	}
	_, err = w.Write(out.Bytes())
	return err
}

func (p *Preprocessor) expandLineForProcess(line string) (string, error) {
	if core, ok := p.soleMacroInvocation(line); ok {
		expanded, err := p.expandLine(core)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(expanded, "\n") {
			expanded = strings.TrimPrefix(expanded, "\n")
		}
		return expanded, nil
	}
	return p.expandLine(line)
}

func (p *Preprocessor) soleMacroInvocation(line string) (string, bool) {
	trim := strings.TrimSpace(stripLineComment(line))
	if trim == "" {
		return "", false
	}
	name, rest, ok := splitIdentPrefix(trim)
	if !ok {
		return "", false
	}
	if _, ok := p.fn[name]; ok {
		if strings.HasPrefix(rest, "(") {
			end, ok := scanParenEnd(rest)
			if !ok {
				return "", false
			}
			if strings.TrimSpace(rest[end:]) == "" {
				return trim, true
			}
		}
		return "", false
	}
	if _, ok := p.obj[name]; ok {
		if strings.TrimSpace(rest) == "" {
			return trim, true
		}
	}
	return "", false
}

func splitIdentPrefix(s string) (name string, rest string, ok bool) {
	if s == "" || !isIdentStart(s[0]) {
		return "", "", false
	}
	i := 1
	for i < len(s) && isIdentPart(s[i]) {
		i++
	}
	return s[:i], s[i:], true
}

func scanParenEnd(s string) (int, bool) {
	if s == "" || s[0] != '(' {
		return 0, false
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				return i + 1, true
			}
		} else if ch == '"' || ch == '\'' {
			quote := ch
			i++
			for i < len(s) {
				ch = s[i]
				if ch == '\\' {
					i++
					if i < len(s) {
						i++
					}
					continue
				}
				if ch == quote {
					break
				}
				i++
			}
		}
	}
	return 0, false
}

func stripLineComment(s string) string {
	if i := strings.Index(s, "//"); i >= 0 {
		return s[:i]
	}
	return s
}

type lineReader struct {
	r         *bufio.Reader
	lastHasNL bool
}

func newLineReader(r io.Reader) *lineReader {
	return &lineReader{r: bufio.NewReader(r)}
}

func (lr *lineReader) next() (line string, hasNL bool, ok bool, err error) {
	s, err := lr.r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", false, false, err
	}
	if len(s) == 0 && err == io.EOF {
		return "", false, false, io.EOF
	}
	hasNL = strings.HasSuffix(s, "\n")
	if hasNL {
		s = s[:len(s)-1]
	}
	lr.lastHasNL = hasNL
	return s, hasNL, true, nil
}

func readDirectiveLine(firstLine string, firstLineNo int, lr *lineReader) (fullLine string, endLineNo int, nextLine string, nextLineNo int, ok bool, endedAtEOF bool, err error) {
	line := firstLine
	lineNo := firstLineNo
	var b strings.Builder
	for {
		if !lineContinues(line) {
			b.WriteString(line)
			return b.String(), lineNo, "", 0, true, false, nil
		}
		b.WriteString(stripLineContinuation(line))
		next, _, ok, err := lr.next()
		if err != nil {
			if err == io.EOF {
				return b.String(), lineNo, "", 0, true, true, nil
			}
			return "", 0, "", 0, false, false, err
		}
		if !ok {
			return b.String(), lineNo, "", 0, true, true, nil
		}
		lineNo++
		if !isContinuationLine(next) {
			return b.String(), lineNo-1, next, lineNo, true, false, nil
		}
		b.WriteByte('\n')
		line = next
	}
}

func lineContinues(s string) bool {
	i := strings.LastIndexFunc(s, func(r rune) bool {
		return r != ' ' && r != '\t'
	})
	return i >= 0 && s[i] == '\\'
}

func stripLineContinuation(s string) string {
	i := strings.LastIndexFunc(s, func(r rune) bool {
		return r != ' ' && r != '\t'
	})
	if i >= 0 && s[i] == '\\' {
		return strings.TrimRight(s[:i], " \t")
	}
	return s
}

func isContinuationLine(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == ' ' || s[0] == '\t'
}

func preserveDirectiveLines(cmd string) bool {
	switch cmd {
	case "include":
		return false
	default:
		return true
	}
}

func (p *Preprocessor) handleDirective(out *bytes.Buffer, filename string, lineNo int, trim string, cond *condStack) error {
	// Allow directives even when inactive, but only those that control conditionals
	fields := splitDirective(trim)

	switch fields.cmd {
	case "include":
		if !cond.Active() {
			return nil
		}
		path, ok := parseIncludeArg(fields.arg)
		if !ok {
			return fmt.Errorf("%s:%d: bad #include syntax: %q", shortPath(filename), lineNo, trim)
		}
		bs, resolved, err := p.readInclude(path, filename)
		if err != nil {
			return fmt.Errorf("%s:%d: include %q: %w", shortPath(filename), lineNo, path, err)
		}
		if err := p.Process(resolved, bytes.NewReader(bs), out); err != nil {
			return err
		}
		return nil

	case "define":
		if !cond.Active() {
			return nil
		}
		name, params, body, ok := parseDefineDirective(fields.arg)
		if !ok {
			return fmt.Errorf("%s:%d: bad #define: %q", shortPath(filename), lineNo, trim)
		}
		if params == nil {
			if p.isDefined(name) {
				return fmt.Errorf("redefinition of macro")
			}
			p.DefineObject(name, body)
		} else {
			if p.isDefined(name) {
				return fmt.Errorf("redefinition of macro")
			}
			p.DefineFunc(name, params, body)
		}
		return nil

	case "undef":
		if !cond.Active() {
			return nil
		}
		name := strings.TrimSpace(fields.arg)
		delete(p.obj, name)
		delete(p.fn, name)
		return nil

	case "ifdef":
		name := strings.TrimSpace(fields.arg)
		cond.Push(p.isDefined(name), lineNo)
		return nil

	case "ifndef":
		name := strings.TrimSpace(fields.arg)
		cond.Push(!p.isDefined(name), lineNo)
		return nil

	case "if":
		// Minimal: treat non-zero / defined symbol as true
		expr := strings.TrimSpace(fields.arg)
		cond.Push(p.evalIfExpr(expr), lineNo)
		return nil

	case "elif":
		expr := strings.TrimSpace(fields.arg)
		cond.Elif(p.evalIfExpr(expr))
		return nil

	case "else":
		cond.Else()
		return nil

	case "endif":
		cond.Pop()
		return nil

	default:
		// Unknown directives are ignored when inactive, error when active to catch typos
		if !cond.Active() {
			return nil
		}
		return fmt.Errorf("%s:%d: unknown directive %q", shortPath(filename), lineNo, fields.cmd)
	}
}

func (p *Preprocessor) isDefined(name string) bool {
	_, ok1 := p.obj[name]
	_, ok2 := p.fn[name]
	return ok1 || ok2
}

func (p *Preprocessor) evalIfExpr(expr string) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false
	}
	// Support: defined(NAME)
	if m := reDefined.FindStringSubmatch(expr); len(m) == 2 {
		return p.isDefined(m[1])
	}
	// Support: NAME (true if defined and value not "0" or empty)
	if p.isDefined(expr) {
		v := strings.TrimSpace(p.obj[expr])
		return v != "" && v != "0"
	}
	// Support: numeric literal
	if expr == "1" {
		return true
	}
	return false
}

var reDefined = regexp.MustCompile(`^defined$begin:math:text$\(\[\_A\-Za\-z\]\[\_A\-Za\-z0\-9\]\*\)$end:math:text$$`)

// expandLine expands macros using a token stream with pushback, similar to Go asm.
// It limits expansions to 100 to avoid infinite recursion.
func (p *Preprocessor) expandLine(line string) (string, error) {
	exp := expander{p: p}
	return exp.expand(line)
}

type expander struct {
	p          *Preprocessor
	stack      []inputChunk
	expansions int
}

type inputChunk struct {
	s string
	i int
}

func (e *expander) expand(line string) (string, error) {
	e.stack = []inputChunk{{s: line}}
	var b bytes.Buffer
	for {
		ch, ok := e.next()
		if !ok {
			break
		}
		if ch == '"' || ch == '\'' {
			b.WriteByte(ch)
			e.copyString(&b, ch)
			continue
		}
		if ch == '/' {
			if e.peekIs('/') {
				b.WriteByte(ch)
				b.WriteByte(e.mustNext())
				e.copyLineComment(&b)
				continue
			}
			if e.peekIs('*') {
				b.WriteByte(ch)
				b.WriteByte(e.mustNext())
				e.copyBlockComment(&b)
				continue
			}
		}
		if isIdentStart(ch) {
			name := e.readIdent(ch)
			if macro, ok := e.p.fn[name]; ok {
				if e.peekIs('(') {
					e.next()
					args, ok := e.readArgs()
					if ok {
						repl := applyFnMacroTokens(macro, args)
						if err := e.pushExpansion(repl); err != nil {
							return "", err
						}
						continue
					}
				}
				b.WriteString(name)
				continue
			}
			if val, ok := e.p.obj[name]; ok {
				if err := e.pushExpansion(val); err != nil {
					return "", err
				}
				continue
			}
			b.WriteString(name)
			continue
		}
		b.WriteByte(ch)
	}
	return b.String(), nil
}

func (e *expander) pushExpansion(s string) error {
	e.expansions++
	if e.expansions > 100 {
		return errors.New("recursive macro invocation")
	}
	if s != "" {
		e.stack = append(e.stack, inputChunk{s: s})
	}
	return nil
}

func (e *expander) next() (byte, bool) {
	for len(e.stack) > 0 {
		top := &e.stack[len(e.stack)-1]
		if top.i >= len(top.s) {
			e.stack = e.stack[:len(e.stack)-1]
			continue
		}
		ch := top.s[top.i]
		top.i++
		return ch, true
	}
	return 0, false
}

func (e *expander) mustNext() byte {
	ch, _ := e.next()
	return ch
}

func (e *expander) peekIs(b byte) bool {
	ch, ok := e.peek(0)
	return ok && ch == b
}

func (e *expander) peek(offset int) (byte, bool) {
	off := offset
	for i := len(e.stack) - 1; i >= 0; i-- {
		chunk := e.stack[i]
		pos := chunk.i
		if pos >= len(chunk.s) {
			continue
		}
		remain := len(chunk.s) - pos
		if off < remain {
			return chunk.s[pos+off], true
		}
		off -= remain
	}
	return 0, false
}

func (e *expander) readIdent(first byte) string {
	var b strings.Builder
	b.WriteByte(first)
	for {
		ch, ok := e.peek(0)
		if !ok || !isIdentPart(ch) {
			break
		}
		e.next()
		b.WriteByte(ch)
	}
	return b.String()
}

func (e *expander) readArgs() ([]string, bool) {
	var args []string
	var cur bytes.Buffer
	depth := 1
	for {
		ch, ok := e.next()
		if !ok {
			return nil, false
		}
		if ch == '"' || ch == '\'' {
			cur.WriteByte(ch)
			e.copyString(&cur, ch)
			continue
		}
		if ch == '/' {
			if e.peekIs('/') {
				cur.WriteByte(ch)
				cur.WriteByte(e.mustNext())
				e.copyLineComment(&cur)
				continue
			}
			if e.peekIs('*') {
				cur.WriteByte(ch)
				cur.WriteByte(e.mustNext())
				e.copyBlockComment(&cur)
				continue
			}
		}
		if ch == '(' {
			depth++
			cur.WriteByte(ch)
			continue
		}
		if ch == ')' {
			depth--
			if depth == 0 {
				args = append(args, strings.TrimSpace(cur.String()))
				return args, true
			}
			cur.WriteByte(ch)
			continue
		}
		if ch == ',' && depth == 1 {
			args = append(args, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}
}

func (e *expander) copyString(b *bytes.Buffer, quote byte) {
	for {
		ch, ok := e.next()
		if !ok {
			return
		}
		b.WriteByte(ch)
		if ch == '\\' {
			if next, ok := e.next(); ok {
				b.WriteByte(next)
			}
			continue
		}
		if ch == quote {
			return
		}
	}
}

func (e *expander) copyLineComment(b *bytes.Buffer) {
	for {
		ch, ok := e.next()
		if !ok {
			return
		}
		b.WriteByte(ch)
		if ch == '\n' {
			return
		}
	}
}

func (e *expander) copyBlockComment(b *bytes.Buffer) {
	for {
		ch, ok := e.next()
		if !ok {
			return
		}
		b.WriteByte(ch)
		if ch == '*' && e.peekIs('/') {
			b.WriteByte(e.mustNext())
			return
		}
	}
}

func applyFnMacroTokens(m FnMacro, args []string) string {
	argMap := make(map[string]string, len(m.Params))
	for i, p := range m.Params {
		if i < len(args) {
			argMap[p] = args[i]
		} else {
			argMap[p] = ""
		}
	}
	return replaceIdents(m.Body, argMap)
}

// parseParenArgs expects s starting with "(" and returns args and consumed length.
func replaceIdents(s string, repl map[string]string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		ch := s[i]
		if ch == '"' || ch == '\'' {
			quote := ch
			b.WriteByte(ch)
			i++
			for i < len(s) {
				ch = s[i]
				b.WriteByte(ch)
				i++
				if ch == '\\' && i < len(s) {
					b.WriteByte(s[i])
					i++
					continue
				}
				if ch == quote {
					break
				}
			}
			continue
		}
		if ch == '/' && i+1 < len(s) && s[i+1] == '/' {
			b.WriteString(s[i:])
			break
		}
		if ch == '/' && i+1 < len(s) && s[i+1] == '*' {
			b.WriteByte(ch)
			i++
			b.WriteByte(s[i])
			i++
			for i < len(s) {
				ch = s[i]
				b.WriteByte(ch)
				i++
				if ch == '*' && i < len(s) && s[i] == '/' {
					b.WriteByte(s[i])
					i++
					break
				}
			}
			continue
		}
		if isIdentStart(ch) {
			j := i + 1
			for j < len(s) && isIdentPart(s[j]) {
				j++
			}
			name := s[i:j]
			if val, ok := repl[name]; ok {
				b.WriteString(val)
			} else {
				b.WriteString(name)
			}
			i = j
			continue
		}
		b.WriteByte(ch)
		i++
	}
	return b.String()
}

func isIdentStart(b byte) bool {
	return b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isIdentPart(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

// ---------------- Lex-style drain (tests) ----------------

type lexLine struct {
	text  string
	hasNL bool
}

func lexDrain(input string) (string, error) {
	pp := NewPreprocessor()
	cond := newCondStack()
	lines := splitLinesKeepNewline(input)

	var buf strings.Builder
	for i := 0; i < len(lines); i++ {
		line := lines[i].text
		hasNL := lines[i].hasNL

		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			first := firstNonSpaceIndex(line)
			if first >= 0 && idx > first {
				return "", errors.New("'#' must be first item on line")
			}
		}

		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "#") {
			startLineNo := i + 1
			full := line
			for lineContinues(full) {
				full = stripLineContinuation(full)
				if i+1 >= len(lines) {
					break
				}
				i++
				full += "\n" + lines[i].text
			}
			fullTrim := strings.TrimSpace(full)
			fields := splitDirective(fullTrim)
			if fields.cmd == "define" && !lines[i].hasNL {
				_, _, body, ok := parseDefineDirective(fields.arg)
				if ok && strings.TrimSpace(body) != "" {
					return "", errors.New("no newline after macro definition")
				}
			}
			if err := pp.handleDirective(&bytes.Buffer{}, "<lex>", startLineNo, fullTrim, cond); err != nil {
				return "", err
			}
			continue
		}

		if !cond.Active() {
			continue
		}

		if err := lexProcessExpanded(&buf, pp, cond, line, hasNL); err != nil {
			return "", err
		}
	}
	if cond.Depth() != 0 {
		return "", errors.New("unclosed #ifdef or #ifndef")
	}
	return buf.String(), nil
}

// LexDrain tokenizes and expands input using the lex-style test harness.
// It returns the dot-separated token stream with newline tokens.
func LexDrain(input string) (string, error) {
	return lexDrain(input)
}

func lexProcessExpanded(buf *strings.Builder, p *Preprocessor, cond *condStack, line string, hasNL bool) error {
	expanded, err := p.expandLine(line)
	if err != nil {
		return err
	}
	parts := strings.Split(expanded, "\n")
	for i := 0; i < len(parts); i++ {
		emitNL := false
		if i < len(parts)-1 {
			emitNL = true
		} else {
			emitNL = hasNL
		}
		if err := lexProcessLine(buf, p, cond, parts[i], emitNL); err != nil {
			return err
		}
	}
	return nil
}

func lexProcessLine(buf *strings.Builder, p *Preprocessor, cond *condStack, line string, emitNL bool) error {
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(trim, "#") {
		if err := p.handleDirective(&bytes.Buffer{}, "<lex>", 0, trim, cond); err != nil {
			return err
		}
		return nil
	}
	if !cond.Active() {
		return nil
	}
	for _, tok := range lexTokens(line) {
		appendToken(buf, tok)
	}
	if emitNL {
		appendToken(buf, "\n")
	}
	return nil
}

func splitLinesKeepNewline(s string) []lexLine {
	if s == "" {
		return nil
	}
	lines := make([]lexLine, 0, 16)
	for {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			lines = append(lines, lexLine{text: s, hasNL: false})
			break
		}
		lines = append(lines, lexLine{text: s[:i], hasNL: true})
		s = s[i+1:]
		if s == "" {
			break
		}
	}
	return lines
}

func firstNonSpaceIndex(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return i
		}
	}
	return -1
}

func isDirectivePrefix(s string) bool {
	if !strings.HasPrefix(s, "#") {
		return false
	}
	s = strings.TrimSpace(s[1:])
	if s == "" {
		return false
	}
	cmd := strings.Fields(s)
	if len(cmd) == 0 {
		return false
	}
	switch cmd[0] {
	case "include", "define", "undef", "ifdef", "ifndef", "if", "elif", "else", "endif":
		return true
	default:
		return false
	}
}

func appendToken(buf *strings.Builder, tok string) {
	if tok == "" {
		return
	}
	if buf.Len() > 0 {
		buf.WriteByte('.')
	}
	buf.WriteString(tok)
}

func lexTokens(line string) []string {
	toks := make([]string, 0, 8)
	for i := 0; i < len(line); {
		ch := line[i]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			i++
			continue
		}
		if ch == '/' && i+1 < len(line) && line[i+1] == '/' {
			break
		}
		if ch == '/' && i+1 < len(line) && line[i+1] == '*' {
			i += 2
			for i+1 < len(line) {
				if line[i] == '*' && line[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			start := i
			quote := ch
			i++
			for i < len(line) {
				ch = line[i]
				i++
				if ch == '\\' && i < len(line) {
					i++
					continue
				}
				if ch == quote {
					break
				}
			}
			toks = append(toks, line[start:i])
			continue
		}
		if lexIsIdentStart(ch) {
			j := i + 1
			for j < len(line) && lexIsIdentPart(line[j]) {
				j++
			}
			toks = append(toks, line[i:j])
			i = j
			continue
		}
		if ch >= '0' && ch <= '9' {
			j := i + 1
			if i+1 < len(line) && line[i] == '0' {
				switch line[i+1] {
				case 'x', 'X':
					j = i + 2
					for j < len(line) && isHexDigit(line[j]) {
						j++
					}
				case 'b', 'B':
					j = i + 2
					for j < len(line) && (line[j] == '0' || line[j] == '1') {
						j++
					}
				case 'o', 'O':
					j = i + 2
					for j < len(line) && line[j] >= '0' && line[j] <= '7' {
						j++
					}
				default:
					for j < len(line) && line[j] >= '0' && line[j] <= '9' {
						j++
					}
				}
			} else {
				for j < len(line) && line[j] >= '0' && line[j] <= '9' {
					j++
				}
			}
			toks = append(toks, line[i:j])
			i = j
			continue
		}
		toks = append(toks, string(ch))
		i++
	}
	return toks
}

// LexTokens exposes the lex-style tokenizer for a single line.
func LexTokens(line string) []string {
	return lexTokens(line)
}

func lexIsIdentStart(b byte) bool {
	return isIdentStart(b) || b == '.' || b == '·'
}

func lexIsIdentPart(b byte) bool {
	return lexIsIdentStart(b) || (b >= '0' && b <= '9')
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// ---------------- Directive parsing helpers ----------------

type directiveFields struct {
	cmd string
	arg string
}

func splitDirective(trim string) directiveFields {
	// trim begins with '#'
	trim = strings.TrimSpace(trim[1:])
	if trim == "" {
		return directiveFields{}
	}
	sp := strings.Fields(trim)
	cmd := sp[0]
	arg := strings.TrimSpace(trim[len(cmd):])
	return directiveFields{cmd: cmd, arg: arg}
}

func parseIncludeArg(arg string) (string, bool) {
	arg = strings.TrimSpace(arg)
	if len(arg) >= 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
		return arg[1 : len(arg)-1], true
	}
	if len(arg) >= 2 && arg[0] == '<' && arg[len(arg)-1] == '>' {
		return arg[1 : len(arg)-1], true
	}
	return "", false
}

func parseDefineDirective(arg string) (name string, params []string, body string, ok bool) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return "", nil, "", false
	}

	// parse name
	if !isIdentStart(arg[0]) {
		return "", nil, "", false
	}
	i := 1
	for i < len(arg) && isIdentPart(arg[i]) {
		i++
	}
	name = arg[:i]
	rest := arg[i:]

	// function-like only if '(' immediately follows name
	if strings.HasPrefix(rest, "(") {
		j := strings.Index(rest, ")")
		if j < 0 {
			return "", nil, "", false
		}
		paramStr := rest[1:j]
		body = trimLeftSpaceTab(rest[j+1:])
		if strings.TrimSpace(paramStr) == "" {
			return name, []string{}, body, true
		}
		raw := strings.Split(paramStr, ",")
		params = make([]string, 0, len(raw))
		for _, r := range raw {
			params = append(params, strings.TrimSpace(r))
		}
		return name, params, body, true
	}

	// object-like: NAME body...
	body = trimLeftSpaceTab(arg[len(name):])
	return name, nil, body, true
}

func trimLeftSpaceTab(s string) string {
	return strings.TrimLeft(s, " \t")
}

func ParseDefine(s string) (name, value string) {
	if i := strings.IndexByte(s, '='); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, "1"
}

// ---------------- Include resolution ----------------

func (p *Preprocessor) readInclude(path string, includingFile string) ([]byte, string, error) {
	resolved, err := p.resolveAsFile(path, includingFile)
	if err != nil {
		return nil, "", err
	}
	bs, err := os.ReadFile(resolved)
	return bs, resolved, err
}

func (p *Preprocessor) resolveAsFile(path string, includingFile string) (string, error) {
	// If path is absolute or relative to including file
	if filepath.IsAbs(path) {
		if fileExists(path) {
			return filepath.Clean(path), nil
		}
		return "", os.ErrNotExist
	}

	// 1) relative to including file directory
	if includingFile != "" && includingFile != "<stdin>" {
		base := filepath.Dir(includingFile)
		cand := filepath.Join(base, path)
		if fileExists(cand) {
			return filepath.Clean(cand), nil
		}
	}

	// 2) include dirs
	for _, dir := range p.IncludeDirs {
		cand := filepath.Join(dir, path)
		if fileExists(cand) {
			return filepath.Clean(cand), nil
		}
	}
	return "", fmt.Errorf("cannot resolve include %q", path)
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func shortPath(p string) string {
	// nicer errors
	if p == "" {
		return p
	}
	return filepath.Base(p)
}

// ---------------- Conditionals ----------------

type condStack struct {
	// Each level stores: parentActive, thisBranchTaken, thisActive
	stack []condFrame
}

type condFrame struct {
	parentActive bool
	taken        bool
	active       bool
	line         int
}

func newCondStack() *condStack  { return &condStack{} }
func (c *condStack) Depth() int { return len(c.stack) }

func (c *condStack) Active() bool {
	if len(c.stack) == 0 {
		return true
	}
	return c.stack[len(c.stack)-1].active
}

func (c *condStack) Push(cond bool, line int) {
	parent := c.Active()
	active := parent && cond
	c.stack = append(c.stack, condFrame{
		parentActive: parent,
		taken:        active, // if active, branch is taken
		active:       active,
		line:         line,
	})
}

func (c *condStack) Elif(cond bool) {
	if len(c.stack) == 0 {
		return
	}
	top := &c.stack[len(c.stack)-1]
	if !top.parentActive {
		top.active = false
		return
	}
	if top.taken {
		top.active = false
		return
	}
	top.active = cond
	if cond {
		top.taken = true
	}
}

func (c *condStack) Else() {
	if len(c.stack) == 0 {
		return
	}
	top := &c.stack[len(c.stack)-1]
	if !top.parentActive {
		top.active = false
		return
	}
	top.active = !top.taken
	top.taken = true
}

func (c *condStack) Pop() {
	if len(c.stack) == 0 {
		return
	}
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *condStack) UnclosedLine() int {
	if len(c.stack) == 0 {
		return 0
	}
	return c.stack[len(c.stack)-1].line
}
