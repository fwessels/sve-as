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
	"math/bits"
	"strconv"
	"strings"
)

func Assemble(ins string) (opcode, opcode2 uint32, err error) {
	mnem := strings.Fields(ins)[0]
	args := strings.Fields(ins)[1:]
	for i := range args {
		args[i] = strings.TrimSpace(strings.ReplaceAll(args[i], ",", ""))
	}

	switch mnem {
	case "add":
		if ok, rd, rn, rm, shift, imm := is_r_rr(args); ok {
			templ := "sf	0	0	0	1	0	1	1	shift	0	Rm	imm6	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "shift", fmt.Sprintf("%0*s", 2, strconv.FormatUint(uint64(shift), 2)))
			return assem_r_rr(templ, rd, rn, rm, "imm6", imm), 0, nil
		} else if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			templ := "sf	0	0	1	0	0	0	1	0	sh	imm12	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_ri(templ, rd, rn, "imm12", imm, shift), 0, nil
		} else if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	0	size	1	Zm	0	0	0	0	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		} else if ok, zd, zn, imm, shift, T := is_z_zi(args); ok {
			if zd != zn {
				return assem_prefixed_z_z(ins, zd, zn)
			} else {
				templ := "0	0	1	0	0	1	0	1	size	1	0	0	0	0	0	1	1	sh	imm8	Zdn"
				templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
				templ = strings.ReplaceAll(templ, "sh", strconv.Itoa(shift))
				templ = strings.ReplaceAll(templ, "Zdn", "Zd")
				return assem_z_i(templ, zd, "imm8", imm), 0, nil
			}
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	0	0	0	0	0	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "udiv":
		if ok, rd, rn, rm, _, _ := is_r_rr(args); ok {
			templ := "sf	0	0	1	1	0	1	0	1	1	0	Rm	0	0	0	0	1	0	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_rr(templ, rd, rn, rm, "", 0), 0, nil
		}
	case "subs":
		if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			templ := "sf	1	1	1	0	0	0	1	0	sh	imm12	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_ri(templ, rd, rn, "imm12", imm, shift), 0, nil
		}
	case "addvl":
		if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			templ := "0	0	0	0	0	1	0	0	0	0	1	Rn	0	1	0	1	0	imm6	Rd"
			return assem_r_ri(templ, rd, rn, "imm6", imm, shift), 0, nil
		}
	case "mul":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	0	size	1	Zm	0	1	1	0	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	0	0	0	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		} else if ok, rd, rn, rm, shift, imm := is_r_rr(args); ok && shift == 0 && imm == 0 {
			templ := "sf	0	0	1	1	0	1	1	0	0	0	Rm	0	Ra	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			ra := 31
			return assem_r_rrr(templ, rd, rn, rm, ra), 0, nil
		}
	case "madd":
		if ok, rd, rn, rm, ra := is_r_rrr(args); ok {
			templ := "sf	0	0	1	1	0	1	1	0	0	0	Rm	0	Ra	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_rrr(templ, rd, rn, rm, ra), 0, nil
		}
	case "msub":
		if ok, rd, rn, rm, ra := is_r_rrr(args); ok {
			templ := "sf	0	0	1	1	0	1	1	0	0	0	Rm	1	Ra	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_rrr(templ, rd, rn, rm, ra), 0, nil
		}
	case "mneg":
		if ok, rd, rn, rm, shift, imm := is_r_rr(args); ok && shift == 0 && imm == 0 {
			templ := "sf	0	0	1	1	0	1	1	0	0	0	Rm	1	Ra	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			ra := 31
			return assem_r_rrr(templ, rd, rn, rm, ra), 0, nil
		}
	case "tst":
		if ok, rn, rm := is_rr(args); ok {
			// equivalent to "ands xzr, <xn>, <xm>{, <shift> #<amount>}"
			templ := "sf	1	1	0	1	0	1	0	shift	0	Rm	imm6	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "shift", "0\t0")
			rd := getR("xzr")
			return assem_r_rr(templ, rd, rn, rm, "imm6", 0), 0, nil
		} else if ok, rn, imm := is_ri(args); ok {
			if immr, imms := parseBitfieldConst(uint64(imm)); immr != 0xffffffff {
				templ := "sf	1	1	1	0	0	1	0	0	N	immr	imms	Rn	1	1	1	1	1"
				templ = strings.ReplaceAll(templ, "sf", "1")
				templ = strings.ReplaceAll(templ, "N	immr	imms", "imm13")
				imm13 := getImm13(imms, immr, "d")
				return assem_ri(templ, rn, "imm13", int(imm13), 0), 0, nil
			}
		}
	case "and":
		if ok, zd, zn, zm, _ := is_z_zz(args); ok {
			return assem_z_zz("0	0	0	0	0	1	0	0	0	0	1	Zm	0	0	1	1	0	0	Zn	Zd", zd, zn, zm), 0, nil
		} else if ok, zdn, zn, imm, T := is_z_zimm(args); ok && zdn == zn {
			if immr, imms := parseBitfieldConst(uint64(imm)); immr != 0xffffffff {
				templ := "0	0	0	0	0	1	0	1	1	0	0	0	0	0	imm13	Zdn"
				templ = strings.ReplaceAll(templ, "Zdn", "Zd")
				imm13 := getImm13(imms, immr, T)
				return assem_z_i(templ, zdn, "imm13", int(imm13)), 0, nil
			}
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	1	0	1	0	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		} else if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			_ = shift
			if immr, imms := parseBitfieldConst(uint64(imm)); immr != 0xffffffff {
				templ := "sf	0	0	1	0	0	1	0	0	N	immr	imms	Rn	Rd"
				templ = strings.ReplaceAll(templ, "sf", "1")
				templ = strings.ReplaceAll(templ, "N	immr	imms", "imm13")
				imm13 := getImm13(imms, immr, "d")
				return assem_r_ri(templ, rd, rn, "imm13", int(imm13), 0), 0, nil
			}
		}
	case "ands":
		if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			_ = shift // fix
			if immr, imms := parseBitfieldConst(uint64(imm)); immr != 0xffffffff {
				templ := "sf	1	1	1	0	0	1	0	0	N	immr	imms	Rn	Rd"
				templ = strings.ReplaceAll(templ, "sf", "1")
				templ = strings.ReplaceAll(templ, "N	immr	imms", "imm13")
				imm13 := getImm13(imms, immr, "d")
				return assem_r_ri(templ, rd, rn, "imm13", int(imm13), 0), 0, nil
			}
		}
	case "eor":
		if ok, zd, zn, zm, _ := is_z_zz(args); ok {
			return assem_z_zz("0	0	0	0	0	1	0	0	1	0	1	Zm	0	0	1	1	0	0	Zn	Zd", zd, zn, zm), 0, nil
		} else if ok, zdn, zn, imm, T := is_z_zimm(args); ok && zdn == zn {
			if immr, imms := parseBitfieldConst(uint64(imm)); immr != 0xffffffff {
				templ := "0	0	0	0	0	1	0	1	0	1	0	0	0	0	imm13	Zdn"
				templ = strings.ReplaceAll(templ, "Zdn", "Zd")
				imm13 := getImm13(imms, immr, T)
				return assem_z_i(templ, zdn, "imm13", int(imm13)), 0, nil
			}
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	1	0	0	1	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "tbl":
		if ok, zd, zn1, zn2, zm, T := is_z_zz_z(args); ok && zn2 == zn1+1 {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	0	1	0	1	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn1, zm), 0, nil
		} else if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	0	1	1	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "dupm": // duplicate with (contiguous) bit mask
		if ok, zd, imm, T := is_z_i(args); ok {
			if immr, imms := parseBitfieldConst(uint64(imm)); immr != 0xffffffff {
				templ := "0	0	0	0	0	1	0	1	1	1	0	0	0	0	imm13	Zd"
				imm13 := getImm13(imms, immr, T)
				return assem_z_i(templ, zd, "imm13", int(imm13)), 0, nil
			}
		}
	case "dup":
		if ok, zd, zn, imm, T := is_z_zindexed(args); ok {
			templ := "0	0	0	0	0	1	0	1	imm2	1	tsz	0	0	1	0	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "tsz", getTypeSpecifier(T))
			return assem_z_zi(templ, zd, zn, "imm2", imm), 0, nil
		} else if ok, zd, imm, T := is_z_i(args); ok {
			templ := "0	0	1	0	0	1	0	1	size	1	1	1	0	0	0	1	1	sh	imm8	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			sh := ""
			if imm >= -128 && imm <= 127 {
				sh = "0"
				if imm < 0 {
					imm = 0x100 + imm
				}
			} else if imm >= -128*256 && imm <= 127*256 && imm%256 == 0 {
				sh = "1"
				if imm < 0 {
					imm = 0x10000 + imm
				}
				imm = imm >> 8
			}
			if sh != "" {
				templ = strings.ReplaceAll(templ, "sh", sh)
				return assem_z_i(templ, zd, "imm8", imm), 0, nil
			}
		}
	case "mov":
		if ok, pd, pn, T := is_p_p(args); ok && strings.ToLower(T) == "b" {
			templ := "0	0	1	0	0	1	0	1	1	0	0	0	Pm	0	1	Pg	0	Pn	0	Pd"
			// MOV <Pd>.B, <Pn>.B
			// is equivalent to
			// ORR <Pd>.B, <Pn>/Z, <Pn>.B, <Pn>.B
			return assem_p_p_p_p(templ, pd, pn, pn, pn), 0, nil
		} else if ok, zd, rn, T := is_z_r(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	0	0	0	0	0	0	1	1	1	0	Rn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_r(templ, zd, rn), 0, nil
		} else if ok, rd, imm, shift := is_r_i(args); ok {
			_ = shift
			// Using MOV (wide immediate) here (which is an alias for MOVZ)
			templ := "sf	1	0	1	0	0	1	0	1	hw	imm16	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "hw", "0\t0")
			return assem_r_i(templ, rd, "imm16", imm), 0, nil
		} else if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			// MOV <Zd>.<T>, <Pv>/M, <Zn>.<T>
			//   is equivalent to
			// SEL <Zd>.<T>, <Pv>, <Zn>.<T>, <Zd>.<T>
			//   and is the preferred disassembly when Zd == Zm.
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	1	1	Pv	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_zz_4(templ, zd, pg, zn, zd), 0, nil
		} else if ok, rd, rn := is_r_r(args); ok {
			// MOV <Xd>, <Xm>
			// is equivalent to
			// ORR <Xd>, XZR, <Xm>
			templ := "sf	0	1	0	1	0	1	0	0	0	0	Rm	0	0	0	0	0	0	1	1	1	1	1	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "Rm", "Rn")
			return assem_r_ri(templ, rd, rn, "imm6", 0, 0), 0, nil
		}
	case "mvn":
		if ok, rd, rn := is_r_r(args); ok {
			// MVN <Xd>, <Xm>{, <shift> #<amount>}
			// is equivalent to
			// ORN <Xd>, XZR, <Xm>{, <shift> #<amount>}
			templ := "sf	0	1	0	1	0	1	0	shift	1	Rm	imm6	1	1	1	1	1	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "shift", "0\t0")
			templ = strings.ReplaceAll(templ, "Rm", "Rn")
			return assem_r_ri(templ, rd, rn, "imm6", 0, 0), 0, nil
		}
	case "cmp":
		if ok, rd, imm, shift := is_r_i(args); ok && 0 <= imm && imm < 4096 && (shift == 0 || shift == 12) {
			// CMP <Xn|SP>, #<imm>{, <shift>}
			// is equivalent to
			// SUBS XZR, <Xn|SP>, #<imm> {, <shift>}
			templ := "sf	1	1	1	0	0	0	1	0	sh	imm12	Rn	1	1	1	1	1"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "Rn", "Rd")
			if shift == 12 {
				templ = strings.ReplaceAll(templ, "sh", "1") // LSL #12
			} else {
				templ = strings.ReplaceAll(templ, "sh", "0") // LSL #0
			}
			return assem_r_i(templ, rd, "imm12", imm), 0, nil
		} else if ok, rd, rn := is_r_r(args); ok && len(args) == 2 {
			// CMP <Xn>, <Xm>{, <shift> #<amount>}
			// is equivalent to
			// SUBS XZR, <Xn>, <Xm> {, <shift> #<amount>}
			templ := "sf	1	1	0	1	0	1	1	shift	0	Rm	imm6	Rn	1	1	1	1	1"
			templ = strings.ReplaceAll(templ, "sf", "1")
			// 00	LSL
			// 01	LSR
			// 10	ASR
			// 11	RESERVED
			templ = strings.ReplaceAll(templ, "shift", "00")
			templ = strings.ReplaceAll(templ, "Rn", "Rd")
			templ = strings.ReplaceAll(templ, "Rm", "Rn")

			return assem_r_ri(templ, rd, rn, "imm6", 0, 0), 0, nil
		}
	case "adr":
		if ok, rd, imm, shift := is_r_i(args); ok {
			_ = shift
			if imm < 0 {
				imm = (1 << 21) + imm
			}
			templ := "0	immlo	1	0	0	0	0	immhi	Rd"
			templ = strings.ReplaceAll(templ, "immlo", fmt.Sprintf("%0*s", 2, strconv.FormatUint(uint64(imm&3), 2)))
			return assem_r_i(templ, rd, "immhi", imm>>2), 0, nil
		}
	case "ldr":
		if ok, zt, xn, imm := is_z_bi(args); ok && -256 <= imm && imm < 256 {
			templ := "1	0	0	0	0	1	0	1	1	0	imm9h	0	1	0	imm9l	Rn	Zt"
			return assem_z_bi(templ, zt, xn, imm), 0, nil
		} else if ok, pt, xn, imm := is_p_bi(args); ok && -256 <= imm && imm < 256 {
			templ := "1	0	0	0	0	1	0	1	1	0	imm9h	0	0	0	imm9l	Rn	0	Pt"
			return assem_p_bi(templ, pt, xn, imm), 0, nil
		} else if ok, rt, rn, rm, option, amount := is_r_br(args); ok {
			if option == 3 {
				templ := "1	x	1	1	1	0	0	0	0	1	1	Rm	option	S	1	0	Rn	Rt"
				templ = strings.ReplaceAll(templ, "x", "1")
				templ = strings.ReplaceAll(templ, "Rt", "Rd")
				templ = strings.ReplaceAll(templ, "option", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(option), 2)))
				s := -1
				if amount == 0 {
					s = 0
				} else if amount == 3 {
					s = 1
				}
				if s != -1 {
					templ = strings.ReplaceAll(templ, "S", fmt.Sprintf("%0*s", 1, strconv.FormatUint(uint64(s), 2)))
					return assem_r_rr(templ, rt, rn, rm, "", 0), 0, nil
				}
			}
		} else if ok, rt, rn, imm12 := is_r_bi(args); ok {
			if imm12&7 == 0 {
				templ := "1	x	1	1	1	0	0	1	0	1	imm12	Rn	Rt"
				templ = strings.ReplaceAll(templ, "Rt", "Rd")
				templ = strings.ReplaceAll(templ, "x", "1")
				return assem_r_ri(templ, rt, rn, "imm12", imm12/8, 0), 0, nil
			}
		}
	case "str":
		if ok, zt, xn, imm := is_z_bi(args); ok && -256 <= imm && imm < 256 {
			templ := "1	1	1	0	0	1	0	1	1	0	imm9h	0	1	0	imm9l	Rn	Zt"
			return assem_z_bi(templ, zt, xn, imm), 0, nil
		} else if ok, pt, xn, imm := is_p_bi(args); ok && -256 <= imm && imm < 256 {
			templ := "1	1	1	0	0	1	0	1	1	0	imm9h	0	0	0	imm9l	Rn	0	Pt"
			return assem_p_bi(templ, pt, xn, imm), 0, nil
		} else if ok, rt, rn, rm, option, amount := is_r_br(args); ok {
			if option == 3 {
				templ := "1	x	1	1	1	0	0	0	0	0	1	Rm	option	S	1	0	Rn	Rt"
				templ = strings.ReplaceAll(templ, "x", "1")
				templ = strings.ReplaceAll(templ, "Rt", "Rd")
				templ = strings.ReplaceAll(templ, "option", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(option), 2)))
				s := -1
				if amount == 0 {
					s = 0
				} else if amount == 3 {
					s = 1
				}
				if s != -1 {
					templ = strings.ReplaceAll(templ, "S", fmt.Sprintf("%0*s", 1, strconv.FormatUint(uint64(s), 2)))
					return assem_r_rr(templ, rt, rn, rm, "", 0), 0, nil
				}
			}
		} else if ok, rt, rn, imm12 := is_r_bi(args); ok {
			if imm12&7 == 0 {
				templ := "1	x	1	1	1	0	0	1	0	0	imm12	Rn	Rt"
				templ = strings.ReplaceAll(templ, "Rt", "Rd")
				templ = strings.ReplaceAll(templ, "x", "1")
				return assem_r_ri(templ, rt, rn, "imm12", imm12/8, 0), 0, nil
			}
		}
	case "ld1d":
		if ok, zt, pg, rn, rm, _, _ := is_z_p_rr(args); ok {
			templ := "1	0	1	0	0	1	0	1	1	1	1	Rm	0	1	0	Pg	Rn	Zt"
			return assem_z_p_rr(templ, zt, pg, rn, rm), 0, nil
		}
	case "st1b":
		if ok, zt, pg, rn, zm, xs, T := is_z_p_bz(args); ok {
			if T == "s" {
				templ := "1	1	1	0	0	1	0	0	0	1	0	Zm	1	xs	0	Pg	Rn	Zt"
				return assem_z_p_bz(templ, zt, pg, rn, zm, xs), 0, nil
			}
		} else if ok, zt, pg, rn, imm, T := is_z_p_bi(args); ok {
			templ := "1	1	1	0	0	1	0	0	0	size	0	imm4	1	1	1	Pg	Rn	Zt"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_bi(templ, zt, pg, rn, "imm4", imm), 0, nil
		}
	case "st1d":
		if ok, zt, pg, rn, rm, shift, T := is_z_p_rr(args); ok && shift == 3 && strings.ToLower(T) == "d" {
			templ := "1	1	1	0	0	1	0	1	1	1	1	Rm	0	1	0	Pg	Rn	Zt"
			return assem_z_p_rr(templ, zt, pg, rn, rm), 0, nil
		}
	case "st1w":
		if ok, zt, pg, rn, zm, xs, T := is_z_p_bz(args); ok {
			if T == "s" {
				templ := "1	1	1	0	0	1	0	1	0	1	0	Zm	1	xs	0	Pg	Rn	Zt"
				return assem_z_p_bz(templ, zt, pg, rn, zm, xs), 0, nil
			}
		} else if ok, zt, pg, rn, rm, shift, T := is_z_p_rr(args); ok && shift == 2 {
			templ := "1	1	1	0	0	1	0	1	0	1	sz	Rm	0	1	0	Pg	Rn	Zt"
			if strings.ToLower(T) == "s" {
				templ = strings.ReplaceAll(templ, "sz", "0")
			}
			if strings.ToLower(T) == "d" {
				templ = strings.ReplaceAll(templ, "sz", "1")
			}
			return assem_z_p_rr(templ, zt, pg, rn, rm), 0, nil
		}
	case "st1h":
		if ok, zt, pg, rn, zm, xs, T := is_z_p_bz(args); ok {
			if T == "s" {
				templ := "1	1	1	0	0	1	0	0	1	1	0	Zm	1	xs	0	Pg	Rn	Zt"
				return assem_z_p_bz(templ, zt, pg, rn, zm, xs), 0, nil
			}
		}
	case "ld1w":
		if ok, zt, pg, rn, zm, xs, T := is_z_p_bz(args); ok {
			if T == "s" {
				templ := "1	0	0	0	0	1	0	1	0	xs	0	Zm	0	1	0	Pg	Rn	Zt"
				return assem_z_p_bz(templ, zt, pg, rn, zm, xs), 0, nil
			}
		} else if ok, zt, pg, rn, rm, shift, T := is_z_p_rr(args); ok && shift == 2 {
			var templ string
			if strings.ToLower(T) == "s" {
				templ = "1	0	1	0	0	1	0	1	0	1	0	Rm	0	1	0	Pg	Rn	Zt"
			}
			if strings.ToLower(T) == "d" {
				templ = "1	0	1	0	0	1	0	1	0	1	1	Rm	0	1	0	Pg	Rn	Zt"
			}
			if strings.ToLower(T) == "q" {
				templ = "1	0	1	0	0	1	0	1	0	0	0	Rm	1	0	0	Pg	Rn	Zt"
			}
			return assem_z_p_rr(templ, zt, pg, rn, rm), 0, nil
		}
	case "ld1rw":
		if ok, zt, pg, rn, imm, T := is_z_p_bi(args); ok {
			var templ string
			if T == "s" {
				templ = "1	0	0	0	0	1	0	1	0	1	imm6	1	1	0	Pg	Rn	Zt"
			} else if T == "d" {
				templ = "1	0	0	0	0	1	0	1	0	1	imm6	1	1	1	Pg	Rn	Zt"
			}
			if templ != "" {
				return assem_z_p_bi(templ, zt, pg, rn, "imm6", imm/4), 0, nil
			}
		}
	case "ld1h":
		if ok, zt, pg, rn, zm, xs, T := is_z_p_bz(args); ok {
			if strings.ToLower(T) == "s" {
				templ := "1	0	0	0	0	1	0	0	1	xs	0	Zm	0	1	0	Pg	Rn	Zt"
				return assem_z_p_bz(templ, zt, pg, rn, zm, xs), 0, nil
			}
		}
	case "ld1b":
		if ok, zt, pg, rn, zm, xs, T := is_z_p_bz(args); ok {
			if strings.ToLower(T) == "s" {
				templ := "1	0	0	0	0	1	0	0	0	xs	0	Zm	0	1	0	Pg	Rn	Zt"
				return assem_z_p_bz(templ, zt, pg, rn, zm, xs), 0, nil
			}
		} else if ok, zt, pg, rn, rm, T := is_zt4_p_rr(args); ok {
			if strings.ToLower(T) == "b" && zt&3 == 0 && pg >= 8 && pg <= 15 {
				templ := "1	0	1	0	0	0	0	0	0	0	0	Rm	1	0	0	PNg	Rn	Zt	0	0"
				return assem_zt4_p_rr(templ, zt>>2, pg, rn, rm), 0, nil
			}
		}
	case "ld4b":
		if ok, zt, pg, rn, rm, T := is_zt4_p_rr(args); ok {
			if strings.ToLower(T) == "b" && pg >= 0 && pg <= 7 {
				templ := "1	0	1	0	0	1	0	0	0	1	1	Rm	1	1	0	Pg	Rn	Zt"
				return assem_z_p_rr(templ, zt, pg, rn, rm), 0, nil
			}
		}
	case "ldrh":
		if ok, rt, rn, imm12 := is_r_bi(args); ok {
			templ := "0	1	1	1	1	0	0	1	0	1	imm12	Rn	Rt"
			templ = strings.ReplaceAll(templ, "Rt", "Rd")
			templ = strings.ReplaceAll(templ, "sh", "")
			return assem_r_ri(templ, rt, rn, "imm12", imm12/2, 0), 0, nil
		}
	case "lsr":
		if ok, rd, rn, imm, _ := is_r_ri(args); ok {
			templ := "sf	1	0	1	0	0	1	1	0	N	immr	x	1	1	1	1	1	Rn	Rd"
			// see https://docsmirror.github.io/A64/2023-06/lsr_ubfm.html
			// LSR <Xd>, <Xn>, #<shift>
			//   is equivalent to
			// UBFM <Xd>, <Xn>, #<shift>, #63
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "N", "1")
			templ = strings.ReplaceAll(templ, "x", "1") // x bit is set for compat with 'as'
			return assem_r_ri(templ, rd, rn, "immr", imm, 0), 0, nil
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	0	0	1	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		} else if ok, zd, pg, zn, imm, T := is_z_p_zimm(args); ok && imm > 0 {
			templ := "0	0	0	0	0	1	0	0	tszh	0	0	0	0	0	1	1	0	0	Pg	tszl	imm3	Zdn"
			imm3, tsz := computeShiftSpecifier(uint(imm), true, T)
			templ = strings.ReplaceAll(templ, "tszh", tsz[:2])
			templ = strings.ReplaceAll(templ, "tszl", tsz[2:])
			if zd == zn {
				return assem_z_p_zi(templ, zd, pg, "imm3", imm3), 0, nil
			} else {
				// we need to use a prefix
				return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
			}
		} else if ok, zd, zn, imm, T := is_z_zimm(args); ok && imm > 0 {
			templ := "0	0	0	0	0	1	0	0	tszh	1	tszl	imm3	1	0	0	1	0	1	Zn	Zd"
			imm3, tsz := computeShiftSpecifier(uint(imm), true, T)
			templ = strings.ReplaceAll(templ, "tszh", tsz[:2])
			templ = strings.ReplaceAll(templ, "tszl", tsz[2:])
			return assem_z_zi(templ, zd, zn, "imm3", imm3), 0, nil
		}
	case "lsl":
		if ok, rd, rn, imm, _ := is_r_ri(args); ok {
			templ := "sf	1	0	1	0	0	1	1	0	N	immr	imms	Rn	Rd"
			// see https://docsmirror.github.io/A64/2023-06/lsl_ubfm.html
			// LSL <Xd>, <Xn>, #<shift>
			//   is equivalent to
			// UBFM <Xd>, <Xn>, #(-<shift> MOD 64), #(63-<shift>)
			immr := (uint(-imm) % 64)
			imms := 63 - imm
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "N", "1")
			templ = strings.ReplaceAll(templ, "imms", fmt.Sprintf("%0*s", 6, strconv.FormatInt(int64(imms), 2)))
			return assem_r_ri(templ, rd, rn, "immr", int(immr), 0), 0, nil
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	0	1	1	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		} else if ok, zd, zn, imm, T := is_z_zimm(args); ok {
			imm3, tsz := computeShiftSpecifier(uint(imm), false, T)
			templ := "0	0	0	0	0	1	0	0	tszh	1	tszl	imm3	1	0	0	1	1	1	Zn	Zd"
			templ = strings.ReplaceAll(templ, "tszh", tsz[:2])
			templ = strings.ReplaceAll(templ, "tszl", tsz[2:])
			return assem_z_zi(templ, zd, zn, "imm3", imm3), 0, nil
		}
	case "rdvl":
		if ok, rd, imm, shift := is_r_i(args); ok {
			_ = shift
			templ := "0	0	0	0	0	1	0	0	1	0	1	1	1	1	1	1	0	1	0	1	0	imm6	Rd"
			return assem_r_i(templ, rd, "imm6", imm), 0, nil
		}
	case "ptrue":
		if ok, pd, T := is_p(args); ok && len(args) <= 2 {
			templ := "0	0	1	0	0	1	0	1	size	0	1	1	0	0	0	1	1	1	0	0	0	pattern	0	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			pattern := "ALL"
			if len(args) == 2 {
				pattern = args[1]
			}
			switch strings.ToUpper(pattern) {
			case "POW2":
				templ = strings.ReplaceAll(templ, "pattern", "00000")
			case "VL1":
				templ = strings.ReplaceAll(templ, "pattern", "00001")
			case "VL2":
				templ = strings.ReplaceAll(templ, "pattern", "00010")
			case "VL3":
				templ = strings.ReplaceAll(templ, "pattern", "00011")
			case "VL4":
				templ = strings.ReplaceAll(templ, "pattern", "00100")
			case "VL5":
				templ = strings.ReplaceAll(templ, "pattern", "00101")
			case "VL6":
				templ = strings.ReplaceAll(templ, "pattern", "00110")
			case "VL7":
				templ = strings.ReplaceAll(templ, "pattern", "00111")
			case "VL8":
				templ = strings.ReplaceAll(templ, "pattern", "01000")
			case "VL16":
				templ = strings.ReplaceAll(templ, "pattern", "01001")
			case "VL32":
				templ = strings.ReplaceAll(templ, "pattern", "01010")
			case "VL64":
				templ = strings.ReplaceAll(templ, "pattern", "01011")
			case "VL128":
				templ = strings.ReplaceAll(templ, "pattern", "01100")
			case "VL256":
				templ = strings.ReplaceAll(templ, "pattern", "01101")
			case "#uimm5":
				templ = strings.ReplaceAll(templ, "pattern", "0111x")
			// TODO: fix this
			// case "#uimm5":
			// 	templ = strings.ReplaceAll(templ, "pattern", "101x1")
			// case "#uimm5":
			// 	templ = strings.ReplaceAll(templ, "pattern", "10110")
			// case "#uimm5":
			// 	templ = strings.ReplaceAll(templ, "pattern", "1x0x1")
			// case "#uimm5":
			// 	templ = strings.ReplaceAll(templ, "pattern", "1x010")
			// case "#uimm5":
			// 	templ = strings.ReplaceAll(templ, "pattern", "1xx00")
			case "MUL4":
				templ = strings.ReplaceAll(templ, "pattern", "11101")
			case "MUL3":
				templ = strings.ReplaceAll(templ, "pattern", "11110")
			case "ALL":
				templ = strings.ReplaceAll(templ, "pattern", "11111")
			}
			return assem_p(templ, pd), 0, nil
		}
	case "eor3":
		if ok, zd, zn, zm, za, _ := is_z_zzz(args); ok {
			templ := "0	0	0	0	0	1	0	0	0	0	1	Zm	0	0	1	1	1	0	Zk	Zdn"
			if zd == zn {
				return assem_z2_zz(templ, zd, zm, za), 0, nil
			}
		}
	case "mad":
		if ok, zdn, pg, zm, za, T := is_z2_p_zz(args); ok {
			templ := "0	0	0	0	0	1	0	0	size	0	Zm	1	1	0	Pg	Za	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_zz(templ, zdn, pg, zm, za), 0, nil
		}
	case "mls":
		if ok, zda, pg, zn, zm, T := is_z2_p_zz(args); ok {
			templ := "0	0	0	0	0	1	0	0	size	0	Zm	0	1	1	Pg	Zn	Zda"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			templ = strings.ReplaceAll(templ, "Zda", "Zdn")
			templ = strings.ReplaceAll(templ, "Zm", "Za")
			templ = strings.ReplaceAll(templ, "Zn", "Zm")
			return assem_z2_p_zz(templ, zda, pg, zn, zm), 0, nil
		}
	case "compact":
		if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	0	0	0	1	1	0	0	Pg	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_z(templ, zd, pg, zn), 0, nil
		}
	case "zip1":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	1	1	0	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "zip2":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	1	1	0	0	1	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "uzp1":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	1	1	0	1	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "uzp2":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	1	1	0	1	1	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "trn1":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	1	1	1	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "trn2":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	1	1	1	0	1	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "rev":
		if ok, zd, zn, T := is_z_z(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	1	1	0	0	0	0	0	1	1	1	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_z(templ, zd, zn), 0, nil
		} else if ok, pg, pn, T := is_p_p(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	1	0	1	0	0	0	1	0	0	0	0	0	Pn	0	Pd"
			templ = strings.ReplaceAll(templ, "Pd", "Pg")
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_p_p(templ, pg, pn), 0, nil
		}
	case "revb":
		if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	0	1	0	0	1	0	0	Pg	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_z(templ, zd, pg, zn), 0, nil
		}
	case "revh":
		if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	0	1	0	1	1	0	0	Pg	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_z(templ, zd, pg, zn), 0, nil
		}
	case "revw":
		if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	0	1	1	0	1	0	0	Pg	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_z(templ, zd, pg, zn), 0, nil
		}
	case "sdot":
		if ok, zda, zn, zm, Td, T := is_z_zz_2t(args); ok {
			templ := "0	1	0	0	0	1	0	0	size	0	Zm	0	0	0	0	0	0	Zn	Zda"
			if Td == "d" && T == "h" {
				templ = strings.ReplaceAll(templ, "size", "11")
				return assem_z_zz2(templ, zda, zn, zm), 0, nil
			} else if Td == "s" && T == "b" {
				templ = strings.ReplaceAll(templ, "size", "10")
				return assem_z_zz2(templ, zda, zn, zm), 0, nil
			}
		}
	case "fcvt":
		if ok, zd, pg, zn, Td, Tn := is_z_p_z_tt(args); ok {
			if Td == "s" && Tn == "h" {
				templ := "0	1	1	0	0	1	0	1	1	0	0	0	1	0	0	1	1	0	1	Pg	Zn	Zd"
				return assem_z_p_z(templ, zd, pg, zn), 0, nil
			}
		}
	case "fmul":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			if T != "b" {
				templ := "0	1	1	0	0	1	0	1	size	0	Zm	0	0	0	0	1	0	Zn	Zd"
				templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
				return assem_z_zz(templ, zd, zn, zm), 0, nil
			}
		}
	case "asr":
		if ok, rd, rn, imm, _ := is_r_ri(args); ok {
			templ := "sf	0	0	1	0	0	1	1	0	N	immr	x	1	1	1	1	1	Rn	Rd"
			// see https://docsmirror.github.io/A64/2023-06/asr_sbfm.html
			// ASR <Xd>, <Xn>, #<shift>
			// is equivalent to
			// SBFM <Xd>, <Xn>, #<shift>, #63
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "N", "1")
			templ = strings.ReplaceAll(templ, "x", "1") // x bit is set for compat with 'as'
			return assem_r_ri(templ, rd, rn, "immr", imm, 0), 0, nil
		} else if ok, zd, zn, imm, T := is_z_zimm(args); ok {
			templ := "0	0	0	0	0	1	0	0	tszh	1	tszl	imm3	1	0	0	1	0	0	Zn	Zd"
			imm3, tsz := computeShiftSpecifier(uint(imm), true, T)
			templ = strings.ReplaceAll(templ, "tszh", tsz[:2])
			templ = strings.ReplaceAll(templ, "tszl", tsz[2:])
			return assem_z_zi(templ, zd, zn, "imm3", imm3), 0, nil
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	0	0	0	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "scvtf":
		if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			if T == "s" {
				templ := "0	1	1	0	0	1	0	1	1	0	0	1	0	1	0	0	1	0	1	Pg	Zn	Zd"
				return assem_z_p_z(templ, zd, pg, zn), 0, nil
			}
		}
	case "fmla":
		if ok, zda, pg, zn, zm, T := is_z_p_zz2(args); ok {
			if T != "b" {
				templ := "0	1	1	0	0	1	0	1	size	1	Zm	0	0	0	Pg	Zn	Zda"
				templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
				return assem_z_p_zz(templ, zda, pg, zn, zm), 0, nil
			}
		}
	case "sel":
		if ok, zd, pv, zn, zm, T := is_z_p_zz_4(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	1	1	Pv	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_zz_4(templ, zd, pv, zn, zm), 0, nil
		}
	case "splice":
		if ok, zdn, pv, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	1	1	0	0	1	0	0	Pv	Zm	Zdn" // Destructive variant
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			templ = strings.ReplaceAll(templ, "Pv", "Pg") // both are 3 bits
			return assem_z2_p_z(templ, zdn, pv, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "asrr":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	1	0	0	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "bic":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	1	0	1	1	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "clasta":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	1	0	0	0	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "clastb":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	1	0	0	1	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "lslr":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	1	1	1	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "lsrr":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	1	0	1	1	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "orr":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	0	0	1	1	Zm	0	0	1	1	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		} else if ok, zdn, zn, imm, T := is_z_zimm(args); ok && zdn == zn {
			if immr, imms := parseBitfieldConst(uint64(imm)); immr != 0xffffffff {
				templ := "0	0	0	0	0	1	0	1	0	0	0	0	0	0	imm13	Zdn"
				templ = strings.ReplaceAll(templ, "Zdn", "Zd")
				imm13 := getImm13(imms, immr, T)
				return assem_z_i(templ, zdn, "imm13", int(imm13)), 0, nil
			}
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	1	0	0	0	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "clz":
		if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	1	0	0	1	1	0	1	Pg	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_z(templ, zd, pg, zn), 0, nil
		}
	case "sabd":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	0	1	1	0	0	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "sdiv":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			if strings.ToLower(T) == "d" || strings.ToLower(T) == "s" {
				// sdiv only defined for 64- and 32-bit
				templ := "0	0	0	0	0	1	0	0	size	0	1	0	1	0	0	0	0	0	Pg	Zm	Zdn"
				templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
				return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
			}
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			if strings.ToLower(T) == "d" || strings.ToLower(T) == "s" {
				// sdiv only defined for 64- and 32-bit
				return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
			}
		} else if ok, rd, rn, rm, shift, imm := is_r_rr(args); ok && shift == 0 && imm == 0 {
			templ := "sf	0	0	1	1	0	1	0	1	1	0	Rm	0	0	0	0	1	1	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "shift", fmt.Sprintf("%0*s", 2, strconv.FormatUint(uint64(shift), 2)))
			return assem_r_rr(templ, rd, rn, rm, "", 0), 0, nil
		}
	case "sdivr":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			if strings.ToLower(T) == "d" || strings.ToLower(T) == "s" {
				// sdivr only defined for 64- and 32-bit
				templ := "0	0	0	0	0	1	0	0	size	0	1	0	1	1	0	0	0	0	Pg	Zm	Zdn"
				templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
				return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
			}
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			if strings.ToLower(T) == "d" || strings.ToLower(T) == "s" {
				// sdivr only defined for 64- and 32-bit
				return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
			}
		}
	case "smin":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	0	1	0	1	0	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "smulh":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	1	0	0	1	0	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "sub":
		if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			templ := "sf	1	0	1	0	0	0	1	0	sh	imm12	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_ri(templ, rd, rn, "imm12", imm, shift), 0, nil
		} else if ok, rd, rn, rm, shift, imm := is_r_rr(args); ok {
			templ := "sf	1	0	0	1	0	1	1	shift	0	Rm	imm6	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "shift", fmt.Sprintf("%0*s", 2, strconv.FormatUint(uint64(shift), 2)))
			return assem_r_rr(templ, rd, rn, rm, "imm6", imm), 0, nil
		} else if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	0	size	1	Zm	0	0	0	0	0	1	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		} else if ok, zd, zn, imm, shift, T := is_z_zi(args); ok && imm < 256 {
			if zd != zn {
				return assem_prefixed_z_z(ins, zd, zn)
			} else {
				templ := "0	0	1	0	0	1	0	1	size	1	0	0	0	0	1	1	1	sh	imm8	Zdn"
				templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
				templ = strings.ReplaceAll(templ, "sh", strconv.Itoa(shift))
				templ = strings.ReplaceAll(templ, "Zdn", "Zd")
				return assem_z_i(templ, zd, "imm8", imm), 0, nil
			}
		} else if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	0	0	0	0	1	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "subr":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	0	0	0	1	1	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "uabd":
		if ok, zdn, pg, zm, T := is_z_p_zz(args); !is_zeroing(args[1]) && ok {
			templ := "0	0	0	0	0	1	0	0	size	0	0	1	1	0	1	0	0	0	Pg	Zm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z2_p_z(templ, zdn, pg, zm), 0, nil
		} else if ok, zd, pg, zn, _, T := is_prefixed_z_p_zz(args); ok {
			return assem_prefixed_z_p_z(ins, args[1], zd, pg, zn, T)
		}
	case "movprfx":
		if ok, zd, zn, T := is_z_z(args); ok {
			if T == "" {
				templ := "0	0	0	0	0	1	0	0	0	0	1	0	0	0	0	0	1	0	1	1	1	1	Zn	Zd"
				return assem_z_z(templ, zd, zn), 0, nil
			}
		} else if ok, zd, pg, zn, T := is_z_p_z(args); ok {
			return assem_prefixed_z_p_z("", args[1], zd, pg, zn, T)
		}
	case "histcnt":
		if ok, zd, pg, zn, zm, T := is_z_p_zz_4(args); ok && pg < 8 && (strings.ToLower(T) == "s" || strings.ToLower(T) == "d") {
			templ := "0	1	0	0	0	1	0	1	size	1	Zm	1	1	0	Pg	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_p_zz_4(templ, zd, pg, zn, zm), 0, nil
		}
	case "histseg":
		if ok, zd, zn, zm, T := is_z_zz(args); ok && strings.ToLower(T) == "b" {
			templ := "0	1	0	0	0	1	0	1	size	1	Zm	1	0	1	0	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), 0, nil
		}
	case "match":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	1	0	0	0	1	0	1	size	1	Zm	1	0	0	Pg	Zn	0	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		}
	case "nmatch":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	1	0	0	0	1	0	1	size	1	Zm	1	0	0	Pg	Zn	1	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		}
	case "cmpeq":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	0	1	0	0	1	0	0	size	0	Zm	1	0	1	Pg	Zn	0	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		} else if ok, pd, pg, zn, imm, T := is_p_p_zi(args); ok && -16 <= imm && imm <= 15 {
			templ := "0	0	1	0	0	1	0	1	size	0	imm5	1	0	0	Pg	Zn	0	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if imm < 0 {
				imm = (1 << 5) + imm
			}
			return assem_p_p_zi(templ, pd, pg, zn, "imm5", imm), 0, nil
		}
	case "cmpne":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	0	1	0	0	1	0	0	size	0	Zm	1	0	1	Pg	Zn	1	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		}
	case "cmphs", "cmpls":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	0	1	0	0	1	0	0	size	0	Zm	0	0	0	Pg	Zn	0	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if mnem == "cmpls" {
				// CMPLS <Pd>.<T>, <Pg>/Z, <Zm>.<T>, <Zn>.<T>
				// is equivalent to
				// CMPHS <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
				zn, zm = zm, zn // swap arguments
			}
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		}
	case "cmphi", "cmplo":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	0	1	0	0	1	0	0	size	0	Zm	0	0	0	Pg	Zn	1	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if mnem == "cmplo" {
				// CMPLO <Pd>.<T>, <Pg>/Z, <Zm>.<T>, <Zn>.<T>
				// is equivalent to
				// CMPHI <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
				zn, zm = zm, zn // swap arguments
			}
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		}
	case "cmpge", "cmple":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	0	1	0	0	1	0	0	size	0	Zm	1	0	0	Pg	Zn	0	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if mnem == "cmple" {
				// CMPLE <Pd>.<T>, <Pg>/Z, <Zm>.<T>, <Zn>.<T>
				// is equivalent to
				// CMPGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
				zn, zm = zm, zn // swap arguments
			}
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		}
	case "cmpgt", "cmplt":
		if ok, pd, pg, zn, zm, T := is_p_p_zz(args); ok {
			templ := "0	0	1	0	0	1	0	0	size	0	Zm	1	0	0	Pg	Zn	1	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if mnem == "cmplt" {
				// CMPLT <Pd>.<T>, <Pg>/Z, <Zm>.<T>, <Zn>.<T>
				// is equivalent to
				// CMPGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
				zn, zm = zm, zn // swap arguments
			}
			return assem_p_p_zz(templ, pd, pg, zn, zm), 0, nil
		} else if ok, pd, pg, zn, imm, T := is_p_p_zi(args); ok && -16 <= imm && imm <= 15 {
			templ := "0	0	1	0	0	1	0	1	size	0	imm5	0	0	0	Pg	Zn	1	Pd"
			if mnem == "cmplt" {
				templ = "0	0	1	0	0	1	0	1	size	0	imm5	0	0	1	Pg	Zn	0	Pd"
			}
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if imm < 0 {
				imm = (1 << 5) + imm
			}
			return assem_p_p_zi(templ, pd, pg, zn, "imm5", imm), 0, nil
		}
	case "ptest":
		if ok, pg, pn, _ := is_p_p(args); ok {
			templ := "0	0	1	0	0	1	0	1	op	S	0	1	0	0	0	0	1	1	Pg	0	Pn	0	opc2"
			// op	S	opc2
			// 0	1	0000	PTEST
			templ = strings.ReplaceAll(templ, "opc2", "0000")
			templ = strings.ReplaceAll(templ, "S", "1")
			templ = strings.ReplaceAll(templ, "op", "0")
			return assem_p_p(templ, pg, pn), 0, nil
		}
	case "pmullb", "pmullt":
		if ok, zd, zn, zm, Td, T := is_z_zz_2t(args); ok {
			templ := "0	1	0	0	0	1	0	1	0	0	0	Zm	0	1	1	0	1	0	Zn	Zd"
			if mnem == "pmullt" {
				templ = "0	1	0	0	0	1	0	1	0	0	0	Zm	0	1	1	0	1	1	Zn	Zd"
			}
			if Td == "q" && T == "d" {
				return assem_z_zz(templ, zd, zn, zm), 0, nil
			}
		}
	case "nop":
		templ := "1	1	0	1	0	1	0	1	0	0	0	0	0	0	1	1	0	0	1	0	0	0	0	0	0	0	0	1	1	1	1	1"
		templ = strings.ReplaceAll(templ, "\t", "")
		code, _ := strconv.ParseUint(templ, 2, 32)
		return uint32(code), 0, nil
	case "ret":
		templ := "1	1	0	1	0	1	1	0	0	1	0	1	1	1	1	1	0	0	0	0	0	0	Rn	0	0	0	0	0"
		templ = strings.ReplaceAll(templ, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(30, 2)))
		templ = strings.ReplaceAll(templ, "\t", "")
		code, _ := strconv.ParseUint(templ, 2, 32)
		return uint32(code), 0, nil
	case "index":
		if ok, zd, imm1, imm2, T := is_z_ii(args); ok && -16 <= imm1 && imm1 < 16 && -16 <= imm2 && imm2 < 16 {
			templ := "0	0	0	0	0	1	0	0	size	1	imm5b	0	1	0	0	0	0	imm5	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if imm1 < 0 {
				imm1 = (1 << 5) + imm1
			}
			if imm2 < 0 {
				imm2 = (1 << 5) + imm2
			}
			templ = strings.ReplaceAll(templ, "imm5b", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(imm2), 2)))
			return assem_z_i(templ, zd, "imm5", imm1), 0, nil
		} else if ok, zd, imm, rm, T := is_z_ir(args); ok && -16 <= imm && imm < 16 {
			templ := "0	0	0	0	0	1	0	0	size	1	Rm	0	1	0	0	1	0	imm5	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if imm < 0 {
				imm = (1 << 5) + imm
			}
			return assem_z_ir(templ, zd, "imm5", imm, rm), 0, nil
		} else if ok, zd, rn, imm, T := is_z_ri(args); ok && -16 <= imm && imm < 16 {
			templ := "0	0	0	0	0	1	0	0	size	1	imm5	0	1	0	0	0	1	Rn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			if imm < 0 {
				imm = (1 << 5) + imm
			}
			return assem_z_ri(templ, zd, rn, "imm5", imm), 0, nil
		}
	case "insr":
		if ok, zdn, rm, T := is_z_r(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	0	1	0	0	0	0	1	1	1	0	Rm	Zdn"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			templ = strings.ReplaceAll(templ, "Zdn", "Zd")
			templ = strings.ReplaceAll(templ, "Rm", "Rn")
			return assem_z_r(templ, zdn, rm), 0, nil
		}
	case "csel":
		if ok, rd, rn, rm, cond := is_r_rr_cond(args); ok {
			_ = cond
			templ := "sf	0	0	1	1	0	1	0	1	0	0	Rm	cond	0	0	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "cond", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(cond), 2)))

			return assem_r_rr(templ, rd, rn, rm, "", 0), 0, nil
		}
	}

	return 0, 0, fmt.Errorf("unhandled instruction: %s", ins)
}

