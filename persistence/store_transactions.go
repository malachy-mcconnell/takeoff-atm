package persistence

import (
	"encoding/csv"
	"github.com/malachy-mcconnell/takeoff-atm/domain"
	"io"
	"log"
	"os"
)

const pathTransactions = "../data/transactions.csv"

type Transactions []domain.Transaction

func (t Transactions) AppendCSVRow(row []string) error {
	transaction, err := domain.NewTransactionFromCSVRow(row)
	if err != nil {
		return err
	}
	t.Append(transaction)
	return nil
}

// NOT USED, Really. Combine.
func (t Transactions) Append(transaction domain.Transaction) {
	t = append(t, transaction)
}

func RecordTransaction(t domain.Transaction) error {
	// open file for writing,
	// write it
	// close file

	// TODO: Make scalable; essentially a write lock (mutex but better to use a channel and one go routine)
	// so, think: a CSVWriterChannel, send the details to the channel, log on failure
	// I mean if persistence is asynchronous, what to do when a save fails? [switch ATM off?]
	return nil
}

func LoadTransactions(ID string) (Transactions, error) {
	// open file, read line by line, keep the ones that match the ID
	f, err := os.Open(pathTransactions)
	if err != nil {
		log.Fatal("Unable to read input file "+pathTransactions, err)
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
			log.Fatal("Unable to parse file as CSV for "+pathTransactions, err)
		}
		if row[0] == ID {
			err = accountTransactions.AppendCSVRow(row)
			if err != nil {
				return accountTransactions, err
			}
		}
	}
	return accountTransactions, nil
}
