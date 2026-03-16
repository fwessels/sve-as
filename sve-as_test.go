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
		{"    WORD $0x53047ef7 // lsr w23, w23, #4"},
		{"    WORD $0x5ac00b39 // rev w25, w25"},
		{"    WORD $0x0b0f01ce // add w14, w14, w15"},
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
		{"    WORD $0xd61f01e0 // br x15"},
		{"    WORD $0xd63f0020 // blr x1"},
		{"    WORD $0x79400801 // ldrh w1, [x0, #4]"},
		{"    WORD $0x79400003 // ldrh w3, [x0]"},
		{"    WORD $0x50002682 // adr x2, #1234"},
		{"    WORD $0x10000801 // adr x1, #256"},
		{"    WORD $0x10ffe002 // adr x2, #-1024"},
		{"    WORD $0x10ffefe3 // adr x3, #-516"},
		{"    WORD $0xf9000041 // str x1, [x2]"},
		{"    WORD $0xf807be8a // str x10, [x20, #123]!"},
		{"    WORD $0xf93ffe8a // str x10, [x20, #32760]"},
		{"    WORD $0xf87b690c // ldr  x12, [x8, x27]"},
		{"    WORD $0xf8767a37 // ldr x23, [x17, x22, lsl #3]"},
		{"    WORD $0xf82d7894 // str x20, [x4, x13, lsl #3]"},
		{"    WORD $0xf90007e2 // str x2, [sp, #8]"},
		{"    WORD $0xf81f83fd // stur x29, [sp, #-8]"},
		{"    WORD $0x7804b3ef // sturh w15, [sp, #75]"},
		{"    WORD $0xf85f83fd // ldur x29, [sp, #-8]"},
		{"    WORD $0x785fe087 // ldurh w7, [x4, #-2]"},
		{"    WORD $0x39000401 // strb x1, [x0, #1]"},
		{"    WORD $0x39000801 // strb x1, [x0, #2]"},
		{"    WORD $0x79000401 // strh x1, [x0, #2]"},
		{"    WORD $0x79000801 // strh x1, [x0, #4]"},
		{"    WORD $0xb9000401 // strw x1, [x0, #4]"},
		{"    WORD $0xb9000801 // strw x1, [x0, #8]"},
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
		{"    WORD $0xb847b56a // ldrw x10, [x11], #123"},   // additional 'word' variant
		{"    WORD $0xb84ffd8b // ldrw x11, [x12, #255]!"},  // additional 'word' variant
		{"    WORD $0xb97ffdac // ldrw x12, [x13, #16380]"}, // additional 'word' variant
		{"    WORD $0xb89fcc6b // ldrsw	x11, [x3, #-4]!"},
		{"    WORD $0x789fec6b // ldrsh x11, [x3, #-2]!"},
		{"    WORD $0x389ffc6b // ldrsb x11, [x3, #-1]!"},
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
		//                        and x10, x11, #0xffffffffffffffff -- illegal (all-ones is not encodable as ARM logical immediate)
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
		{"    WORD $0xa9050ce2 // stp x2, x3, [x7, #80]"},
		{"    WORD $0xa94614e4 // ldp x4, x5, [x7, #96]"},
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
		{"    WORD $0x9e66016a // fmov x10, d11"},
		//
		// vector instructions
		{"    WORD $0x04723240 // mov z0.d, z18.d"},
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
		{"    WORD $0x4451ad6a // addp z10.h, p3/m, z10.h, z11.h"},
		{"    WORD $0x4491a441 // addp z1.s, p1/m, z1.s, z2.s"},
		{"    WORD $0x44d1bc1f // addp z31.d, p7/m, z31.d, z0.d"},
		{"    WORD $0x046c6027 // mul z7.h, z1.h, z12.h"},
		{"    WORD $0x0490054b // mul z11.s, p1/M, z11.s, z10.s"},
		{"    WORD $0x0450058d // mul z13.h, p1/M, z13.h, z12.h"},
		{"    WORD $0x05253065 // tbl z5.b, z3.b, z5.b"},
		{"    WORD $0x05283086 // tbl z6.b, z4.b, z8.b"},
		{"    WORD $0x052b2927 // tbl z7.b, { z9.b, z10.b }, z11.b"},
		{"    WORD $0x05252c65 // tbx z5.b, z3.b, z5.b"},
		{"    WORD $0x05282c86 // tbx z6.b, z4.b, z8.b"},
		{"    WORD $0x04a33080 // eor z0.d, z4.d, z3.d"},
		{"    WORD $0x05212042 // dup z2.b, z2.b[0]"},
		{"    WORD $0x053820c6 // dup z6.d, z6.d[1]"},
		{"    WORD $0x053820c6 // mov z6.d, z6.d[1]"},
		{"    WORD $0x05f820c6 // dup z6.d, z6.d[7]"},
		{"    WORD $0x05a03883 // dup z3.s, w4"},
		{"    WORD $0x05a03883 // mov z3.s, w4"},
		{"    WORD $0x05e038c5 // dup z5.d, x6"},
		{"    WORD $0x05e038c5 // mov z5.d, x6"},
		{"    WORD $0x04fc94c7 // lsr z7.d, z6.d, #4"},
		{"    WORD $0x04fc94a8 // lsr z8.d, z5.d, #4"},
		{"    WORD $0x04233880 // eor3 z0.d, z0.d, z3.d, z4.d"},
		{"    WORD $0x04633880 // bcax z0.d, z0.d, z3.d, z4.d"},
		{"    WORD $0x0420e3e0 // cntb x0"},
		{"    WORD $0x0460e3e0 // cnth x0"},
		{"    WORD $0x04a0e3e0 // cntw x0"},
		{"    WORD $0x04e0e3e0 // cntd x0"},
		{"    WORD $0x0420e000 // cntb x0, pow2"},
		{"    WORD $0x0420e3e0 // cntb x0, all"},
		{"    WORD $0x0422e3e0 // cntb x0, all, mul #3"},
		{"    WORD $0x0470c3e0 // inch z0.h"},
		{"    WORD $0x04b0c3e0 // incw z0.s"},
		{"    WORD $0x04f0c3e0 // incd z0.d"},
		{"    WORD $0x0470c000 // inch z0.h, pow2"},
		{"    WORD $0x0472c3e0 // inch z0.h, all, mul #3"},
		{"    WORD $0x0430e3e0 // incb x0"},
		{"    WORD $0x0470e3e0 // inch x0"},
		{"    WORD $0x04b0e3e0 // incw x0"},
		{"    WORD $0x04f0e3e0 // incd x0"},
		{"    WORD $0x0430e000 // incb x0, pow2"},
		{"    WORD $0x0432e3e0 // incb x0, all, mul #3"},
		{"    WORD $0x0430f3e0 // decb x0"},
		{"    WORD $0x0470f3e0 // dech x0"},
		{"    WORD $0x04b0f3e0 // decw x0"},
		{"    WORD $0x04f0f3e0 // decd x0"},
		{"    WORD $0x0470d3e0 // dech z0.h"},
		{"    WORD $0x04b0d3e0 // decw z0.s"},
		{"    WORD $0x04f0d3e0 // decd z0.d"},
		{"    WORD $0x0430f000 // decb x0, pow2"},
		{"    WORD $0x0432f3e0 // decb x0, all, mul #3"},
		{"    WORD $0x252c8800 // incp x0, p0.b"},
		{"    WORD $0x256c8841 // incp x1, p2.h"},
		{"    WORD $0x25ac8800 // incp x0, p0.s"},
		{"    WORD $0x25ec8800 // incp x0, p0.d"},
		{"    WORD $0x252d8800 // decp x0, p0.b"},
		{"    WORD $0x256d8800 // decp x0, p0.h"},
		{"    WORD $0x25ad8800 // decp x0, p0.s"},
		{"    WORD $0x25ed8800 // decp x0, p0.d"},
		{"    WORD $0x256c8000 // incp z0.h, p0.h"},
		{"    WORD $0x25ac8000 // incp z0.s, p0.s"},
		{"    WORD $0x25ec8000 // incp z0.d, p0.d"},
		{"    WORD $0x256d8000 // decp z0.h, p0.h"},
		{"    WORD $0x25ad8000 // decp z0.s, p0.s"},
		{"    WORD $0x25ed8000 // decp z0.d, p0.d"},
		{"    WORD $0x4503b441 // bdep z1.b, z2.b, z3.b"},
		{"    WORD $0x4549b507 // bdep z7.h, z8.h, z9.h"},
		{"    WORD $0x4586b4a4 // bdep z4.s, z5.s, z6.s"},
		{"    WORD $0x45c3b441 // bdep z1.d, z2.d, z3.d"},
		{"    WORD $0x4503b041 // bext z1.b, z2.b, z3.b"},
		{"    WORD $0x4549b107 // bext z7.h, z8.h, z9.h"},
		{"    WORD $0x4586b0a4 // bext z4.s, z5.s, z6.s"},
		{"    WORD $0x45c3b041 // bext z1.d, z2.d, z3.d"},
		{"    WORD $0x4503b841 // bgrp z1.b, z2.b, z3.b"},
		{"    WORD $0x4549b907 // bgrp z7.h, z8.h, z9.h"},
		{"    WORD $0x4586b8a4 // bgrp z4.s, z5.s, z6.s"},
		{"    WORD $0x45c3b841 // bgrp z1.d, z2.d, z3.d"},
		{"    WORD $0x042b3549 // xar z9.b, z9.b, z10.b, #5"},
		{"    WORD $0x043d34c5 // xar z5.h, z5.h, z6.h, #3"},
		{"    WORD $0x04713483 // xar z3.s, z3.s, z4.s, #15"},
		{"    WORD $0x04f93441 // xar z1.d, z1.d, z2.d, #7"},
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
		{"    WORD $0x449a06dc // udot z28.s, z22.b, z26.b"},
		{"    WORD $0x44830441 // udot z1.s, z2.b, z3.b"},
		{"    WORD $0x449f07ff // udot z31.s, z31.b, z31.b"},
		{"    WORD $0x44cc056a // udot z10.d, z11.h, z12.h"},
		{"    WORD $0x04480d6a // smax z10.h, p3/m, z10.h, z11.h"},
		{"    WORD $0x04880441 // smax z1.s, p1/m, z1.s, z2.s"},
		{"    WORD $0x04c81c1f // smax z31.d, p7/m, z31.d, z0.d"},
		{"    WORD $0x04490d6a // umax z10.h, p3/m, z10.h, z11.h"},
		{"    WORD $0x04890441 // umax z1.s, p1/m, z1.s, z2.s"},
		{"    WORD $0x04c91c1f // umax z31.d, p7/m, z31.d, z0.d"},
		{"    WORD $0x044b0d6a // umin z10.h, p3/m, z10.h, z11.h"},
		{"    WORD $0x048b0441 // umin z1.s, p1/m, z1.s, z2.s"},
		{"    WORD $0x04cb1c1f // umin z31.d, p7/m, z31.d, z0.d"},
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
		{"    WORD $0x04c08260 // asr z0.d, p0/m, z0.d, #13"},
		{"    WORD $0x04408320 // asr z0.s, p0/m, z0.s, #7"},
		{"    WORD $0x040083a0 // asr z0.h, p0/m, z0.h, #3"},
		{"    WORD $0x040081e0 // asr z0.b, p0/m, z0.b, #1"},
		{"    WORD $0x0480942a // asr z10.d, p5/m, z10.d, #63"},
		{"    WORD $0x6594a231 // scvtf z17.s, p0/m, z17.s"},
		{"    WORD $0x65d6a020 // scvtf z0.d, p0/m, z1.d"},
		{"    WORD $0x65d4a020 // scvtf z0.s, p0/m, z1.d"},
		{"    WORD $0x65d0a020 // scvtf z0.d, p0/m, z1.s"},
		{"    WORD $0x65d7a020 // ucvtf z0.d, p0/m, z1.d"},
		{"    WORD $0x6595a020 // ucvtf z0.s, p0/m, z1.s"},
		{"    WORD $0x65d5a020 // ucvtf z0.s, p0/m, z1.d"},
		{"    WORD $0x65d1a020 // ucvtf z0.d, p0/m, z1.s"},
		{"    WORD $0x65d7ae8f // ucvtf z15.d, p3/m, z20.d"},
		{"    WORD $0x65b2023f // fmla z31.s, p0/M, z17.s, z18.s"},
		{"    WORD $0x65e22020 // fmls z0.d, p0/m, z1.d, z2.d"},
		{"    WORD $0x65a22020 // fmls z0.s, p0/m, z1.s, z2.s"},
		{"    WORD $0x65622020 // fmls z0.h, p0/m, z1.h, z2.h"},
		{"    WORD $0x65f42dea // fmls z10.d, p3/m, z15.d, z20.d"},
		{"    WORD $0x65e24020 // fnmla z0.d, p0/m, z1.d, z2.d"},
		{"    WORD $0x65a24020 // fnmla z0.s, p0/m, z1.s, z2.s"},
		{"    WORD $0x65a55c1f // fnmla z31.s, p7/m, z0.s, z5.s"},
		{"    WORD $0x65c20020 // fadd z0.d, z1.d, z2.d"},
		{"    WORD $0x65820020 // fadd z0.s, z1.s, z2.s"},
		{"    WORD $0x65c08020 // fadd z0.d, p0/m, z0.d, z1.d"},
		{"    WORD $0x65808020 // fadd z0.s, p0/m, z0.s, z1.s"},
		{"    WORD $0x65408020 // fadd z0.h, p0/m, z0.h, z1.h"},
		{"    WORD $0x65c20420 // fsub z0.d, z1.d, z2.d"},
		{"    WORD $0x65820420 // fsub z0.s, z1.s, z2.s"},
		{"    WORD $0x65c18020 // fsub z0.d, p0/m, z0.d, z1.d"},
		{"    WORD $0x65818020 // fsub z0.s, p0/m, z0.s, z1.s"},
		{"    WORD $0x65cd8020 // fdiv z0.d, p0/m, z0.d, z1.d"},
		{"    WORD $0x658d8020 // fdiv z0.s, p0/m, z0.s, z1.s"},
		{"    WORD $0x65c68020 // fmax z0.d, p0/m, z0.d, z1.d"},
		{"    WORD $0x65868020 // fmax z0.s, p0/m, z0.s, z1.s"},
		{"    WORD $0x65c78020 // fmin z0.d, p0/m, z0.d, z1.d"},
		{"    WORD $0x65878020 // fmin z0.s, p0/m, z0.s, z1.s"},
		{"    WORD $0x65c28020 // fmul z0.d, p0/m, z0.d, z1.d"},
		{"    WORD $0x65828020 // fmul z0.s, p0/m, z0.s, z1.s"},
		// floating-point unary (predicated)
		{"    WORD $0x04dca020 // fabs z0.d, p0/m, z1.d"},
		{"    WORD $0x049ca020 // fabs z0.s, p0/m, z1.s"},
		{"    WORD $0x045ca020 // fabs z0.h, p0/m, z1.h"},
		{"    WORD $0x04dda020 // fneg z0.d, p0/m, z1.d"},
		{"    WORD $0x049da020 // fneg z0.s, p0/m, z1.s"},
		{"    WORD $0x65cda020 // fsqrt z0.d, p0/m, z1.d"},
		{"    WORD $0x658da020 // fsqrt z0.s, p0/m, z1.s"},
		// rounding instructions
		{"    WORD $0x65c0a020 // frintn z0.d, p0/m, z1.d"},
		{"    WORD $0x6580a020 // frintn z0.s, p0/m, z1.s"},
		{"    WORD $0x65c3a020 // frintz z0.d, p0/m, z1.d"},
		{"    WORD $0x6583a020 // frintz z0.s, p0/m, z1.s"},
		{"    WORD $0x65c4a020 // frinta z0.d, p0/m, z1.d"},
		{"    WORD $0x6584a020 // frinta z0.s, p0/m, z1.s"},
		{"    WORD $0x65c2a020 // frintm z0.d, p0/m, z1.d"},
		{"    WORD $0x6582a020 // frintm z0.s, p0/m, z1.s"},
		{"    WORD $0x65c1a020 // frintp z0.d, p0/m, z1.d"},
		{"    WORD $0x6581a020 // frintp z0.s, p0/m, z1.s"},
		// fmad (multiply-add with different operand order than fmla)
		{"    WORD $0x65e28020 // fmad z0.d, p0/m, z1.d, z2.d"},
		{"    WORD $0x65a28020 // fmad z0.s, p0/m, z1.s, z2.s"},
		// fcvtzs (float to signed integer)
		{"    WORD $0x65dea020 // fcvtzs z0.d, p0/m, z1.d"},
		{"    WORD $0x659ca020 // fcvtzs z0.s, p0/m, z1.s"},
		{"    WORD $0x65d8a020 // fcvtzs z0.s, p0/m, z1.d"},
		{"    WORD $0x65dca020 // fcvtzs z0.d, p0/m, z1.s"},
		{"    WORD $0x65dfa020 // fcvtzu z0.d, p0/m, z1.d"},
		{"    WORD $0x659da020 // fcvtzu z0.s, p0/m, z1.s"},
		{"    WORD $0x65d9a020 // fcvtzu z0.s, p0/m, z1.d"},
		{"    WORD $0x65dda020 // fcvtzu z0.d, p0/m, z1.s"},
		{"    WORD $0x659dbc1f // fcvtzu z31.s, p7/m, z0.s"},
		// cntp (predicate popcount)
		{"    WORD $0x25e08020 // cntp x0, p0, p1.d"},
		{"    WORD $0x25a08020 // cntp x0, p0, p1.s"},
		{"    WORD $0x25608020 // cntp x0, p0, p1.h"},
		{"    WORD $0x25208020 // cntp x0, p0, p1.b"},
		// floating-point compares
		{"    WORD $0x65c16400 // fcmeq p0.d, p1/z, z0.d, z1.d"},
		{"    WORD $0x65c16410 // fcmne p0.d, p1/z, z0.d, z1.d"},
		{"    WORD $0x65c14400 // fcmge p0.d, p1/z, z0.d, z1.d"},
		{"    WORD $0x65c14410 // fcmgt p0.d, p1/z, z0.d, z1.d"},
		{"    WORD $0x65816410 // fcmne p0.s, p1/z, z0.s, z1.s"},
		{"    WORD $0x65416410 // fcmne p0.h, p1/z, z0.h, z1.h"},
		// non-trivial register numbers
		{"    WORD $0x65c08d45 // fadd z5.d, p3/m, z5.d, z10.d"},
		{"    WORD $0x658d9e8f // fdiv z15.s, p7/m, z15.s, z20.s"},
		{"    WORD $0x04dcb41f // fabs z31.d, p5/m, z0.d"},
		{"    WORD $0x04d6a020 // abs z0.d, p0/m, z1.d"},
		{"    WORD $0x0496a020 // abs z0.s, p0/m, z1.s"},
		{"    WORD $0x0456a020 // abs z0.h, p0/m, z1.h"},
		{"    WORD $0x0416a020 // abs z0.b, p0/m, z1.b"},
		{"    WORD $0x04d7a020 // neg z0.d, p0/m, z1.d"},
		{"    WORD $0x0497a020 // neg z0.s, p0/m, z1.s"},
		{"    WORD $0x0457a020 // neg z0.h, p0/m, z1.h"},
		{"    WORD $0x0417a020 // neg z0.b, p0/m, z1.b"},
		{"    WORD $0x04dea020 // not z0.d, p0/m, z1.d"},
		{"    WORD $0x049ea020 // not z0.s, p0/m, z1.s"},
		{"    WORD $0x045ea020 // not z0.h, p0/m, z1.h"},
		{"    WORD $0x041ea020 // not z0.b, p0/m, z1.b"},
		{"    WORD $0x04d6ae8f // abs z15.d, p3/m, z20.d"},
		{"    WORD $0x0497bc1f // neg z31.s, p7/m, z0.s"},
		{"    WORD $0x041eb72a // not z10.b, p5/m, z25.b"},
		{"    WORD $0x25a088a3 // cntp x3, p2, p5.s"},
		{"    WORD $0x25e09d5e // cntp x30, p7, p10.d"},
		{"    WORD $0x65946d55 // fcmne p5.s, p3/z, z10.s, z20.s"},
		{"    WORD $0x65ef88aa // fmad z10.d, p2/m, z5.d, z15.d"},
		{"    WORD $0x65c52020 // fminnmv d0, p0, z1.d"},
		{"    WORD $0x65852020 // fminnmv s0, p0, z1.s"},
		{"    WORD $0x65452020 // fminnmv h0, p0, z1.h"},
		{"    WORD $0x65c42020 // fmaxnmv d0, p0, z1.d"},
		{"    WORD $0x65842020 // fmaxnmv s0, p0, z1.s"},
		{"    WORD $0x65442020 // fmaxnmv h0, p0, z1.h"},
		{"    WORD $0x65852e8f // fminnmv s15, p3, z20.s"},
		{"    WORD $0x65c43c1f // fmaxnmv d31, p7, z0.d"},
		{"    WORD $0x65d82020 // fadda d0, p0, d0, z1.d"},
		{"    WORD $0x65982020 // fadda s0, p0, s0, z1.s"},
		{"    WORD $0x65582020 // fadda h0, p0, h0, z1.h"},
		{"    WORD $0x65982e8f // fadda s15, p3, s15, z20.s"},
		{"    WORD $0x65d83c1f // fadda d31, p7, d31, z0.d"},
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
		{"    WORD $0x2558c083 // pfirst p3.b, p4, p3.b"},
		{"    WORD $0x2599c4c5 // pnext p5.s, p6, p5.s"},
		{"    WORD $0x2598e084 // ptrue p4.s, VL4"},
		{"    WORD $0x2598e3e3 // ptrue p3.s"},
		{"    WORD $0x25a31c41 // whilelo p1.s, x2, x3"},
		{"    WORD $0x25a618b4 // whilehi p4.s, x5, x6"},
		{"    WORD $0x25a91907 // whilehs p7.s, x8, x9"},
		{"    WORD $0x25ac1d7a // whilels p10.s, x11, x12"},
		{"    WORD $0x25a31051 // whilegt p1.s, x2, x3"},
		{"    WORD $0x25a31041 // whilege p1.s, x2, x3"},
		{"    WORD $0x25a31451 // whilele p1.s, x2, x3"},
		{"    WORD $0x25a31441 // whilelt p1.s, x2, x3"},
		{"    WORD $0x05b441ef // rev   p15.s, p15.s"},
		{"    WORD $0x258554a6 // mov   p6.b, p5.b"},
		{"    WORD $0xa540a861 // ld1w { z1.s }, p2/z, [x3]"},
		{"    WORD $0xa567b4c4 // ld1w { z4.d }, p5/z, [x6, #7, MUL VL]"},
		{"    WORD $0xa5183d28 // ld1w { z8.q }, p7/z, [x9, #-8, MUL VL]"},
		{"    WORD $0xa0404882 // ld1w { z2.s, z3.s }, p10/z, [x4]"},
		{"    WORD $0xa04050c4 // ld1w { z4.s-z5.s }, p12/z, [x6]"},
		{"    WORD $0xa047c124 // ld1w { z4.s, z5.s, z6.s, z7.s }, p8/z, [x9, #28, MUL VL]"},
		{"    WORD $0xa040c968 // ld1w { z8.s-z11.s }, p10/z, [x11]"},                        // consecutive
		{"    WORD $0xa567fd25 // ld4w { z5.s, z6.s, z7.s, z8.s }, p7/z, [x9, #28, MUL VL]"}, // interleaved
		{"    WORD $0x855c5482 // ld1w  { z2.s }, p5/z, [x4, z28.s, sxtw]"},
		{"    WORD $0xa54b4450 // ld1w  { z16.s }, p1/z, [x2, x11, lsl #2]"},
		{"    WORD $0x8540de9b // ld1rw { z27.s }, p7/z, [x20]"},
		{"    WORD $0x84cb594a // ld1h  { z10.s }, p6/z, [x10, z11.s, sxtw]"},
		{"    WORD $0xa400a173 // ld1b  { z19.b }, p0/z, [x11]"},
		{"    WORD $0xa420a173 // ld1b  { z19.h }, p0/z, [x11]"},
		{"    WORD $0xa440a173 // ld1b  { z19.s }, p0/z, [x11]"},
		{"    WORD $0xa460a173 // ld1b  { z19.d }, p0/z, [x11]"},
		{"    WORD $0x84155e94 // ld1b  { z20.s }, p7/z, [x20, z21.s, uxtw]"},
		{"    WORD $0xa01f87d8 // ld1b  { z24.b-z27.b }, p9/z, [x30, x31]"},                // identical to following instruction
		{"    WORD $0xa01f87d8 // ld1b  { z24.b, z25.b, z26.b, z27.b }, p9/z, [x30, x31]"}, // consecutive
		{"    WORD $0xa472c66d // ld4b  { z13.b, z14.b, z15.b, z16.b }, p1/z, [x19, x18]"}, // interleaved
		{"    WORD $0xa472c66d // ld4b  { z13.b-z16.b }, p1/z, [x19, x18]"},                // alternative syntax
		{"    WORD $0xa427cca0 // ld2b  { z0.b, z1.b }, p3/z, [x5, x7]"},
		{"    WORD $0xa427cca0 // ld2b  { z0.b-z1.b }, p3/z, [x5, x7]"},
		{"    WORD $0xa421f984 // ld2b  { z4.b-z5.b }, p6/z, [x12, #2, MUL VL]"},
		{"    WORD $0xa449c900 // ld3b  { z0.b, z1.b, z2.b }, p2/z, [x8, x9]"},
		{"    WORD $0xa449c900 // ld3b  { z0.b-z2.b }, p2/z, [x8, x9]"},
		{"    WORD $0xa441e466 // ld3b  { z6.b-z8.b }, p1/z, [x3, #3, MUL VL]"},
		{"    WORD $0xa4a2f0c8 // ld2h  { z8.h-z9.h }, p4/z, [x6, #4, MUL VL]"},
		{"    WORD $0xa4c2ecec // ld3h  { z12.h-z14.h }, p3/z, [x7, #6, MUL VL]"},
		{"    WORD $0xa524f522 // ld2w  { z2.s-z3.s }, p5/z, [x9, #8, MUL VL]"},
		{"    WORD $0xa544e889 // ld3w  { z9.s-z11.s }, p2/z, [x4, #12, MUL VL]"},
		{"    WORD $0xa5a7e504 // ld2d  { z4.d-z5.d }, p1/z, [x8, #14, MUL VL]"},
		{"    WORD $0xa5c7f84f // ld3d  { z15.d-z17.d }, p6/z, [x2, #21, MUL VL]"},
		{"    WORD $0xe42b6ca6 // st2b  { z6.b, z7.b }, p3, [x5, x11]"},
		{"    WORD $0xe42b6ca6 // st2b  { z6.b-z7.b }, p3, [x5, x11]"},
		{"    WORD $0xe431f502 // st2b  { z2.b-z3.b }, p5, [x8, #2, MUL VL]"},
		{"    WORD $0xe44d70c9 // st3b  { z9.b-z11.b }, p4, [x6, x13]"},
		{"    WORD $0xe451e8ec // st3b  { z12.b-z14.b }, p2, [x7, #3, MUL VL]"},
		{"    WORD $0xe46f7920 // st4b  { z0.b-z3.b }, p6, [x9, x15]"},
		{"    WORD $0xe471e464 // st4b  { z4.b-z7.b }, p1, [x3, #4, MUL VL]"},
		{"    WORD $0xe4b2ecaa // st2h  { z10.h-z11.h }, p3, [x5, #4, MUL VL]"},
		{"    WORD $0xe4d2f486 // st3h  { z6.h-z8.h }, p5, [x4, #6, MUL VL]"},
		{"    WORD $0xe4f2f10c // st4h  { z12.h-z15.h }, p4, [x8, #8, MUL VL]"},
		{"    WORD $0xe534f8e4 // st2w  { z4.s-z5.s }, p6, [x7, #8, MUL VL]"},
		{"    WORD $0xe554e469 // st3w  { z9.s-z11.s }, p1, [x3, #12, MUL VL]"},
		{"    WORD $0xe574ecc8 // st4w  { z8.s-z11.s }, p3, [x6, #16, MUL VL]"},
		{"    WORD $0xe5b7f522 // st2d  { z2.d-z3.d }, p5, [x9, #14, MUL VL]"},
		{"    WORD $0xe5d7e886 // st3d  { z6.d-z8.d }, p2, [x4, #21, MUL VL]"},
		{"    WORD $0xe5f7f0ec // st4d  { z12.d-z15.d }, p4, [x7, #28, MUL VL]"},
		// Group 1: contiguous ld1b/ld1h/st1b/st1h with scalar register offset
		{"    WORD $0xa40954e3 // ld1b  { z3.b }, p5/z, [x7, x9]"},
		{"    WORD $0xa42954e3 // ld1b  { z3.h }, p5/z, [x7, x9]"},
		{"    WORD $0xa4a954e3 // ld1h  { z3.h }, p5/z, [x7, x9, lsl #1]"},
		{"    WORD $0xa4c954e3 // ld1h  { z3.s }, p5/z, [x7, x9, lsl #1]"},
		{"    WORD $0xe40954e3 // st1b  { z3.b }, p5, [x7, x9]"},
		{"    WORD $0xe42954e3 // st1b  { z3.h }, p5, [x7, x9]"},
		{"    WORD $0xe4a954e3 // st1h  { z3.h }, p5, [x7, x9, lsl #1]"},
		{"    WORD $0xe4c954e3 // st1h  { z3.s }, p5, [x7, x9, lsl #1]"},
		// Group 2: contiguous ld1h/ld1d/st1h/st1d with scalar immediate offset
		{"    WORD $0xa4a3b4e3 // ld1h  { z3.h }, p5/z, [x7, #3, MUL VL]"},
		{"    WORD $0xa4ceb4e3 // ld1h  { z3.s }, p5/z, [x7, #-2, MUL VL]"},
		{"    WORD $0xa5e3b4e3 // ld1d  { z3.d }, p5/z, [x7, #3, MUL VL]"},
		{"    WORD $0xe4a3f4e3 // st1h  { z3.h }, p5, [x7, #3, MUL VL]"},
		{"    WORD $0xe4cef4e3 // st1h  { z3.s }, p5, [x7, #-2, MUL VL]"},
		{"    WORD $0xe543f4e3 // st1w  { z3.s }, p5, [x7, #3, MUL VL]"},
		{"    WORD $0xe54ef4e3 // st1w  { z3.s }, p5, [x7, #-2, MUL VL]"},
		{"    WORD $0xe5e3f4e3 // st1d  { z3.d }, p5, [x7, #3, MUL VL]"},
		// Group 4: SME2 consecutive multi-vector stores (predicate-as-counter p8-p15 = pn0-pn7)
		{"    WORD $0xa0600000 // st1b  {z0.b-z1.b}, p8, [x0, #0, MUL VL]"},
		{"    WORD $0xa0608000 // st1b  {z0.b-z3.b}, p8, [x0, #0, MUL VL]"},
		{"    WORD $0xa0678124 // st1b  {z4.b-z7.b}, p8, [x9, #28, MUL VL]"},
		{"    WORD $0xa0370622 // st1b  {z2.b-z3.b}, p9, [x17, x23]"},
		{"    WORD $0xa0398e64 // st1b  {z4.b-z7.b}, p11, [x19, x25]"},
		{"    WORD $0xa0603fea // st1h  {z10.h-z11.h}, p15, [x31, #0, MUL VL]"},
		{"    WORD $0xa061b9ac // st1h  {z12.h-z15.h}, p14, [x13, #4, MUL VL]"},
		{"    WORD $0xa02c2160 // st1h  {z0.h-z1.h}, p8, [x11, x12, lsl #1]"},
		{"    WORD $0xa02ca160 // st1h  {z0.h-z3.h}, p8, [x11, x12, lsl #1]"},
		{"    WORD $0xa06059ec // st1w  {z12.s-z13.s}, p14, [x15, #0, MUL VL]"},
		{"    WORD $0xa02cc2bc // st1w  {z28.s-z31.s}, p8, [x21, x12, lsl #2]"},
		{"    WORD $0xa0606164 // st1d  { z4.d-z5.d }, p8, [x11, #0, MUL VL]"},
		{"    WORD $0xa02be2d0 // st1d  { z16.d-z19.d }, p8, [x22, x11, lsl #3]"},
		// Group 3: sign-extending loads
		{"    WORD $0xa5c954e3 // ld1sb { z3.h }, p5/z, [x7, x9]"},
		{"    WORD $0xa5a954e3 // ld1sb { z3.s }, p5/z, [x7, x9]"},
		{"    WORD $0xa58954e3 // ld1sb { z3.d }, p5/z, [x7, x9]"},
		{"    WORD $0xa5c3b4e3 // ld1sb { z3.h }, p5/z, [x7, #3, MUL VL]"},
		{"    WORD $0xa52954e3 // ld1sh { z3.s }, p5/z, [x7, x9, lsl #1]"},
		{"    WORD $0xa50954e3 // ld1sh { z3.d }, p5/z, [x7, x9, lsl #1]"},
		{"    WORD $0xa523b4e3 // ld1sh { z3.s }, p5/z, [x7, #3, MUL VL]"},
		{"    WORD $0xa48954e3 // ld1sw { z3.d }, p5/z, [x7, x9, lsl #2]"},
		{"    WORD $0xa483b4e3 // ld1sw { z3.d }, p5/z, [x7, #3, MUL VL]"},
		// Group 4: broadcast loads
		{"    WORD $0x844094e3 // ld1rb { z3.b }, p5/z, [x7]"},
		{"    WORD $0x8446b4e3 // ld1rb { z3.h }, p5/z, [x7, #6]"},
		{"    WORD $0x84c3b4e3 // ld1rh { z3.h }, p5/z, [x7, #6]"},
		{"    WORD $0x84c3d4e3 // ld1rh { z3.s }, p5/z, [x7, #6]"},
		{"    WORD $0x85c3f4e3 // ld1rd { z3.d }, p5/z, [x7, #24]"},
		// Group 5+6: multi-element h/w/d with register offset and ld4 unified
		{"    WORD $0xa4a7cca4 // ld2h  { z4.h, z5.h }, p3/z, [x5, x7]"},
		{"    WORD $0xa4c7cca4 // ld3h  { z4.h, z5.h, z6.h }, p3/z, [x5, x7]"},
		{"    WORD $0xa4e7cca4 // ld4h  { z4.h, z5.h, z6.h, z7.h }, p3/z, [x5, x7]"},
		{"    WORD $0xa5a7cca4 // ld2d  { z4.d, z5.d }, p3/z, [x5, x7]"},
		{"    WORD $0xa5c7cca4 // ld3d  { z4.d, z5.d, z6.d }, p3/z, [x5, x7]"},
		{"    WORD $0xa5e7cca4 // ld4d  { z4.d, z5.d, z6.d, z7.d }, p3/z, [x5, x7]"},
		{"    WORD $0xa567cca4 // ld4w  { z4.s, z5.s, z6.s, z7.s }, p3/z, [x5, x7]"},
		{"    WORD $0xe4a76ca4 // st2h  { z4.h, z5.h }, p3, [x5, x7]"},
		{"    WORD $0xe4c76ca4 // st3h  { z4.h, z5.h, z6.h }, p3, [x5, x7]"},
		{"    WORD $0xe4e76ca4 // st4h  { z4.h, z5.h, z6.h, z7.h }, p3, [x5, x7]"},
		{"    WORD $0xe5a76ca4 // st2d  { z4.d, z5.d }, p3, [x5, x7]"},
		{"    WORD $0xe5c76ca4 // st3d  { z4.d, z5.d, z6.d }, p3, [x5, x7]"},
		{"    WORD $0xe5e76ca4 // st4d  { z4.d, z5.d, z6.d, z7.d }, p3, [x5, x7]"},
		{"    WORD $0x04018f06 // lsr   z6.h, p3/m, z6.h, #8"},
		{"    WORD $0x04419607 // lsr   z7.s, p5/m, z7.s, #16"},
		{"    WORD $0x04419b0b // lsr   z11.s, p6/m, z11.s, #8"},
		{"    WORD $0x046696d2 // lsr   z18.s, z22.s, #26"},
		{"    WORD $0x047096d2 // lsr   z18.s, z22.s, #16"},
		{"    WORD $0x046896d2 // lsr   z18.s, z22.s, #24"},
		{"    WORD $0x047896d2 // lsr   z18.s, z22.s, #8"},
		{"    WORD $0x0499bf6a // clz   z10.s, p7/m, z27.s"},
		{"    WORD $0x0419a861 // clz   z1.b, p2/m, z3.b"},
		{"    WORD $0x0459aca4 // clz   z4.h, p3/m, z5.h"},
		{"    WORD $0x04d9b107 // clz   z7.d, p4/m, z8.d"},
		{"    WORD $0x049aa440 // cnt   z0.s, p1/m, z2.s"},
		{"    WORD $0x04c130a3 // uaddv d3, p4, z5.d"},
		{"    WORD $0x04c034c4 // saddv d4, p5, z6.d"},
		{"    WORD $0x04c830a3 // smaxv d3, p4, z5.d"},
		{"    WORD $0x04c834c4 // smaxv d4, p5, z6.d"},
		{"    WORD $0x04c82507 // smaxv d7, p1, z8.d"},
		{"    WORD $0x04c930a3 // umaxv d3, p4, z5.d"},
		{"    WORD $0x04c934c4 // umaxv d4, p5, z6.d"},
		{"    WORD $0x04c92507 // umaxv d7, p1, z8.d"},
		{"    WORD $0x04ca30a3 // sminv d3, p4, z5.d"},
		{"    WORD $0x04ca34c4 // sminv d4, p5, z6.d"},
		{"    WORD $0x04ca2507 // sminv d7, p1, z8.d"},
		{"    WORD $0x04cb30a3 // uminv d3, p4, z5.d"},
		{"    WORD $0x04cb34c4 // uminv d4, p5, z6.d"},
		{"    WORD $0x04cb2507 // uminv d7, p1, z8.d"},
		// {" WORD $0x00000000 // uaddv h0, p0, z0.h"},
		// {" WORD $0x00000000 // uaddv s0, p0, z0.h"},
		// {" WORD $0x00000000 // uaddv b0, p0, z0.b"},
		// {" WORD $0x00000000 // uaddv h0, p0, z0.b"},
		// {" WORD $0x00000000 // uaddv s0, p0, z0.b"},
		{"    WORD $0xa5ec456a // ld1d  {z10.d}, p1/z, [x11, x12, lsl #3]"},
		{"    WORD $0xe5ef40c1 // st1d  {z1.d}, p0, [x6, x15, lsl #3]"},
		{"    WORD $0xe56f44d5 // st1w  {z21.d}, p1, [x6, x15, lsl #2]"},
		{"    WORD $0xe55d8447 // st1w  {z7.s}, p1, [x2, z29.s, uxtw]"},
		{"    WORD $0xe4dd8547 // st1h  {z7.s}, p1, [x10, z29.s, uxtw]"},
		{"    WORD $0xe45d8547 // st1b  {z7.s}, p1, [x10, z29.s, uxtw]"},
		// gather loads: scalar + 64-bit vector offset (Zm.D)
		{"    WORD $0xc44edc74 // ld1b  {z20.d}, p7/z, [x3, z14.d]"},
		{"    WORD $0xc4cfd895 // ld1h  {z21.d}, p6/z, [x4, z15.d]"},
		{"    WORD $0xc4f0d4b6 // ld1h  {z22.d}, p5/z, [x5, z16.d, lsl #1]"},
		{"    WORD $0xc551d0d7 // ld1w  {z23.d}, p4/z, [x6, z17.d]"},
		{"    WORD $0xc572ccf8 // ld1w  {z24.d}, p3/z, [x7, z18.d, lsl #2]"},
		{"    WORD $0xc5d3c919 // ld1d  {z25.d}, p2/z, [x8, z19.d]"},
		{"    WORD $0xc5f4c53a // ld1d  {z26.d}, p1/z, [x9, z20.d, lsl #3]"},
		// scatter stores: scalar + 64-bit vector offset (Zm.D)
		{"    WORD $0xe404bdb4 // st1b  {z20.d}, p7, [x13, z4.d]"},
		{"    WORD $0xe484b9d3 // st1h  {z19.d}, p6, [x14, z4.d]"},
		{"    WORD $0xe4a4b5f2 // st1h  {z18.d}, p5, [x15, z4.d, lsl #1]"},
		{"    WORD $0xe504b211 // st1w  {z17.d}, p4, [x16, z4.d]"},
		{"    WORD $0xe524ae30 // st1w  {z16.d}, p3, [x17, z4.d, lsl #2]"},
		{"    WORD $0xe584aa4f // st1d  {z15.d}, p2, [x18, z4.d]"},
		{"    WORD $0xe5a4a66e // st1d  {z14.d}, p1, [x19, z4.d, lsl #3]"},
		{"    WORD $0xe440e28d // st1b  {z13.s}, p0, [x20]"},
		{"    WORD $0xe5800281 // str   p1, [x20]"},
		{"    WORD $0x45016802 // pmullb z2.q, z0.d, z1.d"},
		{"    WORD $0x454c696a // pmullb z10.h, z11.b, z12.b"},
		{"    WORD $0x45d66ab4 // pmullb z20.d, z21.s, z22.s"},
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
		{"    WORD $0x4522e061 // aese z1.b, z1.b, z3.b"},
		{"    WORD $0x4522e4c4 // aesd z4.b, z4.b, z6.b"},
		{"    WORD $0x4520e009 // aesmc z9.b, z9.b"},
		{"    WORD $0x4520e407 // aesimc z7.b, z7.b"},
		{"    WORD $0x05611dac // ext z12.b, {z13.b, z14.b}, #15"},
		{"    WORD $0x05220a30 // ext z16.b, z16.b, z17.b, #18"},
		{"    WORD $0x05b0a000 // clasta x0, p0, x0, z0.s"},
		{"    WORD $0x05b1a421 // clastb x1, p1, x1, z1.s"},
		{"    WORD $0x05a0a000 // lasta x0, p0, z0.s"},
		{"    WORD $0x05a1a000 // lastb x0, p0, z0.s"},
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
		{"    DWORD $0x0488046104912441 // smax z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x0489046104912441 // umax z1.s, p1/M, z2.s, z3.s"},
		{"    DWORD $0x048b046104912441 // umin z1.s, p1/M, z2.s, z3.s"},
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

