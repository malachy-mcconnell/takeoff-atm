package service

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var allowedCommands = []string{
	"",
	"?",
	"help",
	"atm_balance",
	"reset",
	"end",
	"authorize",
	"withdraw",
	"deposit",
	"balance",
	"history",
	"logout",
}

// GetUserInput Reads a command line from the keyboard and returns it
// as an array of lowercase strings
func GetUserInput() ([]string, error) {
	fmt.Print("> ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return []string{""}, errors.Join(errors.New("An error occurred while reading input. Please try again"), err)
	}
	input = strings.ToLower(strings.TrimSuffix(input, "\n"))
	instruction := strings.Split(input, " ")
	err = validate(instruction)
	if err != nil {
		return []string{""}, err
	}
	return instruction, nil
}

func validate(instruction []string) error {
	allowed := false
	for i := 0; i < len(allowedCommands); i++ {
		if allowedCommands[i] == instruction[0] {
			allowed = true
			break
		}
	}

	if !allowed {
		err := errors.New("Command not recognized. Enter '?' or 'help' for instructions.")
		return err
	}
	return nil
}
