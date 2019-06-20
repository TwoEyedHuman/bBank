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

func withdraw(acct Account, db *sql.DB, amt, withdrawDate time.Time) (status int) {
    if getBalance(acct, db, xDate) < amt {
        status = 1
        return
    }
    sqlStrParam := `
        select fromAcc
            ,toAcc
            ,amt
            ,nullified
            ,xDate
            ,effInterestRate
        from transactions
        where toaccid = $1
        and nullified = false
        order by xDate desc
        ;
    `
    curr := Currency{"US Dollars", "USD", "$"}
    rows, err :=db.Query(sqlStrParam, acct.acctNum)
    var xtns []Transaction
    run_sum := 0
    for rows.Next() {
        var fromAcc, toAcc int
        var amt float64
        var nullified bool
        var xDate time.Time
        var effInterestRate float64
        err = rows.Scan(&fromAcc, &toAcc, &amt, &nullified, &xDate, &effInterestRate)
        if err != nil {
            print("Error pulling transaction for withdraw calculation.")
            panic(err)
        }
        xtn := Transaction{fromAcc, toAcc, amt, nullified, xDate, effInterestRate}
        xtns = append(xtns, xtn)
    }
}

func sliceSumUntil(xtns []Transaction, amt float64) {
    run_sum := 0
    for _, xtn := range xtns {
        run_sum += xtn.amt
        if run_sum >= amt {
            
        }
    }
}

func deposit(acct Account, db *sql.DB, amt, xDate time.Time, effInterestRate float64) (status int) {
    sqlStrParam := `
        insert into transactions (fromAccId, toAccId, amount, xDate, nullified, effInterestRate) values ($1, $2, $3, $4, $5, $6);
    `

    _, err := db.Exec(sqlStrParam, 1, acct.acctNum, amt,  xDate.Format(RFC3339FullDate), false, effInterestRate)

    if err != nil {
        print("Error pushing deposit to server.")
        status = 0
        panic(err)
    }
    status = 1
    return
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

func getBalance(acct Account, db *sql.DB, pullDate time.Time) (balance float64) {
    // pull in transactions that went into the account that havent been nullified
    sqlStrParam := `
        select toAccId
            ,xDate
            ,amount
            ,effInterestRate
        from transactions
        where toaccid = $1
        and nullified = false;`

    rows, err := db.Query(sqlStrParam, acct.acctNum)  // run the query

    // go through each valid transaction, and calculate its current value, adding to the balance
    balance = 0
    for rows.Next() {
        var toaccid int
        var xDate time.Time
        var amount float64
        var effInterestRate float64
        err = rows.Scan(&toaccid, &xDate, &amount, &effInterestRate)
        if err != nil {
            print(fmt.Sprintf("Error pulling transaction for account %d.\n", acct.acctNum))
            panic(err)
        }
        balance += calcInterest(amount, effInterestRate, xDate, pullDate, "year")
    }
   return
}

func main() {
    argsWithoutProg := os.Args[1:]

    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+ "password=%s dbname=%s sslmode=disable", host, port, user, argsWithoutProg, dbname)
    db, err := sql.Open("postgres", psqlInfo)

    if err != nil {
        panic(err)
    }

    defer db.Close()

    err = db.Ping()

    if err != nil {
        fmt.Println("Error on DB ping.")
        panic(err)
    }

    specAcc := Account{354, "Savings", Currency{"US Dollars", "USD", "$"}}

    fmt.Println(getBalance(specAcc, db, time.Now()))
}
