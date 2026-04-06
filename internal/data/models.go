package data

import (
	"database/sql"
	"errors"
)

type Models struct {
	Persons             PersonModel
	Customers           CustomerModel
	Accounts            AccountModel
	AccountOwnerships   AccountOwnershipModel
	JournalEntries      JournalEntryModel
	LedgerEntries       LedgerEntryModel
	AccountTransactions AccountTransactionModel
	Loans               LoanModel
	GLAccounts          GLAccountModel
	Users               UserModel
	Employees           EmployeeModel
}

func (m Models) NewModels(db *sql.DB) Models {
	return Models{
		Persons:             PersonModel{DB: db},
		Customers:           CustomerModel{DB: db},
		Accounts:            AccountModel{DB: db},
		AccountOwnerships:   AccountOwnershipModel{DB: db},
		JournalEntries:      JournalEntryModel{DB: db},
		LedgerEntries:       LedgerEntryModel{DB: db},
		AccountTransactions: AccountTransactionModel{DB: db},
		Loans:               LoanModel{DB: db},
		GLAccounts:          GLAccountModel{DB: db},
		Users:               UserModel{DB: db},
		Employees:           EmployeeModel{DB: db},
	}
}

// ErrRecordNotFound is returned when a database query does not find a matching record
var ErrRecordNotFound = errors.New("record not found")

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func (f Filters) sortColumn() string {
	switch f.Sort {
	case "first_name":
		return "first_name"
	case "last_name":
		return "last_name"
	case "created_at":
		return "created_at"
	default:
		return "id"
	}
}

func (f Filters) sortDirection() string {
	if len(f.Sort) > 0 && f.Sort[0] == '-' {
		return "DESC"
	}
	return "ASC"
}
