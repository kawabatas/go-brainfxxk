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

func computeJumpTable(p *Program) (map[int]int, error) {
	pc := 0
	programSize := len(p.instructions)
	jt := make(map[int]int)

	for pc < programSize {
		char := p.instructions[pc]
		if char == '[' {
			bracketNesting := 1
			seek := pc

			for {
				if bracketNesting == 0 {
					break
				}
				seek++
				if seek >= programSize {
					break
				}

				if p.instructions[seek] == ']' {
					bracketNesting--
				} else if p.instructions[seek] == '[' {
					bracketNesting++
				}
			}

			if bracketNesting != 0 {
				return jt, fmt.Errorf("unmatched '[' at pc=%d", pc)
			}
			jt[pc] = seek
			jt[seek] = pc
		}
		pc++
	}

	return jt, nil
}

func runOptInterpreter(p *Program) error {
	memory := make([]byte, MEMORY_SIZE)
	pc := 0
	dataPtr := 0

	jumpTable, err := computeJumpTable(p)
	if err != nil {
		return err
	}

	for pc < len(p.instructions) {
		char := p.instructions[pc]
		switch char {
		case '>': // Move the pointer to the right
			dataPtr++
		case '<': // Move the pointer to the left
			dataPtr--
		case '+': // Increment the memory cell at the pointer
			memory[dataPtr]++
		case '-': // Decrement the memory cell at the pointer
			memory[dataPtr]--
		case '.': // Output the character signified by the cell at the pointer
			fmt.Printf("%c", memory[dataPtr])
		case ',': // Input a character and store it in the cell at the pointer
			buf := make([]byte, 1)
			_, err := os.Stdin.Read(buf)
			if err != nil {
				return err
			}
			memory[dataPtr] = buf[0]
		case '[': // Jump past the matching ] if the cell at the pointer is 0
			if memory[dataPtr] != 0 {
				break
			}
			pc = jumpTable[pc]
		case ']': // Jump back to the matching [ if the cell at the pointer is nonzero
			if memory[dataPtr] == 0 {
				break
			}
			pc = jumpTable[pc]
		default:
			return fmt.Errorf("bad char '%c' instruction at pc=%d", char, pc)
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
