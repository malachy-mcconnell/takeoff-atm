package persistence

import (
	"github.com/malachy-mcconnell/takeoff-atm/domain"
	"os"
)

const pathATMBalance = "./data/atm.csv"
const openingATMBalance = domain.USD(1000000)

func WriteATMBalance(ATM domain.ATM) error {
	f, err := os.Create(pathATMBalance) // Always overwrite existing
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(ATM.Balance.ToString()) // writing...
	if err != nil {
		return err
	}
	return nil
}

func SetATMBalance(ATM *domain.ATM) error {
	data, err := os.ReadFile(pathATMBalance)
	if err != nil {
		return err
	}
	dataString := string(data)
	if len(dataString) == 0 {
		return ResetATMBalance(ATM)
	}
	ATM.Balance, err = domain.USDFromString(string(data))
	if err != nil {
		return err
	}
	return nil
}

func ResetATMBalance(ATM *domain.ATM) error {
	ATM.Balance = openingATMBalance
	return WriteATMBalance(*ATM)
}
