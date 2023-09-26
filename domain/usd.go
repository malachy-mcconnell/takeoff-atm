package domain

import (
	"fmt"
	"strconv"
)

// To ensure a currency number with precisely two decimal places
// and handle any rounding that might be needed.
// This code works with cents as integers and converts to dollars for display only
type USD int32

func (u USD) ToString() string {
	return fmt.Sprintf("%0.2f", float32(u)/100)
}

func USDFromString(value string) (USD, error) {
	temp, err := strconv.ParseFloat(value, 32)
	if err != nil {
		// TODO: Wrap error messages from ParseFloat to be more meaningful
		return USD(0), err
	}
	return USD((float32(temp) * 100)), nil
}

func (u USD) MultipleOf(z int32) bool {
	return u%USD(z) == 0
}
