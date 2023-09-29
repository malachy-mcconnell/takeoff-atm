## ATM CLI

A CLI to mimick an ATM.

To run the code, from the root directory
```bash
go run .
```
or similar, starts the program running.

There is a binary, `bin/takeoff-atm` that was compiled on an Apple Mac peronal
computer, which may or may not run on other Apple Mac personal computers.


### User commands
```bash
authorize <accoount id> <pin>
withdraw <amount (number, multiple of twenty, eg 100)>
deposit <amount>
balance
history
logout
```
Users are automatically logged out afer two minutes of inactivity


### Admin commands
Note that the bank state is preserved between runs: account balances, atm balance, 
transaction history. To reset to all defaults, use the `reset` command.
```bash
atm_balance
reset
end
```
There is no authorization needed to run admin commands.

Log entries are written to `data/atm.log` and persist even after `reset`, 
so delete the log file manually if you want to remove log entries.


### Run tests

Only the code in the `bank/` folder has tests. Run them like this

```bash
~/Dev/takeoff-atm>cd bank 
~/Dev/takeoff-atm/bank>
~/Dev/takeoff-atm/bank>
~/Dev/takeoff-atm/bank>go test
Running Suite: Domain Suite - /Users/malachy/Dev/takeoff-atm/bank
=================================================================
Random Seed: 1695936800

Will run 3 of 3 specs
•••

Ran 3 of 3 Specs in 0.000 seconds
SUCCESS! -- 3 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS
ok  	github.com/malachy-mcconnell/takeoff-atm/bank	0.234s
~/Dev/takeoff-atm/bank>
```