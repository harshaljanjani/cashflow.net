package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	// concurrency control: best practice to run concurrent co-routines
	n := 5 // number of concurrent transactions
	amount := int64(10) // amount of transfer
	errs := make(chan error) // connect co-routines (without explicit locking)
	results := make(chan TransferTxResult)  
	for i := 0; i < n; i++{
		// start new go  routine
		go func(){
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID: account2.ID,
				Amount: amount,
			})
			errs <- err
			results <- result
			// send error back to the main go routine and check from there
		}()
	}
		// check result
		// check if k is unique with every transaction
		for i := 0; i < n; i++{
			err := <- errs
			require.NoError(t, err)
			result := <- results
			require.NotEmpty(t, result)

			// check transfer
			transfer := result.Transfer
			require.NotEmpty(t, transfer)
			require.Equal(t, account1.ID, transfer.FromAccountID)
			require.Equal(t, account2.ID, transfer.ToAccountID)
			require.Equal(t, amount, transfer.Amount)
			require.NotZero(t,transfer.ID)
			require.NotZero(t,transfer.CreatedAt)

			_, err = store.GetTransfer(context.Background(), transfer.ID)
			require.NoError(t, err)

			// FromEntry
			fromEntry := result.FromEntry
			require.NotEmpty(t, fromEntry)
			require.Equal(t, account1.ID, fromEntry.AccountID)
			require.Equal(t,-amount,fromEntry.Amount)
			require.NotZero(t, fromEntry.ID)
			require.NotZero(t, fromEntry.CreatedAt)

			_, err = store.GetEntry(context.Background(), fromEntry.ID)
			require.NoError(t, err)

			// ToEntry
			toEntry := result.ToEntry
			require.NotEmpty(t, toEntry)
			require.Equal(t, account2.ID, toEntry.AccountID)
			require.Equal(t,amount,toEntry.Amount)
			require.NotZero(t, fromEntry.ID)
			require.NotZero(t, fromEntry.CreatedAt)
			_, err = store.GetEntry(context.Background(), toEntry.ID)
			require.NoError(t, err)
			
			// writing tests first instead (trying out TDD): check account balance
			// things to keep in mind: concurrency control and deadlock prevention
			fromAccount := result.FromAccount
			require.NotEmpty(t, fromAccount)
			require.Equal(t, account1.ID, fromAccount)
			// toAccount
			toAccount := result.ToAccount
			require.NotEmpty(t, toAccount)
			require.Equal(t, account1.ID, toAccount)

			// check balance
			diff1 := account1.Balance - fromAccount.Balance // account1 -> before transaction
			// account2 -> after transaction
			diff2 := toAccount.Balance - account2.Balance
			require.Equal(t, diff1, diff2)
			require.True(t, diff1 % amount == 0)
			// each time the transaction is made, the balance of account1 would be decreased by (1 * amount) -> 1 * amount, 2 * amount, 3 * amount ...
			k := int(diff1 / amount) 
			require.True(t, k >= 1 && k <= n)
			
			// TODO: check k value is unique or not
		}
	}