// TestEvalIntExpr tests the integer expression evaluator used by getImm.
func TestEvalIntExpr(t *testing.T) {
	cases := []struct {
		expr string
		want int
		ok   bool
	}{
		{"4+8", 12, true},
		{"16-4", 12, true},
		{"3*4", 12, true},
		{"24/2", 12, true},
		{"25%13", 12, true},
		{"1<<3", 8, true},  // left shift
		{"16>>1", 8, true}, // right shift
		{"(3+1)*4", 16, true},
		{"2*4+1", 9, true},
		{"1+2*4", 9, true}, // precedence: multiply before add
		{"0x10+8", 24, true},
		{"0x10-0x4", 12, true},
		{"-4+16", 12, true},
		{"(8+4)", 12, true},
		{"", 0, false},
		{"abc", 0, false},
		{"4/0", 0, false},
	}
	for _, tc := range cases {
		ok, got := evalIntExpr(tc.expr)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("evalIntExpr(%q): got (%d, %v), want (%d, %v)", tc.expr, got, ok, tc.want, tc.ok)
		}
	}
}

// TestGetImmExpr tests that getImm handles arithmetic expressions after macro expansion.
func TestGetImmExpr(t *testing.T) {
	cases := []struct {
		imm  string
		want int
	}{
		{"#4+8", 12},   // e.g. #OFFSET+8 with OFFSET=4
		{"#16-4", 12},  // subtraction
		{"#2*6", 12},   // multiplication
		{"#1<<4", 16},  // shift
		{"#(4+8)", 12}, // parenthesised
		{"#0x4+8", 12}, // hex plus decimal
	}
	for _, tc := range cases {
		ok, got := getImm(tc.imm)
		if !ok || got != tc.want {
			t.Errorf("getImm(%q): got (%d, %v), want (%d, true)", tc.imm, got, ok, tc.want)
		}
	}
}

