package persistence

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/malachy-mcconnell/takeoff-atm/domain"
	"io"
	"log"
	"os"
)

// const pathAccounts = "./data/accounts.csv"
const pathInitialAccounts = "./data/accounts-initial.csv"

// LoadAccountByID Loads one account. In this version we do not create accounts, and
// we do not change PINs. We do adjust balances; but those are written to transactions.csv
// so accounts initial is never written to, and balances come from the transactions.csv
func LoadAccountByID(ID string) (*domain.Account, error) {
	f, err := os.Open(pathInitialAccounts)
	if err != nil {
		fmt.Println(err.Error())
		return &domain.Account{}, errors.Join(
			errors.New("Unable to read input file "+pathInitialAccounts),
			err,
		)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Unable to parse file as CSV for "+pathInitialAccounts, err)
			return &domain.Account{}, errors.Join(
				errors.New("Unable to retrieve account "+ID),
				err,
			)
		}

		if row[0] == ID {
			openingBalance, err := balanceFromTransactionLogOrInitial(ID, row[2])
			if err != nil {
				return &domain.Account{}, errors.Join(
					errors.New("Unable to create opening balance for account "+ID),
					err,
				)
			}
			return domain.NewAccount(row[0], row[1], openingBalance), nil
		}
	}
	return &domain.Account{}, errors.New("Account ID " + ID + " not found.")
}

// balanceFromTransactionLogOrInitial is inefficient because we load all transactions
// just to get the opening balance of one account. Consider more efficient storage (eg database)
func balanceFromTransactionLogOrInitial(ID string, initial string) (domain.USD, error) {
	allTransactions, err := LoadTransactions(ID)
	if err != nil {
		return domain.USD(0), err
	}
	count := len(allTransactions)
	if count == 0 {
		return domain.USDFromString(initial)
	}
	return allTransactions[count-1].BalanceAfter, nil
}