func is_zeroing(predicate string) bool {
	return strings.HasSuffix(strings.ToUpper(predicate), "/Z")
}

func getR(r string) int {
	if len(r) > 0 && (r[0] == 'x' || r[0] == 'w') {
		if r[1:] == "zr" {
			return 31 // https://stackoverflow.com/questions/42788696/why-might-one-use-the-xzr-register-instead-of-the-literal-0-on-armv8
		} else if num, err := strconv.ParseInt(r[1:], 10, 32); err == nil && num < 32 {
			return int(num)
		}
	}
	return -1
}

func getCond(cond string) int {
	switch strings.ToLower(cond) {
	case "eq":
		return 0
	case "ne":
		return 1
	case "cs":
		return 2
	case "cc":
		return 3
	case "mi":
		return 4
	case "pl":
		return 5
	case "vs":
		return 6
	case "vc":
		return 7
	case "hi":
		return 8
	case "ls":
		return 9
	case "ge":
		return 10
	case "lt":
		return 11
	case "gt":
		return 12
	case "le":
		return 13
	case "al":
		return 14
	case "nv":
		return 15
	default:
		return -1
	}
}

func getP(r string) int {
	if len(r) > 0 && r[0] == 'p' {
		if num, err := strconv.ParseInt(r[1:], 10, 32); err == nil && num < 16 {
			return int(num)
		}
	}
	return -1
}

