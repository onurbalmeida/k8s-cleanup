package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/onurbalmeida/k8s-cleanup/internal/engine"
	"github.com/onurbalmeida/k8s-cleanup/internal/helpers"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type cleanupRecord struct {
	Resource  string        `json:"resource"`
	Namespace string        `json:"namespace"`
	Name      string        `json:"name"`
	State     string        `json:"state"`
	Age       time.Duration `json:"age"`
	Deleted   bool          `json:"deleted"`
	DryRun    bool          `json:"dryRun"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"ts"`
}

var (
	dryRun               bool
	olderThan            string
	kinds                []string
	namespace            string
	allNS                bool
	excludeNS            []string
	labelSelector        string
	fieldSelector        string
	includeCompleted     bool
	includeFailed        bool
	includeEvicted       bool
	protectLabelKV       string
	concurrency          int
	output               string
	auditFile            string
	exitNonZeroOnChanges bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Scan and delete old Pods and Jobs",
	Long:  "Scans namespaces and deletes Pods/Jobs that match filters and exceed the given age threshold. Defaults to dry-run for safety.",
	RunE: func(cmd *cobra.Command, args []string) error {
		applyDefaults()
		syncFromViper()

		dur, err := time.ParseDuration(olderThan)
		if err != nil {
			return fmt.Errorf("invalid --older-than: %w", err)
		}

		cfg, err := clientConfig()
		if err != nil {
			return err
		}
		cs, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}

		pk, pv := helpers.ParseKV(protectLabelKV)

		nsList := []string(nil)
		if !allNS {
			if namespace == "" {
				nsList = []string{"default"}
			} else {
				nsList = []string{namespace}
			}
		}

		eng := engine.New(cs, engine.Config{
			OlderThan:         dur,
			Kinds:             kinds,
			AllNamespaces:     allNS,
			Namespaces:        nsList,
			ExcludeNamespaces: excludeNS,
			LabelSelector:     labelSelector,
			FieldSelector:     fieldSelector,
			IncludeCompleted:  includeCompleted,
			IncludeFailed:     includeFailed,
			IncludeEvicted:    includeEvicted,
			ProtectKey:        pk,
			ProtectVal:        pv,
		})

		cands, err := eng.FindCandidates(cmd.Context())
		if err != nil {
			return err
		}

		writer, closer, err := prepareAudit(auditFile)
		if err != nil {
			return err
		}
		if closer != nil {
			defer closer()
		}

		if concurrency < 1 {
			concurrency = 1
		}
		workCh := make(chan engine.Candidate)
		resCh := make(chan cleanupRecord)
		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for c := range workCh {
					rec := cleanupRecord{
						Resource:  c.Kind,
						Namespace: c.Namespace,
						Name:      c.Name,
						State:     c.State,
						Age:       c.Age,
						DryRun:    dryRun,
						Deleted:   false,
						Timestamp: time.Now(),
					}
					if dryRun {
						resCh <- rec
						continue
					}
					if err := eng.Delete(context.Background(), c); err != nil {
						rec.Error = err.Error()
						resCh <- rec
						continue
					}
					rec.Deleted = true
					resCh <- rec
				}
			}()
		}

		go func() {
			for _, c := range cands {
				workCh <- c
			}
			close(workCh)
			wg.Wait()
			close(resCh)
		}()

		errs, deleted := 0, 0
		var results []cleanupRecord
		for r := range resCh {
			results = append(results, r)
			if r.Error != "" {
				errs++
				log.Error().Str("kind", r.Resource).Str("ns", r.Namespace).Str("name", r.Name).Str("state", r.State).Dur("age", r.Age).Msg("delete failed")
			} else if r.DryRun {
				log.Info().Str("kind", r.Resource).Str("ns", r.Namespace).Str("name", r.Name).Str("state", r.State).Dur("age", r.Age).Msg("would delete")
			} else if r.Deleted {
				deleted++
				log.Info().Str("kind", r.Resource).Str("ns", r.Namespace).Str("name", r.Name).Str("state", r.State).Dur("age", r.Age).Msg("deleted")
			}
			if writer != nil {
				enc, _ := json.Marshal(r)
				_, _ = writer.Write(enc)
				_, _ = writer.WriteString("\n")
			}
		}
		if writer != nil {
			_ = writer.Flush()
		}

		switch strings.ToLower(output) {
		case "json":
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
		default:
		}

		if dryRun && exitNonZeroOnChanges && len(cands) > 0 {
			setExitCode(2)
		} else if !dryRun && errs > 0 {
			setExitCode(3)
		} else if !dryRun && deleted > 0 {
			setExitCode(2)
		}
		return nil
	},
}

func applyDefaults() {
	viper.SetDefault("dryRun", true)
	viper.SetDefault("olderThan", "24h")
	viper.SetDefault("kinds", []string{"pod", "job"})
	viper.SetDefault("allNamespaces", false)
	viper.SetDefault("excludeNamespaces", []string{"kube-system", "kube-public"})
	viper.SetDefault("completed", true)
	viper.SetDefault("failed", true)
	viper.SetDefault("evicted", true)
	viper.SetDefault("protectLabel", "keep=true")
	viper.SetDefault("concurrency", 10)
	viper.SetDefault("output", "text")
	viper.SetDefault("auditFile", "")
	viper.SetDefault("exitNonZeroOnChanges", false)
}

