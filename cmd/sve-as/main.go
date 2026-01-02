package main

import (
	"bufio"
	"bytes"
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

			opcode, opcode2, err := sve_as.Assemble(strings.Split(instruction, "/*")[0])
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

		assembled.WriteString(line + "\n")
	}

	out = assembled.String()
	return
}

func asm2s(buf []byte, plan9s bool) (out string, err error) {
	assembled := strings.Builder{}
	scanner := bufio.NewScanner(bytes.NewReader(buf))

	for lineno := 0; scanner.Scan(); lineno++ {
		line := scanner.Text()
		fmt.Println(lineno, line)
		parts := strings.Split(line, "//")
		var comments string
		if len(parts) == 2 {
			line = strings.TrimRightFunc(parts[0], func(r rune) bool { return unicode.IsSpace(r) })
			comments = parts[0][len(line):] + "//" + parts[1]
		}
		if strings.TrimSpace(line) == "" ||
			strings.ToLower(line) != line ||
			strings.HasPrefix(strings.TrimSpace(line), "//") ||
			strings.HasSuffix(line, ":") {
		} else {
			opcode, opcode2, err := sve_as.Assemble(line)
			if err != nil {
				fmt.Printf("'%s'\n", line)
				fmt.Println(err)
				os.Exit(2)
			}
			if opcode2 == 0 {
				line = fmt.Sprintf("    WORD $0x%08x", opcode)
			} else {
				oc64 := uint64(opcode2)<<32 | uint64(opcode)
				line = fmt.Sprintf("    DWORD $0x%016x", oc64)
			}
		}
		assembled.WriteString(line + comments + "\n")
	}

	out = assembled.String()

	if plan9s {
		// Get GOROOT the same way the go tool does
		goroot := runtime.GOROOT()
		includeDir := filepath.Join(goroot, "pkg", "include")

		var srccode, objcode *os.File
		if srccode, err = os.CreateTemp("", "asm2s-*.s"); err != nil {
			return
		}
		if objcode, err = os.CreateTemp("", "asm2s-*.o"); err != nil {
			return
		}
		defer os.Remove(srccode.Name())
		defer os.Remove(objcode.Name())

		if err = os.WriteFile(srccode.Name(), []byte(out), 0666); err != nil {
			return
		}
		cmd := exec.Command(
			"go", "tool", "asm",
			"-o", objcode.Name(), "-I", includeDir,
			srccode.Name(),
		)
		if err = cmd.Run(); err != nil {
			return
		}

		// Capture stdout + stderr
		var objdump []byte
		if objdump, err = exec.Command("go", "tool", "objdump", objcode.Name()).
			CombinedOutput(); err != nil {
			return
		}

		// replace opcodes with plan9s instructions
		scanner := bufio.NewScanner(bytes.NewReader(objdump))
		for lineno := 0; scanner.Scan(); lineno++ {
			line := scanner.Text()
			flds := strings.Fields(line)
			if len(flds) >= 4 {
				opcode := flds[2]
				if opcode != "00000000" {
					instr := strings.Join(flds[3:], " ")
					out = strings.ReplaceAll(out,
						fmt.Sprintf("WORD $0x%s", opcode), instr)
				}
			}
		}
	}

	return
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sve-as <filename>")
		os.Exit(1)
	}

	fname := strings.ToLower(os.Args[1])
	isAsm, isS := strings.HasSuffix(fname, ".asm"), strings.HasSuffix(fname, ".s")
	if !isAsm && !isS {
		fmt.Println("Usage: sve-as <filename.s/.asm>")
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
			if processed, err = asm2s(buf, true); err != nil {
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
