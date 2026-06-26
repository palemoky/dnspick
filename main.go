package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"golang.org/x/term"

	"github.com/palemoky/dnspick/internal/buildinfo"
	"github.com/palemoky/dnspick/internal/dnsbench"
	"github.com/palemoky/dnspick/internal/i18n"
	"github.com/palemoky/dnspick/internal/ui"
	"github.com/palemoky/dnspick/internal/updater"
)

var (
	domainsStr       string
	serversStr       string
	queriesPerDomain int
	queryTimeout     time.Duration
	maxConcurrency   int
	noSystemDNS      bool
	langFlag         string
	jsonOutput       bool
)

var rootCmd = &cobra.Command{
	Use:           "dnspick",
	Version:       buildinfo.Version,
	RunE:          runBenchmark,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(buildinfo.String())
	},
}

var updateCmd = &cobra.Command{
	Use:  "update",
	RunE: runUpdate,
}

// setup wires up localized help text, flags and subcommands. It must run after
// the active language has been selected so that --help reflects --lang/$LANG.
func setup() {
	m := i18n.L()

	rootCmd.Short = m.CmdRootShort
	rootCmd.Long = m.CmdRootLong
	versionCmd.Short = m.CmdVersionShort
	updateCmd.Short = m.CmdUpdateShort

	rootCmd.SetVersionTemplate("{{.Version}}\n")

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&domainsStr, "domains", "d", "", m.FlagDomains)
	flags.StringVarP(&serversStr, "servers", "s", "", m.FlagServers)
	flags.IntVarP(&queriesPerDomain, "queries", "q", 3, m.FlagQueries)
	flags.DurationVarP(&queryTimeout, "timeout", "t", 2*time.Second, m.FlagTimeout)
	flags.IntVarP(&maxConcurrency, "concurrency", "c", 16, m.FlagConcurrency)
	flags.BoolVar(&noSystemDNS, "no-system-dns", false, m.FlagNoSystemDNS)
	flags.StringVar(&langFlag, "lang", "", m.FlagLang)
	flags.BoolVar(&jsonOutput, "json", false, m.FlagJSON)

	rootCmd.AddCommand(versionCmd, updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), updater.DefaultTimeout)
	defer cancel()

	m := i18n.L()
	fmt.Printf(m.UpdateChecking, buildinfo.Version)
	latest, updated, err := updater.Update(ctx, buildinfo.Version)
	if err != nil {
		return fmt.Errorf("%s %w", m.UpdateFailed, err)
	}
	if !updated {
		fmt.Printf(m.UpdateUpToDate, latest)
		return nil
	}
	fmt.Printf(m.UpdateDone, latest)
	return nil
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	m := i18n.L()

	// Domains: use the custom list when -d is given (classified as Custom),
	// otherwise fall back to the built-in categorized list.
	domains := dnsbench.DefaultDomains
	if cmd.Flags().Changed("domains") {
		domains = dnsbench.ParseDomains(domainsStr)
	}
	if len(domains) == 0 {
		return fmt.Errorf("%s", m.ErrNoDomains)
	}

	// Servers: the custom list when -s is given, otherwise the built-in list;
	// in both cases the system default DNS is appended unless disabled.
	servers := dnsbench.DefaultServers
	if cmd.Flags().Changed("servers") {
		servers = dnsbench.ParseServers(serversStr)
		if len(servers) == 0 {
			return fmt.Errorf("%s", m.ErrNoServers)
		}
	}
	if !noSystemDNS {
		if sys := dnsbench.DetectSystemDNS(m.SystemDNSName, m.SystemDNSNameN); len(sys) > 0 {
			servers = append(append([]dnsbench.Server{}, servers...), sys...)
		}
	}

	opts := dnsbench.Options{
		Servers:     servers,
		Domains:     domains,
		Queries:     queriesPerDomain,
		Timeout:     queryTimeout,
		Concurrency: maxConcurrency,
	}

	// JSON mode: stdout carries only the JSON document, status goes to stderr,
	// and the live progress UI is skipped so the output stays pipe-friendly.
	if jsonOutput {
		fmt.Fprintf(os.Stderr, m.BenchStarting, len(servers), len(domains))
		results := dnsbench.Run(opts, nil)
		return ui.WriteJSON(os.Stdout, results, queriesPerDomain, len(domains))
	}

	// Kick off a non-blocking check for a newer release; it runs concurrently
	// with the benchmark and the notice (if any) is printed at the end.
	updateCh := startUpdateCheck()

	fmt.Printf(m.BenchStarting, len(servers), len(domains))

	tracker := ui.NewStatusTracker(domains, len(servers), queriesPerDomain)
	tracker.Start()
	results := dnsbench.Run(opts, tracker.Progress)
	tracker.Stop()

	fmt.Println(m.ResultsHeader)
	ui.PrintResultsTable(results)

	fmt.Println(m.RecommendHeader)
	ui.PrintRecommendations(results)

	autoUpdate(updateCh)
	return nil
}

