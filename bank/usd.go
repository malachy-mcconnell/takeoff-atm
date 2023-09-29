package bank

import (
	"fmt"
	"math"
	"strconv"
)

// USD To ensure a currency number with precisely two decimal places
// and handle any rounding that might be needed.
// This code works with cents as integers and converts to dollars for display only
type USD int32

func (u USD) ToString() string {
	return fmt.Sprintf("%0.2f", u.ToFloat())
}

func (u USD) ToFloat() float32 {
	return float32(u) / 100
}

func USDFromString(value string) (USD, error) {
	// TODO: Check decimal places count -- must be 2 or less
	// TODO: Better to strip ,. and parse to int; needs to have dot or add 00
	temp, err := strconv.ParseFloat(value, 32)
	if err != nil {
		// TODO: Wrap error messages from ParseFloat to be more meaningful
		return USD(0), err
	}
	return USD(float32(math.Round(temp * 100))), nil
}

func (u USD) MultipleOf(z USD) bool {
	return u%z == 0
}
