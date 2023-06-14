package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	// concurrency control: best practice to run concurrent co-routines
	// writing logs (before)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	n := 5 // number of concurrent transactions
	amount := int64(10) // amount of transfer
	errs := make(chan error) // connect co-routines (without explicit locking)
	results := make(chan TransferTxResult)  
	for i := 0; i < n; i++{
		// start new go routine
		go func(){
			// removed adding value to the context
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
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
		// idea: create a map, add k values with every transaction
		// if k exists in the map already, fail test
		uniqueMap := make(map[int]bool)
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

			// writing logs (after each transaction)
			fmt.Println(">> after:", fromAccount.Balance, toAccount.Balance)
			// check balance
			diff1 := account1.Balance - fromAccount.Balance // account1 -> before transaction
			// account2 -> after transaction
			diff2 := toAccount.Balance - account2.Balance
			require.Equal(t, diff1, diff2)
			require.True(t, diff1 % amount == 0)
			// each time the transaction is made, the balance of account1 would be decreased by (1 * amount) -> 1 * amount, 2 * amount, 3 * amount ...
			k := int(diff1 / amount) 
			require.True(t, k >= 1 && k <= n)
			
			// check uniqueness of k
			require.NotContains(t, uniqueMap, k)
			uniqueMap[k] = true
		}

		// check final updated balance
		updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
		require.NoError(t, err)
		updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
		require.NoError(t, err)
		
		// writing logs (after all transactions)
		fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
		require.Equal(t, account1.Balance - int64(n) * amount, updatedAccount1.Balance)
		require.Equal(t, account2.Balance + int64(n) * amount, updatedAccount2.Balance)
	}

	// deadlock avoidance: found a possible deadlock scenario -> half the transactions going from account1 to account2, and the other half going the other way around
	// expected: no change in balance
	// outcome: deadlock
	func TestTransferTxDeadlock(t *testing.T) {
		store := NewStore(testDB)
	
		account1 := createRandomAccount(t)
		account2 := createRandomAccount(t)
		// concurrency control: best practice to run concurrent co-routines
		// writing logs (before)
		fmt.Println(">> before:", account1.Balance, account2.Balance)
	
		n := 10 // number of concurrent transactions
		amount := int64(10) // amount of transfer
		errs := make(chan error) // connect co-routines (without explicit locking)  
		for i := 0; i < n/2; i++{
			fromAccountID := account1.ID
			toAccountID := account2.ID
			// idea is to implement half the transactions from account1 to account2, and the other half from account2 to account1
			if i % 2 == 1{
				fromAccountID = account2.ID
				toAccountID = account1.ID
			}
			// start new go routine
			go func(){
				// removed adding value to the context
				ctx := context.Background()
				_ , err := store.TransferTx(ctx, TransferTxParams{
					FromAccountID: fromAccountID,
					ToAccountID: toAccountID,
					Amount: amount,
				})
				errs <- err
				// send error back to the main go routine and check from there
			}()
		}
			// check result
			for i := 0; i < n; i++{
				err := <- errs
				require.NoError(t, err)
			}
	
			// check final updated balance
			updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
			require.NoError(t, err)
			updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
			require.NoError(t, err)
			
			// writing logs (after all transactions)
			fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
			require.Equal(t, account1.Balance, updatedAccount1.Balance)
			require.Equal(t, account2.Balance, updatedAccount2.Balance)
		}