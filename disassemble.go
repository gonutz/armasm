package armasm

import (
	"bytes"
	"errors"
	"fmt"
)

func Disassemble(code []byte) (string, error) {
	nextARM := func() (uint32, bool) {
		if len(code) >= 4 {
			instr := 0 +
				uint32(code[0]) +
				uint32(code[1])<<8 +
				uint32(code[2])<<16 +
				uint32(code[3])<<24
			code = code[4:]
			return instr, true
		}
		return 0, false
	}

	conditions := []string{
		"EQ", "NE", "CS", "CC",
		"MI", "PL", "VS", "VC",
		"HI", "LS", "GE", "LT",
		"GT", "LE", "", "ERROR",
	}
	const (
		AND = iota
		EOR
		SUB
		RSB
		ADD
		ADC
		SBC
		RSC
		TST
		TEQ
		CMP
		CMN
		ORR
		MOV
		BIC
		MVN
	)
	opCodes := []string{
		"AND", "EOR", "SUB", "RSB",
		"ADD", "ADC", "SBC", "RSC",
		"TST", "TEQ", "CMP", "CMN",
		"ORR", "MOV", "BIC", "MVN",
	}

	var buf bytes.Buffer
	printf := func(format string, a ...interface{}) {
		fmt.Fprintf(&buf, format, a...)
	}
	for {
		instr, ok := nextARM()
		if !ok {
			break
		}

		printf("%08X ", instr)
		for i := 31; i >= 0; i-- {
			printf("%d", instr&(1<<uint(i))>>uint(i))
			if i%8 == 0 {
				printf(" ")
			}
		}

		cond := instr & 0xF0000000 >> 28
		if cond == 0xF {
			return string(buf.Bytes()), errors.New("invalid condition 0xF")
		}

		if instr&0x0C000000 == 0x04000000 {
			// Single Data Transfer.
			if instr&(1<<20) == 0 {
				printf("STR")
			} else {
				printf("LDR")
			}
			printf(conditions[cond])
			if instr&(1<<22) != 0 {
				printf("B")
			}
			if instr&(1<<21) != 0 {
				printf("T")
			}
			printf(" R%d, ", (instr&0x0000F000)>>12)
			// Address follows.
			printf("[R%d", (instr&0x000F0000)>>16)
			preIndexed := instr&(1<<24) != 0
			if !preIndexed {
				printf("]")
			}
			if instr&(1<<25) == 0 {
				// Immediate offset.
				printf(", #")
				if instr&(1<<23) == 0 {
					// Up/down bit is 0 -> negative offset
					printf("-")
				}
				printf("0x%X", instr&0xFFF)
			} else {
				// Offset is a register.
				// TODO shifts except for register shift
			}
			if preIndexed {
				printf("]")
				if instr&(1<<21) != 0 {
					printf("!")
				}
			}
		} else if instr&0x0C000000 == 0x00000000 {
			// Data Processing / PSR Transfer
			opCode := (instr & 0x01E00000) >> 21
			printf(opCodes[opCode])
			printf(conditions[cond])

			if opCode == MOV || opCode == MVN {
				if instr&(1<<20) != 0 {
					printf("S")
				}
				printf(" R%d, ", (instr&0x0000F000)>>12)
			} else if opCode == CMP || opCode == CMN ||
				opCode == TEQ || opCode == TST {
				printf(" R%d, ", (instr&0x000F0000)>>16)
			} else {
				if instr&(1<<20) != 0 {
					printf("S")
				}
				printf(" R%d, R%d, ", (instr&0x0000F000)>>12, (instr&0x000F0000)>>16)
			}

			if instr&(1<<25) == 0 {
				// Operand 2 is a register.
				printf("reg")
			} else {
				// Operand 2 is an immediate value.
				value := instr & 0xFF
				rotate := (instr & 0xF00) >> 8
				printf("#0x%X", rotateRight(value, rotate*2))
			}
		}

		printf("\n")
	}
	return string(buf.Bytes()), nil
}

func rotateRight(value, by uint32) uint32 {
	by %= 32
	mask := (uint32(1) << by) - 1
	return value>>by + ((value & mask) << (32 - by))
}