func getPdes(reg string) (_ int, T string) {
	if r := strings.Split(reg, ".")[0]; len(r) > 0 && r[0] == 'p' {
		if num, err := strconv.ParseInt(r[1:], 10, 32); err == nil && num < 16 {
			if len(strings.Split(reg, ".")) == 2 {
				T = strings.Split(reg, ".")[1]
				return int(num), T
			}
		}
	}
	return -1, ""
}

func getZ(reg string) (_ int, T string, index int) {
	if r := strings.Split(reg, ".")[0]; len(r) > 0 && r[0] == 'z' {
		if num, err := strconv.ParseInt(r[1:], 10, 32); err == nil {
			if len(strings.Split(reg, ".")) == 2 {
				T = strings.Split(reg, ".")[1]
				if len(T) > 1 && len(strings.Split(T, "[")) > 1 {
					indexNum := strings.ReplaceAll(strings.Split(T, "[")[1], "]", "")
					T = strings.Split(T, "[")[0]
					if inum, err := strconv.ParseInt(indexNum, 10, 32); err == nil {
						index = int(inum)
					}
				}
			}
			if int(num) < 32 {
				return int(num), T, index
			}
		}
	}
	return -1, "", -1
}

func getImm(imm string) (bool, int) {
	if len(imm) > 3 && imm[:3] == "#0x" {
		imm = imm[3:]
		if num, err := strconv.ParseUint(imm, 16, 64); err == nil {
			return true, int(num)
		}
	} else if len(imm) > 2 && imm[:2] == "#-" {
		imm = imm[1:]
		if num, err := strconv.ParseInt(imm, 10, 64); err == nil {
			return true, int(num)
		}
	} else if len(imm) > 0 && imm[0] == '#' {
		imm = imm[1:]
		if num, err := strconv.ParseUint(imm, 10, 64); err == nil {
			return true, int(num)
		}
	}
	return false, 0
}

