package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sve-as <filename>")
		os.Exit(1)
	}

	fmt.Println("Processing", os.Args[1])

	if buf, err := os.ReadFile(os.Args[1]); err != nil {
		fmt.Println("Error reading file: ", err)
		os.Exit(1)
	} else {
		_, containsDWordsMap := assemble(buf, nil)
		assembled, _ := assemble(buf, &containsDWordsMap)
		if err := os.WriteFile(os.Args[1], []byte(assembled), 0644); err != nil {
			fmt.Println("Error writing file: ", err)
			os.Exit(1)
		}
	}
}
