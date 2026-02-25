package preprocessor

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLex(t *testing.T) {
	for _, tt := range lexTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lexDrain(tt.input)
			if err != nil {
				t.Fatalf("lex error: %v", err)
			}
			if diff := cmp.Diff(tt.output, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBadLex(t *testing.T) {
	for _, tt := range badLexTests {
		t.Run(tt.error, func(t *testing.T) {
			_, err := lexDrain(tt.input)
			if err == nil {
				t.Fatalf("expected error %q", tt.error)
			}
			if diff := cmp.Diff(tt.error, err.Error()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// From: https://tip.golang.org/src/cmd/asm/internal/lex/lex_test.go

func lines(a ...string) string {
	return strings.Join(a, "\n") + "\n"
}

type lexTest struct {
	name   string
	input  string
	output string
}

var lexTests = []lexTest{
	{
		"empty",
		"",
		"",
	},
	{
		"simple",
		"1 (a)",
		"1.(.a.)",
	},
	{
		"simple define",
		lines(
			"#define A 1234",
			"A",
		),
		"1234.\n",
	},
	{
		"define without value",
		"#define A",
		"",
	},
	{
		"macro without arguments",
		"#define A() 1234\n" + "A()\n",
		"1234.\n",
	},
	{
		"macro with just parens as body",
		"#define A () \n" + "A\n",
		"(.).\n",
	},
	{
		"macro with parens but no arguments",
		"#define A (x) \n" + "A\n",
		"(.x.).\n",
	},
	{
		"macro with arguments",
		"#define A(x, y, z) x+z+y\n" + "A(1, 2, 3)\n",
		"1.+.3.+.2.\n",
	},
	{
		"argumented macro invoked without arguments",
		lines(
			"#define X() foo ",
			"X()",
			"X",
		),
		"foo.\n.X.\n",
	},
	{
		"multiline macro without arguments",
		lines(
			"#define A 1\\",
			"\t2\\",
			"\t3",
			"before",
			"A",
			"after",
		),
		"before.\n.1.\n.2.\n.3.\n.after.\n",
	},
	{
		"multiline macro with arguments",
		lines(
			"#define A(a, b, c) a\\",
			"\tb\\",
			"\tc",
			"before",
			"A(1, 2, 3)",
			"after",
		),
		"before.\n.1.\n.2.\n.3.\n.after.\n",
	},
	{
		"LOAD macro",
		lines(
			"#define LOAD(off, reg) \\",
			"\tMOVBLZX	(off*4)(R12),	reg \\",
			"\tADDB	reg,		DX",
			"",
			"LOAD(8, AX)",
		),
		"\n.\n.MOVBLZX.(.8.*.4.).(.R12.).,.AX.\n.ADDB.AX.,.DX.\n",
	},
	{
		"nested multiline macro",
		lines(
			"#define KEYROUND(xmm, load, off, r1, r2, index) \\",
			"\tMOVBLZX	(BP)(DX*4),	R8 \\",
			"\tload((off+1), r2) \\",
			"\tMOVB	R8,		(off*4)(R12) \\",
			"\tPINSRW	$index, (BP)(R8*4), xmm",
			"#define LOAD(off, reg) \\",
			"\tMOVBLZX	(off*4)(R12),	reg \\",
			"\tADDB	reg,		DX",
			"KEYROUND(X0, LOAD, 8, AX, BX, 0)",
		),
		"\n.MOVBLZX.(.BP.).(.DX.*.4.).,.R8.\n.\n.MOVBLZX.(.(.8.+.1.).*.4.).(.R12.).,.BX.\n.ADDB.BX.,.DX.\n.MOVB.R8.,.(.8.*.4.).(.R12.).\n.PINSRW.$.0.,.(.BP.).(.R8.*.4.).,.X0.\n",
	},
	{
		"taken #ifdef",
		lines(
			"#define A",
			"#ifdef A",
			"#define B 1234",
			"#endif",
			"B",
		),
		"1234.\n",
	},
	{
		"not taken #ifdef",
		lines(
			"#ifdef A",
			"#define B 1234",
			"#endif",
			"B",
		),
		"B.\n",
	},
	{
		"taken #ifdef with else",
		lines(
			"#define A",
			"#ifdef A",
			"#define B 1234",
			"#else",
			"#define B 5678",
			"#endif",
			"B",
		),
		"1234.\n",
	},
	{
		"not taken #ifdef with else",
		lines(
			"#ifdef A",
			"#define B 1234",
			"#else",
			"#define B 5678",
			"#endif",
			"B",
		),
		"5678.\n",
	},
	{
		"nested taken/taken #ifdef",
		lines(
			"#define A",
			"#define B",
			"#ifdef A",
			"#ifdef B",
			"#define C 1234",
			"#else",
			"#define C 5678",
			"#endif",
			"#endif",
			"C",
		),
		"1234.\n",
	},
	{
		"nested taken/not-taken #ifdef",
		lines(
			"#define A",
			"#ifdef A",
			"#ifdef B",
			"#define C 1234",
			"#else",
			"#define C 5678",
			"#endif",
			"#endif",
			"C",
		),
		"5678.\n",
	},
	{
		"nested not-taken/would-be-taken #ifdef",
		lines(
			"#define B",
			"#ifdef A",
			"#ifdef B",
			"#define C 1234",
			"#else",
			"#define C 5678",
			"#endif",
			"#endif",
			"C",
		),
		"C.\n",
	},
	{
		"nested not-taken/not-taken #ifdef",
		lines(
			"#ifdef A",
			"#ifdef B",
			"#define C 1234",
			"#else",
			"#define C 5678",
			"#endif",
			"#endif",
			"C",
		),
		"C.\n",
	},
	{
		"nested #define",
		lines(
			"#define A #define B THIS",
			"A",
			"B",
		),
		"THIS.\n",
	},
	{
		"nested #define with args",
		lines(
			"#define A #define B(x) x",
			"A",
			"B(THIS)",
		),
		"THIS.\n",
	},
	/* This one fails. See comment in Slice.Col.
	{
		"nested #define with args",
		lines(
			"#define A #define B (x) x",
			"A",
			"B(THIS)",
		),
		"x.\n",
	},
	*/
}

type badLexTest struct {
	input string
	error string
}

var badLexTests = []badLexTest{
	{
		"3 #define foo bar\n",
		"'#' must be first item on line",
	},
	{
		"#ifdef foo\nhello",
		"unclosed #ifdef or #ifndef",
	},
	{
		"#ifndef foo\nhello",
		"unclosed #ifdef or #ifndef",
	},
	{
		"#ifdef foo\nhello\n#else\nbye",
		"unclosed #ifdef or #ifndef",
	},
	{
		"#define A() A()\nA()",
		"recursive macro invocation",
	},
	{
		"#define A a\n#define A a\n",
		"redefinition of macro",
	},
	{
		"#define A a",
		"no newline after macro definition",
	},
}