func syncFromViper() {
	dryRun = viper.GetBool("dryRun")
	olderThan = viper.GetString("olderThan")
	kinds = viper.GetStringSlice("kinds")
	namespace = viper.GetString("namespace")
	allNS = viper.GetBool("allNamespaces")
	excludeNS = viper.GetStringSlice("excludeNamespaces")
	labelSelector = viper.GetString("labelSelector")
	fieldSelector = viper.GetString("fieldSelector")
	includeCompleted = viper.GetBool("completed")
	includeFailed = viper.GetBool("failed")
	includeEvicted = viper.GetBool("evicted")
	protectLabelKV = viper.GetString("protectLabel")
	concurrency = viper.GetInt("concurrency")
	output = viper.GetString("output")
	auditFile = viper.GetString("auditFile")
	exitNonZeroOnChanges = viper.GetBool("exitNonZeroOnChanges")
}

func clientConfig() (*rest.Config, error) {
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides).ClientConfig()
}

func prepareAudit(path string) (*bufio.Writer, func(), error) {
	if path == "" {
		return nil, nil, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	w := bufio.NewWriter(f)
	closeFn := func() {
		_ = w.Flush()
		_ = f.Close()
	}
	return w, closeFn, nil
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", true, "Simulate without deleting")
	runCmd.Flags().StringVar(&olderThan, "older-than", "24h", "Age threshold (e.g., 30m, 24h, 7d)")
	runCmd.Flags().StringSliceVar(&kinds, "kind", []string{"pod", "job"}, "Resource kinds: pod,job")
	runCmd.Flags().StringVar(&namespace, "namespace", "default", "Target namespace")
	runCmd.Flags().BoolVar(&allNS, "all-namespaces", false, "Process all namespaces")
	runCmd.Flags().StringSliceVar(&excludeNS, "exclude-ns", []string{"kube-system", "kube-public"}, "Namespaces to exclude")
	runCmd.Flags().StringVar(&labelSelector, "label-selector", "", "Label selector")
	runCmd.Flags().StringVar(&fieldSelector, "field-selector", "", "Field selector")
	runCmd.Flags().BoolVar(&includeCompleted, "completed", true, "Include Completed/Succeeded")
	runCmd.Flags().BoolVar(&includeFailed, "failed", true, "Include Failed")
	runCmd.Flags().BoolVar(&includeEvicted, "evicted", true, "Include Evicted (pods)")
	runCmd.Flags().StringVar(&protectLabelKV, "protect", "keep=true", "Protect resources with this label (key[=value])")
	runCmd.Flags().IntVar(&concurrency, "concurrency", 10, "Concurrent deletions")
	runCmd.Flags().StringVar(&output, "output", "text", "Output format: text|json")
	runCmd.Flags().StringVar(&auditFile, "audit-file", "", "Write NDJSON audit events to file")
	runCmd.Flags().BoolVar(&exitNonZeroOnChanges, "exit-nonzero-on-changes", false, "Exit with code 2 if there are candidates (dry-run)")

	_ = viper.BindPFlag("dryRun", runCmd.Flags().Lookup("dry-run"))
	_ = viper.BindPFlag("olderThan", runCmd.Flags().Lookup("older-than"))
	_ = viper.BindPFlag("kinds", runCmd.Flags().Lookup("kind"))
	_ = viper.BindPFlag("namespace", runCmd.Flags().Lookup("namespace"))
	_ = viper.BindPFlag("allNamespaces", runCmd.Flags().Lookup("all-namespaces"))
	_ = viper.BindPFlag("excludeNamespaces", runCmd.Flags().Lookup("exclude-ns"))
	_ = viper.BindPFlag("labelSelector", runCmd.Flags().Lookup("label-selector"))
	_ = viper.BindPFlag("fieldSelector", runCmd.Flags().Lookup("field-selector"))
	_ = viper.BindPFlag("completed", runCmd.Flags().Lookup("completed"))
	_ = viper.BindPFlag("failed", runCmd.Flags().Lookup("failed"))
	_ = viper.BindPFlag("evicted", runCmd.Flags().Lookup("evicted"))
	_ = viper.BindPFlag("protectLabel", runCmd.Flags().Lookup("protect"))
	_ = viper.BindPFlag("concurrency", runCmd.Flags().Lookup("concurrency"))
	_ = viper.BindPFlag("output", runCmd.Flags().Lookup("output"))
	_ = viper.BindPFlag("auditFile", runCmd.Flags().Lookup("audit-file"))
	_ = viper.BindPFlag("exitNonZeroOnChanges", runCmd.Flags().Lookup("exit-nonzero-on-changes"))

	rootCmd.AddCommand(runCmd)
}
