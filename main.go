package main

import (
	"errors"
	"fmt"
	"github.com/malachy-mcconnell/takeoff-atm/bank"
	"github.com/malachy-mcconnell/takeoff-atm/persistence"
	"github.com/malachy-mcconnell/takeoff-atm/service"
	"log"
	"os"
	"time"
)

const instructions = "Later I will write instructions"

func main() {

	// START HERE
	// Logging and history display.
	// UNIT TEST? eg one test?

	var sessionAccount *bank.Account
	var atm *bank.ATM
	var logoutTimer *time.Timer
	var instruction []string

	var continueFlag bool
	var breakFlag bool
	var err error

	var transacting = false
	var logoutAfterTransaction = false
	var logFileHandle *os.File

	logFileHandle, err = establishLogger()
	defer logFileHandle.Close()
	log.Println("ATM switched on.")

	atm = bank.NewATM()
	sessionAccount = bank.NilAccount()

	err = persistence.SetATMBalance(atm)
	if err != nil {
		fmt.Println(err.Error())
	}

	for true {
		instruction, err = service.GetUserInput()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		if sessionAccount.Authorized() {
			resetTimer(logoutTimer)
		}

		continueFlag, breakFlag, err = processNonAuthInstructions(instruction, atm)
		if err != nil {
			fmt.Println(err.Error())
			log.Println(err.Error())
		}
		if continueFlag {
			continue
		}
		if breakFlag {
			log.Println("ATM switched off ('end' command).")
			break // end for loop
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
				log.Printf("Authorization error %s.\n", err.Error())
				fmt.Println("Authorization failed.")
				continue
			}

			log.Printf("Authorized account %s.\n", sessionAccount.ID)
			fmt.Printf("%s successfully authorized.\n", sessionAccount.ID)
			logoutTimer = startLogoutTimer(sessionAccount, &transacting, &logoutAfterTransaction)
			continue
		}

		if sessionAccount.ID == "" || !sessionAccount.Authorized() {
			fmt.Println("Authorization required.")
			continue
		}

		if instruction[0] == "withdraw" {
			transacting = true
			processWithdrawal(instruction, sessionAccount, atm)
			if logoutAfterTransaction && sessionAccount.Authorized() {
				sessionAccount.Logout()
				log.Printf("Automatic logout of account %s.\n", sessionAccount.ID)
				fmt.Println("Account is logged out after two minutes idle and transaction is complete.")
			}
			transacting = false
			continue
		}

		transacting = true
		processAuthInstructions(instruction, sessionAccount, atm, logoutTimer)
		transacting = false
		if logoutAfterTransaction && sessionAccount.Authorized() {
			sessionAccount.Logout()
			fmt.Println("Account is logged out after two minutes idle and transaction is complete. Account " + sessionAccount.ID)
		}
	}
}

func establishLogger() (*os.File, error) {
	f, err := os.OpenFile("./data/atm.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error opening log file: %v", err))
	}
	log.SetOutput(f)
	return f, nil
}

func auth(instruction []string) (sessionAccount *bank.Account, err error) {
	if len(instruction) != 3 {
		err = errors.New("Command 'authorize' must include account ID and PIN, space separated")
		return bank.NilAccount(), err
	}

	sessionAccount, err = persistence.LoadAccountByID(instruction[1])
	if err != nil {
		err = errors.New("Account ID not matched. Please check and try again. " + err.Error())
		return bank.NilAccount(), err
	}

	// Login if pin matches
	err = sessionAccount.Authorize(instruction[2])
	if err != nil {
		return bank.NilAccount(), err
	}

	return sessionAccount, nil
}

func startLogoutTimer(sessionAccount *bank.Account, transacting *bool, logoutAfterTransaction *bool) *time.Timer {
	logoutTimer := time.NewTimer(time.Minute * 2)

	// This can fire mid-transaction, which we really don't want.
	// Hoop jumping, to only log out if not transacting, otherwise set "logout after transaction"
	// The pointers of the go func() are in the closure of the go func
	go func() {
		<-logoutTimer.C
		if !*transacting {
			if sessionAccount.Authorized() {
				sessionAccount.Logout()
				fmt.Println("Account is logged out after two minutes idle.")
			}
		} else {
			*logoutAfterTransaction = true
		}
	}()

	return logoutTimer
}

func resetTimer(timer *time.Timer) {
	if timer != nil {
		// Stop the timer, drain it if needed, and then reset it
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(time.Minute * 2)
	}
}

func processNonAuthInstructions(instruction []string, atm *bank.ATM) (cont bool, stop bool, err error) {
	err = nil
	cont = true
	stop = false

	if instruction[0] == "" {
		fmt.Println("Enter '?' or 'help' for instructions.")
		return cont, stop, err
	}

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
		err = errors.Join(
			persistence.ResetATMBalance(atm),
			persistence.ResetTransactionsStorage(),
		)
		if err == nil {
			println("ATM balance reset, transactions records deleted, account opening balances reset.")
			log.Println("---------------")
			log.Println("---  reset  ---")
			log.Println("---------------")
		}
		return cont, stop, err
	}

	if instruction[0] == "atm_balance" {
		fmt.Printf("ATM balance is %s\n", atm.Balance.ToString())
		return cont, stop, nil
	}

	return false, false, nil
}

