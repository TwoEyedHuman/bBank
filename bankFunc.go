package main

import (
    "database/sql"
    "strconv"
    "fmt"
    "os"
    "time"
    "strings"
    "sort"
    "math"
    "bufio"
    "os/user"
)

/////////////////// User Interface Functions /////////////////////

func cmdHandler(cmd string, db *sql.DB) (retVal int) {
    // cmd : the string of the user input
    // db : connection to the database

    cmd_tkn := strings.Split(strings.Trim(cmd, "\n"), " ")  // tokenize command for easy parsing

    if cmd_tkn[0] == "balance" {  // balance acctId
        if len(cmd_tkn) == 2 {
            acctId, _ := strconv.Atoi(cmd_tkn[1])
            dispBalance(acctId, db)
            retVal = 0
        } else {
            dispError("Incorrect parameters supplied for balance request.")
        }
    } else if cmd_tkn[0] == "deposit" {  // deposit acctId amt interestRate
        if len(cmd_tkn) == 4 {
            acctId, _ := strconv.Atoi(cmd_tkn[1])
            amt, _ := strconv.ParseFloat(cmd_tkn[2], 64)
            intRate, _ := strconv.ParseFloat(cmd_tkn[3], 64)
            retVal = deposit(acctId, db, amt, time.Now(), intRate)
        } else {
            dispError("Incorrect parameters supplied for deposit request.")
        }
    } else if cmd_tkn[0] == "withdraw" {  // withdraw acctId amt
        if len(cmd_tkn) == 3 {
            acctId, _ := strconv.Atoi(cmd_tkn[1])
            amt, _ := strconv.ParseFloat(cmd_tkn[2], 64)
            retVal = withdraw(acctId, db, amt, time.Now())
        } else {
            dispError("Incorrect parameters supplied for withdraw request.")
        }
    } else if cmd_tkn[0] == "exit" || cmd_tkn[0] == "quit" {
        retVal = 1
    } else {
        dispError("Invalid command. Try again.")
    }

    return
}

func getNewXtnId(db *sql.DB) (nextXtnId int) {
    sqlStr := "select max(xtnId) from transactions;"

    rows, err := db.Query(sqlStr)

    if err != nil {
        print("Error pulling transaction ID")
        panic(err)
    }

    for rows.Next() {
        var maxXtnId int
        err = rows.Scan(&maxXtnId)
        if err == nil {
            nextXtnId = maxXtnId + 1
        }
    }

    return
}

func dispError(str string) {
    fmt.Println("-------------------------------------")
    fmt.Sprintf("ERROR: %s\n", str)
    fmt.Println("-------------------------------------")
}

func dispBalance(acctId int, db *sql.DB) {
    fmt.Println("--------------------------------------")
    fmt.Print(fmt.Sprintf("Account ID: %d\n Balance: %.2f\n", acctId, getBalance(acctId, db,  time.Now())))
    fmt.Println("--------------------------------------")
}

func buildXtns(acctId int, db *sql.DB) ([]Transaction) {
    sqlStrParam := `
        select xtnId
            ,fromAccId
            ,toAccId
            ,amount
            ,xDate
            ,effInterestRate
        from transactions
        where toaccid = $1
        and nullified = false
        order by xDate desc
        ;
    `
    curr := Currency{"US Dollars", "USD", "$"}
    rows, err :=db.Query(sqlStrParam, acctId)

    if err != nil {
        panic(err)
    }

    // build a slice of Transactions for this account
    var xtns []Transaction
    for rows.Next() {
        var xtnId int
        var fromAccId, toAccId int
        var amt float64
        var xDate time.Time
        var effInterestRate float64
        err = rows.Scan(&xtnId, &fromAccId, &toAccId, &amt, &xDate, &effInterestRate)
        if err != nil {
            print("Error pulling transaction for withdraw calculation.")
            panic(err)
        }
        xtn := Transaction{xtnId, fromAccId, toAccId, amt, false, xDate, curr, effInterestRate}
        xtns = append(xtns, xtn)
    }
    return xtns
}

func nullifyXtn(xtnId int, db *sql.DB) int {
    // xtnId : id number of the transaction
    // db : connection to the database

    sqlStrParam := `
        update transactions
        set nullify = true
        where xtnId = $1
        ;
    `
    _, err := db.Exec(sqlStrParam, xtnId)
    if err != nil {
        return 1
    } else {
        return 0
    }
}

func nullifyXtns(xtnIds []int, db *sql.DB) {
    for _, xtnId := range xtnIds {
        err := nullifyXtn(xtnId, db)
        if err == 1 {
            _ = nullifyXtn(xtnId, db)
        }
    }
}

