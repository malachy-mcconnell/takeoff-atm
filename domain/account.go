package domain

import (
	"errors"
	"fmt"
)

const overdraftFee = USD(5)

var notAuthorized = errors.New("Not Authorized")

type Account struct {
	ID         string
	pin        string
	authorized bool
	balance    USD
}

func NewAccount(ID string, pin string, balance USD) Account {
	return Account{
		ID:         ID,
		pin:        pin,
		authorized: false,
		balance:    balance,
	}
}

func (a Account) Authorize(challengePin string) error {
	if a.pin == challengePin {
		a.authorized = true
		return nil
	}
	return errors.New("PIN does not match account id. Please try again.")
}

func (a Account) Authorized() bool {
	return a.authorized
}

func (a Account) Overdrawn() bool {
	// Authorization to access balance?
	return a.balance < 0
}

func (a Account) ValidateWithdrawl(amount USD) (adjustedAmount USD, fee USD, err error) {
	fee = USD(0)
	if amount > a.balance {
		// TODO Only check the adjusted amount is a multiple of 20 _if_ this is an ATM withdrawl
		adjustedAmount = a.balance - (a.balance % billSize)
		err = errors.New(
			// TODO: Better to offer user a choice: Accept $5 fee and go overdrawn, or withdraw a lesser amount, or cancel withdrawl
			fmt.Sprintf("Unable to dispense full amount requested at this time. Dispensing %s instead.", amount.ToString()),
		)
		if adjustedAmount > a.balance {
			fee = overdraftFee
		}
		return adjustedAmount, fee, err
	}
	// Else amount is 'account' valid
	return amount, fee, nil
}

func (a Account) Update(amount USD) error {
	if a.authorized {
		// Assumes overdraft logic already applied
		a.balance = a.balance + amount
		return nil
	}
	return notAuthorized
}

// TODO: Redundant. Use Update() instead
func (a Account) ApplyFee(amount USD) error {
	if a.authorized {
		// Assumes overdraft logic already applied
		a.balance = a.balance - amount
		return nil
	}
	return notAuthorized
}

func (a Account) ReadBalance() (USD, error) {
	if a.authorized {
		return a.balance, nil
	}
	return USD(0), notAuthorized
}

func (a Account) logout() {
	a.authorized = false
}
