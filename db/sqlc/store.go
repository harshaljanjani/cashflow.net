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
	Amount 		  int64 `json:"amount"`
}

// output params of the transfer transaction
type TransferTxResult struct {
	Transfer 	Transfer `json:"transfer"`
	// after balance updation (entries in accounts, and transaction entries for materializing changes)
	FromAccount Account  `json:"from_account"`
	ToAccount 	Account  `json:"to_account"`
	FromEntry 	Entry	 `json:"from_entry"`
	ToEntry 	Entry 	 `json:"to_entry"`
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error){
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error{
		// yet to implement callback function
		return nil
	})
	return result, err
}