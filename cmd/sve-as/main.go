package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	sve_as "github.com/fwessels/sve-as"
	"github.com/fwessels/sve-as/internal/preprocessor"
)

func assemble(buf []byte, hasDWordsMap *map[string]bool) (out string, containsDWordsMap map[string]bool) {
	containsDWordsMap = make(map[string]bool)

	assembled := strings.Builder{}
	scanner := bufio.NewScanner(bytes.NewReader(buf))

	r := regexp.MustCompile(`^TEXT ·([^\(]+)\(SB\)`)
	align, routineName := "", ""

	for scanner.Scan() {
		line := scanner.Text()

		matches := r.FindStringSubmatch(line)
		if len(matches) > 1 {
			routineName = matches[1] // Contains the extracted name
			if hasDWordsMap != nil && (*hasDWordsMap)[routineName] {
				align = strings.Repeat(" ", 9)
			} else {
				align = ""
			}
		}

		if strings.HasPrefix(line, "//") {
			// Intentionally ignore (skip full line of comments)
		} else if regexp.MustCompile(`(?:WORD \$0x[0-9a-f]{8}|DWORD \$0x[0-9a-f]{16})\s*//`).MatchString(line) {
			instruction := strings.Split(line, "//")[1]
			ins := strings.Split(instruction, "/*")[0]
			if pt, ok := passThrough(ins); ok {
				line = "    " + pt
			} else {
				opcode, opcode2, err := sve_as.Assemble(ins)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if opcode2 == 0 {
					line = fmt.Sprintf("    WORD $0x%08x %s//%s", opcode, align, instruction)
				} else {
					oc64 := uint64(opcode2)<<32 | uint64(opcode)
					line = fmt.Sprintf("    DWORD $0x%016x //%s", oc64, instruction)
					containsDWordsMap[routineName] = true
				}
			}
		}

		assembled.WriteString(line + "\n")
	}

	out = assembled.String()
	return
}

// Check for instructions to pass through and/or translate into plan9s equivalents
func passThrough(ins string) (string, bool) {
	allCaps := func(s string) (hasLetter bool) {
		for _, r := range s {
			if unicode.IsLetter(r) {
				hasLetter = true
				if !unicode.IsUpper(r) {
					return false
				}
			}
		}
		return
	}
	reg2Plan9s := func(reg string) string {
		if strings.HasPrefix(reg, "x") {
			return strings.ReplaceAll(reg, "x", "R")
		}
		return reg
	}
	if strings.TrimSpace(ins) == "" {
		return "", false
	}
	mnem := strings.Fields(ins)[0]
	args := strings.Fields(ins)[1:]
	for i := range args {
		args[i] = strings.TrimSpace(strings.ReplaceAll(args[i], ",", ""))
	}

	switch strings.ToLower(mnem) {
	case "ldr", "str":
		if len(args) == 2 && strings.HasSuffix(args[1], "(fp)]") {
			lbl := args[1]
			lbl = strings.ReplaceAll(lbl, "(fp)", "(FP)")
			lbl = strings.NewReplacer("[", "", "]", "").Replace(lbl)
			if mnem == "ldr" {
				return "MOVD" + " " + lbl + ", " + reg2Plan9s(args[0]), true
			} else {
				return "MOVD" + " " + reg2Plan9s(args[0]) + ", " + lbl, true
			}
		}

	case "adr":
		if allCaps(mnem) {
			return strings.TrimSpace(ins), true
		} else if len(args) == 2 {
			lbl := args[1]
			if strings.HasPrefix(lbl, "$·") && strings.HasSuffix(lbl, "(sb)") {
				// for absolute addresses, use MOVD instruction
				lbl = strings.ReplaceAll(lbl, "(sb)", "(SB)")
				return "MOVD" + " " + lbl + ", " + reg2Plan9s(args[0]), true
			} else {
				// for PC-relative addresses
				return strings.ToUpper(mnem) + " " + lbl + ", " + reg2Plan9s(args[0]), true
			}
		}

	case "movd":
		if allCaps(mnem) {
			return strings.TrimSpace(ins), true
		}

	case "b", "beq", "bne", "bcc", "blo", "bcs", "bmi", "bpl", "bvs", "bvc", "bhi", "bls", "bge", "blt", "bgt", "ble", "bal", "bnv",
		"b.eq", "b.ne", "b.cc", "b.lo", "b.cs", "b.mi", "b.pl", "b.vs", "b.vc", "b.hi", "b.ls", "b.ge", "b.lt", "b.gt", "b.le", "b.al", "b.nv":
		if allCaps(mnem) {
			return strings.TrimSpace(ins), true
		} else {
			return strings.ToUpper(mnem) + " " + strings.Join(strings.Fields(ins)[1:], " "), true
		}

	case "bl":
		if allCaps(mnem) {
			return strings.TrimSpace(ins), true
		}

	case "tbz", "tbnz":
		if allCaps(mnem) {
			return strings.TrimSpace(ins), true
		} else if len(args) == 3 {
			return strings.ToUpper(mnem) + " $" + strings.ReplaceAll(args[1], "#", "") + ", " + reg2Plan9s(args[0]) + ", " + strings.Join(args[2:], " "), true
		}

	case "cbz", "cbnz":
		if allCaps(mnem) {
			return strings.TrimSpace(ins), true
		} else if len(args) == 2 {
			return strings.ToUpper(mnem) + " " + reg2Plan9s(args[0]) + ", " + strings.Join(args[1:], " "), true
		}

	case "jmp":
		if allCaps(mnem) {
			return strings.TrimSpace(ins), true
		}
	}

	return "", false
}

