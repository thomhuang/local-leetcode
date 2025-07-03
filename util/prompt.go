package util

import (
	"bufio"
	"fmt"
)

func PromptWithValidation(scanner *bufio.Scanner, prompt string, validate func(string) (bool, string)) string {
	for {
		fmt.Println(prompt)
		input := ReadInput(scanner)
		fmt.Println()

		if valid, errorMsg := validate(input); valid {
			return input
		} else {
			fmt.Println(errorMsg)
			fmt.Println()
		}
	}
}

func ReadInput(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}
