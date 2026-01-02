package main

import (
	"testing"
)

func TestAsm2s(t *testing.T) {
	got, err := asm2s([]byte(asm), false)
	if err != nil {
		t.Errorf("%v", err)
	} else if got != opcodes {
		t.Errorf("got: %v; want: %v\n", got, opcodes)
	}
	got, err = asm2s([]byte(asm), true)
	if err != nil {
		t.Errorf("%v", err)
	} else if got != plan9s {
		t.Errorf("got: %v; want: %v\n", got, plan9s)
	}
}

const (
	asm = `
TEXT ·snippets(SB), $0-8
    add x0, x1, x2
    ret
`
	opcodes = `
TEXT ·snippets(SB), $0-8
    WORD $0x8b020020
    WORD $0xd65f03c0
`
	plan9s = `
TEXT ·snippets(SB), $0-8
    ADD R2, R1, R0
    RET
`
)
