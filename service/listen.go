package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GetUserInput Reads a command line from the keyboard and returns it
// as an array of lowercase strings
func GetUserInput() ([]string, error) {
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)
		return []string{""}, err
	}
	input = strings.ToLower(strings.TrimSuffix(input, "\n"))
	return strings.Split(input, " "), nil
}
