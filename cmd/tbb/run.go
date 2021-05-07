package main

import (
	"fmt"
	"os"

	"github.com/KinyaElGrande/TBB/node"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the TBB Node and its HTTP API",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, _ := cmd.Flags().GetString("dataDir")
			// ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)

			fmt.Println("Launching TBB NODE and its HTTP API")

			bootstrap := node.NewPeerNode(
				"127.0.0.1",
				8000,
				true,
				false,
			)

			n := node.New(dataDir, port, bootstrap)

			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)

	return runCmd
}
