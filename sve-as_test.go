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
	}

	for i, tc := range testCases {
		ins := strings.TrimSpace(strings.Split(tc.ins, "//")[1])
		oc, err := Assemble(ins)
		if err != nil {
			t.Errorf("TestSveAssembler(%d): `%s`: %v", i, ins, err)
		} else {
			opcode := fmt.Sprintf("%08x", oc)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestSveAssembler(%d): `%s`: got: 0x%s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				oc2, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 32)
				if err == nil {
					fmt.Printf("%32s\n", strconv.FormatUint(uint64(oc), 2))
					fmt.Printf("%32s\n", strconv.FormatUint(uint64(oc2), 2))
				}
			}
		}
	}
}
