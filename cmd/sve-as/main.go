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
			if pt, ok := sve_as.PassThrough(ins); ok {
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
		} else if pt, ok := sve_as.PassThrough(line); ok {
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

	opcodes := assembled.String()

	if !toPlan9s {
		out = opcodes
	} else {
		// Get GOROOT the same way the go tool does
		goroot := runtime.GOROOT()
		includeDir := filepath.Join(goroot, "pkg", "include")

		var srccode, objcode, disas *os.File
		if srccode, err = os.CreateTemp("", "asm2s-*.s"); err != nil {
			return
		}
		if objcode, err = os.CreateTemp("", "asm2s-*.o"); err != nil {
			return
		}
		if disas, err = os.CreateTemp("", "asm2s-*.disas"); err != nil {
			return
		}
		defer os.Remove(srccode.Name())
		defer os.Remove(objcode.Name())
		defer os.Remove(disas.Name())

		if err = os.WriteFile(srccode.Name(), []byte(opcodes), 0666); err != nil {
			return
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
			return
		}
		getNextOpcode := func(scan *bufio.Scanner) (ophex, instr string) {
			for scan.Scan() {
				flds := strings.Fields(scan.Text())
				if len(flds) >= 4 {
					if _, err := hex.DecodeString(flds[2]); err == nil {
						ophex = flds[2]
						instr = strings.Join(flds[3:], " ")
						return
					}
				}
			}
			return "", ""
		}
		plan9s := strings.Builder{}
		scanner := bufio.NewScanner(bytes.NewReader([]byte(opcodes)))
		scanObjdump := bufio.NewScanner(bytes.NewReader(objdump))
		// replace opcodes with plan9s instructions
		for lineno := 1; scanner.Scan(); lineno++ {
			line := scanner.Text()
			if pt, ok := sve_as.PassThrough(line); ok {
				_, instr := getNextOpcode(scanObjdump)
				// fmt.Println(pt, "|", instr)
				if strings.Fields(pt)[0] == "MOVD" && strings.Fields(instr)[0] == "ADRP" {
					// MOVD $·const(SB), R3 becomes two instructions:
					//   ....  90000003        ADRP 0(PC), R3          [0:8]R_ADDRARM64:<unlinkable>.const
					//   ....  91000063        ADD $0, R3, R3
					_, instr := getNextOpcode(scanObjdump)
					if strings.Fields(instr)[0] != "ADD" {
						panic("out of sync")
					}
				} else if strings.Fields(pt)[0] == "B" && strings.Fields(instr)[0] == "JMP" ||
					strings.Fields(pt)[0] == "BLO" && strings.Fields(instr)[0] == "BCC" {
					// synonyms -- accept
				} else if strings.Fields(pt)[0] != strings.Fields(instr)[0] {
					panic(fmt.Sprintf("out of sync: %s vs %s", strings.Join(strings.Fields(pt), " "), strings.Join(strings.Fields(instr), " ")))
				}
			} else if strings.HasPrefix(strings.TrimSpace(line), "WORD $0x") {
				ophex, instr := getNextOpcode(scanObjdump)
				if strings.TrimSpace(line)[len("WORD $0x"):len("WORD $0x")+8] == ophex {
					if instr == "?" {
						// NOP -- keep existing line
					} else {
						line = "    " + instr
					}
				} else {
					panic("out of sync")
				}
			} else if strings.HasPrefix(strings.TrimSpace(line), "DWORD $0x") {
				panic("handle case")
			}
			plan9s.WriteString(line + "\n")
		}

		out = plan9s.String()
	}

	return
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
