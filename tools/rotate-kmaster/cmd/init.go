package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/kmaster"
	"github.com/spf13/cobra"
)

var initOutput string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate the initial K_master_v1 key and current pointer",
	RunE: func(cmd *cobra.Command, args []string) error {
		if initOutput == "" {
			return fmt.Errorf("--output required")
		}
		if err := os.MkdirAll(initOutput, 0o700); err != nil {
			return err
		}
		k, err := kmaster.GenerateLocalKey()
		if err != nil {
			return err
		}
		keyPath := filepath.Join(initOutput, "k_master_v1")
		curPath := filepath.Join(initOutput, "current")
		if _, err := os.Stat(keyPath); err == nil {
			return fmt.Errorf("refusing to overwrite existing %s", keyPath)
		}
		if err := os.WriteFile(keyPath, k, 0o600); err != nil {
			return err
		}
		if err := os.WriteFile(curPath, []byte("v1"), 0o600); err != nil {
			return err
		}
		fmt.Printf("initialized: %s (v1)\n", initOutput)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "", "Output directory (will hold k_master_v1 + current)")
	_ = initCmd.MarkFlagRequired("output")
}
