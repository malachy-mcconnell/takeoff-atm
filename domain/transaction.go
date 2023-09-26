package domain

import (
	"errors"
	"time"
)

type Transaction struct {
	AccountID    string
	dateTime     time.Time
	amount       USD
	BalanceAfter USD
}

func NewTransaction(amount USD, account Account) Transaction {
	return Transaction{
		AccountID:    account.ID,
		dateTime:     time.Now(),
		amount:       amount,
		BalanceAfter: account.Balance,
	}
}

func NewTransactionFromCSVRow(row []string) (Transaction, error) {
	dateTime, err := time.Parse("", row[1])
	amount, err := USDFromString(row[1])
	balanceAfter, err := USDFromString(row[2])
	if err != nil {
		return Transaction{}, errors.Join(errors.New("Invalid data in CSV row"), err)
	}
	return Transaction{
		AccountID:    row[0],
		dateTime:     dateTime,
		amount:       amount,
		BalanceAfter: balanceAfter,
	}, nil
}

func (t Transaction) humanReadableTabulated() string {
	return ""
}
