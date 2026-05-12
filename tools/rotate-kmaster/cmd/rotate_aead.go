package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lee-mcfaul2/pii-tokenizer/internal/kmaster"
	"github.com/spf13/cobra"
)

var rotateDir string

var rotateAEADCmd = &cobra.Command{
	Use:   "rotate-aead",
	Short: "Append a new K_master version to the aead-local key directory and advance current",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rotateDir == "" {
			return fmt.Errorf("--dir required")
		}
		highest, err := highestVersion(rotateDir)
		if err != nil {
			return err
		}
		next := highest + 1
		k, err := kmaster.GenerateLocalKey()
		if err != nil {
			return err
		}
		newPath := filepath.Join(rotateDir, fmt.Sprintf("k_master_v%d", next))
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("refusing to overwrite existing %s", newPath)
		}
		if err := os.WriteFile(newPath, k, 0o600); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(rotateDir, "current"), []byte(fmt.Sprintf("v%d", next)), 0o600); err != nil {
			return err
		}
		fmt.Printf("rotated: %s (now v%d)\n", rotateDir, next)
		return nil
	},
}

func init() {
	rotateAEADCmd.Flags().StringVarP(&rotateDir, "dir", "d", "", "Key directory (the same path used by aead-local backend)")
	_ = rotateAEADCmd.MarkFlagRequired("dir")
}

func highestVersion(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	highest := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "k_master_v") {
			continue
		}
		v, err := strconv.Atoi(strings.TrimPrefix(name, "k_master_v"))
		if err != nil {
			continue
		}
		if v > highest {
			highest = v
		}
	}
	if highest == 0 {
		return 0, fmt.Errorf("no k_master_v* files in %s; run `init` first", dir)
	}
	return highest, nil
}