// imms: imms is the number of bits **set**
// immr: immr is the number of bits to **rotate**
func getImm13(imms, immr uint32, T string) (imm13 uint32) {
	imm13 = (imms - 1)
	if T == "b" {
		imm13 |= ((8-immr)&7)<<6 | 0x30
	} else if T == "h" {
		imm13 |= ((16-immr)&15)<<6 | 0x20
	} else if T == "s" {
		imm13 |= ((32 - immr) & 31) << 6
	} else if T == "d" {
		imm13 |= 1<<12 | ((64-immr)&63)<<6
	} else {
		panic("unimplemented")
	}
	return
}

// find a (lsb, width) pair for BFC
// lsb must be in [0, 63], width must be in [1, 64 - lsb]
// return (0xffffffff, 0) if v is not a binary like 0...01...10...0
func parseBitfieldConst(v uint64) (lsb, width uint32) {
	// BFC is not applicable with zero
	if v != 0 {
		// find the lowest set bit, for example l=2 for 0x3ffffffc
		lsb = uint32(bits.TrailingZeros64(v))
		// m-1 represents the highest set bit index, for example m=30 for 0x3ffffffc
		m := 64 - uint32(bits.LeadingZeros64(v))
		// check if v is a binary like 0...01...10...0
		if uint64(1<<m)-(1<<lsb) == v {
			// it must be m > l for non-zero v
			return lsb, m - lsb
		}
	}
	// invalid
	return 0xffffffff, 0
}

