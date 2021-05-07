package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const flagDataDir = "datadir"
// const flagIP = "ip"
const flagPort = "port"

func main() {
	var tbbCmd = &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(runCmd())

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path where all data will/is stored")

	cmd.MarkFlagRequired(flagDataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
