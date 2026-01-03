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
		{"    WORD $0x91008022 // add x2, x1, #0x20, lsl #0"},
		{"    WORD $0x91408022 // add x2, x1, #0x20, lsl #12"},
		{"    WORD $0x8b208022 // add x2, x1, w0, sxtb"},
		{"    WORD $0x8b208022 // add x2, x1, x0, sxtb"},
		{"    WORD $0x8b202822 // add x2, x1, w0, uxth #2"},
		{"    WORD $0x8b202822 // add x2, x1, x0, uxth #2"},
		{"    WORD $0xab220c20 // adds x0, x1, x2, uxtb #3"},
		{"    WORD $0xcb233041 // sub x1, x2, x3, uxth #4"},
		{"    WORD $0xeb245462 // subs x2, x3, x4, uxtw #5"},
		{"    WORD $0xf1000400 // subs x0, x0, #1"},
		{"    WORD $0xd101f050 // sub x16, x2, #124"},
		{"    WORD $0xcb050129 // sub x9, x9, x5"},
		{"    WORD $0xd346fc00 // lsr x0, x0, #6"},
		{"    WORD $0xd37ae400 // lsl x0, x0, #6"},
		{"    WORD $0x9ac22420 // lsrv x0, x1, x2"},
		{"    WORD $0x9ac22420 // lsr x0, x1, x2"},
		{"    WORD $0x9ac22020 // lslv x0, x1, x2"},
		{"    WORD $0x9ac22020 // lsl x0, x1, x2"},
		{"    WORD $0x9343fc41 // asr x1, x2, #3"},
		{"    WORD $0xf10008df // cmp x6, #2"},
		{"    WORD $0xf11348ff // cmp x7, #1234"},
		{"    WORD $0xf11348ff // cmp x7, #1234, lsl #0"},
		{"    WORD $0xf17ffd1f // cmp x8, #4095, lsl #12"},
		{"    WORD $0xf17ffd1f // cmp x8, #4095, lsl #12"},
		{"    WORD $0xb17ffd1f // cmn x8, #4095, lsl #12"},
		{"    WORD $0xeb1501bf // cmp x13, x21"},
		{"    WORD $0xeb1509bf // cmp x13, x21, lsl #2"},
		{"    WORD $0xab1501bf // cmn x13, x21"},
		{"    WORD $0xab5509bf // cmn x13, x21, lsr #2"},
		{"    WORD $0xea00001f // tst x0, x0"},
		{"    WORD $0xf24024df // tst x6, #0x3ff"},
		{"    WORD $0x04225022 // addvl x2, x2, #1"},
		{"    WORD $0x04bf5050 // rdvl x16, #2"},
		{"    WORD $0x9ac10800 // udiv x0, x0, x1"},
		{"    WORD $0xd29fffe1 // mov x1, #0xffff"},
		{"    WORD $0xd2bfffe1 // mov x1, #0xffff, lsl #16"},
		{"    WORD $0xd2dfffe1 // mov x1, #0xffff, lsl #32"},
		{"    WORD $0xd2ffffe1 // mov x1, #0xffff, lsl #48"},
		{"    WORD $0xd29fffe1 // mov x1, #0xffff"},
		{"    WORD $0xd2bfffe1 // mov x1, #0xffff0000"},
		{"    WORD $0xd2dfffe1 // mov x1, #0xffff00000000"},
		{"    WORD $0xd2ffffe1 // mov x1, #0xffff000000000000"},
		{"    WORD $0xf29fffe1 // movk x1, #0xffff"},
		{"    WORD $0xf2bfffe1 // movk x1, #0xffff, lsl #16"},
		{"    WORD $0xf2dfffe1 // movk x1, #0xffff, lsl #32"},
		{"    WORD $0xf2ffffe1 // movk x1, #0xffff, lsl #48"},
		{"    WORD $0xf29fffe1 // movk x1, #0xffff"},
		{"    WORD $0xf2bfffe1 // movk x1, #0xffff0000"},
		{"    WORD $0xf2dfffe1 // movk x1, #0xffff00000000"},
		{"    WORD $0xf2ffffe1 // movk x1, #0xffff000000000000"},
		{"    WORD $0x929fffe1 // movn x1, #0xffff"},
		{"    WORD $0x92bfffe1 // movn x1, #0xffff0000"},
		{"    WORD $0x92dfffe1 // movn x1, #0xffff00000000"},
		{"    WORD $0x92ffffe1 // movn x1, #0xffff000000000000"},
		{"    WORD $0x929fffe1 // mov x1, #0xffffffffffff0000"},
		{"    WORD $0x92bfffe1 // mov x1, #0xffffffff0000ffff"},
		{"    WORD $0x92dfffe1 // mov x1, #0xffff0000ffffffff"},
		{"    WORD $0x92ffffe1 // mov x1, #0xffffffffffff"},
		{"    WORD $0xaa0103ea // mov x10, x1"},
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
		{"    WORD $0xf807be8a // str x10, [x20, #123]!"},
		{"    WORD $0xf93ffe8a // str x10, [x20, #32760]"},
		{"    WORD $0xf8767a37 // ldr x23, [x17, x22, lsl #3]"},
		{"    WORD $0xf82d7894 // str x20, [x4, x13, lsl #3]"},
		{"    WORD $0xf90007e2 // str x2, [sp, #8]"},
		{"    WORD $0xf940fdd4 // ldr x20, [x14, #0x1f8]"},
		{"    WORD $0xf8408480 // ldr x0, [x4], #8"},
		{"    WORD $0xf8408c80 // ldr x0, [x4, #8]!"},
		{"    WORD $0xf97ffc80 // ldr x0, [x4, #32760]"},
		{"    WORD $0x384014a1 // ldrb x1, [x5], #1"},
		{"    WORD $0x38401ca1 // ldrb x1, [x5, #1]!"},
		{"    WORD $0x397ffca1 // ldrb x1, [x5, #4095]"},
		{"    WORD $0x784024c2 // ldrh x2, [x6], #2"},
		{"    WORD $0x784ffce3 // ldrh x3, [x7, #255]!"},
		{"    WORD $0x797ffd04 // ldrh x4, [x8, #8190]"},
		{"    WORD $0x92400d6a // and x10, x11, #0xf"},
		{"    WORD $0xf2401d6a // ands x10, x11, #0xff"},
		{"    WORD $0xd2402d6a // eor x10, x11, #0xfff"},
		{"    WORD $0xb2403d6a // orr x10, x11, #0xffff"},
		{"    WORD $0x92404d6a // and x10, x11, #0xfffff"},
		{"    WORD $0x92405d6a // and x10, x11, #0xffffff"},
		{"    WORD $0x92406d6a // and x10, x11, #0xfffffff"},
		{"    WORD $0x92407d6a // and x10, x11, #0xffffffff"},
		{"    WORD $0x92408d6a // and x10, x11, #0xfffffffff"},
		{"    WORD $0x92409d6a // and x10, x11, #0xffffffffff"},
		{"    WORD $0x9240ad6a // and x10, x11, #0xfffffffffff"},
		{"    WORD $0x9240bd6a // and x10, x11, #0xffffffffffff"},
		{"    WORD $0x9240cd6a // and x10, x11, #0xfffffffffffff"},
		{"    WORD $0x9240dd6a // and x10, x11, #0xffffffffffffff"},
		{"    WORD $0x9240ed6a // and x10, x11, #0xfffffffffffffff"},
		{"    WORD $0x9240fd6a // and x10, x11, #0xffffffffffffffff"},
		{"    WORD $0x8a031041 // and x1, x2, x3, lsl #4"},
		{"    WORD $0xea431041 // ands x1, x2, x3, lsr #4"},
		{"    WORD $0xca831041 // eor x1, x2, x3, asr #4"},
		{"    WORD $0xca231041 // eon x1, x2, x3, lsl #4"},
		{"    WORD $0xaa031041 // orr x1, x2, x3, lsl #4"},
		{"    WORD $0xaa231041 // orn x1, x2, x3, lsl #4"},
		{"    WORD $0xaa2313e1 // mvn x1, x3, lsl #4"},
		{"    WORD $0xaa060be9 // orr x9, xzr, x6, lsl #2"},
		{"    WORD $0xaa0603e9 // orr x9, xzr, x6, lsl #0"},
		{"    WORD $0xaa0603e9 // orr x9, xzr, x6"},
		{"    WORD $0xaa0603e9 // mov x9, x6"}, // alias of above
		{"    WORD $0xf2400b8e // ands x14, x28, #7"},
		{"    WORD $0x9a830041 // csel x1, x2, x3, eq"}, // eq = none
		{"    WORD $0x9a831041 // csel x1, x2, x3, ne"}, // ne = any
		{"    WORD $0x9a832041 // csel x1, x2, x3, cs"}, // cs = hs, nlast
		{"    WORD $0x9a833041 // csel x1, x2, x3, cc"}, // cc = lo, ul, last
		{"    WORD $0x9a834041 // csel x1, x2, x3, mi"}, // mi = first
		{"    WORD $0x9a835041 // csel x1, x2, x3, pl"}, // pl = nfrst
		{"    WORD $0x9a836041 // csel x1, x2, x3, vs"},
		{"    WORD $0x9a837041 // csel x1, x2, x3, vc"},
		{"    WORD $0x9a838041 // csel x1, x2, x3, hi"}, // hi = pmore
		{"    WORD $0x9a839041 // csel x1, x2, x3, ls"}, // ls = plast
		{"    WORD $0x9a83a041 // csel x1, x2, x3, ge"}, // ge = tcont
		{"    WORD $0x9a83b041 // csel x1, x2, x3, lt"}, // lt = tstop
		{"    WORD $0x9a83c041 // csel x1, x2, x3, gt"},
		{"    WORD $0x9a83d041 // csel x1, x2, x3, le"},
		{"    WORD $0x9a83e041 // csel x1, x2, x3, al"},
		{"    WORD $0x9a83f041 // csel x1, x2, x3, nv"},
		{"    WORD $0x9a8c056a // csinc x10, x11, x12, eq"},
		{"    WORD $0x9a8c156a // csinc x10, x11, x12, ne"},
		{"    WORD $0x9a8b156a // cinc x10, x11, eq"}, // eq = none
		{"    WORD $0x9a8b056a // cinc x10, x11, ne"}, // ne = any
		{"    WORD $0x9a9f27f4 // cset x20, cc"},
		{"    WORD $0x9a9f37f4 // cset x20, cs"},
		{"    WORD $0xda830441 // csneg x1, x2, x3, eq"},
		{"    WORD $0xda831441 // csneg x1, x2, x3, ne"},
		{"    WORD $0xda821441 // cneg x1, x2, eq"},
		{"    WORD $0xda820441 // cneg x1, x2, ne"},
		{"    WORD $0xda833041 // csinv x1, x2, x3, cc"},
		{"    WORD $0xda822041 // cinv x1, x2, cc"},
		{"    WORD $0xda9f23e1 // csetm x1, cc"},
		{"    WORD $0x9b020c20 // madd x0, x1, x2, x3"},
		{"    WORD $0x9b027c20 // mul x0, x1, x2"},
		{"    WORD $0x9ac20c20 // sdiv x0, x1, x2"},
		{"    WORD $0x9b0cb56a // msub x10, x11, x12, x13"},
		{"    WORD $0x9b0cfd6a // mneg x10, x11, x12"},
		{"    WORD $0xcb0203e1 // neg x1, x2"},
		{"    WORD $0xcb820fe1 // neg x1, x2, asr #3"},
		{"    WORD $0xeb4c7fea // negs x10, x12, lsr #31"},
		{"    WORD $0xda0d03ec // ngc x12, x13"},
		{"    WORD $0xfa0f03ee // ngcs x14, x15"},
		{"    WORD $0xdac02041 // abs x1, x2"},
		{"    WORD $0xdac01462 // cls x2, x3"},
		{"    WORD $0xdac01083 // clz x3, x4"},
		{"    WORD $0xdac018a4 // ctz x4, x5"},
		{"    WORD $0xdac01cc5 // cnt x5, x6"},
		{"    WORD $0xdac000e6 // rbit x6, x7"},
		{"    WORD $0xdac00d07 // rev x7, x8"},
		{"    WORD $0xdac00528 // rev16 x8, x9"},
		{"    WORD $0xdac00949 // rev32 x9, x10"},
		{"    WORD $0xdac00d6a // rev64 x10, x11"},
		{"    WORD $0x9a030041 // adc x1, x2, x3"},
		{"    WORD $0xba0600a4 // adcs x4, x5, x6"},
		{"    WORD $0xda090107 // sbc x7, x8, x9"},
		{"    WORD $0xfa0c016a // sbcs x10, x11, x12"},
		{"    WORD $0x8a620c20 // bic x0, x1, x2, lsr #3"},
		{"    WORD $0xea6c356a // bics x10, x11, x12, lsr #13"},
		{"    WORD $0x93c20c20 // extr x0, x1, x2, #3"},
		{"    WORD $0x93d6feb4 // extr x20, x21, x22, #63"},
		{"    WORD $0x93cbfd6a // ror x10, x11, #63"},
		{"    WORD $0x9ac42c62 // ror x2, x3, x4"},
		{"    WORD $0x9ac42c62 // rorv x2, x3, x4"},
		{"    WORD $0x93410145 // sbfiz x5, x10, #63, #1"},
		{"    WORD $0x93492145 // sbfiz x5, x10, #55, #9"},
		{"    WORD $0x934a2545 // sbfiz x5, x10, #54, #10"},
		{"    WORD $0x937ff945 // sbfiz x5, x10, #1, #63"},
		{"    WORD $0x93400145 // sbfx x5, x10, #0, #1"},
		{"    WORD $0x9340f945 // sbfx x5, x10, #0, #63"},
		{"    WORD $0x934a2945 // sbfx x5, x10, #10, #1"},
		{"    WORD $0x934a2d45 // sbfx x5, x10, #10, #2"},
		{"    WORD $0x934aed45 // sbfx x5, x10, #10, #50"},
		{"    WORD $0x934af945 // sbfx x5, x10, #10, #53"},
		{"    WORD $0x937ef945 // sbfx x5, x10, #62, #1"},
		{"    WORD $0x934afd45 // asr x5, x10, #10"},
		{"    WORD $0x937efd45 // asr x5, x10, #62"},
		{"    WORD $0x937ffd45 // asr x5, x10, #63"},
		{"    WORD $0x93401d45 // sxtb x5, x10"},
		{"    WORD $0x93403d45 // sxth x5, x10"},
		{"    WORD $0x93407d45 // sxtw x5, x10"},
		{"    WORD $0xd3410145 // lsl   x5, x10, #63"},
		{"    WORD $0xd3410145 // ubfiz x5, x10, #63, #1"},
		{"    WORD $0xd37ff945 // lsl   x5, x10, #1"},
		{"    WORD $0xd37ff945 // ubfiz x5, x10, #1, #63"},
		{"    WORD $0xd34a2545 // lsl   x5, x10, #54"},
		{"    WORD $0xd34a2545 // ubfiz x5, x10, #54, #10"},
		{"    WORD $0xd34a0145 // ubfiz x5, x10, #54, #1"},
		{"    WORD $0xd34a0545 // ubfiz x5, x10, #54, #2"},
		{"    WORD $0xd3400145 // ubfx x5, x10, #0, #1"},
		{"    WORD $0xd340f945 // ubfx x5, x10, #0, #63"},
		{"    WORD $0x53001d45 // uxtb x5, x10"},
		{"    WORD $0x53003d45 // uxth x5, x10"},
		{"    WORD $0xb3410545 // bfxil x5, x10, #1, #1"},
		{"    WORD $0xb3415145 // bfxil x5, x10, #1, #20"},
		{"    WORD $0xb341fd45 // bfxil x5, x10, #1, #63"},
		{"    WORD $0xb376d945 // bfxil x5, x10, #54, #1"},
		{"    WORD $0xb376fd45 // bfxil x5, x10, #54, #10"},
		{"    WORD $0xb34a03e5 // bfc x5, #54, #1"},
		{"    WORD $0xb37eefe5 // bfc x5, #2, #60"},
		{"    WORD $0xb37eed45 // bfi x5, x10, #2, #60"},
		{"    WORD $0xd4001001 // svc #0x80"},
		{"    WORD $0xc8a07c01 // cas    x0, x1, [x0]"},
		{"    WORD $0xc8e27c23 // casa   x2, x3, [x1]"},
		{"    WORD $0xc8e4fc45 // casal  x4, x5, [x2]"},
		{"    WORD $0xc8a6fc67 // casl   x6, x7, [x3]"},
		{"    WORD $0x08a87c09 // casb   x8, x9, [x0]"},
		{"    WORD $0x08ea7c2b // casab  x10, x11, [x1]"},
		{"    WORD $0x08ecfc4d // casalb x12, x13, [x2]"},
		{"    WORD $0x08aefc6f // caslb  x14, x15, [x3]"},
		{"    WORD $0x48b07c11 // cash   x16, x17, [x0]"},
		{"    WORD $0x48f27c33 // casah  x18, x19, [x1]"},
		{"    WORD $0x48f4fc55 // casalh x20, x21, [x2]"},
		{"    WORD $0x48b6fc77 // caslh  x22, x23, [x3]"},
		{"    WORD $0x482a7c14 // casp   x10, x11, x20, x21, [x0]"},
		{"    WORD $0x486c7c36 // caspa  x12, x13, x22, x23, [x1]"},
		{"    WORD $0x486efc58 // caspal x14, x15, x24, x25, [x2]"},
		{"    WORD $0x4830fc7a // caspl  x16, x17, x26, x27, [x3]"},
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
		{"    WORD $0x046c6027 // mul z7.h, z1.h, z12.h"},
		{"    WORD $0x0490054b // mul z11.s, p1/M, z11.s, z10.s"},
		{"    WORD $0x0450058d // mul z13.h, p1/M, z13.h, z12.h"},
		{"    WORD $0x05253065 // tbl z5.b, z3.b, z5.b"},
		{"    WORD $0x05283086 // tbl z6.b, z4.b, z8.b"},
		{"    WORD $0x052b2927 // tbl z7.b, { z9.b, z10.b }, z11.b"},
		{"    WORD $0x04a33080 // eor z0.d, z4.d, z3.d"},
		{"    WORD $0x05212042 // dup z2.b, z2.b[0]"},
		{"    WORD $0x05a03883 // dup z3.s, w4"},
		{"    WORD $0x05a03883 // mov z3.s, w4"},
		{"    WORD $0x05e038c5 // dup z5.d, x6"},
		{"    WORD $0x05e038c5 // mov z5.d, x6"},
		{"    WORD $0x04fc94c7 // lsr z7.d, z6.d, #4"},
		{"    WORD $0x04fc94a8 // lsr z8.d, z5.d, #4"},
		{"    WORD $0x04233880 // eor3 z0.d, z0.d, z3.d, z4.d"},
		{"    WORD $0x04ccc5ab // mad z11.d, p1/m, z12.d, z13.d"},
		{"    WORD $0x04cd658b // mls z11.d, p1/m, z12.d, z13.d"},
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
		{"    WORD $0x04a19210 // asr z16.d, z16.d, #0x3f"},
		{"    WORD $0x04ff9210 // asr z16.d, z16.d, #0x1"},
		{"    WORD $0x04619231 // asr z17.s, z17.s, #0x1f"},
		{"    WORD $0x047f9231 // asr z17.s, z17.s, #0x1"},
		{"    WORD $0x04319252 // asr z18.h, z18.h, #0xf"},
		{"    WORD $0x043f9252 // asr z18.h, z18.h, #0x1"},
		{"    WORD $0x04299273 // asr z19.b, z19.b, #7"},
		{"    WORD $0x042f9273 // asr z19.b, z19.b, #1"},
		{"    WORD $0x6594a231 // scvtf z17.s, p0/m, z17.s"},
		{"    WORD $0x65b2023f // fmla z31.s, p0/M, z17.s, z18.s"},
		{"    WORD $0x2538c1e0 // dup z0.b, #15"},
		{"    WORD $0x2538c1e0 // mov z0.b, #15"},
		{"    WORD $0x2538de20 // dup z0.b, #-15"},
		{"    WORD $0x2538de20 // mov z0.b, #-15"},
		{"    WORD $0x25b8f016 // dup z22.s, #-32768"},
		{"    WORD $0x25b8f016 // mov z22.s, #-32768"},
		{"    WORD $0x25b8eff6 // dup z22.s, #0x7f00"},
		{"    WORD $0x25b8eff6 // mov z22.s, #0x7f00"},
		{"    WORD $0x25b8e036 // dup z22.s, #256"},
		{"    WORD $0x25b8e036 // mov z22.s, #256"},
		{"    WORD $0x25b8fff6 // dup z22.s, #-256"},
		{"    WORD $0x25b8fff6 // mov z22.s, #-256"},
		{"    WORD $0x25b8eff6 // dup z22.s, #32512"},
		{"    WORD $0x25b8eff6 // mov z22.s, #32512"},
		{"    WORD $0x2578c0e1 // dup z1.h, #7"},
		{"    WORD $0x2578c0e1 // mov z1.h, #7"},
		{"    WORD $0x25b8c0a2 // dup z2.s, #5"},
		{"    WORD $0x25b8c0a2 // mov z2.s, #5"},
		{"    WORD $0x25f8c163 // dup z3.d, #11"},
		{"    WORD $0x25f8c163 // mov z3.d, #11"},
		{"    WORD $0x2538c816 // dup z22.b, #64"},
		{"    WORD $0x2538c816 // mov z22.b, #64"},
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
		{"    WORD $0x04a19dcd // lsl z13.d, z14.d, #1"},
		{"    WORD $0x04a89dcd // lsl z13.d, z14.d, #8"},
		{"    WORD $0x04ff9dcd // lsl z13.d, z14.d, #63"},
		{"    WORD $0x04619dcd // lsl z13.s, z14.s, #1"},
		{"    WORD $0x04689dcd // lsl z13.s, z14.s, #8"},
		{"    WORD $0x047f9dcd // lsl z13.s, z14.s, #31"},
		{"    WORD $0x04319dcd // lsl z13.h, z14.h, #1"},
		{"    WORD $0x04389dcd // lsl z13.h, z14.h, #8"},
		{"    WORD $0x043f9dcd // lsl z13.h, z14.h, #15"},
		{"    WORD $0x04299dcd // lsl z13.b, z14.b, #1"},
		{"    WORD $0x042f9dcd // lsl z13.b, z14.b, #7"},
		{"    WORD $0x04938ca4 // lsl z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04978ca4 // lslr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04918ca4 // lsr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04958ca4 // lsrr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04980ca4 // orr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x047f32ab // orr z11.d, z21.d, z31.d"},
		{"    WORD $0x048c0ca4 // sabd z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04940f38 // sdiv  z24.s, p3/M, z24.s, z25.s"},
		{"    WORD $0x04960ca4 // sdivr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x048a0ca4 // smin z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04920ca4 // smulh z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04810ca4 // sub z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x04fe068a // sub z10.d, z20.d, z30.d"},
		{"    WORD $0x25a1d11d // sub z29.s, z29.s, #136"},
		{"    WORD $0x04830ca4 // subr z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x048d0ca4 // uabd z4.s, p3/M, z4.s, z5.s"},
		{"    WORD $0x05aadaaa // mov z10.s, p6/m,  z21.s"},
		{"    WORD $0x258898e5 // cmpeq p5.s, p6/z, z7.s, #8"},
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
		{"    WORD $0x258e11bc // cmpgt p12.s, p4/z, z13.s, #14"},
		{"    WORD $0x249791bc // cmpgt p12.s, p4/z, z13.s, z23.s"},
		{"    WORD $0x248e971b // cmplt p11.s, p5/z, z14.s, z24.s"},
		{"    WORD $0x248e971b // cmpgt p11.s, p5/z, z24.s, z14.s"},
		{"    WORD $0x453a9b2c // match p12.b, p6/z, z25.b, z26.b"},
		{"    WORD $0x457c9f7d // nmatch p13.h, p7/z, z27.h, z28.h"},
		{"    WORD $0x4523a041 // histseg z1.b, z2.b, z3.b"},
		{"    WORD $0x45acc96a // histcnt z10.s, p2/z, z11.s, z12.s"},
		{"    WORD $0x45edcd8b // histcnt z11.d, p3/z, z12.d, z13.d"},
		{"    WORD $0x2550d0a0 // ptest p4, p5.b"},
		{"    WORD $0x2598e084 // ptrue p4.s, VL4"},
		{"    WORD $0x2598e3e3 // ptrue p3.s"},
		{"    WORD $0x05b441ef // rev   p15.s, p15.s"},
		{"    WORD $0x258554a6 // mov   p6.b, p5.b"},
		{"    WORD $0x855c5482 // ld1w  { z2.s }, p5/z, [x4, z28.s, sxtw]"},
		{"    WORD $0xa54b4450 // ld1w  { z16.s }, p1/z, [x2, x11, lsl #2]"},
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
		{"    WORD $0xe55d8447 // st1w  { z7.s }, p1, [x2, z29.s, uxtw]"},
		{"    WORD $0xe4dd8547 // st1h  { z7.s }, p1, [x10, z29.s, uxtw]"},
		{"    WORD $0xe45d8547 // st1b  { z7.s }, p1, [x10, z29.s, uxtw]"},
		{"    WORD $0xe440fe8b // st1b  { z11.s }, p7, [x20]"},
		{"    WORD $0xe5800281 // str   p1, [x20]"},
		{"    WORD $0x45016802 // pmullb z2.q, z0.d, z1.d"},
		{"    WORD $0x45036c85 // pmullt z5.q, z4.d, z3.d"},
		{"    WORD $0x0424400c // index z12.b, #0, #4"},
		{"    WORD $0x042f420d // index z13.b, #-16, #15"},
		{"    WORD $0x043041ee // index z14.b, #15, #-16"},
		{"    WORD $0x0477496f // index z15.h, #11, w23"},
		{"    WORD $0x04b849f0 // index z16.s, #15, w24"},
		{"    WORD $0x04f94a11 // index z17.d, #-16, x25"},
		{"    WORD $0x042f4752 // index z18.b, w26, #15"},
		{"    WORD $0x04704773 // index z19.h, w27, #-16"},
		{"    WORD $0x04b14794 // index z20.s, w28, #-15"},
		{"    WORD $0x04ee47b5 // index z21.d, x29, #14"},
		{"    WORD $0x05a43820 // insr z0.s, w1"},
		{"    WORD $0x05e43841 // insr z1.d, x2"},
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
		{"    DWORD $0x0494046104912441 // sdiv  z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x04d6046104d12441 // sdivr z1.d, p1/M, z2.d, z3.d"},
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