func getSizeFromType(T string) string {
	switch strings.ToUpper(T) {
	case "B":
		return "00"
	case "H":
		return "01"
	case "S":
		return "10"
	case "D":
		return "11"
	default:
		fmt.Println("Invalid type: ", T)
		return ""
	}
}

func getTypeSpecifier(T string) string {
	switch strings.ToUpper(T) {
	// 00000	RESERVED
	case "B":
		return "00001" // xxxx1: B
	case "H":
		return "00010" // xxx10: H
	case "S":
		return "00100" // xx100: S
	case "D":
		return "01000" // x1000: D
	case "Q":
		return "10000" // 10000: Q
	default:
		fmt.Println("Invalid type: ", T)
		return ""
	}
}

func getShift(in string) int {
	switch strings.ToUpper(in) {
	case "LSL":
		return 0
	case "LSR":
		return 1
	case "ASR":
		return 2
	// case "RESERVED"
	// return 3
	default:
		fmt.Println("Invalid shift type: ", in)
		return -1
	}
}

func computeShiftSpecifier(imm uint, reverse bool, T string) (int, string) {
	switch strings.ToUpper(T) {
	case "B":
		const esize = 8
		if imm < esize {
			if reverse {
				imm = esize - imm
			}
			return int(imm), "0001"
		}
	case "H":
		const esize = 16
		if imm < esize {
			if reverse {
				imm = esize - imm
			}
			return int(imm & 7), fmt.Sprintf("001%01b", imm>>3)
		}
	case "S":
		const esize = 32
		if imm < esize {
			if reverse {
				imm = esize - imm
			}
			return int(imm & 7), fmt.Sprintf("01%02b", imm>>3)
		}
	case "D":
		const esize = 64
		if imm < esize {
			if reverse {
				imm = esize - imm
			}
			return int(imm & 7), fmt.Sprintf("1%03b", imm>>3)
		}
	}
	panic(fmt.Sprintf("computeTypeSpecifier: invalid immediate %d and %s combination", imm, T))
}

