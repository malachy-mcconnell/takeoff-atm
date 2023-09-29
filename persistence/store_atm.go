package persistence

import (
	"github.com/malachy-mcconnell/takeoff-atm/bank"
	"os"
)

const pathATMBalance = "./data/atm.csv"
const openingATMBalance = bank.USD(1000000)

func WriteATMBalance(ATM bank.ATM) error {
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

func SetATMBalance(ATM *bank.ATM) error {
	data, err := os.ReadFile(pathATMBalance)
	if err != nil {
		return err
	}
	dataString := string(data)
	if len(dataString) == 0 {
		return ResetATMBalance(ATM)
	}
	ATM.Balance, err = bank.USDFromString(string(data))
	if err != nil {
		return err
	}
	return nil
}

func ResetATMBalance(ATM *bank.ATM) error {
	ATM.Balance = openingATMBalance
	return WriteATMBalance(*ATM)
}
