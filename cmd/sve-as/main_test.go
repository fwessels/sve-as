package main

import (
	"strings"
	"testing"
)

func TestAsm2s(t *testing.T) {
	normalize := func(s string) string {
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimLeft(line, " \t")
		}
		return strings.Join(lines, "\n")
	}
	got, err := asm2s([]byte(asm), false)
	if err != nil {
		t.Errorf("%v", err)
	} else if normalize(got) != normalize(opcodes) {
		t.Errorf("got: %v; want: %v\n", got, opcodes)
	}
	got, err = asm2s([]byte(asm), true)
	if err != nil {
		t.Errorf("%v", err)
	} else if normalize(got) != normalize(plan9s) {
		t.Errorf("got: %v; want: %v\n", got, plan9s)
	}
}

const (
	asm = `
TEXT ·snippets(SB), $0-8
    add x0, x1, x2
	cmp x3, x4
	beq done
    mov x2, #0x6e3a
    movk x2, #0x4f5d, lsl #16
    movk x2, #0xfedc, lsl #32
    movk x2, #0x1234, lsl #48
	add z0.s, z0.s, z0.s
done:
	ret
`
	opcodes = `
TEXT ·snippets(SB), $0-8
	WORD $0x8b020020 // add x0, x1, x2
	WORD $0xeb04007f // cmp x3, x4
	BEQ done
	WORD $0xd28dc742 // mov x2, #0x6e3a
	WORD $0xf2a9eba2 // movk x2, #0x4f5d, lsl #16
	WORD $0xf2dfdb82 // movk x2, #0xfedc, lsl #32
	WORD $0xf2e24682 // movk x2, #0x1234, lsl #48
	WORD $0x04a00000 // add z0.s, z0.s, z0.s
done:
	WORD $0xd65f03c0 // ret
`
	plan9s = `
TEXT ·snippets(SB), $0-8
	ADD R2, R1, R0
	CMP R4, R3
	BEQ done
	MOVD $28218, R2
	MOVK $(20317<<16), R2
	MOVK $(65244<<32), R2
	MOVK $(4660<<48), R2
	WORD $0x04a00000 // add z0.s, z0.s, z0.s
done:
	RET
`
)
