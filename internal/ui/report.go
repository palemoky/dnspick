package ui

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/palemoky/dnspick/internal/dnsbench"
	"github.com/palemoky/dnspick/internal/i18n"
)

// CategoryLabel returns the localized display label for a domain category key.
func CategoryLabel(category string) string {
	switch category {
	case dnsbench.CategoryDomestic:
		return i18n.L().CatDomestic
	case dnsbench.CategoryForeign:
		return i18n.L().CatForeign
	case dnsbench.CategoryCustom:
		return i18n.L().CatCustom
	default:
		return category
	}
}

// PrintResultsTable prints a formatted result table using tablewriter.
func PrintResultsTable(results []dnsbench.Result) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header(i18n.L().TableCol)

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for i, r := range results {
		rateStr := fmt.Sprintf("%.1f%% (%d/%d)", r.SuccessRate*100, r.Successes, r.Total)
		if r.SuccessRate < 1.0 {
			rateStr = red(rateStr)
		} else {
			rateStr = green(rateStr)
		}

		name := r.Name
		if r.IsSystem {
			name += i18n.L().SystemSuffix
		}
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			name,
			displayAddress(r),
			r.AvgTime.Round(time.Microsecond).String(),
			rateStr,
			fmt.Sprintf("%.2f", r.Score),
		})
	}
	table.Render()
}

// recommendThreshold is the minimum success rate for a server to be recommended.
// maxRecommendations caps how many are surfaced (the "Top N").
const (
	recommendThreshold = 0.98
	maxRecommendations = 3
)

// Recommendation is a recommended server together with its overall rank (1-based)
// in the full, score-descending result list.
type Recommendation struct {
	Rank int
	dnsbench.Result
}

// topRecommendations selects the reliably-performing servers worth recommending,
// preserving their overall ranking, capped at maxRecommendations.
func topRecommendations(results []dnsbench.Result) []Recommendation {
	var top []Recommendation
	for i, r := range results {
		if r.SuccessRate < recommendThreshold {
			continue
		}
		top = append(top, Recommendation{Rank: i + 1, Result: r})
		if len(top) >= maxRecommendations {
			break
		}
	}
	return top
}

// PrintRecommendations prints the top recommendations.
func PrintRecommendations(results []dnsbench.Result) {
	palette := []*color.Color{
		color.New(color.FgGreen, color.Bold),
		color.New(color.FgYellow),
		color.New(color.FgCyan),
	}

	top := topRecommendations(results)
	for i, best := range top {
		palette[i].Printf("#%d: %s (%s)\n", i+1, best.Name, displayAddress(best.Result))
		fmt.Printf(i18n.L().RecommendLine,
			best.Score, best.AvgTime.Round(time.Microsecond).String(), best.SuccessRate*100)
	}
	if len(top) == 0 {
		color.New(color.FgRed).Println(i18n.L().NoGoodDNS)
	}

	if msg, ok := systemDNSVerdict(results); ok {
		c := color.New(color.FgGreen, color.Bold)
		if strings.HasPrefix(msg, "⚠") {
			c = color.New(color.FgYellow, color.Bold)
		}
		fmt.Println()
		c.Println(msg)
	}
}

const (
	// switchLatencyThreshold: if the system DNS is slower than the best by less
	// than this, the gap is treated as insignificant and no switch is suggested.
	switchLatencyThreshold = 15 * time.Millisecond
	// switchSuccessMargin: if the system DNS success rate trails the best by more
	// than this, a switch is suggested even when latency is close.
	switchSuccessMargin = 0.05
)

// VerdictKind classifies the system DNS conclusion in a stable, machine-readable
// way, independent of the localized message. It is part of the --json contract.
type VerdictKind string

const (
	VerdictAllFailed  VerdictKind = "all_failed"  // system DNS failed every query
	VerdictBest       VerdictKind = "best"        // system DNS is already the top server
	VerdictGoodEnough VerdictKind = "good_enough" // not the best, but close enough to keep
	VerdictSwitch     VerdictKind = "switch"      // a clearly better server exists
)

