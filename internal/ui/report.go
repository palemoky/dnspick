package ui

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

		avgTimeStr := r.AvgTime.Round(time.Microsecond).String()
		scoreStr := fmt.Sprintf("%.2f", r.Score)
		if r.Successes == 0 {
			avgTimeStr = red(i18n.L().AllFailed)
			scoreStr = red("-")
		}

		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			name,
			displayAddress(r),
			avgTimeStr,
			rateStr,
			scoreStr,
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
// UDP, mirroring the https:// already carried by DoH addresses.
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

// resolveTableLimit is the maximum number of top-ranked servers shown in the
// resolution details table, keeping the output readable for large server lists.
const resolveTableLimit = 10

// showDNSServerCol controls whether the "DNS Server" column is displayed
// in the merged resolution table. Set to true to enable.
const showDNSServerCol = false

// PrintResolutions prints the DNS resolution details and port connectivity
// test results in a merged table. The table is grouped by DNS server, then by domain,
// with each row showing an IP address (in ip:port format), its connectivity status,
// and latency. The best (lowest latency) IP per domain is highlighted in green.
// Skips output entirely when no port-enabled domains exist in the results.
func PrintResolutions(results []dnsbench.Result, ports []int) {
	// Check if any result has resolution data; skip if none.
	hasResolutions := false
	hasPortResults := false
	for _, r := range results {
		if len(r.Resolutions) > 0 {
			hasResolutions = true
		}
		if len(r.PortResults) > 0 {
			hasPortResults = true
		}
	}
	if !hasResolutions || !hasPortResults {
		return
	}

	// Determine the port column header: use the single port number when there
	// is exactly one port, otherwise fall back to a generic "Port" label.
	portHeader := "Port"
	if len(ports) == 1 {
		portHeader = strconv.Itoa(ports[0])
	}

	m := i18n.L()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// --- Merged Resolution & Port Connectivity table ---
	fmt.Println(m.ResolveHeader)

	mergedTable := tablewriter.NewWriter(os.Stdout)
	allHeaders := []string{"DNS Server", "DNS ADDRESS", "Domain", "IP Address", portHeader, "Latency"}
	colStart := 1
	if showDNSServerCol {
		colStart = 0
	}
	mergedTable.Header(allHeaders[colStart:])

	limit := resolveTableLimit
	if limit > len(results) {
		limit = len(results)
	}

	// Build per-domain IP list with port results, deduplicating (domain, IP:port) pairs.
	type domainIP struct {
		Domain string
		IP     string
	}
	domainIPs := make(map[string][]string) // domain -> ordered unique IP:port keys
	domainOrder := []string{}              // preserve first-seen domain order
	seenDI := make(map[domainIP]bool)
	for _, r := range results {
		for _, res := range r.Resolutions {
			if _, ok := domainIPs[res.Domain]; !ok {
				domainOrder = append(domainOrder, res.Domain)
			}
			for _, ip := range res.IPs {
				for _, port := range ports {
					key := net.JoinHostPort(ip, strconv.Itoa(port))
					di := domainIP{res.Domain, key}
					if !seenDI[di] {
						seenDI[di] = true
						domainIPs[res.Domain] = append(domainIPs[res.Domain], key)
					}
				}
			}
		}
	}

	// Find the fastest (lowest latency) IP:port per domain.
	bestIPPerDomain := make(map[string]string)
	for domain, ipKeys := range domainIPs {
		var bestKey string
		var bestLatency time.Duration
		for _, ipKey := range ipKeys {
			for _, r := range results {
				if pr, ok := r.PortResults[ipKey]; ok && pr.OK {
					if bestKey == "" || pr.Duration < bestLatency {
						bestKey = ipKey
						bestLatency = pr.Duration
					}
					break
				}
			}
		}
		if bestKey != "" {
			bestIPPerDomain[domain] = bestKey
		}
	}

	// Group results by DNS server, then by domain.
	type serverDomainData struct {
		ServerName string
		Address    string
		Domains    map[string][]string // domain -> IP:port keys resolved by this server
	}
	serverDataMap := make(map[string]*serverDomainData)
	serverOrder := []string{} // preserve server order

	for i := range limit {
		r := &results[i]
		name := r.Name
		if r.IsSystem {
			name += m.SystemSuffix
		}

		if _, ok := serverDataMap[name]; !ok {
			serverOrder = append(serverOrder, name)
			serverDataMap[name] = &serverDomainData{
				ServerName: name,
				Address:    r.Address,
				Domains:    make(map[string][]string),
			}
		}

		for _, res := range r.Resolutions {
			if len(res.IPs) > 0 {
				var keys []string
				for _, ip := range res.IPs {
					for _, port := range ports {
						keys = append(keys, net.JoinHostPort(ip, strconv.Itoa(port)))
					}
				}
				serverDataMap[name].Domains[res.Domain] = keys
			}
		}
	}

	// Pre-compute max content width per column for dynamic separators.
	colWidths := []int{
		utf8.RuneCountInString("DNS Server"),
		utf8.RuneCountInString("DNS ADDRESS"),
		utf8.RuneCountInString("Domain"),
		utf8.RuneCountInString("IP Address"),
		utf8.RuneCountInString(portHeader),
		utf8.RuneCountInString("Latency"),
	}
	for sName, sdata := range serverDataMap {
		if w := utf8.RuneCountInString(sName); w > colWidths[0] {
			colWidths[0] = w
		}
		if w := utf8.RuneCountInString(sdata.Address); w > colWidths[1] {
			colWidths[1] = w
		}
		for dom, ipKeys := range sdata.Domains {
			if w := utf8.RuneCountInString(dom); w > colWidths[2] {
				colWidths[2] = w
			}
			for _, ipKey := range ipKeys {
				if w := utf8.RuneCountInString(ipKey); w > colWidths[3] {
					colWidths[3] = w
				}
				for _, r := range results {
					if pr, ok := r.PortResults[ipKey]; ok {
						if pr.OK {
							if w := utf8.RuneCountInString(pr.Duration.Round(time.Millisecond).String()); w > colWidths[5] {
								colWidths[5] = w
							}
						}
						break
					}
				}
			}
		}
	}

	// Output rows grouped by server and domain.
	firstServer := true
	for _, serverName := range serverOrder {
		data := serverDataMap[serverName]
		domains := data.Domains

		// Use the original domainOrder to maintain consistency.
		var finalDomains []string
		for _, d := range domainOrder {
			if _, ok := domains[d]; ok {
				finalDomains = append(finalDomains, d)
			}
		}

		// Print a visible line before each server (except the first).
		if !firstServer {
			sep := make([]string, len(colWidths)-colStart)
			for i, w := range colWidths[colStart:] {
				sep[i] = strings.Repeat("─", w)
			}
			mergedTable.Append(sep)
		}
		firstServer = false

		// Servers with no successful resolutions: show a single error row.
		if len(finalDomains) == 0 {
			mergedTable.Append([]string{
				serverName,
				data.Address,
				"-",
				red(m.AllFailed),
				"-",
				"-",
			}[colStart:])
			continue
		}

		firstRowForServer := true
		for _, domain := range finalDomains {
			ipKeys := domains[domain]
			bestKey := bestIPPerDomain[domain]

			for idx, ipKey := range ipKeys {
				status := "-"
				latency := "-"
				isBestIP := (ipKey == bestKey)

				// Get port result for this IP:port key
				for _, r := range results {
					if pr, ok := r.PortResults[ipKey]; ok {
						if pr.OK {
							status = m.PortOK
							latency = pr.Duration.Round(time.Millisecond).String()
						} else {
							status = red(m.PortFail)
						}
						break
					}
				}

				// Prepare cell values
				serverCell := ""
				serverIPCell := ""
				if firstRowForServer {
					serverCell = serverName
					serverIPCell = data.Address
					firstRowForServer = false
				}

				domainCell := ""
				if idx == 0 {
					domainCell = domain
				}

				ipCell := ipKey
				statusCell := status
				latencyCell := latency

				// Highlight best IP row: IP, Domain, and DNS Server columns
				if isBestIP {
					ipCell = green(ipCell)
					statusCell = green(statusCell)
					latencyCell = green(latencyCell)
				}

				mergedTable.Append([]string{
					serverCell,
					serverIPCell,
					domainCell,
					ipCell,
					statusCell,
					latencyCell,
				}[colStart:])
			}
		}

	}

	mergedTable.Render()
}