// TestAssembleImmExpr verifies full assembly with expression immediates.
// Each entry assembles to the same opcode as a plain-literal equivalent.
func TestAssembleImmExpr(t *testing.T) {
	cases := []struct {
		expr  string // instruction with expression immediate
		plain string // equivalent instruction with plain literal
	}{
		// add x8, x8, #64  →  opcode 0x91010108
		{"add x8, x8, #60+4", "add x8, x8, #64"},
		// add x2, x1, #0x20, lsl #0  →  0x91008022
		{"add x2, x1, #0x10+0x10, lsl #0", "add x2, x1, #0x20, lsl #0"},
		// sub x16, x2, #124  →  0xd101f050
		{"sub x16, x2, #120+4", "sub x16, x2, #124"},
	}
	for _, tc := range cases {
		oc1, oc2_1, err1 := Assemble(tc.expr)
		oc2, oc2_2, err2 := Assemble(tc.plain)
		if err1 != nil {
			t.Errorf("Assemble(%q): %v", tc.expr, err1)
			continue
		}
		if err2 != nil {
			t.Errorf("Assemble(%q): %v", tc.plain, err2)
			continue
		}
		if oc1 != oc2 || oc2_1 != oc2_2 {
			t.Errorf("Assemble(%q) = 0x%08x, want 0x%08x (from %q)", tc.expr, oc1, oc2, tc.plain)
		}
	}
}
