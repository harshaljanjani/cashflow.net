package db

import (
	"context"
	"database/sql"
	"fmt"
)

// store provides all functions to execute db operations (individually) and transactions (combination of individual db operations)
// extend the functionality of Queries struct (only supports single transaction)
type Store struct{
	// embed/composition instead of inheritance
	*Queries
	db *sql.DB
}

// creates a new store
func NewStore(db *sql.DB) *Store{
	return &Store{
		db:db,
		Queries: New(db), 
	}
}

// implemented the following
// executes a function within the database transaction (TODO: implement rollback)
// 1) takes a context and a callback function as an input
// 2) start new db transaction
// 3) create a new Queries object with that transaction
// 4) call the callback function with the created queries
// 5) commit or rollback based on error returned

// don't want external package to call it directly : execTx
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error{
	tx, err := store.db.BeginTx(ctx, nil) // set custom isolation level with &sql.TxOptions{}
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil{
		if rbErr := tx.Rollback(); rbErr != nil{
			return fmt.Errorf("tx err %v, rb err: %v", err, rbErr)
		}
		return err
	}	
	return tx.Commit()
}

// money transfer function: TransferTx performs a money transfer from one account to another
// It creates a transfer record, add account entries, and update account's balance within a single canned transaction

// input params of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// output params of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// unused variable: var txKey = struct{}{}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error){
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error{
		// accessing variable result from outside the scope of this function (callback becomes closure (no generics in Go))
		
		// 1) create transfer record
		var err error

		// txName := ctx.Value(txKey)
		// write locks
		// logs: fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, TransferTxParams{
			FromAccountID:arg.FromAccountID,
			ToAccountID:arg.ToAccountID,
			Amount:arg.Amount,
		})
		if err != nil{
			return err
		}

		// 2) account entries creation
		// logs: fmt.Println(txName, "create entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount, // money is moving out
		})
		if err != nil{
			return err
		}

		// logs: fmt.Println(txName, "create entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount, // money is moving in
		})
		if err != nil{
			return err
		}

		// old: 3) get account => update balance
		// moving money out of the fromAccount
		// logs: fmt.Println(txName, "get account 1 for update")
		// account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		// if err != nil{
		// 	return err
		// }

		// new: better way of implementing the UpdateBalance operation with a single handler
		// logs: fmt.Println(txName, "update balance of account 1")
		// added deadlock avoidance mechanism:
		// always update account with smaller AccountID first
		if(arg.FromAccountID < arg.ToAccountID){
			// old: moving money into the toAccount
			// fmt.Println(txName, "get account 2 for update")
			// account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
			// if err != nil{
			// 	return err
			// }
			// new: better way of implementing the UpdateBalance operation with a single handler
			// fmt.Println(txName, "update balance of account 2")
			result.FromAccount, result.ToAccount, err = addMoney(ctx,q,arg.FromAccountID,-arg.Amount,arg.ToAccountID,arg.Amount)
			if err != nil{
				return err
			}
			
		} else {
			// update ToAccount first (arg.ToAccountID < arg.FromAccountID)
			result.ToAccount, result.FromAccount, err = addMoney(ctx,q,arg.ToAccountID,arg.Amount,arg.FromAccountID,-arg.Amount)
			if err != nil{
				return err
			}
		}
		return nil
	})
	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error){
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID1,
		Amount: amount1,
	})
	if err != nil{
		return // return account1, account2, err
	}
	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID2,
		Amount: amount2,
	})
	return // return account1, account2, err
}