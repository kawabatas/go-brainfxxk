package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	MEMORY_SIZE = 30000
)

type opKind int

const (
	INVALID_OP opKind = iota
	INC_PTR
	DEC_PTR
	INC_DATA
	DEC_DATA
	READ_STDIN
	WRITE_STDOUT
	JUMP_IF_DATA_ZERO
	JUMP_IF_DATA_NOT_ZERO
)

type op struct {
	opKind   opKind
	argument int
}

type OptimizedProgram struct {
	instructions []op
}

type Program struct {
	instructions []byte
}

func parseFromReader(reader io.Reader) (*Program, error) {
	program := &Program{}
	buf := make([]byte, 1)

	for {
		_, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch buf[0] {
		case '>', '<', '+', '-', '.', ',', '[', ']':
			program.instructions = append(program.instructions, buf[0])
		}
	}

	return program, nil
}

func translateProgram(p *Program) (*OptimizedProgram, error) {
	pc := 0
	programSize := len(p.instructions)
	optProgram := &OptimizedProgram{}
	var openBracketStack []int

	for pc < programSize {
		char := p.instructions[pc]
		if char == '[' {
			// Place a jump op with a placeholder 0 offset. It will be patched-up to
			// the right offset when the matching ']' is found.
			openBracketStack = append(openBracketStack, len(optProgram.instructions))
			optProgram.instructions = append(optProgram.instructions, op{opKind: JUMP_IF_DATA_ZERO, argument: 0})
			pc++
		} else if char == ']' {
			if len(openBracketStack) == 0 {
				return nil, fmt.Errorf("unmatched closing ']' at pc=%d", pc)
			}
			openBracketOffset := openBracketStack[len(openBracketStack)-1] // top
			openBracketStack = openBracketStack[:len(openBracketStack)-1]  // pop
			optProgram.instructions[openBracketOffset].argument = len(optProgram.instructions)
			optProgram.instructions = append(optProgram.instructions, op{opKind: JUMP_IF_DATA_NOT_ZERO, argument: openBracketOffset})
			pc++
		} else {
			start := pc
			for pc < programSize && p.instructions[pc] == char {
				pc++
			}
			repeatNum := pc - start

			kind := INVALID_OP
			switch char {
			case '>':
				kind = INC_PTR
			case '<':
				kind = DEC_PTR
			case '+':
				kind = INC_DATA
			case '-':
				kind = DEC_DATA
			case ',':
				kind = READ_STDIN
			case '.':
				kind = WRITE_STDOUT
			}
			optProgram.instructions = append(optProgram.instructions, op{opKind: kind, argument: repeatNum})
		}
	}
	return optProgram, nil
}

func runOptInterpreter(p *Program) error {
	memory := make([]byte, MEMORY_SIZE)
	pc := 0
	dataPtr := 0

	optProgram, err := translateProgram(p)
	if err != nil {
		return err
	}

	for pc < len(optProgram.instructions) {
		op := optProgram.instructions[pc]
		kind := op.opKind
		switch kind {
		case INC_PTR: // Move the pointer to the right
			dataPtr += op.argument
		case DEC_PTR: // Move the pointer to the left
			dataPtr -= op.argument
		case INC_DATA: // Increment the memory cell at the pointer
			memory[dataPtr] += byte(op.argument)
		case DEC_DATA: // Decrement the memory cell at the pointer
			memory[dataPtr] -= byte(op.argument)
		case WRITE_STDOUT: // Output the character signified by the cell at the pointer
			i := 0
			for i < op.argument {
				i++
				fmt.Printf("%c", memory[dataPtr])
			}
		case READ_STDIN: // Input a character and store it in the cell at the pointer
			i := 0
			for i < op.argument {
				i++
				buf := make([]byte, 1)
				_, err := os.Stdin.Read(buf)
				if err != nil {
					return err
				}
				memory[dataPtr] = buf[0]
			}
		case JUMP_IF_DATA_ZERO: // Jump past the matching ] if the cell at the pointer is 0
			if memory[dataPtr] != 0 {
				break
			}
			pc = op.argument
		case JUMP_IF_DATA_NOT_ZERO: // Jump back to the matching [ if the cell at the pointer is nonzero
			if memory[dataPtr] == 0 {
				break
			}
			pc = op.argument
		default:
			return fmt.Errorf("INVALID_OP encountered on pc=%d", pc)
		}
		pc++
	}
	return nil
}

// see https://esolangs.org/wiki/Brainfuck
func main() {
	// bfFilePath := "./testdata/1to5.bf"
	// bfFilePath := "./testdata/mandelbrot.bf"
	bfFilePath := "./testdata/factor.bf" // stdin 179424691

	input, err := os.Open(bfFilePath)
	if err != nil {
		log.Fatal(err)
	}

	program, err := parseFromReader(input)
	if err != nil {
		log.Fatal(err)
	}

	if err := runOptInterpreter(program); err != nil {
		log.Fatal(err)
	}
}
