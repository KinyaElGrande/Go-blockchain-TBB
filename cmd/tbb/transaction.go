package main

import (
	"fmt"
	"os"

	"github.com/KinyaElGrande/TBB/database"
	"github.com/spf13/cobra"
)

const flagFrom = "from"
const flagTo = "to"
const flagValue = "value"
const flagData = "data"

func transactionCmd() *cobra.Command {
	var transactionsCmd = &cobra.Command{
		Use:   "tx",
		Short: "Interact with Transactions (add ...)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	transactionsCmd.AddCommand(transactionAddCmd())

	return transactionsCmd
}

func transactionAddCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Adds a new transaction to database",
		Run: func(cmd *cobra.Command, args []string) {
			from, _ := cmd.Flags().GetString(flagFrom)
			to, _ := cmd.Flags().GetString(flagTo)
			value, _ := cmd.Flags().GetUint(flagValue)
			data, _ := cmd.Flags().GetString(flagData)

			fromAcc := database.NewAccount(from)
			toAcc := database.NewAccount(to)

			transaction := database.NewTransaction(
				fromAcc,
				toAcc,
				value,
				data,
			)

			state, err := database.NewStateFromDisk()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			defer state.Close()

			// Add the transaction to an IN-Memory array (pool)
			err = state.Add(transaction)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			_, err = state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Println("Transaction successfully added to the ledger")
		},
	}

	cmd.Flags().String(flagFrom, "", "From what account to send tokens")
	cmd.MarkFlagRequired(flagFrom)

	cmd.Flags().String(flagTo, "", "To what account to send tokens")
	cmd.MarkFlagRequired(flagTo)

	cmd.Flags().Uint(flagValue, 0, "How many Tokens you send")
	cmd.MarkFlagRequired(flagValue)

	cmd.Flags().String(flagData, "", "Transaction data")

	return cmd
}
