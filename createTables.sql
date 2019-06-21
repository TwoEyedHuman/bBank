drop table if exists transactions;

create table transactions (
    xtnId INTEGER PRIMARY KEY,
    fromAccId INTEGER NOT NULL,
    toAccId INTEGER NOT NULL,
    amount DECIMAL(18,8),
    xDate DATE,
    nullified BOOLEAN,
    effInterestRate DECIMAL(10,4)
)
;