func is_p(args []string) (ok bool, pd int, T string) {
	if len(args) >= 1 {
		pd, T = getPdes(args[0])
		if pd != -1 {
			return true, pd, T
		}
	}
	return false, -1, ""
}

func is_rr(args []string) (ok bool, rn, rm int) {
	if len(args) == 2 {
		rn, rm = getR(args[0]), getR(args[1])
		if rn != -1 && rm != -1 {
			return true, rn, rm
		}
	}
	return
}

func is_r_rr(args []string) (ok bool, rd, rn, rm, shift, imm int) {
	if len(args) == 3 {
		rd, rn, rm = getR(args[0]), getR(args[1]), getR(args[2])
		if rd != -1 && rn != -1 && rm != -1 {
			return true, rd, rn, rm, 0, 0
		}
	} else if len(args) == 5 {
		rd, rn, rm = getR(args[0]), getR(args[1]), getR(args[2])
		shift = getShift(args[3])
		ok, imm = getImm(args[4])
		return
	}
	return
}

func is_r_rr_cond(args []string) (ok bool, rd, rn, rm, cond int) {
	if len(args) == 4 {
		rd, rn, rm = getR(args[0]), getR(args[1]), getR(args[2])
		cond = getCond(args[3])
		if rd != -1 && rn != -1 && rm != -1 && cond != -1 {
			return true, rd, rn, rm, cond
		}
	}
	return
}

func is_r_rrr(args []string) (ok bool, rd, rn, rm, ra int) {
	if len(args) == 4 {
		rd, rn, rm, ra = getR(args[0]), getR(args[1]), getR(args[2]), getR(args[3])
		if rd != -1 && rn != -1 && rm != -1 && ra != -1 {
			return true, rd, rn, rm, ra
		}
	}
	return
}

func is_r_r(args []string) (ok bool, rd, rn int) {
	if len(args) >= 2 {
		rd, rn = getR(args[0]), getR(args[1])
		if rd != -1 && rn != -1 {
			return true, rd, rn
		}
	}
	return
}

func is_r_ri(args []string) (ok bool, rd, rn, imm, shift int) {
	if len(args) >= 3 {
		rd, rn = getR(args[0]), getR(args[1])
		if rd != -1 && rn != -1 {
			if ok, imm := getImm(args[2]); ok {
				if len(args) >= 4 && args[3] == "LSL" && args[4] == "#12" {
					return true, rd, rn, imm, 1
				}
				return true, rd, rn, imm, 0
			}
		}
	}
	return
}

func is_r_bi(args []string) (ok bool, rt, xn, imm int) {
	if len(args) >= 2 {
		rt = getR(args[0])
		if rt != -1 && args[1][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
			if xn, imm = getMemAddrImm(args[1:]); xn != -1 {
				return true, rt, xn, imm
			}
		}
	}
	return
}

func is_r_br(args []string) (ok bool, rt, rn, rm, option, amount int) {
	if len(args) >= 2 {
		rt = getR(args[0])
		if rt != -1 && args[1][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
			if rn, rm, option, amount = getMemAddrRegister(args[1:]); rn != -1 && rm != -1 {
				return true, rt, rn, rm, option, amount
			}
		}
	}
	return
}

func is_z_z(args []string) (ok bool, zd, zn int, T string) {
	if len(args) == 2 {
		var t1, t2 string
		zd, t1, _ = getZ(args[0])
		zn, t2, _ = getZ(args[1])
		if zd != -1 && zn != -1 && t1 == t2 {
			return true, zd, zn, t1
		}
	}
	return
}

func is_z_zz(args []string) (ok bool, zd, zn, zm int, T string) {
	if len(args) == 3 {
		var t1, t2, t3 string
		zd, t1, _ = getZ(args[0])
		zn, t2, _ = getZ(args[1])
		zm, t3, _ = getZ(args[2])
		if zd != -1 && zn != -1 && zm != -1 && t1 == t2 && t2 == t3 {
			return true, zd, zn, zm, t1
		}
	}
	return
}

func is_z_zi(args []string) (ok bool, zd, zn, imm, shift int, T string) {
	if len(args) >= 3 {
		var t2 string
		zd, T, _ = getZ(args[0])
		zn, t2, _ = getZ(args[1])
		if zd != -1 && zn != -1 && T == t2 {
			if ok, imm := getImm(args[2]); ok {
				if len(args) >= 4 && args[3] == "LSL" && args[4] == "#8" {
					return true, zd, zn, imm, 1, T
				}
				return true, zd, zn, imm, 0, T
			}
		}
	}
	return
}

func is_z_zz_2t(args []string) (ok bool, zd, zn, zm int, Td, T string) {
	if len(args) == 3 {
		var td, t2, t3 string
		zd, td, _ = getZ(args[0])
		zn, t2, _ = getZ(args[1])
		zm, t3, _ = getZ(args[2])
		if zd != -1 && zn != -1 && zm != -1 && t2 == t3 {
			return true, zd, zn, zm, td, t2
		}
	}
	return
}

func is_z_p_z(args []string) (ok bool, zd, pg, zn int, T string) {
	if len(args) == 3 {
		var t1, t2 string
		zd, t1, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])

		if zd != -1 && zn != -1 && pg != -1 && t1 == t2 {
			return true, zd, pg, zn, t1
		}
	}
	return
}

func is_z_p_z_tt(args []string) (ok bool, zd, pg, zn int, Td, Tn string) {
	if len(args) == 3 {
		zd, Td, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, Tn, _ = getZ(args[2])

		if zd != -1 && zn != -1 && pg != -1 {
			return true, zd, pg, zn, Td, Tn
		}
	}
	return
}

func is_p_p(args []string) (ok bool, pg, pn int, T string) {
	if len(args) == 2 {
		pg = getP(args[0])
		if pg == -1 {
			pg, T = getPdes(args[0])
		}
		pn = getP(args[1])
		var T2 string
		if pn == -1 {
			pn, T2 = getPdes(args[1])
		}
		if pg != -1 && pn != -1 &&
			(T == T2 || T == "" && strings.ToLower(T2) == "b" /* for ptest p4, p5.b */) {
			return true, pg, pn, T
		}
	}
	return
}

func is_p_p_zz(args []string) (ok bool, pd, pg, zn, zm int, T string) {
	if len(args) == 4 {
		var t1, t2, t3 string
		pd, t1 = getPdes(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])
		zm, t3, _ = getZ(args[3])
		if pd != -1 && pg != -1 && zn != -1 && zm != -1 && t1 == t2 && t2 == t3 {
			return true, pd, pg, zn, zm, t2
		}
	}
	return
}

func is_p_p_zi(args []string) (ok bool, pd, pg, zn, imm int, T string) {
	if len(args) == 4 {
		var t1, t2 string
		pd, t1 = getPdes(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])
		if ok, imm := getImm(args[3]); ok {
			if pd != -1 && pg != -1 && zn != -1 && t1 == t2 {
				return true, pd, pg, zn, imm, t1
			}
		}
	}
	return
}

func is_z_p_zz(args []string) (ok bool, zdn, pg, zm int, T string) {
	if len(args) == 4 {
		var t1, t2, t3 string
		var zn int
		zdn, t1, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])
		zm, t3, _ = getZ(args[3])

		if zdn == zn && zdn != -1 && zn != -1 && zm != -1 && pg != -1 && t1 == t2 && t2 == t3 {
			return true, zdn, pg, zm, t1
		}
	}
	return
}

func is_z2_p_zz(args []string) (ok bool, zdn, pg, zm, za int, T string) {
	if len(args) == 4 {
		var t1, t2, t3 string
		zdn, t1, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zm, t2, _ = getZ(args[2])
		za, t3, _ = getZ(args[3])
		if zdn != -1 && zm != -1 && za != -1 && pg != -1 && t1 == t2 && t2 == t3 {
			return true, zdn, pg, zm, za, t1
		}
	}
	return
}

func assem_prefixed_z_z(ins string, zd, zn int) (opcode, opcode2 uint32, err error) {
	//
	// insert 'MOVPRFX (unpredicated)' instruction
	//
	templ := "0	0	0	0	0	1	0	0	0	0	1	0	0	0	0	0	1	0	1	1	1	1	Zn	Zd"
	prfx := assem_z_z(templ, zd, zn)
	if ins == "" {
		return prfx, 0, nil // we're just assembling a prefix instruction
	}

	insPatched := strings.ReplaceAll(ins, fmt.Sprintf("z%d.", zn), fmt.Sprintf("z%d.", zd))
	if oc, oc2, err := Assemble(insPatched); err == nil && oc2 == 0 {
		return prfx, oc, nil
	}
	return 0, 0, fmt.Errorf("unhandled 'MOVPRFX (unpredicated)' instruction: %s", ins)
}

func is_prefixed_z_p_zz(args []string) (ok bool, zd, pg, zn, zm int, T string) {
	if len(args) == 4 {
		var t1, t2, t3 string
		zd, t1, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])
		zm, t3, _ = getZ(args[3])

		if zd != -1 && zn != -1 && zm != -1 && pg != -1 && t1 == t2 && t2 == t3 {
			return true, zd, pg, zn, zm, t1
		}
	}
	return
}

func is_z_p_zz2(args []string) (ok bool, zda, pg, zn, zm int, T string) {
	if len(args) == 4 {
		var t1, t2, t3 string
		zda, t1, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])
		zm, t3, _ = getZ(args[3])

		if zda != -1 && zn != -1 && zm != -1 && pg != -1 && t1 == t2 && t2 == t3 {
			return true, zda, pg, zn, zm, t1
		}
	}
	return
}

