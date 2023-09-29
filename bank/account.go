package bank

import (
	"errors"
)

const overdraftFee = USD(500)

var notAuthorized = errors.New("Not Authorized")

type Account struct {
	ID         string
	pin        string
	authorized bool
	balance    USD
}

func NewAccount(ID string, pin string, balance USD) *Account {
	account := Account{
		ID:         ID,
		pin:        pin,
		authorized: false,
		balance:    balance,
	}
	return &account
}

func NilAccount() *Account {
	nilAccount := Account{}
	return &nilAccount
}

func (a *Account) Authorize(challengePin string) error {
	if a.pin == challengePin {
		a.authorized = true
		return nil
	}
	return errors.New("PIN does not match account id. Please try again.")
}

func (a *Account) Logout() {
	a.ID = ""
	a.pin = ""
	a.authorized = false
	a.balance = 0
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
	err = nil
	if amount >= a.balance {
		// TODO: Better to use a custom error type
		err = errors.New(
			// TODO: Better to offer user a choice: Accept $5 fee and go overdrawn, or withdraw a lesser amount, or cancel withdrawl
			"Overdrawn",
		)
		fee = overdraftFee
	}
	return amount, fee, nil //err
}

func (a *Account) Update(amount USD) error {
	if a.authorized {
		// Assumes overdraft logic already applied
		a.balance = a.balance + amount
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
