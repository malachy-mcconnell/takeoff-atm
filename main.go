package main

import (
	"errors"
	"fmt"
	"github.com/malachy-mcconnell/takeoff-atm/domain"
	"github.com/malachy-mcconnell/takeoff-atm/persistence"
	"github.com/malachy-mcconnell/takeoff-atm/service"
	"time"
)

const instructions = "Later I will write instructions"

func main() {

	var sessionAccount *domain.Account
	var atm *domain.ATM
	var logoutTimer *time.Timer

	var continueFlag bool
	var breakFlag bool
	var err error

	err = persistence.SetATMBalance(atm)
	if err != nil {
		fmt.Println(err.Error())
	}

	for true {
		instruction, _ := service.GetUserInput()

		resetTimer(logoutTimer)

		continueFlag, breakFlag, err = processNonAuthInstructions(instruction, atm)
		if err != nil {
			fmt.Println(err.Error())
		}
		if continueFlag {
			continue
		}
		if breakFlag {
			break
		}

		if instruction[0] == "authorize" {
			if sessionAccount.Authorized() {
				if sessionAccount.ID == instruction[1] {
					fmt.Printf("Already authorized for account %s", sessionAccount.ID)
					continue
				} // else logout, but logout is implicit in the code below
			}

			sessionAccount, err = auth(instruction)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			logoutTimer = startLogoutTimer(sessionAccount)
		}

		if sessionAccount.ID == "" || !sessionAccount.Authorized() {
			fmt.Println("Authorization required.")
			continue
		}

		if instruction[0] == "withdraw" {
			continueFlag = processWithdrawal(instruction, sessionAccount, atm)
			if continueFlag {
				continue
			}
		}

		processAuthInstructions(instruction, sessionAccount, atm)
	}
}

func auth(instruction []string) (sessionAccount *domain.Account, err error) {
	if len(instruction) != 3 {
		err = errors.New("Command 'authorize' must include account ID and PIN, space separated")
		return &domain.Account{}, err
	}

	//fmt.Println("Account login authorize command received. " + instruction[1])
	sessionAccount, err = persistence.LoadAccountByID(instruction[1])
	if err != nil {
		err = errors.New("Account ID not matched. Please check and try again. " + err.Error())
		return &domain.Account{}, err
	}
	// Try to log in, if pin matches:
	err = sessionAccount.Authorize(instruction[2])
	if err != nil {
		return &domain.Account{}, err
	}

	//fmt.Println("Authorized.")
	// If there is no activity for 2 minutes, your program should automatically log out the account
	// Start logout timer (reset logout timer to two mins on next input)
	// TODO NOW: Start a global timer?

	// Start timer
	// So we might not need account.authorized (feels a little weak; less generic)
	return sessionAccount, nil

}

func startLogoutTimer(sessionAccount *domain.Account) *time.Timer {
	logoutTimer := time.NewTimer(time.Minute * 2)
	go func(sessionAccount *domain.Account) {
		<-logoutTimer.C
		fmt.Println("Logout timer fired.")
		sessionAccount.Logout()
	}(sessionAccount)
	return logoutTimer
}

func resetTimer(timer *time.Timer) {
	if timer != nil {
		// Stop the timer, drain and reset it
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(time.Minute * 2)
		fmt.Println("Reset the logout timer due to activity.")
	}
}

func processWithdrawal(instruction []string, sessionAccount *domain.Account, atm *domain.ATM) bool {
	var err error
	var preamble string = ""
	var balance domain.USD
	var amount domain.USD
	var amountAdjusted domain.USD
	var feeAmount domain.USD

	fmt.Println("Account withdraw command received.")
	if atm.Balance == 0 {
		fmt.Println("Unable to process your withdrawal at this time.")
		return true
	}
	if sessionAccount.Overdrawn() {
		fmt.Println("Your account is overdrawn. You may not make withdrawals at this time.")
		return true
	}
	if len(instruction) != 2 {
		fmt.Println("Must include one amount only like this 'withdraw 100'")
		return true
	}
	amount, err = validateAmount(instruction[1], instruction[0])
	if err != nil {
		fmt.Println(err.Error())
		return true
	}
	amountAdjusted, feeAmount, err = validateWithdrawalAmount(amount, atm, sessionAccount)
	if err != nil {
		fmt.Println(err.Error())
		// TODO: Weak. Use error types instead.
		if amountAdjusted == amount {
			return true
		}
		// else continue
	}
	amount = amountAdjusted
	err = atm.Withdraw(amount)
	if err != nil {
		fmt.Println(err.Error())
		return true
	}
	fmt.Printf("Amount dispensed: $%s\n", amount.ToString())

	transact(-1*amount, sessionAccount)
	if feeAmount != 0 {
		preamble = fmt.Sprintf("You have been charged an overdraft fee of $%s. ", feeAmount.ToString())
		transact(-1*feeAmount, sessionAccount)
		feeAmount = domain.USD(0)
	}
	balance, err = sessionAccount.ReadBalance()
	if err != nil {
		fmt.Println("Cannot read new balance from account")
	} else {
		fmt.Printf("%sCurrent balance: $%s\n", preamble, balance.ToString())
	}

	return true
}

