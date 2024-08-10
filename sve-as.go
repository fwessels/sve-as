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
)

func Assemble(ins string) (opcode uint32, err error) {
	mnem := strings.Fields(ins)[0]
	args := strings.Fields(ins)[1:]
	for i := range args {
		args[i] = strings.TrimSpace(strings.ReplaceAll(args[i], ",", ""))
	}

	switch mnem {
	case "add":
		if ok, rd, rn, rm := is_r_rr(args); ok {
			templ := "sf	0	0	0	1	0	1	1	shift	0	Rm	imm6	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "shift", "0\t0")
			return assem_r_rr(templ, rd, rn, rm, "imm6", 0), nil
		} else if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			templ := "sf	0	0	1	0	0	0	1	0	sh	imm12	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_ri(templ, rd, rn, "imm12", imm, shift), nil
		}
	case "udiv":
		if ok, rd, rn, rm := is_r_rr(args); ok {
			templ := "sf	0	0	1	1	0	1	0	1	1	0	Rm	0	0	0	0	1	0	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_rr(templ, rd, rn, rm, "", 0), nil
		}
	case "subs":
		if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			templ := "sf	1	1	1	0	0	0	1	0	sh	imm12	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			return assem_r_ri(templ, rd, rn, "imm12", imm, shift), nil
		}
	case "addvl":
		if ok, rd, rn, imm, shift := is_r_ri(args); ok {
			templ := "0	0	0	0	0	1	0	0	0	0	1	Rn	0	1	0	1	0	imm6	Rd"
			return assem_r_ri(templ, rd, rn, "imm6", imm, shift), nil
		}
	case "tst":
		if ok, rn, rm := is_rr(args); ok {
			// equivalent to "ands xzr, <xn>, <xm>{, <shift> #<amount>}"
			templ := "sf	1	1	0	1	0	1	0	shift	0	Rm	imm6	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "shift", "0\t0")
			rd := getR("xzr")
			return assem_r_rr(templ, rd, rn, rm, "imm6", 0), nil
		}
	case "and":
		if ok, zd, zn, zm, _ := is_z_zz(args); ok {
			return assem_z_zz("0	0	0	0	0	1	0	0	0	0	1	Zm	0	0	1	1	0	0	Zn	Zd", zd, zn, zm), nil
		}
	case "eor":
		if ok, zd, zn, zm, _ := is_z_zz(args); ok {
			return assem_z_zz("0	0	0	0	0	1	0	0	1	0	1	Zm	0	0	1	1	0	0	Zn	Zd", zd, zn, zm), nil
		}
	case "tbl":
		if ok, zd, zn, zm, T := is_z_zz(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	Zm	0	0	1	1	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_zz(templ, zd, zn, zm), nil
		}
	case "dup":
		if ok, zd, zn, imm, T := is_z_zindexed(args); ok {
			templ := "0	0	0	0	0	1	0	1	imm2	1	tsz	0	0	1	0	0	0	Zn	Zd"
			templ = strings.ReplaceAll(templ, "tsz", getTypeSpecifier(T))
			return assem_z_zi(templ, zd, zn, "imm2", imm), nil
		}
	case "mov":
		if ok, zd, rn, T := is_z_r(args); ok {
			templ := "0	0	0	0	0	1	0	1	size	1	0	0	0	0	0	0	0	1	1	1	0	Rn	Zd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			return assem_z_r(templ, zd, rn), nil
		} else if ok, rd, imm := is_r_i(args); ok {
			// Using MOV (wide immediate) here (which is an alias for MOVZ)
			templ := "sf	1	0	1	0	0	1	0	1	hw	imm16	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "hw", "0\t0")
			return assem_r_i(templ, rd, "imm16", imm), nil
		}
	case "ldr":
		if ok, zt, xn, imm := is_z_bi(args); ok {
			templ := "1	0	0	0	0	1	0	1	1	0	imm9h	0	1	0	imm9l	Rn	Zt"
			return assem_z_bi(templ, zt, xn, imm), nil
		}
	case "str":
		if ok, zt, xn, imm := is_z_bi(args); ok {
			templ := "1	1	1	0	0	1	0	1	1	0	imm9h	0	1	0	imm9l	Rn	Zt"
			return assem_z_bi(templ, zt, xn, imm), nil
		}
	case "ld1d":
		if ok, zt, pg, rn, rm := is_z_p_rr(args); ok {
			templ := "1	0	1	0	0	1	0	1	1	1	1	Rm	0	1	0	Pg	Rn	Zt"
			return assem_z_p_rr(templ, zt, pg, rn, rm), nil
		}
	case "st1d":
		if ok, zt, pg, rn, rm := is_z_p_rr(args); ok {
			templ := "1	1	1	0	0	1	0	1	1	1	1	Rm	0	1	0	Pg	Rn	Zt"
			return assem_z_p_rr(templ, zt, pg, rn, rm), nil
		}
	case "lsr":
		if ok, zd, zn, imm, T := is_z_zimm(args); ok {
			tsz := getTypeSpecifier(T)[1:] // drop (5th) MSB bit for Q
			if strings.ToUpper(T) == "D" {
				tsz = tsz[:1] + "111" // set x bit for compat with 'as' (see https://docsmirror.github.io/A64/2023-06/lsr_z_zi.html)
			}
			templ := "0	0	0	0	0	1	0	0	tszh	1	tszl	imm3	1	0	0	1	0	1	Zn	Zd"
			templ = strings.ReplaceAll(templ, "tszh", tsz[:2])
			templ = strings.ReplaceAll(templ, "tszl", tsz[2:])
			return assem_z_zi(templ, zd, zn, "imm3", imm), nil
		} else if ok, rd, rn, imm, _ := is_r_ri(args); ok {
			templ := "sf	1	0	1	0	0	1	1	0	N	immr	x	1	1	1	1	1	Rn	Rd"
			templ = strings.ReplaceAll(templ, "sf", "1")
			templ = strings.ReplaceAll(templ, "N", "1")
			templ = strings.ReplaceAll(templ, "x", "1") // x bit is set for compat with 'as'
			return assem_r_ri(templ, rd, rn, "immr", imm, 0), nil
		}
	case "ptrue":
		if ok, pd, T := is_p(args); ok {
			templ := "0	0	1	0	0	1	0	1	size	0	1	1	0	0	0	1	1	1	0	0	0	pattern	0	Pd"
			templ = strings.ReplaceAll(templ, "size", getSizeFromType(T))
			templ = strings.ReplaceAll(templ, "pattern", "1\t1\t1\t1\t1")
			return assem_p(templ, pd), nil
		}
	case "eor3":
		if ok, zd, zn, zm, za, _ := is_z_zzz(args); ok {
			templ := "0	0	0	0	0	1	0	0	0	0	1	Zm	0	0	1	1	1	0	Zk	Zdn"
			if zd == zn {
				return assem_z2_zz(templ, zd, zm, za), nil
			}
		}
	}
	return 0, fmt.Errorf("unhandled instruction: %s", ins)
}

