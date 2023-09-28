package domain

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Transaction struct {
	AccountID    string
	dateTime     time.Time
	amount       USD
	BalanceAfter USD
}

func NewTransaction(amount USD, account Account) Transaction {
	balance, _ := account.ReadBalance()
	return Transaction{
		AccountID:    account.ID,
		dateTime:     time.Now(),
		amount:       amount,
		BalanceAfter: balance,
	}
}

func NewTransactionFromCSVRow(row []string) (Transaction, error) {
	unixMicroTime, err := strconv.ParseInt(row[1], 10, 64) // time.Parse("", row[1])
	dateTime := time.UnixMicro(unixMicroTime)
	amount, err := USDFromString(row[2])
	balanceAfter, err := USDFromString(row[3])
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

func (t Transaction) AsStringSlice() []string {
	return []string{
		t.AccountID,
		fmt.Sprintf("%d", t.dateTime.UnixMicro()),
		t.amount.ToString(),
		t.BalanceAfter.ToString(),
	}
}

func (t Transaction) humanReadableTabulated() string {
	return ""
}