func NewPreprocessor(fname string) (pp *preprocessor.Preprocessor, err error) {
	pp = preprocessor.NewPreprocessor()
	pp.KeepLineComments = false // true for `// textflag.h:10` references
	if fname != "" {
		pp.IncludeDirs = append(pp.IncludeDirs, filepath.Dir(fname))
	}

	cmd := exec.Command("go", "env", "GOROOT")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	goroot := strings.TrimSpace(out.String())
	runtimePath := filepath.Join(goroot, "src", "runtime")
	pp.IncludeDirs = append(pp.IncludeDirs, runtimePath)

	// // Apply -D defines
	// for _, def := range D {
	// 	name, val := asmpp.ParseDefine(def)
	// 	pp.DefineObject(name, val)
	// }
	return
}

func asm2s(fname string, buf []byte, toPlan9s bool) (out string, err error) {

	var pp *preprocessor.Preprocessor
	if pp, err = NewPreprocessor(fname); err != nil {
		return "", err
	}

	preprocessed := bytes.Buffer{}
	if err := pp.Process(fname, bytes.NewReader(buf), &preprocessed); err != nil {
		return "", err
	}

	assembled := strings.Builder{}
	scanner := bufio.NewScanner(&preprocessed)

	for lineno := 0; scanner.Scan(); lineno++ {
		line := scanner.Text()
		// fmt.Println(lineno, line)
		parts := strings.Split(line, "//")
		var comments string
		if len(parts) == 2 {
			line = strings.TrimRightFunc(parts[0], func(r rune) bool { return unicode.IsSpace(r) })
			comments = parts[0][len(line):] + "//" + parts[1]
		}
		if strings.TrimSpace(line) == "" ||
			strings.ToLower(line) != line /* line contains any upper case letters? */ ||
			strings.HasPrefix(strings.TrimSpace(line), "//") ||
			strings.HasPrefix(strings.TrimSpace(line), "#include") ||
			strings.HasSuffix(line, ":") {
			// pass along verbatim
		} else if pt, ok := passThrough(line); ok {
			line = "    " + pt
		} else {
			opcode, opcode2, err := sve_as.Assemble(line)
			if err != nil {
				fmt.Printf("'%s'\n", line)
				fmt.Println(err)
				os.Exit(2)
			}
			if opcode2 == 0 {
				line = fmt.Sprintf("    WORD $0x%08x // %s", opcode, strings.TrimSpace(line))
			} else {
				oc64 := uint64(opcode2)<<32 | uint64(opcode)
				line = fmt.Sprintf("    DWORD $0x%016x // %s", oc64, strings.TrimSpace(line))
			}
		}
		assembled.WriteString(line + comments + "\n")
	}

	if toPlan9s {
		return translateBackToPlan9s(assembled.String())
	} else {
		return assembled.String(), nil
	}
}

