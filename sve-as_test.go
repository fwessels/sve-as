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
		{"    WORD $0x8b030441 // add x1, x2, x3, lsl #1"},
		{"    WORD $0x8b440862 // add x2, x3, x4, lsr #2"},
		{"    WORD $0x8b850c83 // add x3, x4, x5, asr #3"},
		{"    WORD $0x91010108 // add x8, x8, #64"},
		{"    WORD $0xf1000400 // subs x0, x0, #1"},
		{"    WORD $0xd101f050 // sub x16, x2, #124"},
		{"    WORD $0xd346fc00 // lsr x0, x0, #6"},
		{"    WORD $0xd37ae400 // lsl x0, x0, #6"},
		{"    WORD $0xea00001f // tst x0, x0"},
		{"    WORD $0x04225022 // addvl x2, x2, #1"},
		{"    WORD $0x04bf5050 // rdvl x16, #2"},
		{"    WORD $0x9ac10800 // udiv x0, x0, x1"},
		{"    WORD $0xd2800001 // mov x1, #0"},
		{"    WORD $0xaa2003e0 // mvn x0, x0"},
		{"    WORD $0xd503201f // nop"},
		{"    WORD $0xd65f03c0 // ret"},
		{"    WORD $0x79400801 // ldrh w1, [x0, #4]"},
		{"    WORD $0x79400003 // ldrh w3, [x0]"},
		{"    WORD $0x50002682 // adr x2, #1234"},
		{"    WORD $0x10000801 // adr x1, #256"},
		{"    WORD $0x10ffe002 // adr x2, #-1024"},
		{"    WORD $0x10ffefe3 // adr x3, #-516"},
		{"    WORD $0xf9000041 // str x1, [x2]"},
		{"    WORD $0xf900068a // str x10, [x20, #8]"},

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
		{"    WORD $0x052b2927 // tbl z7.b, { z9.b, z10.b }, z11.b"},
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
		{"    WORD $0x2538de20 // dup z0.b, #-15"},
		{"    WORD $0x25b8f016 // dup z22.s,  #-32768"},
		{"    WORD $0x25b8eff6 // dup z22.s,  #0x7f00"},
		{"    WORD $0x25b8e036 // dup z22.s,  #256"},
		{"    WORD $0x25b8fff6 // dup z22.s,  #-256"},
		{"    WORD $0x25b8eff6 // dup z22.s,  #32512"},
		{"    WORD $0x2578c0e1 // dup z1.h, #7"},
		{"    WORD $0x25b8c0a2 // dup z2.s, #5"},
		{"    WORD $0x25f8c163 // dup z3.d, #11"},
		{"    WORD $0x2538c816 // dup z22.b, #64"},
		{"    WORD $0x043602d6 // add z22.b, z22.b, z22.b"},
		{"    WORD $0x05800600 // and z0.b, z0.b, #1"},
		{"    WORD $0x058006c0 // and z0.b, z0.b, #0x7f"},
		{"    WORD $0x05800e00 // and z0.b, z0.b, #0x80"},
		{"    WORD $0x05803ec0 // and z0.b, z0.b, #0xfe"},
		{"    WORD $0x05800401 // and z1.h, z1.h, #1"},
		{"    WORD $0x058005c1 // and z1.h, z1.h, #0x7fff"},
		{"    WORD $0x05800c01 // and z1.h, z1.h, #0x8000"},
		{"    WORD $0x05807dc1 // and z1.h, z1.h, #0xfffe"},
		{"    WORD $0x05800002 // and z2.s, z2.s, #1"},
		{"    WORD $0x058003c2 // and z2.s, z2.s, #0x7fffffff"},
		{"    WORD $0x05800802 // and z2.s, z2.s, #0x80000000"},
		{"    WORD $0x0580fbc2 // and z2.s, z2.s, #0xfffffffe"},
		{"    WORD $0x05820003 // and z3.d, z3.d, #1"},
		{"    WORD $0x058207c3 // and z3.d, z3.d, #0x7fffffffffffffff"},
		{"    WORD $0x05820803 // and z3.d, z3.d, #0x8000000000000000"},
		{"    WORD $0x0583ffc3 // and z3.d, z3.d, #0xfffffffffffffffe"},
		{"    WORD $0x05803ecb // and z11.b, z11.b, #254"},
		{"    WORD $0x05403ecb // eor z11.b, z11.b, #254"},
		{"    WORD $0x05003ecb // orr z11.b, z11.b, #254"},
		{"    WORD $0x05c03ecb // dupm z11.b, #254"},
		{"    WORD $0x05c07dcc // dupm z12.h, #0xfffe"},
		{"    WORD $0x05c0fbcd // dupm z13.s, #0xfffffffe"},
		{"    WORD $0x05c3ffce // dupm z14.d, #0xfffffffffffffffe"},
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
		{"    WORD $0x047f32ab // orr z11.d, z21.d, z31.d"},
		{"    WORD $0x048c0ca4 // sabd z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04960ca4 // sdivr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x048a0ca4 // smin z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04920ca4 // smulh z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04810ca4 // sub z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04fe068a // sub z10.d, z20.d, z30.d"},
		{"    WORD $0x04830ca4 // subr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x048d0ca4 // uabd z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x05aadaaa // mov z10.s, p6/m,  z21.s"},
		{"    WORD $0x2491b8e5 // cmpeq p5.s, p6/z,  z7.s, z17.s"},
		{"    WORD $0x2494a55f // cmpne p15.s, p1/z, z10.s, z20.s"},
		{"    WORD $0x249218c5 // cmphs p5.s, p6/z,  z6.s, z18.s"},
		{"    WORD $0x249a1b61 // cmpls p1.s, p6/z, z26.s, z27.s"},
		{"    WORD $0x249a1b61 // cmphs p1.s, p6/z, z27.s, z26.s"},
		{"    WORD $0x249218d5 // cmphi p5.s, p6/z,  z6.s, z18.s"},
		{"    WORD $0x249a1b72 // cmplo p2.s, p6/z, z26.s, z27.s"},
		{"    WORD $0x249a1b72 // cmphi p2.s, p6/z, z27.s, z26.s"},
		{"    WORD $0x2495896e // cmpge p14.s, p2/z, z11.s, z21.s"},
		{"    WORD $0x248c8ecd // cmple p13.s, p3/z, z12.s, z22.s"},
		{"    WORD $0x248c8ecd // cmpge p13.s, p3/z, z22.s, z12.s"},
		{"    WORD $0x249791bc // cmpgt p12.s, p4/z, z13.s, z23.s"},
		{"    WORD $0x248e971b // cmplt p11.s, p5/z, z14.s, z24.s"},
		{"    WORD $0x248e971b // cmpgt p11.s, p5/z, z24.s, z14.s"},
		{"    WORD $0x2550d0a0 // ptest p4, p5.b"},
		{"    WORD $0x2598e084 // ptrue p4.s, VL4"},
		{"    WORD $0x2598e3e3 // ptrue p3.s"},
		{"    WORD $0x05b441ef // rev   p15.s, p15.s"},
		{"    WORD $0x258554a6 // mov   p6.b, p5.b"},
		{"    WORD $0x855c5482 // ld1w  { z2.s }, p5/z, [x4, z28.s, sxtw]"},
		{"    WORD $0x8540de9b // ld1rw { z27.s }, p7/z, [x20]"},
		{"    WORD $0x84cb594a // ld1h  { z10.s }, p6/z, [x10, z11.s, sxtw]"},
		{"    WORD $0x84155e94 // ld1b  { z20.s }, p7/z, [x20, z21.s, uxtw]"},
		{"    WORD $0xa01f87d8 // ld1b  { z24.b, z25.b, z26.b, z27.b }, p9/z, [x30, x31]"}, // consecutive
		{"    WORD $0xa472c66d // ld4b  { z13.b, z14.b, z15.b, z16.b }, p1/z, [x19, x18]"}, // interleaved
		{"    WORD $0x04018f06 // lsr   z6.h, p3/m, z6.h, #8"},
		{"    WORD $0x04419607 // lsr   z7.s, p5/m, z7.s, #16"},
		{"    WORD $0x04419b0b // lsr   z11.s, p6/m, z11.s, #8"},
		{"    WORD $0x046696d2 // lsr   z18.s, z22.s, #26"},
		{"    WORD $0x047096d2 // lsr   z18.s, z22.s, #16"},
		{"    WORD $0x046896d2 // lsr   z18.s, z22.s, #24"},
		{"    WORD $0x047896d2 // lsr   z18.s, z22.s, #8"},
		{"    WORD $0x0499bf6a // clz   z10.s, p7/m, z27.s"},
		{"    WORD $0xe5ef40c1 // st1d  { z1.d }, p0, [x6, x15, lsl #3]"},
		{"    WORD $0xe56f44d5 // st1w  { z21.d }, p1, [x6, x15, lsl #2]"},
		{"    WORD $0xe440fe8b // st1b  { z11.s }, p7, [x20]"},
		{"    WORD $0xe5800281 // str   p1, [x20]"},
		{"    WORD $0x45016802 // pmullb z2.q, z0.d, z1.d"},
		{"    WORD $0x45036c85 // pmullt z5.q, z4.d, z3.d"},
		{"    WORD $0x0424400c // index z12.b, #0, #4"},
	}

	for i, tc := range testCases {
		ins := strings.TrimSpace(strings.Split(tc.ins, "//")[1])
		oc, oc2, err := Assemble(ins)
		if err != nil {
			t.Errorf("TestSveAssembler(%d): `%s`: %v", i, ins, err)
		} else if oc2 != 0 {
			oc64 := uint64(oc2)<<32 | uint64(oc)
			opcode := fmt.Sprintf("0x%016x", oc64)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestSveAssembler(%d): `%s`: got: %s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 64)
				if err == nil {
					fmt.Printf("%064s\n", strconv.FormatUint(oc64, 2))
					fmt.Printf("%064s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		} else {
			opcode := fmt.Sprintf("0x%08x ", oc)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestSveAssembler(%d): `%s`: got: %s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
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
		{"    DWORD $0x0480046104902421 // add z1.s, p1/Z, z1.s, z3.s"}, /* /Z should always generate a prefix instruction, even in case of Zdn */
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
			t.Errorf("TestDWords(%d): `%s`: %v", i, ins, err)
		} else if oc2 != 0 {
			oc64 := uint64(oc2)<<32 | uint64(oc)
			opcode := fmt.Sprintf("0x%016x", oc64)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestDWords(%d): `%s`: got: %s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 64)
				if err == nil {
					fmt.Printf("%064s\n", strconv.FormatUint(oc64, 2))
					fmt.Printf("%064s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		} else {
			opcode := fmt.Sprintf("0x%08x ", oc)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestDWords(%d): `%s`: got: %s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 32)
				if err == nil {
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(oc), 2))
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		}
	}
}