// systemEval is the structured outcome of comparing the system DNS against the
// best server. ok (from evalSystemDNS) is false when there is no system DNS.
type systemEval struct {
	kind       VerdictKind
	sys, best  dnsbench.Result
	rank       int // 1-based rank of the system DNS
	latencyGap time.Duration
}

// evalSystemDNS locates the system DNS in the (score-descending) results and
// classifies whether it should be changed. A switch is only suggested when the
// system DNS is clearly slower (latency gap >= switchLatencyThreshold) or clearly
// less reliable, avoiding misleading the user over a few insignificant
// milliseconds. ok is false when the results contain no system DNS.
func evalSystemDNS(results []dnsbench.Result) (systemEval, bool) {
	if len(results) == 0 {
		return systemEval{}, false
	}
	sysRank := -1
	for i := range results {
		if results[i].IsSystem {
			sysRank = i
			break
		}
	}
	if sysRank < 0 {
		return systemEval{}, false
	}

	sys := results[sysRank]
	best := results[0]
	latencyGap := sys.AvgTime - best.AvgTime
	closeEnough := latencyGap < switchLatencyThreshold && best.SuccessRate-sys.SuccessRate <= switchSuccessMargin

	e := systemEval{sys: sys, best: best, rank: sysRank + 1, latencyGap: latencyGap}
	switch {
	case sys.Successes == 0:
		e.kind = VerdictAllFailed
	case sysRank == 0:
		e.kind = VerdictBest
	case closeEnough:
		e.kind = VerdictGoodEnough
	default:
		e.kind = VerdictSwitch
	}
	return e, true
}

// isInternalDNS reports whether addr is a local/internal resolver: an RFC 1918
// or RFC 4193 private address, or a loopback address. Loopback covers stub
// resolvers such as systemd-resolved's 127.0.0.53, which forward to an upstream
// (often VPN/corporate) DNS, so switching away from them can also break internal
// name resolution.
func isInternalDNS(addr string) bool {
	ip := net.ParseIP(strings.TrimSpace(addr))
	return ip != nil && (ip.IsPrivate() || ip.IsLoopback())
}

// displayAddress renders a server address for human-facing output. DoT servers
// are shown with a tls:// scheme so users can tell the protocol apart from plain
// UDP, mirroring the https:// already carried by DoH addresses. DoH3 keeps its
// real https:// endpoint (the actual address a user would configure; HTTP/3 is
// negotiated underneath) — its protocol is conveyed by the server name's "(DoH3)"
// suffix rather than a non-standard h3:// scheme that users might copy verbatim.
func displayAddress(r dnsbench.Result) string {
	if r.Protocol == dnsbench.DOT {
		return "tls://" + r.Address
	}
	return r.Address
}

// systemDNSVerdict produces a localized conclusion on whether the system default
// DNS should be changed. If the results contain no system DNS, ok is false.
// results must be sorted by score in descending order.
func systemDNSVerdict(results []dnsbench.Result) (msg string, ok bool) {
	e, ok := evalSystemDNS(results)
	if !ok {
		return "", false
	}

	m := i18n.L()
	privateNote := ""
	if isInternalDNS(e.sys.Address) {
		privateNote = fmt.Sprintf(m.PrivateDNSNote, e.sys.Address)
	}
	switch e.kind {
	case VerdictAllFailed:
		return fmt.Sprintf(m.VerdictAllFailed, displayAddress(e.sys), e.best.Name, displayAddress(e.best)) + privateNote, true
	case VerdictBest:
		return fmt.Sprintf(m.VerdictBest, displayAddress(e.sys)), true
	case VerdictGoodEnough:
		return fmt.Sprintf(m.VerdictGoodEnough, displayAddress(e.sys), e.rank, e.latencyGap.Round(time.Microsecond)), true
	default:
		return fmt.Sprintf(m.VerdictSwitch, displayAddress(e.sys), e.rank, e.best.Name, displayAddress(e.best),
			e.sys.AvgTime.Round(time.Microsecond), e.best.AvgTime.Round(time.Microsecond)) + privateNote, true
	}
}
