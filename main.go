package main

import (
	"fmt"
	"github.com/malachy-mcconnell/takeoff-atm/domain"
	"github.com/malachy-mcconnell/takeoff-atm/persistence"
	"github.com/malachy-mcconnell/takeoff-atm/service"
)

const instructions = "Later I will write instructions"

func main() {

	var sessionAccount domain.Account
	var atm domain.ATM
	var transactions persistence.Transactions
	//var transaction domain.Transaction
	//var ATMStore persistence.StoreATM
	//var AccountStore persistence.StoreAccounts
	//var transactionStore persistence.StoreTransactions
	//var logger persistence.StoreLogs // Consider logs/slogs
	var err error

	var amount domain.USD
	var feeAmount domain.USD
	
	err = persistence.SetATMBalance(&atm)
	if err != nil {
		fmt.Println(err.Error())
	}

	for true {
		instruction, _ := service.GetUserInput()
		fmt.Printf("The input command was: %s\n", instruction[0])

		if instruction[0] == "?" || instruction[0] == "help" {
			fmt.Println(instructions)
			continue
		}

		if instruction[0] == "end" {
			fmt.Println("End command received.")
			err = persistence.WriteATMBalance(atm)
			if err != nil {
				fmt.Println(err.Error())
			}
			// terminate, or send done to the done channel, and maybe do shut down steps there.
			break
		}

		if instruction[0] == "reset" {
			fmt.Println("Reset command received (reset all storage to initial values (ATM and accounts)).")
			// TODO: Blow out the transactions storage
			// Reset the ATM balance
			err = persistence.ResetATMBalance(&atm)
			if err != nil {
				fmt.Println(err.Error())
			}
			continue
		}

		if instruction[0] == "authorize" {
			if len(instruction) != 3 {
				fmt.Println("Command 'authorize' must include account ID and PIN, space separated")
				continue
			}

			fmt.Println("Account login authorize command received.")
			sessionAccount, err = persistence.LoadAccountByID(instruction[1])
			if err != nil {
				fmt.Println("Account ID not matched. Please check and try again.")
				continue
			}
			// Try to log in, if pin matches:
			err = sessionAccount.Authorize(instruction[2])
			if err != nil {
				fmt.Println("PIN does not match account. Please try again.")
				sessionAccount = domain.Account{} // For security but not strictly needed to reset this to nil-value
				continue
			}
			fmt.Println("Authorized.")
			// If there is no activity for 2 minutes, your program should automatically log out the account
			// Start logout timer (reset logout timer to two mins on next input)

			// Start timer
			// So we might not need account.authorized (feels a little weak; less generic)
			continue
		}

		if sessionAccount.ID == "" || !sessionAccount.Authorized() {
			fmt.Println("Please authorize first.")
			continue
		}

		switch instruction[0] {
		case "withdraw":
			fmt.Println("Account withdraw command received.")
			if sessionAccount.Overdrawn() {
				fmt.Println("Your account is overdrawn. You may not make withdrawals at this time.")
				break
			}
			if len(instruction) != 2 {
				fmt.Println("Must include one amount only like this 'withdraw 100'")
				break
			}
			amount, err = validateAmount(instruction[1], instruction[0])
			if err != nil {
				break
			}
			amount, feeAmount, err = validateWithdrawlAmount(amount, atm, sessionAccount)

			err = atm.Withdraw(amount)
			if err != nil {
				fmt.Println(err.Error())
				break
			}

			transact(-1*amount, sessionAccount)
			if feeAmount != 0 {
				transact(feeAmount, sessionAccount)
			}

		case "deposit":
			fmt.Println("Account deposit command received.")
			if len(instruction) != 2 {
				fmt.Println("Must include one amount only like this 'withdraw 100'")
				break
			}
			amount, err = validateAmount(instruction[1], instruction[0])
			if err != nil {
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
			fmt.Println(sessionAccount.ReadBalance())

		case "history":
			fmt.Println("Account history command received.")
			transactions, err = persistence.LoadTransactions(sessionAccount.ID)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(transactions)

		case "logout":
			fmt.Println("Account logout command received.")
			sessionAccount = domain.Account{}

		default:
			fmt.Println("Instruction not recognised. Enter '?' for list of instructions.")
		}
	}
}

func validateAmount(amount string, instruction string) (domain.USD, error) {
	USDAmount, err := domain.USDFromString(amount)
	if err != nil {
		fmt.Sprintln("Amount must be numeric like this '%s 100'", instruction)
		USDAmount = domain.USD(0)
	}
	return USDAmount, err
}

func validateWithdrawlAmount(amount domain.USD, atm domain.ATM, account domain.Account) (amountAllowed domain.USD, fee domain.USD, err error) {
	amountAllowed, fee, err = atm.ValidateWithdrawal(amount)
	if err != nil {
		fmt.Println(err.Error())
	}
	amountAllowed, fee, err = account.ValidateWithdrawl(amountAllowed)
	if err != nil {
		fmt.Println(err.Error())
	}
	return amountAllowed, fee, err
}

func transact(amount domain.USD, account domain.Account) {
	_ = account.Update(amount)
	// err here isn't really allowed. How to recover? We already gave the user the cash from the machine
	// Propose: Log a fatal error and turn the machine off (maintenance mode)
	// TODO: Implement that ^^ (error here can occur if the two minute timer triggers a logout)
	// So we might think about letting the transaction complete before logging out
	// or remove the need to be authorised just to update an account balance
	_ = persistence.RecordTransaction(domain.NewTransaction(amount, account))
	// TODO: Handle both these errors by logging fatal, and suspend machine (no spec on suspension)
}