func TestZeroing(t *testing.T) {
	testCases := []struct {
		ins string
	}{
		//
		{"    DWORD $0x0480046104902421 // add z1.s, p1/Z, z1.s, z3.s"}, /* /Z should always generate a prefix instruction, even in case of Zdn */
		{"    DWORD $0x04800461         // add z1.s, p1/M, z1.s, z3.s"},
		{"    DWORD $0x0480046104902441 // add z1.s, p1/Z, z2.s, z3.s"},
		{"    DWORD $0x0480046104912441 // add z1.s, p1/M, z2.s, z3.s"},
		//
		{"    DWORD $0x0480046104902421 // add z1.s, p1/z, z1.s, z3.s"}, /* /z should always generate a prefix instruction, even in case of Zdn */
		{"    DWORD $0x04800461         // add z1.s, p1/m, z1.s, z3.s"},
		{"    DWORD $0x0480046104902441 // add z1.s, p1/z, z2.s, z3.s"},
		{"    DWORD $0x0480046104912441 // add z1.s, p1/m, z2.s, z3.s"},
	}

	for i, tc := range testCases {
		ins := strings.TrimSpace(strings.Split(tc.ins, "//")[1])
		oc, oc2, err := Assemble(ins)
		if err != nil {
			t.Errorf("TestZeroing(%d): `%s`: %v", i, ins, err)
		} else if oc2 != 0 {
			oc64 := uint64(oc2)<<32 | uint64(oc)
			opcode := fmt.Sprintf("0x%016x", oc64)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestZeroing(%d): `%s`: got: %s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 64)
				if err == nil {
					fmt.Printf("%064s\n", strconv.FormatUint(oc64, 2))
					fmt.Printf("%064s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		} else {
			opcode := fmt.Sprintf("0x%08x ", oc)
			if !strings.Contains(tc.ins, opcode) {
				t.Errorf("TestZeroing(%d): `%s`: got: %s want: %s", i, ins, opcode, strings.Fields(tc.ins)[1][1:])
				ocWant, err := strconv.ParseUint(strings.Fields(tc.ins)[1][3:], 16, 32)
				if err == nil {
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(oc), 2))
					fmt.Printf("%032s\n", strconv.FormatUint(uint64(ocWant), 2))
				}
			}
		}
	}
}
