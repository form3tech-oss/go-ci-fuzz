package cmd

import (
	"github.com/form3tech-oss/go-ci-fuzz/fuzz"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

const (
	flagFuzzTime = "fuzz-time"
	flagFailFast = "fail-fast"
	flagOut      = "out"
)

var fuzzCmd = &cobra.Command{
	Use:   "fuzz [packages...]",
	Short: "Runs all fuzz targets of packages",
	Long: `Runs all fuzz targets in <packages> in current directory for the duration of --fuzz-time / N where N is the number of fuzz targets.
Continues to the next fuzz target on failure unless --fail-fast is defined.

Failing outputs are written to --out directory if specified. The structure is identical to how corpora is stored locally.
e.g.
out-dir
└── testdata
    └── fuzz
        └── FuzzTarget
            └── 0a7e5e215d8c088d4b9c4993d0189a07e81603fbdf64f2ca44738aa27159acef
`,
	Run:          fuzzRun,
	SilenceUsage: true,
}

func init() {
	fuzzCmd.Flags().StringP(flagOut, "o", "", "directory to write failing outputs to")
	fuzzCmd.Flags().Duration(flagFuzzTime, 10*time.Minute, "fuzzing duration for the whole suite")
	fuzzCmd.Flags().Bool(flagFailFast, false, "exit once failing input is discovered")
}

func fuzzRun(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	quiet, err := cmd.Flags().GetBool(flagQuiet)
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	fuzzTime, err := cmd.Flags().GetDuration(flagFuzzTime)
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	out, err := cmd.Flags().GetString(flagOut)
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	failFast, err := cmd.Flags().GetBool(flagFailFast)
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	proj := &fuzz.Project{
		Directory: wd,
		Quiet:     quiet,
	}

	packages := []string{"."}
	if len(args) > 0 {
		packages = args
	}

	targets, err := proj.ListFuzzTargets(ctx, packages...)
	if err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	if len(targets) == 0 {
		cmd.Println("No fuzz tests found")
		os.Exit(0)
	}

	timePerTarget := time.Duration(fuzzTime.Milliseconds()/int64(len(targets))) * time.Millisecond

	hasFailures := false

	cmd.Printf("go-ci-fuzz: discovered %d targets, each of them will be fuzzed for %s\n", len(targets), timePerTarget)
	for _, target := range targets {
		cmd.Printf("go-ci-fuzz: fuzzing %s for %s\n", target, timePerTarget)
		if err := proj.Fuzz(ctx, target, timePerTarget); err != nil {
			hasFailures = true
			if inputErr, ok := err.(fuzz.FailingInputError); ok {
				if inputErr.File != "" && out != "" {
					srcFile := filepath.Join(proj.Directory, inputErr.File)
					destFile := filepath.Join(out, inputErr.File)
					destFileFolder := filepath.Dir(destFile)

					if err := os.MkdirAll(destFileFolder, 0755); err != nil {
						cmd.PrintErrf("error creating %s directory when copying a failing input from %s: %s\n", destFileFolder, inputErr.File, err)
						os.Exit(1)
					}

					if err := fuzz.CopyFile(destFile, srcFile, 0644); err != nil {
						cmd.PrintErrf("copying a failing input from %s to %s: %s\n", srcFile, destFile, err)
						os.Exit(1)
					}

					cmd.Printf("Found failing input, saving to %s\n", destFile)
				} else {
					cmd.Printf("Found %s, not saving\n", inputErr)
				}
			} else {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			if failFast {
				os.Exit(2)
			}
		}
	}

	if hasFailures {
		os.Exit(2)
	}
}
