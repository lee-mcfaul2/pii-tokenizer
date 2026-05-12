package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "rotate-kmaster",
	Short: "K_master initialization and rotation for pii-tokenizer (aead-local backend)",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd, rotateAEADCmd)
}