func withdraw(acctId int, db *sql.DB, amt float64, withdrawDate time.Time) (status int) {
    // acct : the account that will be withdrawn from
    // db : connection to the database
    // amt : the amount that is to be withdrawn
    // withdrawDate : date that the amount is withdrawn

    //ensure there is enough in the account to cover the withdraw
    if getBalance(acctId, db, withdrawDate) < amt {
        status = 1
        return
    }

    // extract all valid deposits for this account
    xtns := buildXtns(acctId, db)

    // sort the transactions from latest to earliest
    sort.Slice(xtns, func(i, j int) bool {
        if xtns[i].xDate.After(xtns[j].xDate) {
            return true
        }
        if xtns[i].xDate.Before(xtns[j].xDate) {
            return false
        }
        return xtns[i].xtnId > xtns[j].xtnId
    })

    nullifyXtnsList, newXtn := idWithdrawNullXtn(xtns, amt,  withdrawDate)


    nullifyXtns(nullifyXtnsList, db)

    _ = deposit(newXtn.toAcc, db, newXtn.amt, newXtn.xDate, newXtn.effInterestRate)

    status = 0
   return
}

func cumulativeSum(slc []float64) ([]float64) {
    cumSum := make([]float64, len(slc))
    run_sum := 0.0
    for indx, ele := range slc {
        fmt.Sprintf("(%d, %d)\n", indx, ele)
        run_sum += ele
        cumSum[indx] = run_sum
    }
    return cumSum
}

func idWithdrawNullXtn(xtns []Transaction, amt float64, withdrawDate time.Time) ([]int, Transaction) {
    // xtns : a slice of Transactions ordered from latest to earliest
    // amt : the amount that is to be withdrawn
    // withdrawDate : date of withdrawal

    // iterate over each transaction, and calulcate the current value (including interest), and remove transactions until there is enough to cover the withdraw amount
    run_sum := 0.0
    var nullXtnIds []int
    var newXtn Transaction
    for _, xtn := range xtns {
        run_sum += calcInterest(xtn.amt, xtn.effInterestRate, xtn.xDate, withdrawDate, "year")
        if (run_sum > amt) {
            newXtn = Transaction{0, -1, xtns[0].toAcc, run_sum - amt, false, withdrawDate, xtn.currency, xtn.effInterestRate}
            break
        } else {
            nullXtnIds = append(nullXtnIds, xtn.xtnId)
        }
    }
    return nullXtnIds, newXtn
}

//func sliceSumUntil(xtns []Transaction, amt float64) {
//    run_sum := 0
//    for _, xtn := range xtns {
//        run_sum += xtn.amt
//        if run_sum >= amt {
//            
//        }
//    }
//}

func deposit(acctId int, db *sql.DB, amt float64, xDate time.Time, effInterestRate float64) (status int) {
    sqlStrParam := `
        insert into transactions (xtnId, fromAccId, toAccId, amount, xDate, nullified, effInterestRate) values ($1, $2, $3, $4, $5, $6, $7);
    `
    newXtnId := getNewXtnId(db)

    _, err := db.Exec(sqlStrParam, newXtnId, 1, acctId, amt,  xDate.Format(RFC3339FullDate), false, effInterestRate)

    if err != nil {
        print("Error pushing deposit to server.")
        status = 1
        panic(err)
    }
    status = 0
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

func getBalance(acctId int, db *sql.DB, pullDate time.Time) (balance float64) {
    // pull in transactions that went into the account that havent been nullified
    sqlStrParam := `
        select toAccId
            ,xDate
            ,amount
            ,effInterestRate
        from transactions
        where toaccid = $1
        and nullified = false;`

    rows, err := db.Query(sqlStrParam, acctId)  // run the query

    if err != nil {
        panic(err)
    }

    // go through each valid transaction, and calculate its current value, adding to the balance
    balance = 0
    for rows.Next() {
        var toaccid int
        var xDate time.Time
        var amount float64
        var effInterestRate float64
        err = rows.Scan(&toaccid, &xDate, &amount, &effInterestRate)
        if err != nil {
            print(fmt.Sprintf("Error pulling transaction for account %d.\n", acctId))
            panic(err)
        }
        balance += calcInterest(amount, effInterestRate, xDate, pullDate, "year")
    }
   return
}

func establishConn(host string, port int, usr string, pword string, dbname string, sslmode string) (*sql.DB) {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+ "password=%s dbname=%s sslmode=disable", host, port, usr, pword, dbname)
    db, err := sql.Open("postgres", psqlInfo)

    if err != nil {
        panic(err)
    }

    err = db.Ping()

    if err != nil {
        fmt.Println("Error on connecting to database.")
        panic(err)
    }

    err = db.Ping()

    if err != nil {
        panic(err)
    }

    return db
}

func userInterface() {
    user, err := user.Current()

    if err != nil {
        panic(err)
    }

    db := establishConn("localhost", 5432, user.Username, "pword", "drexel", "disable") //usr, pword, database

    reader := bufio.NewReader(os.Stdin)

    fmt.Println("Welcome to bBank!")

    status := 0
    for status == 0 {
        fmt.Print("bank> ")
        usr_in, _ := reader.ReadString('\n')

        usr_in =  strings.Replace(usr_in, "\n", "", -1)

        status = cmdHandler(usr_in, db)
    }
    defer db.Close()
}

