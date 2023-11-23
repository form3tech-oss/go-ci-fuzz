package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	flagQuiet = "quiet"
)

var rootCmd = &cobra.Command{
	Use:   "go-ci-fuzz",
	Short: "Run Go Fuzz targets in CI systems",
	Long: `go-ci-fuzz implements missing functionality in 'go test -fuzz' such as
- running multiple test targets in a single command
- extracting failed outputs
- corpus management
`,
	Example: `go-ci-fuzz fuzz ./... --fuzz-time 10m --out /tmp/failing-inputs`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(fuzzCmd)
	rootCmd.PersistentFlags().Bool(flagQuiet, false, "silences underlying Go CLI StdOut")
}