func is_z_p_zz_4(args []string) (ok bool, zd, pg, zn, zm int, T string) {
	if len(args) == 4 {
		var t1, t2, t3 string
		zd, t1, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])
		zm, t3, _ = getZ(args[3])
		if zd != -1 && zn != -1 && zm != -1 && pg != -1 && t1 == t2 && t2 == t3 {
			return true, zd, pg, zn, zm, t1
		}
	}
	return
}

func is_z_zzz(args []string) (ok bool, zd, zn, zm, za int, T string) {
	if len(args) == 4 {
		var t1, t2, t3, t4 string
		zd, t1, _ = getZ(args[0])
		zn, t2, _ = getZ(args[1])
		zm, t3, _ = getZ(args[2])
		za, t4, _ = getZ(args[3])
		if zd != -1 && zn != -1 && zm != -1 && za != -1 && t1 == t2 && t2 == t3 && t3 == t4 {
			return true, zd, zn, zm, za, t1
		}
	}
	return
}

func is_z_zz_z(args []string) (ok bool, zd, zn1, zn2, zm int, T string) {
	if len(args) == 6 && args[1] == "{" && args[4] == "}" {
		var t1, t2, t3, t4 string
		zd, t1, _ = getZ(args[0])
		zn1, t2, _ = getZ(args[2])
		zn2, t3, _ = getZ(args[3])
		zm, t4, _ = getZ(args[5])
		if zd != -1 && zn1 != -1 && zn2 != -1 && zm != -1 && t1 == t2 && t2 == t3 && t3 == t4 {
			return true, zd, zn1, zn2, zm, t1
		}
	}
	return
}

func is_z_i(args []string) (ok bool, zd, imm int, T string) {
	if len(args) == 2 {
		var t1 string
		zd, t1, _ = getZ(args[0])
		if zd != -1 {
			if ok, imm := getImm(args[1]); ok {
				return true, zd, imm, t1
			}
		}
	}
	return
}

func is_z_ii(args []string) (ok bool, zd, imm1, imm2 int, T string) {
	if len(args) == 3 {
		var t1 string
		zd, t1, _ = getZ(args[0])
		if zd != -1 {
			if ok, imm1 := getImm(args[1]); ok {
				if ok, imm2 := getImm(args[2]); ok {
					return true, zd, imm1, imm2, t1
				}
			}
		}
	}
	return
}

func is_z_zimm(args []string) (ok bool, zd, zn, imm int, T string) {
	if len(args) == 3 {
		var t1, t2 string
		zd, t1, _ = getZ(args[0])
		zn, t2, _ = getZ(args[1])
		if zd != -1 && zn != -1 && t1 == t2 {
			if ok, imm := getImm(args[2]); ok {
				return true, zd, zn, imm, t1
			}
		}
	}
	return
}

func is_z_p_zimm(args []string) (ok bool, zd, pg, zn, imm int, T string) {
	if len(args) == 4 {
		var t1, t2 string
		zd, t1, _ = getZ(args[0])
		pg = getP(strings.Split(args[1], "/")[0]) // drop any trailer
		zn, t2, _ = getZ(args[2])
		if zd != -1 && pg != -1 && zn != -1 && t1 == t2 {
			if ok, imm := getImm(args[3]); ok {
				return true, zd, pg, zn, imm, t1
			}
		}
	}
	return
}

func is_z_zindexed(args []string) (ok bool, zd, zn, index int, T string) {
	if len(args) == 2 {
		var t1, t2 string
		zd, t1, _ = getZ(args[0])
		zn, t2, index = getZ(args[1])
		if zd != -1 && zn != -1 && t1 == t2 {
			return true, zd, zn, index, t1
		}
	}
	return
}

func is_z_r(args []string) (ok bool, zd, rn int, T string) {
	if len(args) == 2 {
		zd, T, _ = getZ(args[0])
		rn = getR(args[1])
		if zd != -1 && rn != -1 {
			return true, zd, rn, T
		}
	}
	return
}

func is_z_ir(args []string) (ok bool, zd, imm, rm int, T string) {
	if len(args) == 3 {
		zd, T, _ = getZ(args[0])
		if ok, imm := getImm(args[1]); ok {
			rm = getR(args[2])
			if zd != -1 && rm != -1 {
				return true, zd, imm, rm, T
			}
		}
	}
	return
}

func is_z_ri(args []string) (ok bool, zd, rn, imm int, T string) {
	if len(args) == 3 {
		zd, T, _ = getZ(args[0])
		rn = getR(args[1])
		if ok, imm := getImm(args[2]); ok {
			if zd != -1 && rn != -1 {
				return true, zd, rn, imm, T
			}
		}
	}
	return
}

func is_r_i(args []string) (ok bool, rd int, imm, shift int) {
	if len(args) == 2 || len(args) == 4 {
		rd = getR(args[0])
		if ok, imm := getImm(args[1]); ok {
			if len(args) == 4 && args[2] == "lsl" {
				if ok, sh := getImm(args[3]); ok {
					if sh == 0 || sh == 12 {
						return true, rd, imm, sh
					}
				}
			} else {
				return true, rd, imm, 0
			}
		}
	}
	return
}

func getMemAddrImm(args []string) (xn, imm int) {
	if args[0][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
		memaddr := strings.Join(args[0:], ", ")
		memaddr = strings.NewReplacer("[", "", "]", "", "MUL, VL", "MUL VL").Replace(memaddr)
		mas := strings.Split(memaddr, ", ")
		if len(mas) >= 1 {
			xn = getR(mas[0])
			if len(mas) == 3 && mas[2] == "MUL VL" {
				if ok, imm := getImm(mas[1]); ok && xn != -1 {
					return xn, imm
				}
			} else if len(mas) == 2 && mas[1][0] == '#' {
				if ok, imm := getImm(mas[1]); ok && xn != -1 {
					return xn, imm
				}
			} else if xn != -1 {
				return xn, 0
			}
		}
	}
	return -1, 0
}

func getMemAddrRegister(args []string) (rn, rm, option, amount int) {
	if args[0][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
		memaddr := strings.Join(args[0:], ", ")
		memaddr = strings.NewReplacer("[", "", "]", "").Replace(memaddr)
		mas := strings.Split(memaddr, ", ")
		if len(mas) >= 2 {
			rn = getR(mas[0])
			if rn != -1 {
				rm = getR(mas[1])
				if rm != -1 {
					mas = mas[2:]
					if len(mas) == 2 && strings.ToLower(mas[0]) == "lsl" {
						if ok, imm := getImm(mas[1]); ok {
							// option	<extend>
							// 010	UXTW
							// 011	LSL
							// 110	SXTW
							// 111	SXTX
							option, amount = 0b011, imm
							return rn, rm, option, amount
						}
					}
				}
			}
		}
	}
	return -1, -1, -1, -1
}

func getMemAddrVectored(args []string) (rn, zm, xs int, T string) {
	if args[0][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
		memaddr := strings.Join(args[0:], ", ")
		memaddr = strings.NewReplacer("[", "", "]", "", "MUL, VL", "MUL VL").Replace(memaddr)
		mas := strings.Split(memaddr, ", ")
		if len(mas) >= 1 {
			rn = getR(mas[0])
			var tm string
			zm, tm, _ = getZ(mas[1])
			xs := 0
			if len(mas) > 2 && strings.ToUpper(mas[2]) == "SXTW" {
				xs = 1
			}
			if rn != -1 && zm != -1 {
				return rn, zm, xs, tm
			}
		}
	}
	return -1, 0, 0, ""
}

func is_z_bi(args []string) (ok bool, zt, xn, imm int) {
	if len(args) > 1 {
		zt, _, _ = getZ(args[0])
		if zt != -1 && args[1][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
			memaddr := strings.Join(args[1:], ", ")
			memaddr = strings.NewReplacer("[", "", "]", "", "MUL, VL", "MUL VL").Replace(memaddr)
			mas := strings.Split(memaddr, ", ")
			if len(mas) >= 1 {
				xn = getR(mas[0])
				if len(mas) == 3 && mas[2] == "MUL VL" {
					if ok, imm := getImm(mas[1]); ok && xn != -1 {
						return true, zt, xn, imm
					}
				} else if len(mas) == 1 && xn != -1 {
					return true, zt, xn, 0
				}
			}
		}
	}
	return
}

func is_p_bi(args []string) (ok bool, pt, xn, imm int) {
	if len(args) > 1 {
		pt = getP(args[0])
		if pt != -1 && args[1][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
			memaddr := strings.Join(args[1:], ", ")
			memaddr = strings.NewReplacer("[", "", "]", "", "MUL, VL", "MUL VL").Replace(memaddr)
			mas := strings.Split(memaddr, ", ")
			if len(mas) >= 1 {
				xn = getR(mas[0])
				if len(mas) == 3 && mas[2] == "MUL VL" {
					if ok, imm := getImm(mas[1]); ok && xn != -1 {
						return true, pt, xn, imm
					}
				} else if xn != -1 {
					return true, pt, xn, 0
				}
			}
		}
	}
	return
}

func is_z_p_bz(args []string) (ok bool, zt, pg, rn, zm, xs int, T string) {
	if len(args) == 7 && args[0] == "{" && args[2] == "}" {
		zt, T, _ = getZ(args[1])
		pg = getP(strings.Split(args[3], "/")[0]) // drop any trailer
		if zt != -1 && pg != -1 && args[4][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
			var tm string
			if rn, zm, xs, tm = getMemAddrVectored(args[4:]); rn != -1 && T == tm {
				return true, zt, pg, rn, zm, xs, T
			}
		}
	}
	return
}

func is_zt4_p_rr(args []string) (ok bool, zt, pg, rn, rm int, T string) {
	if len(args) == 9 && args[0] == "{" && args[5] == "}" {
		zt, T, _ = getZ(args[1])
		for a := 2; a < 5; a++ { // check Z-registers are consecutive
			ztNext, TNext, _ := getZ(args[a])
			if ztNext != zt+1+(a-2) || T != TNext {
				return false, 0, 0, 0, 0, ""
			}
		}
		pg = getP(strings.Split(args[6], "/")[0]) // drop any trailer
		rn = getR(strings.ReplaceAll(args[7], "[", ""))
		rm = getR(strings.ReplaceAll(args[8], "]", ""))
		ok = true
	}
	return
}

