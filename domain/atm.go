package domain

import (
	"errors"
	"fmt"
)

const billSize = USD(2000)
const ATMFee = USD(0)

type ATM struct {
	Balance   USD
	Deposited USD
}

func (a ATM) ValidateWithdrawal(amount USD) (USD, USD, error) {
	var err error
	if amount > a.Balance {
		// There is an assumption that the remaining balance _is_ a multiple of 20
		// It really must be. TODO: Also check the remainder is a multiple of 20
		amount = a.Balance
		err = errors.New(
			fmt.Sprintf("Unable to dispense full amount requested at this time. Dispensing %s instead.", amount.ToString()),
		)
		return amount, ATMFee, err //nil
	}
	if !amount.MultipleOf(billSize) {
		err = errors.New(fmt.Sprintf("Amount must be a multiple of %s", billSize.ToString()))
		return amount, ATMFee, err
	}
	// Else amount is valid
	return amount, ATMFee, nil
}

func (a *ATM) Withdraw(amount USD) error {
	if amount > a.Balance {
		return errors.New("This ATM cannot dispense that amount.")
	}
	a.Balance = a.Balance - amount
	return nil
}

func (a *ATM) Deposit(amount USD) {
	a.Deposited = a.Deposited + amount
}

func (a ATM) End() {
	//Persist balance
	//return persistence.WriteBalance(a)
	// Okay, but not here. No persistence in the domain model
}