func translateBackToPlan9s(opcodes string) (string, error) {
	// Get GOROOT the same way the go tool does
	goroot := runtime.GOROOT()
	includeDir := filepath.Join(goroot, "pkg", "include")

	var srccode, objcode, disas *os.File
	var err error
	if srccode, err = os.CreateTemp("", "asm2s-*.s"); err != nil {
		return "", err
	}
	if objcode, err = os.CreateTemp("", "asm2s-*.o"); err != nil {
		return "", err
	}
	if disas, err = os.CreateTemp("", "asm2s-*.disas"); err != nil {
		return "", err
	}
	defer os.Remove(srccode.Name())
	defer os.Remove(objcode.Name())
	defer os.Remove(disas.Name())

	if err = os.WriteFile(srccode.Name(), []byte(opcodes), 0666); err != nil {
		return "", err
	}
	cmd := exec.Command(
		"go", "tool", "asm",
		"-o", objcode.Name(), "-I", includeDir,
		srccode.Name(),
	)
	if goasm, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("go tool asm failed: %w\noutput:\n%s", err, goasm)
	}

	// Capture stdout + stderr
	var objdump []byte
	if objdump, err = exec.Command("go", "tool", "objdump", objcode.Name()).
		CombinedOutput(); err != nil {
		return "", fmt.Errorf("go tool objdump failed: %w\noutput:\n%s", err, objdump)
	}

	// fmt.Println(string(objdump))
	if err = os.WriteFile(disas.Name(), []byte(objdump), 0666); err != nil {
		return "", err

	}
	type disasOpcode struct {
		ophex string
		instr string
	}
	opcodesByLine := map[int][]disasOpcode{}
	scanObjdump := bufio.NewScanner(bytes.NewReader(objdump))
	for scanObjdump.Scan() {
		flds := strings.Fields(scanObjdump.Text())
		if len(flds) < 4 {
			continue
		}
		if _, err := hex.DecodeString(flds[2]); err != nil {
			continue
		}

		// first field is "<source>:<line>"
		colon := strings.LastIndexByte(flds[0], ':')
		if colon <= 0 || colon+1 >= len(flds[0]) {
			continue
		}
		lineno, err := strconv.Atoi(flds[0][colon+1:])
		if err != nil {
			continue
		}

		opcodesByLine[lineno] = append(opcodesByLine[lineno], disasOpcode{
			ophex: strings.ToLower(flds[2]),
			instr: strings.Join(flds[3:], " "),
		})
	}

	findByOpcode := func(line []disasOpcode, ophex string) (string, bool) {
		ophex = strings.ToLower(ophex)
		for _, d := range line {
			if d.ophex == ophex {
				return d.instr, true
			}
		}
		return "", false
	}

	extractHex := func(line, prefix string, n int) (string, bool) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			return "", false
		}
		hexPart := line[len(prefix):]
		if len(hexPart) < n {
			return "", false
		}
		hexPart = hexPart[:n]
		if _, err := hex.DecodeString(hexPart); err != nil {
			return "", false
		}
		return strings.ToLower(hexPart), true
	}

	plan9s := strings.Builder{}
	scanner := bufio.NewScanner(bytes.NewReader([]byte(opcodes)))
	// replace opcodes with plan9 instructions using objdump source line numbers
	for lineno := 1; scanner.Scan(); lineno++ {
		line := scanner.Text()
		if pt, ok := passThrough(line); ok {
			// Preserve symbolic form (e.g. labels) from the original source.
			line = "    " + pt
		} else if strings.HasPrefix(strings.TrimSpace(line), "WORD $0x") {
			if idx := strings.Index(line, "//"); idx >= 0 {
				if pt, ok := passThrough(line[idx+2:]); ok {
					// Prefer source-comment instruction when it carries labels.
					line = "    " + pt
					plan9s.WriteString(line + "\n")
					continue
				}
			}
			if ophex, ok := extractHex(line, "WORD $0x", 8); ok {
				if instr, found := findByOpcode(opcodesByLine[lineno], ophex); found && instr != "?" {
					line = "    " + instr
				}
			}
		} else if strings.HasPrefix(strings.TrimSpace(line), "DWORD $0x") {
			if ophex, ok := extractHex(line, "DWORD $0x", 16); ok {
				upper := ophex[:8]
				lower := ophex[8:]
				lineOpcodes := opcodesByLine[lineno]
				upperInstr, hasUpper := findByOpcode(lineOpcodes, upper)
				lowerInstr, hasLower := findByOpcode(lineOpcodes, lower)
				if hasLower && lowerInstr != "?" {
					line = "    " + lowerInstr
				}
				if hasUpper && upperInstr != "?" {
					plan9s.WriteString("    " + upperInstr + "\n")
				}
			}
		}
		plan9s.WriteString(line + "\n")
	}

	return plan9s.String(), nil
}

func main() {
	plan9 := flag.Bool("plan9", false, "enable plan9 disassembly for asm mode")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: sve-as [-plan9] <filename.s/.asm> [...]")
		os.Exit(1)
	}

	for _, fname := range args {
		fname = strings.ToLower(fname)
		isAsm, isS := strings.HasSuffix(fname, ".asm"), strings.HasSuffix(fname, ".s")
		if !isAsm && !isS {
			fmt.Println("Usage: sve-as [-plan9] <filename.s/.asm> [...]")
			os.Exit(1)
		}

		if buf, err := os.ReadFile(fname); err != nil {
			fmt.Println("Error reading file: ", err)
			os.Exit(1)
		} else {
			var processed string
			var err error
			if isAsm {
				fmt.Printf("Processing %s", fname)
				fname = strings.ReplaceAll(fname, ".asm", ".s")
				fmt.Printf(" → %s\n", fname)
				if processed, err = asm2s(fname, buf, *plan9); err != nil {
					log.Fatal(err)
				}
			}
			if isS {
				fmt.Println("Processing", fname)
				_, containsDWordsMap := assemble(buf, nil)
				processed, _ = assemble(buf, &containsDWordsMap)
			}
			if err := os.WriteFile(fname, []byte(processed), 0644); err != nil {
				fmt.Println("Error writing file: ", err)
				os.Exit(1)
			}
		}
	}
}
