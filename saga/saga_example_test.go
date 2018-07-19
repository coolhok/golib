package saga_test

import (
	"context"

	"github.com/cjysmat/golib/saga"
	_ "github.com/cjysmat/golib/saga/storage/memory"
)

// This example show how to initialize an Saga execution coordinator(SEC) and add Sub-transaction to it, then start a transfer transaction.
// In transfer transaction we deduce `100` from foo at first, then deposit 100 into `bar`, deduce & deduce wil both success or rollbacked.
func Example_sagaTransaction() {

	// 1. Define sub-transaction method, anonymous method is NOT required, Just define them as normal way.
	DeduceAccount := func(ctx context.Context, account string, amount int) error {
		// Do deduce amount from account, like: account.money - amount
		return nil
	}
	CompensateDeduce := func(ctx context.Context, account string, amount int) error {
		// Compensate deduce amount to account, like: account.money + amount
		return nil
	}
	DepositAccount := func(ctx context.Context, account string, amount int) error {
		// Do deposit amount to account, like: account.money + amount
		return nil
	}
	CompensateDeposit := func(ctx context.Context, account string, amount int) error {
		// Compensate deposit amount from account, like: account.money - amount
		return nil
	}

	// 2. Init SEC as global SINGLETON(this demo not..), and add Sub-transaction definition into SEC.
	saga.AddSubTxDef("deduce", DeduceAccount, CompensateDeduce).
		AddSubTxDef("deposit", DepositAccount, CompensateDeposit)

	// 3. Start a saga to transfer 100 from foo to bar.

	from, to := "foo", "bar"
	amount := 100
	ctx := context.Background()

	var sagaID uint64 = 2
	saga.StartSaga(ctx, sagaID).
		ExecSub("deduce", from, amount).
		ExecSub("deposit", to, amount).
		EndSaga()

	// 4. done.
}