func getR(r string) int {
	if len(r) > 0 && r[0] == 'x' {
		if r[1:] == "zr" {
			return 31 // https://stackoverflow.com/questions/42788696/why-might-one-use-the-xzr-register-instead-of-the-literal-0-on-armv8
		} else if num, err := strconv.ParseInt(r[1:], 10, 32); err == nil {
			return int(num)
		}
	}
	return -1
}

func getP(r string) int {
	if len(r) > 0 && r[0] == 'p' {
		if num, err := strconv.ParseInt(r[1:], 10, 32); err == nil {
			return int(num)
		}
	}
	return -1
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
			return int(num), T, index
		}
	}
	return -1, "", -1
}

func getImm(imm string) (bool, int) {
	if len(imm) > 0 && imm[0] == '#' {
		imm = imm[1:]
		if num, err := strconv.ParseInt(imm, 10, 32); err == nil {
			return true, int(num)
		}
	}
	fmt.Printf("Invalid immediate: %s\n", imm)
	return false, 0
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

func is_p(args []string) (ok bool, pd int, T string) {
	return true, 0, "d"
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

func is_r_rr(args []string) (ok bool, rd, rn, rm int) {
	if len(args) == 3 {
		rd, rn, rm = getR(args[0]), getR(args[1]), getR(args[2])
		if rd != -1 && rn != -1 && rm != -1 {
			return true, rd, rn, rm
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

func is_z_zzz(args []string) (ok bool, zd, zn, zm, za int, T string) {
	if len(args) == 4 {
		var t1, t2, t3, t4 string
		zd, t1, _ = getZ(args[0])
		zn, t2, _ = getZ(args[1])
		zm, t3, _ = getZ(args[2])
		za, t4, _ = getZ(args[3])
		if zd != -1 && zn != -1 && zm != -1 && t1 == t2 && t2 == t3 && t3 == t4 {
			return true, zd, zn, zm, za, t1
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
			if ok, imm = getImm(args[2]); ok {
				return true, zd, zn, imm, t1
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

func is_r_i(args []string) (ok bool, rd int, imm int) {
	if len(args) == 2 {
		rd = getR(args[0])
		if ok, imm = getImm(args[1]); ok {
			return true, rd, imm
		}
	}
	return
}

func is_z_bi(args []string) (ok bool, zt, xn, imm int) {
	if len(args) > 1 {
		zt, _, _ = getZ(args[0])
		if args[1][0] == '[' && strings.HasSuffix(args[len(args)-1], "]") {
			memaddr := strings.Join(args[1:], ", ")
			memaddr = strings.NewReplacer("[", "", "]", "", "MUL, VL", "MUL VL").Replace(memaddr)
			mas := strings.Split(memaddr, ", ")
			if len(mas) >= 1 {
				xn = getR(mas[0])
				if len(mas) == 3 && mas[2] == "MUL VL" {
					if ok, imm = getImm(mas[1]); ok && xn != -1 {
						return true, zt, xn, imm
					}
				} else if xn != -1 {
					return true, zt, xn, 0
				}
			}
		}
	}
	return
}

func is_z_p_rr(args []string) (ok bool, zt, pg, rn, rm int) {
	if len(args) == 8 && args[0] == "{" && args[2] == "}" && strings.ToUpper(args[6]) == "LSL" && args[7] == "#3]" {
		zt, _, _ = getZ(args[1])
		pg = getP(strings.Split(args[3], "/")[0]) // drop any trailer
		rn = getR(strings.ReplaceAll(args[4], "[", ""))
		rm = getR(args[5])
		return true, zt, pg, rn, rm
	}
	return
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
	case "imm16":
		opcode = strings.ReplaceAll(opcode, "imm16", fmt.Sprintf("%0*s", 16, strconv.FormatUint(uint64(imm), 2)))
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
