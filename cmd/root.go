package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nas-manager",
	Short: "NAS script and task management tool",
	Long:  "A CLI tool for managing scripts and tasks on your NAS system",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	loadConfig()
	rootCmd.AddCommand(acmeCmd)
	rootCmd.AddCommand(ddnsCmd)
}