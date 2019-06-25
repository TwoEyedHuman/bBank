drop table if exists transactions;

-- build the transactions table
create table transactions (
    xtnId INTEGER PRIMARY KEY,  -- transaction ID
    fromAccId INTEGER NOT NULL,  -- transaction from
    toAccId INTEGER NOT NULL,  -- transaction to
    amount DECIMAL(18,8),  -- dollar amount of transaction
    xDate DATE,  -- date that the transaction should calculate interest from
    nullified BOOLEAN,  -- indicates if the transaction has been nullified, if true then it is not used in balance calculations
    effInterestRate DECIMAL(10,4)  --  interest rate of the transaction
)
;