func is_z_p_bi(args []string) (ok bool, zt, pg, rn, imm int, T string) {
	if len(args) >= 4 && args[0] == "{" && args[2] == "}" {
		zt, T, _ = getZ(args[1])
		pg = getP(strings.Split(args[3], "/")[0]) // drop any trailer
		if zt != -1 && pg != -1 && args[4][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
			if rn, imm = getMemAddrImm(args[4:]); rn != -1 {
				return true, zt, pg, rn, imm, T
			}
		}
	}
	return
}

func is_z_p_rr(args []string) (ok bool, zt, pg, rn, rm, shift int, T string) {
	if len(args) == 8 && args[0] == "{" && args[2] == "}" && strings.ToLower(args[6]) == "lsl" && (args[7] == "#3]" || args[7] == "#2]") {
		zt, T, _ = getZ(args[1])
		pg = getP(strings.Split(args[3], "/")[0]) // drop any trailer
		rn = getR(strings.ReplaceAll(args[4], "[", ""))
		rm = getR(args[5])
		rplc := strings.NewReplacer("#", "", "]", "")
		var err error
		if shift, err = strconv.Atoi(rplc.Replace(args[7])); err != nil {
			return false, -1, -1, -1, -1, -1, ""
		}
		return true, zt, pg, rn, rm, shift, T
	}
	return
}

func assem_prefixed_z_p_z(ins, arg_1 string, zd, pg, zn int, T string) (opcode, opcode2 uint32, err error) {
	//
	// insert 'MOVPRFX (predicated)' instruction
	//
	templ := "0	0	0	0	0	1	0	0	size	0	1	0	0	0	M	0	0	1	Pg	Zn	Zd"
	templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
	if strings.Contains(strings.ToUpper(arg_1), "/M") {
		templ = strings.ReplaceAll(templ, "M", "1")
	} else if strings.Contains(strings.ToUpper(arg_1), "/Z") {
		templ = strings.ReplaceAll(templ, "M", "0")
	} else {
		return 0, 0, fmt.Errorf("unhandled (prefixed) instruction: %s", ins)
	}

	prfx := assem_z_p_z(templ, zd, pg, zn)
	if ins == "" {
		return prfx, 0, nil // we're just assembling a prefix instruction
	}

	// Make sure we hit "Zdn" path by setting Zd == Zn
	insPatched := strings.ReplaceAll(ins, fmt.Sprintf("z%d.", zn), fmt.Sprintf("z%d.", zd))
	// "/Z" is handled via MOVPRFX (see above), so always use merging behaviour
	rplc := strings.NewReplacer("/Z", "/M", "/z", "/m")
	insPatched = rplc.Replace(insPatched)
	if oc, oc2, err := Assemble(insPatched); err == nil && oc2 == 0 {
		return prfx, oc, nil
	}
	return 0, 0, fmt.Errorf("unhandled 'MOVPRFX (predicated)' instruction: %s", ins)
}

func assem_p(template string, pd int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Pd", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pd), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_r_rr(template string, rd, rn, rm int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Rd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rd), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	opcode = strings.ReplaceAll(opcode, "Rm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rm), 2)))
	switch immPttrn {
	case "":
		// ignore
	case "imm6":
		opcode = strings.ReplaceAll(opcode, "imm6", fmt.Sprintf("%0*s", 6, strconv.FormatUint(uint64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_r_ri(template string, rd, rn int, immPttrn string, imm, shift int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Rd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rd), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	switch immPttrn {
	case "imm6":
		opcode = strings.ReplaceAll(opcode, "imm6", fmt.Sprintf("%0*s", 6, strconv.FormatUint(uint64(imm), 2)))
	case "imm12":
		opcode = strings.ReplaceAll(opcode, "imm12", fmt.Sprintf("%0*s", 12, strconv.FormatInt(int64(imm), 2)))
	case "imm13":
		opcode = strings.ReplaceAll(opcode, "imm13", fmt.Sprintf("%0*s", 13, strconv.FormatInt(int64(imm), 2)))
	case "immr":
		opcode = strings.ReplaceAll(opcode, "immr", fmt.Sprintf("%0*s", 6, strconv.FormatInt(int64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "sh", fmt.Sprintf("%0*s", 1, strconv.FormatInt(int64(shift), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_r_rrr(template string, rd, rn, rm, ra int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Rd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rd), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	opcode = strings.ReplaceAll(opcode, "Rm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rm), 2)))
	opcode = strings.ReplaceAll(opcode, "Ra", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(ra), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_z(template string, zd, zn int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_zz(template string, zd, zn, zm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z2_zz(template string, zdn, zm, zk int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zdn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zdn), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "Zk", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zk), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z2_p_zz(template string, zdn, pg, zm, za int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zdn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zdn), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "Za", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(za), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_zz2(template string, zda, zn, zm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zda", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zda), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_i(template string, zd int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	switch immPttrn {
	case "imm5":
		opcode = strings.ReplaceAll(opcode, "imm5", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(imm), 2)))
	case "imm8":
		opcode = strings.ReplaceAll(opcode, "imm8", fmt.Sprintf("%0*s", 8, strconv.FormatUint(uint64(imm), 2)))
	case "imm13":
		opcode = strings.ReplaceAll(opcode, "imm13", fmt.Sprintf("%0*s", 13, strconv.FormatUint(uint64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_ir(template string, zd int, immPttrn string, imm, rm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	switch immPttrn {
	case "imm5":
		opcode = strings.ReplaceAll(opcode, "imm5", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "Rm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_ri(template string, zd, rn int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	switch immPttrn {
	case "imm5":
		opcode = strings.ReplaceAll(opcode, "imm5", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_zi(template string, zd, zn int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	switch immPttrn {
	case "imm2":
		opcode = strings.ReplaceAll(opcode, "imm2", fmt.Sprintf("%0*s", 2, strconv.FormatUint(uint64(imm), 2)))
	case "imm3":
		opcode = strings.ReplaceAll(opcode, "imm3", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_p_zi(template string, zdn, pg int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zdn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zdn), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	switch immPttrn {
	case "imm2":
		opcode = strings.ReplaceAll(opcode, "imm2", fmt.Sprintf("%0*s", 2, strconv.FormatUint(uint64(imm), 2)))
	case "imm3":
		opcode = strings.ReplaceAll(opcode, "imm3", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_r(template string, zd, rn int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_r_i(template string, rd int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Rd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rd), 2)))
	switch immPttrn {
	case "imm6":
		opcode = strings.ReplaceAll(opcode, "imm6", fmt.Sprintf("%0*s", 6, strconv.FormatUint(uint64(imm), 2)))
	case "imm12":
		opcode = strings.ReplaceAll(opcode, "imm12", fmt.Sprintf("%0*s", 12, strconv.FormatUint(uint64(imm), 2)))
	case "imm16":
		opcode = strings.ReplaceAll(opcode, "imm16", fmt.Sprintf("%0*s", 16, strconv.FormatUint(uint64(imm), 2)))
	case "immhi":
		opcode = strings.ReplaceAll(opcode, "immhi", fmt.Sprintf("%0*s", 19, strconv.FormatUint(uint64(imm), 2)))
	default:
		fmt.Println("Invalid immediate pattern: ", immPttrn)
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_bi(template string, zt, xn, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zt", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zt), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(xn), 2)))
	const bits = 9
	if imm < 0 {
		imm = (1 << bits) + imm
	}
	immstr := fmt.Sprintf("%0*s", bits, strconv.FormatInt(int64(imm), 2))
	opcode = strings.ReplaceAll(opcode, "imm9h", immstr[:6])
	opcode = strings.ReplaceAll(opcode, "imm9l", immstr[6:])
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_p_bi(template string, pt, xn, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Pt", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pt), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(xn), 2)))
	const bits = 9
	if imm < 0 {
		imm = (1 << bits) + imm
	}
	immstr := fmt.Sprintf("%0*s", bits, strconv.FormatInt(int64(imm), 2))
	opcode = strings.ReplaceAll(opcode, "imm9h", immstr[:6])
	opcode = strings.ReplaceAll(opcode, "imm9l", immstr[6:])
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_p_bz(template string, zt, pg, rn, zm, xs int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zt", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zt), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "xs", fmt.Sprintf("%0*s", 1, strconv.FormatUint(uint64(xs), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_p_bi(template string, zt, pg, rn int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zt", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zt), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	switch immPttrn {
	case "imm4":
		opcode = strings.ReplaceAll(opcode, "imm4", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(imm), 2)))
	case "imm6":
		opcode = strings.ReplaceAll(opcode, "imm6", fmt.Sprintf("%0*s", 6, strconv.FormatUint(uint64(imm), 2)))
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_p_rr(template string, zt, pg, rn, rm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zt", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zt), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	opcode = strings.ReplaceAll(opcode, "Rm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_zt4_p_rr(template string, zt, png, rn, rm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zt", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(zt), 2)))
	opcode = strings.ReplaceAll(opcode, "PNg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(png-8), 2)))
	opcode = strings.ReplaceAll(opcode, "Rn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rn), 2)))
	opcode = strings.ReplaceAll(opcode, "Rm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(rm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z2_p_z(template string, zdn, pg, zm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zdn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zdn), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_p_p(template string, pg, pn int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Pn", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pn), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_p_p_p_p(template string, pd, pg, pn, pm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Pd", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pd), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Pn", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pn), 2)))
	opcode = strings.ReplaceAll(opcode, "Pm", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_p_p_zz(template string, pd, pg, zn, zm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Pd", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pd), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_p_p_zi(template string, pd, pg, zn int, immPttrn string, imm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Pd", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(pd), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	switch immPttrn {
	case "imm5":
		opcode = strings.ReplaceAll(opcode, "imm5", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(imm), 2)))
	}
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_p_zz(template string, zda, pg, zn, zm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zda", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zda), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_p_zz_4(template string, zd, p, zn, zm int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	if strings.Contains(opcode, "Pv") {
		opcode = strings.ReplaceAll(opcode, "Pv", fmt.Sprintf("%0*s", 4, strconv.FormatUint(uint64(p), 2)))
	} else {
		opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(p), 2)))
	}
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	opcode = strings.ReplaceAll(opcode, "Zm", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zm), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}

func assem_z_p_z(template string, zd, pg, zn int) uint32 {
	opcode := template
	opcode = strings.ReplaceAll(opcode, "Zd", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zd), 2)))
	opcode = strings.ReplaceAll(opcode, "Pg", fmt.Sprintf("%0*s", 3, strconv.FormatUint(uint64(pg), 2)))
	opcode = strings.ReplaceAll(opcode, "Zn", fmt.Sprintf("%0*s", 5, strconv.FormatUint(uint64(zn), 2)))
	opcode = strings.ReplaceAll(opcode, "\t", "")
	if code, err := strconv.ParseUint(opcode, 2, 32); err != nil {
		panic(err)
	} else {
		return uint32(code)
	}
}
