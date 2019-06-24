package main

import "time"

import _"github.com/lib/pq"
/*
const (
    host = "localhost"
    port = 5432
    user = "blocke"
    dbname = "drexel"
)
*/
const RFC3339FullDate = "2006-01-02"

type Currency struct {
    fullName string
    shortName string
    symbol string
}

type Transaction struct {
    xtnId int
    fromAcc int
    toAcc int
    amt float64
    nullified bool
    xDate time.Time
    currency Currency
    effInterestRate float64  // effective interest rate
}

type Account struct {
    acctNum int
    acctType string
    currency Currency
}


func main() {
    userInterface()
}
