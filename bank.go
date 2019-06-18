package main

import "fmt"
import "time"
import "math"
import "database/sql"
import "os"

import _"github.com/lib/pq"

const (
    host = "localhost"
    port = 5432
    user = ""
    dbname = ""
)

const RFC3339FullDate = "2006-01-02"

type Currency struct {
    fullName string
    shortName string
    symbol string
}

type Transaction struct {
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

func calcInterest(premium float64, interestRate float64, timeStart time.Time, timeEnd time.Time, interestTimeBase string) (calcInterest float64) {
    // calculate time passe din the units of interestTimeBase
    timeRatio := timeEnd.Sub(timeStart).Hours()/(365.0*24.0)
    calcInterest = premium * math.Exp(interestRate * timeRatio)
    return
}

func dispX(xtn Transaction) {
    fmt.Println("--------------------------------------------")
    fmt.Println(xtn.xDate.Format("2006-01-02 15:04:05 Monday"), xtn.currency.symbol, xtn.amt)
}

func getBalance(acct Account, db *sql.DB) (balance float64) {
    sqlStrDepParam := "select sum(amount) as amount from transactions where toaccid = %d and nullified = false;"
    sqlStrDep := fmt.Sprintf(sqlStrDepParam, acct.acctNum)

    sqlStrWthParam := "select sum(amount) as amount from transactions where fromaccid = %d and nullified = false;"
    sqlStrWth := fmt.Sprintf(sqlStrWthParam, acct.acctNum)

    rows, err := db.Query(sqlStrDep)

    var depSum float64
    for rows.Next() {
        err = rows.Scan(&depSum)
        if err != nil {
            depSum=0
        }
    }

    rows, err = db.Query(sqlStrWth)

    var wthSum float64
    for rows.Next() {
        err = rows.Scan(&wthSum)
        if err != nil {
            wthSum = 0
        }
    }

    balance = depSum - wthSum
    return
}

func main() {
    argsWithoutProg := os.Args[1:]

//    sDate, _ := time.Parse(RFC3339FullDate, "2011-01-19")
//    tDate, _ := time.Parse(RFC3339FullDate, "2012-01-19")
//    fmt.Println(calcInterest(100, 0.05, sDate, tDate, "year"))
//    xfr := Transaction{30, false, sDate, Currency{"US Dollars", "USD", "$"}}
//    dispX(xfr)

    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+ "password=%s dbname=%s sslmode=disable", host, port, user, argsWithoutProg, dbname)
    db, err := sql.Open("postgres", psqlInfo)

    if err != nil {
        panic(err)
    }

    defer db.Close()

    err = db.Ping()

    if err != nil {
        panic(err)
    }

    specAcc := Account{354, "Savings", Currency{"US Dollars", "USD", "$"}}

    fmt.Println(getBalance(specAcc, db))
}