// updateCheckTimeout bounds the background "is there a newer release?" check so a
// slow or unreachable network never holds anything up for long.
const updateCheckTimeout = 3 * time.Second

// updateNoticeGrace is how long the final notice waits for a still-pending check
// before giving up, so an unusually fast benchmark doesn't block on the network.
const updateNoticeGrace = 1500 * time.Millisecond

// startUpdateCheck launches a non-blocking check for a newer release and returns
// a channel that yields the result (or nil on any error). It is skipped for
// non-release builds (e.g. "dev"), which are never valid semver, so local builds
// are not nagged on every run.
func startUpdateCheck() <-chan *updater.CheckResult {
	ch := make(chan *updater.CheckResult, 1)
	if !semver.IsValid(buildinfo.Version) {
		ch <- nil
		return ch
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
		defer cancel()
		res, err := updater.Check(ctx, buildinfo.Version)
		if err != nil {
			ch <- nil
			return
		}
		ch <- res
	}()
	return ch
}

// autoUpdate acts on the background update check. When a newer release is found
// it prints a notice and updates in place automatically. In a non-interactive
// context (piped/CI) it does not self-modify, printing a passive hint instead so
// scripted runs stay reproducible. It waits at most updateNoticeGrace for a
// still-pending check; a pending or failed check does nothing.
func autoUpdate(ch <-chan *updater.CheckResult) {
	var res *updater.CheckResult
	select {
	case res = <-ch:
	case <-time.After(updateNoticeGrace):
		return // check still running; don't block this run
	}
	if res == nil || !res.HasUpdate {
		return
	}

	m := i18n.L()
	// Non-interactive (piped/CI): just hint, never self-modify unprompted.
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf(m.UpdateAvailable, res.Current, res.Latest, res.URL)
		return
	}

	fmt.Printf(m.UpdateAutoNotice, res.Current, res.Latest)
	ctx, cancel := context.WithTimeout(context.Background(), updater.DefaultTimeout)
	defer cancel()
	latest, updated, err := updater.Update(ctx, res.Current)
	if err != nil {
		fmt.Printf("%s %v\n", m.UpdateFailed, err)
		return
	}
	if updated {
		fmt.Printf(m.UpdateDone, latest)
	}
}

func main() {
	// Resolve the language before building commands so that help text honors
	// --lang. Cobra renders help without running PreRun hooks, so the flag is
	// scanned manually here from the raw arguments.
	i18n.Set(i18n.Detect(langFromArgs(os.Args[1:])))
	setup()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// langFromArgs extracts the value of --lang from raw CLI arguments, supporting
// both "--lang zh" and "--lang=zh" forms. Returns "" when absent.
func langFromArgs(args []string) string {
	for i, a := range args {
		switch {
		case a == "--lang":
			if i+1 < len(args) {
				return args[i+1]
			}
		case strings.HasPrefix(a, "--lang="):
			return strings.TrimPrefix(a, "--lang=")
		}
	}
	return ""
}