func processAuthInstructions(instruction []string, sessionAccount *domain.Account, atm *domain.ATM) {
	var err error
	var amount domain.USD
	var transactions persistence.Transactions

	switch instruction[0] {
	case "deposit":
		fmt.Println("Account deposit command received.")
		if len(instruction) != 2 {
			fmt.Println("Must include one amount only like this 'withdraw 100'")
			break
		}
		amount, err = validateAmount(instruction[1], instruction[0])
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		atm.Deposit(amount)
		transact(amount, sessionAccount)

	case "balance":
		fmt.Println("Account balance command received.")
		if len(instruction) != 1 {
			fmt.Println("Must be a single instruction, like this 'balance'")
			break
		}
		balance, err := sessionAccount.ReadBalance()
		if err != nil {
			fmt.Printf("Error reading account balance %s\n", err.Error())
		}
		fmt.Println(balance.ToString())

	case "history":
		fmt.Println("Account history command received.")
		transactions, err = persistence.LoadTransactions(sessionAccount.ID)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(transactions)

	case "logout":
		fmt.Println("Account logout command received.")
		sessionAccount.Logout()

	default:
		fmt.Println("Instruction not recognised. Enter '?' for list of instructions.")
	}
}

func processNonAuthInstructions(instruction []string, atm *domain.ATM) (cont bool, stop bool, err error) {
	err = nil
	cont = true
	stop = false

	if instruction[0] == "?" || instruction[0] == "help" {
		fmt.Println(instructions)
		return cont, stop, err
	}

	// TODO: Admin inputs need authorization too
	if instruction[0] == "end" {
		fmt.Println("End command received.")
		cont = false
		stop = true
		return cont, stop, persistence.WriteATMBalance(*atm)
	}

	if instruction[0] == "reset" {
		fmt.Println("Reset command received (reset all storage to initial values (ATM and accounts)).")
		// TODO: Blow out the transactions storage
		// Reset the ATM balance and continue
		return cont, stop, persistence.ResetATMBalance(atm)
	}

	if instruction[0] == "atm_balance" {
		fmt.Printf("ATM balance is %s\n", atm.Balance.ToString())
		return cont, stop, nil
	}

	return cont, stop, err
}

func validateAmount(amount string, instruction string) (domain.USD, error) {
	USDAmount, err := domain.USDFromString(amount)
	if err != nil {
		err = errors.New(
			fmt.Sprintln("Amount must be numeric like this '%s 100'", instruction),
		)
		USDAmount = domain.USD(0)
	}
	return USDAmount, err
}

func validateWithdrawalAmount(amount domain.USD, atm *domain.ATM, account *domain.Account) (amountAllowed domain.USD, fee domain.USD, err error) {
	amountAllowed, fee, err = atm.ValidateWithdrawal(amount)
	if err != nil {
		if amount == amountAllowed {
			// Then the error is blocking.
			// TODO: Better to use custom error types
			return amountAllowed, fee, err
		}
		fmt.Println(err.Error())
	}

	amountAllowed, fee, err = account.ValidateWithdrawl(amountAllowed)
	if err != nil {
		fmt.Println(err.Error())
	}
	return amountAllowed, fee, nil
}

func transact(amount domain.USD, account *domain.Account) {
	err := account.Update(amount)
	if err != nil {
		fmt.Println("DEBUG: transact() failed to update account")
	}
	// err here isn't really allowed. How to recover? We already gave the user the cash from the machine
	// Propose: Log a fatal error and turn the machine off (maintenance mode)
	// TODO: Implement that ^^ (error here can occur if the two minute timer triggers a logout)
	// So we might think about letting the transaction complete before logging out
	// or remove the need to be authorised just to update an account balance
	_ = persistence.RecordTransaction(domain.NewTransaction(amount, *account))
	if err != nil {
		fmt.Println("DEBUG: transact() failed to persist the transaction")
	}
	// TODO: Handle both these errors by logging fatal, and suspend machine (no spec on how to suspend the machine)
}
