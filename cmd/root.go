package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nas-manager",
	Short: "CLI tool for DDNS and certificate management",
	Long:  "A CLI tool for DDNS and certificate management.\n\nFeatures:\n- DDNS Management: Update Cloudflare DNS records with current public IP\n- ACME Certificates: Issue/renew Let's Encrypt certificates via Cloudflare DNS",
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