func processWithdrawal(instruction []string, sessionAccount *bank.Account, atm *bank.ATM) {
	var err error
	var preamble = ""
	var balance bank.USD
	var amount bank.USD
	var amountAdjusted bank.USD
	var feeAmount bank.USD

	if atm.Balance == 0 {
		fmt.Println("Unable to process your withdrawal at this time.")
		return
	}
	if sessionAccount.Overdrawn() {
		fmt.Println("Your account is overdrawn. You may not make withdrawals at this time.")
		return
	}
	if len(instruction) != 2 {
		fmt.Println("Must include one amount only like this 'withdraw 100'")
		return
	}
	amount, err = validateAmount(instruction[1], instruction[0])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	amountAdjusted, feeAmount, err = validateWithdrawalAmount(amount, atm, sessionAccount)
	if err != nil {
		fmt.Println(err.Error())
		// TODO: Weak. Use error types instead. (to identify an error that prevents withdrawl, rather than an error which adjusts the withrawalable amount)
		if amountAdjusted == amount {
			log.Printf("Invalid withdrawal amount %s\n", instruction[1])
			return
		}
		// else continue to withdraw the adjusted amount
	}
	amount = amountAdjusted
	err = atm.Withdraw(amount)
	if err != nil {
		log.Printf("Withdrawal error. Account %s. Amount%s.\n", sessionAccount.ID, instruction[1])
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Amount dispensed: $%s\n", amount.ToString())

	transact(-1*amount, sessionAccount)
	if feeAmount != 0 {
		log.Printf("Fee applied (no ledger). Account %s. Fee amount %s.\n", sessionAccount.ID, feeAmount.ToString())
		preamble = fmt.Sprintf("You have been charged an overdraft fee of $%s. ", feeAmount.ToString())
		transact(-1*feeAmount, sessionAccount)
		feeAmount = bank.USD(0)
	}
	balance, err = sessionAccount.ReadBalance()
	if err != nil {
		log.Printf("Error reading balance of account %s.\n", sessionAccount.ID)
		fmt.Println("Cannot read new balance from account")
	} else {
		fmt.Printf("%sCurrent balance: $%s\n", preamble, balance.ToString())
	}

	return
}

func processAuthInstructions(instruction []string, sessionAccount *bank.Account, atm *bank.ATM, logoutTimer *time.Timer) {
	var err error
	var amount bank.USD
	var transactions persistence.Transactions

	switch instruction[0] {
	case "deposit":
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
		log.Printf("Accepted deposit (no ledger). Account %s. Amount %s.\n", sessionAccount.ID, amount.ToString())
		displayCurrentBalance(sessionAccount)

	case "balance":
		if len(instruction) != 1 {
			fmt.Println("Must be a single instruction, like this 'balance'")
			break
		}
		displayCurrentBalance(sessionAccount)

	case "history":
		transactions, err = persistence.LoadTransactions(sessionAccount.ID)
		if err != nil {
			log.Printf("Error creating transaciton history. Account %s. %s\n", sessionAccount.ID, err.Error())
			fmt.Println(err.Error())
		}
		if len(transactions) == 0 {
			fmt.Println("No history found")
			break
		}
		fmt.Printf(transactions.AsTextTable())

	case "logout":
		if !sessionAccount.Authorized() {
			fmt.Println("No account is currently authorized.")
			break
		}
		if !logoutTimer.Stop() {
			<-logoutTimer.C
		}
		ID := sessionAccount.ID
		sessionAccount.Logout()
		fmt.Printf("Account %s logged out.\n", ID)
		fmt.Printf("Account %s logged out.", ID)
	}
}

func validateAmount(amount string, instruction string) (bank.USD, error) {
	USDAmount, err := bank.USDFromString(amount)
	if err != nil {
		err = errors.New(
			fmt.Sprintf("Amount must be numeric like this '%s 100'\n", instruction),
		)
		USDAmount = bank.USD(0)
	}
	return USDAmount, err
}

func validateWithdrawalAmount(amount bank.USD, atm *bank.ATM, account *bank.Account) (amountAllowed bank.USD, fee bank.USD, err error) {
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

func transact(amount bank.USD, account *bank.Account) {
	err := account.Update(amount)
	if err != nil {
		log.Printf("transact() failed to update account %s. %s\n", account.ID, err.Error())
	}
	// err here isn't really allowed. How to recover? We already gave the user the cash from the machine
	// Propose: Log a fatal error and turn the machine off (maintenance mode)
	// TODO: Implement that ^^
	_ = persistence.RecordTransaction(bank.NewTransaction(amount, *account))
	if err != nil {
		log.Printf("transact() failed to persist the transaction %s. %s\n", account.ID, err.Error())
	}
	// TODO: Handle both these errors by logging fatal, and suspend machine (no spec on how to suspend the machine)
}

func displayCurrentBalance(sessionAccount *bank.Account) {
	currentBalance, err := sessionAccount.ReadBalance()
	if err != nil {
		log.Printf("Error reading balance of account %s. %s\n", sessionAccount.ID, err.Error())
		fmt.Printf("Error reading account balance %s\n", err.Error())
		return
	}
	fmt.Printf("Current balance: %s\n", currentBalance.ToString())
}
