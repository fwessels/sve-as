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
#include "textflag.h"

TEXT ·snippets(SB), $0-8
	ldr x1, [arg0+0(fp)]
	ldr x2, [arg1+8(fp)]
    add x0, x1, x2
loop:
	cmp x3, x4
	beq done
	adr x1, loop
	adr x3, $·const(sb)
    mov x2, #0x6e3a
    movk x2, #0x4f5d, lsl #16
    movk x2, #0xfedc, lsl #32
    movk x2, #0x1234, lsl #48
	add z0.s, z0.s, z0.s
done:
    str x2, [ret+16(fp)]
	ret

DATA ·const+0x000(SB)/8, $0x0102030405060708
GLOBL ·const(SB), (NOPTR+RODATA), $16
`
	opcodes = `
#include "textflag.h"

TEXT ·snippets(SB), $0-8
	MOVD arg0+0(FP), R1
	MOVD arg1+8(FP), R2
	WORD $0x8b020020 // add x0, x1, x2
loop:
	WORD $0xeb04007f // cmp x3, x4
	BEQ done
	ADR loop, R1
	MOVD $·const(SB), R3
	WORD $0xd28dc742 // mov x2, #0x6e3a
	WORD $0xf2a9eba2 // movk x2, #0x4f5d, lsl #16
	WORD $0xf2dfdb82 // movk x2, #0xfedc, lsl #32
	WORD $0xf2e24682 // movk x2, #0x1234, lsl #48
	WORD $0x04a00000 // add z0.s, z0.s, z0.s
done:
	MOVD R2, ret+16(FP)
	WORD $0xd65f03c0 // ret

DATA ·const+0x000(SB)/8, $0x0102030405060708
GLOBL ·const(SB), (NOPTR+RODATA), $16
`
	plan9s = `
#include "textflag.h"

TEXT ·snippets(SB), $0-8
	MOVD arg0+0(FP), R1
	MOVD arg1+8(FP), R2
	ADD R2, R1, R0
loop:
	CMP R4, R3
	BEQ done
	ADR loop, R1
	MOVD $·const(SB), R3
	MOVD $28218, R2
	MOVK $(20317<<16), R2
	MOVK $(65244<<32), R2
	MOVK $(4660<<48), R2
	WORD $0x04a00000 // add z0.s, z0.s, z0.s
done:
	MOVD R2, ret+16(FP)
	RET

DATA ·const+0x000(SB)/8, $0x0102030405060708
GLOBL ·const(SB), (NOPTR+RODATA), $16
`
)
