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

func runSimpleInterpreter(p *Program) error {
	memory := make([]byte, MEMORY_SIZE)
	pc := 0
	dataPtr := 0

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

			bracketNesting := 1
			savedPC := pc

			for {
				if bracketNesting == 0 {
					break
				}
				pc++
				if pc >= len(p.instructions) {
					break
				}

				if p.instructions[pc] == ']' {
					bracketNesting--
				} else if p.instructions[pc] == '[' {
					bracketNesting++
				}
			}

			if bracketNesting != 0 {
				return fmt.Errorf("unmatched '[' at pc=%d", savedPC)
			}
		case ']': // Jump back to the matching [ if the cell at the pointer is nonzero
			if memory[dataPtr] == 0 {
				break
			}

			bracketNesting := 1
			savedPC := pc

			for (bracketNesting != 0) && (pc > 0) {
				pc--
				if p.instructions[pc] == '[' {
					bracketNesting--
				} else if p.instructions[pc] == ']' {
					bracketNesting++
				}
			}

			if bracketNesting != 0 {
				return fmt.Errorf("unmatched ']' at pc=%d", savedPC)
			}
		default:
			return fmt.Errorf("bad char '%c' instruction at pc=%d", char, pc)
		}
		pc++
	}
	return nil
}

// see https://esolangs.org/wiki/Brainfuck
func main() {
	bfFilePath := "./testdata/1to5.bf"
	// bfFilePath := "./testdata/mandelbrot.bf"
	// bfFilePath := "./testdata/factor.bf" // stdin 179424691

	input, err := os.Open(bfFilePath)
	if err != nil {
		log.Fatal(err)
	}

	program, err := parseFromReader(input)
	if err != nil {
		log.Fatal(err)
	}

	if err := runSimpleInterpreter(program); err != nil {
		log.Fatal(err)
	}
}
