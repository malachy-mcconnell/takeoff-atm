package persistence

import (
	"encoding/csv"
	"github.com/malachy-mcconnell/takeoff-atm/bank"
	"io"
	"log"
	"os"
)

const pathTransactions = "./data/transactions.csv"

type Transactions []bank.Transaction

func appendCSVRow(t Transactions, row []string) (Transactions, error) {
	transaction, err := bank.NewTransactionFromCSVRow(row)
	if err != nil {
		return Transactions{}, err
	}
	return append(t, transaction), nil
}

// RecordTransaction open file for writing, append this one transaction and close file
// TODO: Make scalable; essentially a write lock (mutex but better to use a channel and one go routine)
// so, think: a CSVWriterChannel, send the details to the channel, log on failure
// I mean if persistence is asynchronous, what to do when a save fails? [switch ATM off?]
func RecordTransaction(t bank.Transaction) error {
	f, err := os.OpenFile(pathTransactions, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Unable to open transactions file for writing "+pathTransactions, err)
	}
	defer f.Close()

	csvWriter := csv.NewWriter(f)
	csvWriter.Write(t.AsStringSlice())
	csvWriter.Flush()
	return csvWriter.Error()
}

func LoadTransactions(ID string) (Transactions, error) {
	// open file, read line by line, keep the ones that match the ID
	f, err := os.Open(pathTransactions)
	if err != nil {
		log.Println("Unable to read input file "+pathTransactions, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	accountTransactions := Transactions{}
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Unable to parse file as CSV for "+pathTransactions, err)
		}
		if row[0] == ID {
			accountTransactions, err = appendCSVRow(accountTransactions, row)
			if err != nil {
				return accountTransactions, err
			}
		}
	}
	return accountTransactions, nil
}

func ResetTransactionsStorage() error {
	return os.Truncate(pathTransactions, 0)
}

// AsTextTable <date> <time> <amount> <balance after transaction>
func (t Transactions) AsTextTable() string {
	table := "\n"
	for _, transaction := range t {
		table += transaction.AsTextRow()
	}
	return table
}
