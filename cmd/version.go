package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = ""
	BuiltBy = ""
)

var (
	versionOutput string
	versionShort  bool
	verbose       bool
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	BuiltBy string `json:"builtBy"`
	Go      string `json:"go"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "By default prints a single line like: 'k8s-cleanup version v0.1.0'. Use --verbose or --output json for more details.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if Date == "" {
			Date = time.Now().Format(time.RFC3339)
		}
		if versionShort && versionOutput == "text" && !verbose {
			fmt.Println(Version)
			return nil
		}

		vi := versionInfo{
			Version: Version,
			Commit:  Commit,
			Date:    Date,
			BuiltBy: BuiltBy,
			Go:      runtime.Version(),
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		}

		switch versionOutput {
		case "json":
			b, _ := json.MarshalIndent(vi, "", "  ")
			fmt.Println(string(b))
		default:
			if verbose {
				fmt.Printf("k8s-cleanup %s (%s)\nBuilt: %s by %s\nGo: %s %s/%s\n",
					vi.Version, vi.Commit, vi.Date, vi.BuiltBy, vi.Go, vi.OS, vi.Arch)
			} else {
				fmt.Printf("k8s-cleanup version %s\n", vi.Version)
			}
		}
		return nil
	},
}

func init() {
	versionCmd.Flags().StringVarP(&versionOutput, "output", "o", "text", "Output format: text|json")
	versionCmd.Flags().BoolVar(&versionShort, "short", false, "Print only the version number")
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed fields (commit, date, Go)")
	rootCmd.AddCommand(versionCmd)

	// also support `k8s-cleanup --version` one-liner
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("k8s-cleanup version {{.Version}}\n")
}
