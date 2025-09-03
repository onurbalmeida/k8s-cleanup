package cmd

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	verbosity string
	exitCode  = 0
)

func setExitCode(code int) {
	if code > exitCode {
		exitCode = code
	}
}

var rootCmd = &cobra.Command{
	Use:           "k8s-cleanup",
	Short:         "Clean up old pods and jobs in Kubernetes",
	Long:          "k8s-cleanup scans your Kubernetes cluster and removes stale Pods and Jobs, with safe defaults, dry-run by default, and rich filters.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level := strings.ToLower(verbosity)
		switch level {
		case "trace":
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		default:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
		zerolog.DurationFieldUnit = time.Second

		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.SetConfigName("cleanup")
			viper.SetConfigType("yaml")
			viper.AddConfigPath(".")
			if home, err := os.UserHomeDir(); err == nil {
				viper.AddConfigPath(home)
				viper.AddConfigPath(home + "/.config/k8s-cleanup")
			}
		}
		_ = viper.ReadInConfig()
		return nil
	},
	Example: `  # Dry-run pods and jobs older than 24h in all namespaces
  k8s-cleanup run --all-namespaces --older-than 24h

  # Actually delete completed pods older than 7d in a namespace
  k8s-cleanup run --namespace my-ns --completed --failed=false --dry-run=false --older-than 7d

  # Use label selector and write NDJSON audit
  k8s-cleanup run --all-namespaces --label-selector app=myapp --audit-file audit.ndjson

  # Show version info
  k8s-cleanup version

  # Generate shell completion
  k8s-cleanup completion zsh`,
}

func Execute() int {
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("k8s-cleanup version {{.Version}}\n")

	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() +
		"\nEnvironment variables:\n  KUBECONFIG\tPath to kubeconfig file(s)\n")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (YAML)")
	rootCmd.PersistentFlags().StringVar(&verbosity, "log-level", "info", "Log level: trace|debug|info|warn|error")
	_ = viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))

	cobra.AddTemplateFunc("runtimeOS", func() string { return runtime.GOOS })
	cobra.AddTemplateFunc("runtimeArch", func() string { return runtime.GOARCH })

	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("command failed")
		if exitCode < 3 {
			exitCode = 3
		}
	}
	return exitCode
}
