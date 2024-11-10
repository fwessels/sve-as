/*
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sve_as

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestSveAssembler(t *testing.T) {
	testCases := []struct {
		ins string
	}{
		// scalar instructions
		{"    WORD $0x8b0f01ce // add x14, x14, x15"},
		{"    WORD $0x8b0f0129 // add x9, x9, x15"},
		{"    WORD $0x91010108 // add x8, x8, #64"},
		{"    WORD $0xf1000400 // subs x0, x0, #1"},
		{"    WORD $0xd346fc00 // lsr x0, x0, #6"},
		{"    WORD $0xd37ae400 // lsl x0, x0, #6"},
		{"    WORD $0xea00001f // tst x0, x0"},
		{"    WORD $0x04225022 // addvl x2, x2, #1"},
		{"    WORD $0x04bf5050 // rdvl x16, #2"},
		{"    WORD $0x9ac10800 // udiv x0, x0, x1"},
		{"    WORD $0xd2800001 // mov   x1, #0"},
		//
		// vector instructions
		{"    WORD $0x05e039e2 // mov z2.d, x15"},
		{"    WORD $0x85804425 // ldr z5, [x1, #1, MUL VL]"},
		{"    WORD $0x85804026 // ldr z6, [x1]"},
		{"    WORD $0x85800141 // ldr p1, [x10]"},
		{"    WORD $0xe58041c0 // str z0, [x14]"},
		{"    WORD $0xe58045c1 // str z1, [x14, #1, MUL VL]"},
		{"    WORD $0x042230c6 // and z6.d, z6.d, z2.d"},
		{"    WORD $0x042230a5 // and z5.d, z5.d, z2.d"},
		{"    WORD $0x04e30041 // add z1.d, z2.d, z3.d"},
		{"    WORD $0x04c00461 // add z1.d, p1/M, z1.d, z3.d"},
		{"    WORD $0x0490054b // mul z11.s, p1/M, z11.s, z10.s"},
		{"    WORD $0x0450058d // mul z13.h, p1/M, z13.h, z12.h"},
		{"    WORD $0x05253065 // tbl z5.b, z3.b, z5.b"},
		{"    WORD $0x05283086 // tbl z6.b, z4.b, z8.b"},
		{"    WORD $0x04a33080 // eor z0.d, z4.d, z3.d"},
		{"    WORD $0x05212042 // dup z2.b, z2.b[0]"},
		{"    WORD $0x04fc94c7 // lsr z7.d, z6.d, #4"},
		{"    WORD $0x04fc94a8 // lsr z8.d, z5.d, #4"},
		{"    WORD $0x04233880 // eor3 z0.d, z0.d, z3.d, z4.d"},
		{"    WORD $0x05e18441 // compact z1.d, p1, z2.d"},
		{"    WORD $0x05e36041 // zip1 z1.d, z2.d, z3.d"},
		{"    WORD $0x05a36441 // zip2 z1.s, z2.s, z3.s"},
		{"    WORD $0x05a668a4 // uzp1 z4.s, z5.s, z6.s"},
		{"    WORD $0x05a96d07 // uzp2 z7.s, z8.s, z9.s"},
		{"    WORD $0x05a37041 // trn1 z1.s, z2.s, z3.s"},
		{"    WORD $0x05a674a4 // trn2 z4.s, z5.s, z6.s"},
		{"    WORD $0x05f83841 // rev z1.d, z2.d"},
		{"    WORD $0x05e48861 // revb z1.d, p2/M, z3.d"},
		{"    WORD $0x05e594c4 // revh z4.d, p5/M, z6.d"},
		{"    WORD $0x05e694c4 // revw z4.d, p5/M, z6.d"},
		{"    WORD $0x449a02dc // sdot z28.s, z22.b, z26.b"},
		{"    WORD $0x6589a231 // fcvt z17.s, p0/m, z17.h"},
		{"    WORD $0x65910a52 // fmul z18.s, z18.s, z17.s"},
		{"    WORD $0x047c9231 // asr z17.s, z17.s, #0x4"},
		{"    WORD $0x6594a231 // scvtf z17.s, p0/m, z17.s"},
		{"    WORD $0x65b2023f // fmla z31.s, p0/M, z17.s, z18.s"},
		{"    WORD $0x2538c1e0 // dup z0.b, #15"},
		{"    WORD $0x2578c0e1 // dup z1.h, #7"},
		{"    WORD $0x25b8c0a2 // dup z2.s, #5"},
		{"    WORD $0x25f8c163 // dup z3.d, #11"},
		{"    WORD $0x05800020 // and z0.b, z0.b, #1"},
		{"    WORD $0x05800021 // and z1.h, z1.h, #1"},
		{"    WORD $0x05800022 // and z2.s, z2.s, #1"},
		{"    WORD $0x05800023 // and z3.d, z3.d, #1"},
		{"    WORD $0x05801fcb // and z11.b, z11.b, #254"},
		{"    WORD $0x05ac856a // splice z10.s, p1/M, z10.s, z11.s"},
		{"    WORD $0x05acc56a // sel z10.s, p1/M, z11.s, z12.s"},
		{"    WORD $0x049a0ca4 // and z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04908ca4 // asr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04948ca4 // asrr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x049b0ca4 // bic z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x05a88ca4 // clasta z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x05a98ca4 // clastb z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04990ca4 // eor z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04938ca4 // lsl z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04978ca4 // lslr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04918ca4 // lsr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04958ca4 // lsrr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04980ca4 // orr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x048c0ca4 // sabd z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04960ca4 // sdivr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x048a0ca4 // smin z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04920ca4 // smulh z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04810ca4 // sub z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04830ca4 // subr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x048d0ca4 // uabd z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x05aadaaa // mov z10.s, p6/m,  z21.s"},
	}

	for i, tc := range testCases {
		ins := strings.TrimSpace(strings.Split(tc.ins, "//")[1])
		oc, oc2, err := Assemble(ins)
		if err != nil || oc2 != 0 {
			t.Errorf("TestSveAssembler(%d): `%s`: %v", i, ins, err)
		} else {
			opcode := fmt.Sprintf("%08x", oc)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestSveAssembler(%d): `%s`: got: 0x%s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 32)
				if err == nil {
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(oc), 2))
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		}
	}
}

func TestDWords(t *testing.T) {
	testCases := []struct {
		ins string
	}{
		{"    WORD $0x04913bfc // movprfx z28.s, p6/m, z31.s"},
		{"    WORD $0x04502c46 // movprfx z6.h, p3/z, z2.h"},
		{"    WORD $0x049034c7 // movprfx z7.s, p5/z, z6.s"},
		{"    WORD $0x0420bc5a // movprfx z26, z2"},
		{"    WORD $0x04913447 // movprfx z7.s, p5/m, z2.s"},
		{"    WORD $0x0491384b // movprfx z11.s, p6/m, z2.s"},
		{"    WORD $0x0420bf9e // movprfx z30, z28"},
		{"    WORD $0x0420bcff // movprfx z31, z7"},
		{"    WORD $0x049124ff // movprfx z31.s, p1/m, z7.s"},
		{"    WORD $0x04903bfc // movprfx z28.s, p6/z, z31.s"},
		//
		{"    WORD $0x04800441 // add z1.s, p1/M, z1.s, z2.s"},
		{"    WORD $0x04912441 // movprfx z1.s, p1/M, z2.s"},
		{"    WORD $0x04800461 // add z1.s, p1/M, z1.s, z3.s"},
		//
		// Merges ...
		{"    DWORD $0x0480046104912441 // add z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0490046104912441 // mul z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x049a046104912441 // and z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0499046104912441 // eor z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0491846104912441 // lsr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0493846104912441 // lsl z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0490846104912441 // asr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0494846104912441 // asrr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x05ac846104912441 // splice z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x049b046104912441 // bic z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x05a8846104912441 // clasta z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x05a9846104912441 // clastb z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0497846104912441 // lslr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0495846104912441 // lsrr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0498046104912441 // orr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x048c046104912441 // sabd z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0496046104912441 // sdivr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x048a046104912441 // smin z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0492046104912441 // smulh z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0481046104912441 // sub z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0483046104912441 // subr z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x048d046104912441 // uabd z1.s, p1/M, z2.s, z3.s"},
		//
		// Zeroing ...
		{"    DWORD $0x0480046104902441 // add z1.s, p1/Z, z2.s, z3.s"},
		//
		{"    DWORD $0x0401858504112425 // lsr z5.b, p1/m, z1.b, #4"},
		{"    DWORD $0x04018f0604512c46 // lsr z6.h, p3/m, z2.h, #8"},
		{"    DWORD $0x0441960704913467 // lsr z7.s, p5/m, z3.s, #16"},
		{"    DWORD $0x04c19c0804d13c88 // lsr z8.d, p7/m, z4.d, #32"},
	}

	for i, tc := range testCases {
		ins := strings.TrimSpace(strings.Split(tc.ins, "//")[1])
		oc, oc2, err := Assemble(ins)
		if err != nil {
			t.Errorf("TestSveAssembler(%d): `%s`: %v", i, ins, err)
		} else if oc2 != 0 {
			oc64 := uint64(oc2)<<32 | uint64(oc)
			opcode := fmt.Sprintf("%016x", oc64)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestDWords(%d): `%s`: got: 0x%s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 64)
				if err == nil {
					fmt.Printf("%064s\n", strconv.FormatUint(oc64, 2))
					fmt.Printf("%064s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		} else {
			opcode := fmt.Sprintf("%08x", oc)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestDWords(%d): `%s`: got: 0x%s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 32)
				if err == nil {
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(oc), 2))
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		}
	}
}
