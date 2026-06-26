// Package i18n provides lightweight, language-switchable user-facing strings.
// The active language defaults to English and can be overridden via the
// --lang flag or the LC_ALL/LC_MESSAGES/LANG environment variables.
package i18n

import (
	"os"
	"strings"
	"sync/atomic"
)

// Lang identifies a supported UI language.
type Lang string

const (
	EN Lang = "en"
	ZH Lang = "zh"
)

// Messages holds every user-facing string for one language. Fields documented
// with %-verbs are fmt templates applied at the call site.
type Messages struct {
	// Commands.
	CmdRootShort    string
	CmdRootLong     string
	CmdVersionShort string
	CmdUpdateShort  string

	// Flag usages.
	FlagDomains     string
	FlagServers     string
	FlagQueries     string
	FlagTimeout     string
	FlagConcurrency string
	FlagNoSystemDNS string
	FlagLang        string
	FlagJSON        string

	// update command.
	UpdateAvailable  string // non-interactive startup hint: (current, latest, releaseURL)
	UpdateAutoNotice string // interactive startup notice before auto-updating: (current, latest)
	UpdateChecking   string // "Current version: %s, checking for updates...\n"
	UpdateFailed     string // printed before the error detail
	UpdateUpToDate   string // "Already on the latest version (%s).\n"
	UpdateDone       string // "✓ Updated to %s.\n"

	// benchmark run.
	ErrNoDomains    string
	ErrNoServers    string
	BenchStarting   string // "...%d DNS servers ... %d domains...\n\n"
	ResultsHeader   string
	RecommendHeader string

	// results table & recommendations.
	TableCol      []string // #, DNS Server, Address, Avg Latency, Success Rate, Score
	SystemSuffix  string   // appended to the system DNS row, e.g. " (current)"
	RecommendLine string   // "    Score: %.2f, avg latency: %s, success rate: %.1f%%\n"
	NoGoodDNS     string

	// system DNS verdict.
	VerdictAllFailed  string // (sysAddr, bestName, bestAddr)
	VerdictBest       string // (sysAddr)
	VerdictGoodEnough string // (sysAddr, rank, latencyGap)
	VerdictSwitch     string // (sysAddr, rank, bestName, bestAddr, sysAvg, bestAvg)
	PrivateDNSNote    string // (sysAddr) — appended when system DNS is an RFC-1918 address

	// progress.
	ProgressPercent string // "  Progress: %d%%\n"
	ProgressLine    string // "Progress: %d%% (%d/%d)"
	StatusCol       string

	// domain categories & system DNS naming.
	CatDomestic    string
	CatForeign     string
	CatCustom      string
	SystemDNSName  string // single system DNS
	SystemDNSNameN string // "Current default DNS %d"
}

var en = &Messages{
	CmdRootShort:    "Pick the best DNS server for your network",
	CmdRootLong:     "dnspick benchmarks a set of DNS servers by concurrently querying a list of common domains, then recommends the fastest and most reliable one for your network.",
	CmdVersionShort: "Print version information",
	CmdUpdateShort:  "Check for and update to the latest version",

	FlagDomains:     "Comma-separated custom domains to test (defaults to the built-in domestic/foreign list)",
	FlagServers:     "Comma-separated custom DNS servers to test, e.g. \"1.1.1.1, tls://dns.google, https://dns.google/dns-query, h3://cloudflare-dns.com/dns-query\" (defaults to the built-in list)",
	FlagQueries:     "Number of queries per domain",
	FlagTimeout:     "Timeout per query",
	FlagConcurrency: "Maximum number of servers tested concurrently",
	FlagNoSystemDNS: "Do not detect or test the current system default DNS",
	FlagLang:        "UI language: en or zh (defaults to $LANG)",
	FlagJSON:        "Output machine-readable JSON to stdout (suppresses the progress UI)",

	UpdateAvailable:  "\n📦 A new version is available: %s → %s\n   Run 'dnspick update' to upgrade. Release notes: %s\n",
	UpdateAutoNotice: "\n📦 A new version is available: %s → %s, updating automatically...\n",
	UpdateChecking:   "Current version: %s, checking for updates...\n",
	UpdateFailed:     "update failed:",
	UpdateUpToDate:   "Already on the latest version (%s).\n",
	UpdateDone:       "✓ Updated to %s.\n",

	ErrNoDomains:    "error: no valid domains to test.",
	ErrNoServers:    "error: no valid servers to test.",
	BenchStarting:   "dnspick: benchmarking %d DNS servers against %d domains...\n\n",
	ResultsHeader:   "\n--- Benchmark Results ---",
	RecommendHeader: "\n--- Top 3 Recommendations ---",

	TableCol:      []string{"#", "DNS Server", "Address", "Avg Latency", "Success Rate", "Score"},
	SystemSuffix:  " (current)",
	RecommendLine: "    Score: %.2f, avg latency: %s, success rate: %.1f%%\n",
	NoGoodDNS:     "No sufficiently reliable DNS server found; please check your network connection.",

	VerdictAllFailed:  "⚠️  Current default DNS (%s) failed every query; consider switching to #1 %s (%s).",
	VerdictBest:       "✅ Current default DNS (%s) is already the best; no change needed.",
	VerdictGoodEnough: "✅ Current default DNS (%s) is good enough (ranked #%d, only %s slower); no change needed.",
	VerdictSwitch:     "⚠️  Current default DNS (%s) ranked #%d; consider switching to #1 %s (%s): avg latency %s → %s.",
	PrivateDNSNote:    "\n⚠️  Note: %s appears to be an internal/corporate DNS server. Switching to an external DNS may break resolution of private hostnames.",

	ProgressPercent: "  Progress: %d%%\n",
	ProgressLine:    "Progress: %d%% (%d/%d)",
	StatusCol:       "Status",

	CatDomestic:    "Domestic",
	CatForeign:     "Foreign",
	CatCustom:      "Custom",
	SystemDNSName:  "Current default DNS",
	SystemDNSNameN: "Current default DNS %d",
}

var zh = &Messages{
	CmdRootShort:    "为你的网络选出最优 DNS 服务器",
	CmdRootLong:     "dnspick 通过对一组常用域名进行并发测试，为你的网络环境推荐最快、最稳定的 DNS 服务器。",
	CmdVersionShort: "显示版本信息",
	CmdUpdateShort:  "检查并更新到最新版本",

	FlagDomains:     "自定义测试域名列表，以逗号分隔（默认使用内置国内/国外域名）",
	FlagServers:     "自定义 DNS 服务器列表，以逗号分隔，例如 \"1.1.1.1, tls://dns.google, https://dns.google/dns-query, h3://cloudflare-dns.com/dns-query\"（默认使用内置列表）",
	FlagQueries:     "每个域名的查询次数",
	FlagTimeout:     "单次查询超时时间",
	FlagConcurrency: "同时测试的服务器数量上限",
	FlagNoSystemDNS: "不检测、不测试当前系统默认 DNS",
	FlagLang:        "界面语言：en 或 zh（默认跟随 $LANG）",
	FlagJSON:        "以机器可读的 JSON 输出到 stdout（不显示进度界面）",

	UpdateAvailable:  "\n📦 发现新版本：%s → %s\n   运行 'dnspick update' 升级。版本说明：%s\n",
	UpdateAutoNotice: "\n📦 发现新版本：%s → %s，正在自动更新...\n",
	UpdateChecking:   "当前版本: %s，正在检查更新...\n",
	UpdateFailed:     "更新失败:",
	UpdateUpToDate:   "已是最新版本 (%s)。\n",
	UpdateDone:       "✓ 已更新到 %s。\n",

	ErrNoDomains:    "错误: 没有有效的测试域名。",
	ErrNoServers:    "错误: 没有有效的测试服务器。",
	BenchStarting:   "DNS 选优工具: 开始对 %d 个 DNS 服务器、%d 个域名进行综合基准测试...\n\n",
	ResultsHeader:   "\n--- 综合测试结果 ---",
	RecommendHeader: "\n--- 最佳 DNS 推荐 (Top 3) ---",

	TableCol:      []string{"#", "DNS 服务器", "地址", "平均延迟", "成功率", "综合评分"},
	SystemSuffix:  " (当前)",
	RecommendLine: "    综合评分: %.2f, 平均延迟: %s, 成功率: %.1f%%\n",
	NoGoodDNS:     "没有找到表现足够好的 DNS 服务器，请检查网络连接。",

	VerdictAllFailed:  "⚠️  当前默认 DNS (%s) 查询全部失败，建议切换到 #1 %s (%s)。",
	VerdictBest:       "✅ 当前默认 DNS (%s) 已是最优，无需调整。",
	VerdictGoodEnough: "✅ 当前默认 DNS (%s) 已足够好（排名第 %d，仅慢 %s），无需调整。",
	VerdictSwitch:     "⚠️  当前默认 DNS (%s) 排名第 %d，建议切换到 #1 %s (%s)：平均延迟 %s → %s。",
	PrivateDNSNote:    "\n⚠️  注意：%s 看起来是内网或企业 DNS，切换为外部 DNS 可能导致内部域名无法解析。",

	ProgressPercent: "  测试进度: %d%%\n",
	ProgressLine:    "测试进度: %d%% (%d/%d)",
	StatusCol:       "状态",

	CatDomestic:    "国内",
	CatForeign:     "国外",
	CatCustom:      "自定义",
	SystemDNSName:  "当前默认 DNS",
	SystemDNSNameN: "当前默认 DNS %d",
}

// active is the currently selected catalog. Defaults to English so that code
// paths and tests that never call Set behave deterministically.
var active atomic.Pointer[Messages]

func init() { active.Store(en) }

// L returns the active language's message catalog.
func L() *Messages { return active.Load() }

// Set switches the active language.
func Set(l Lang) {
	if l == ZH {
		active.Store(zh)
		return
	}
	active.Store(en)
}

// Detect resolves a language from an explicit value (e.g. the --lang flag),
// falling back to the LC_ALL/LC_MESSAGES/LANG environment variables. Anything
// that is not recognizably Chinese resolves to English.
func Detect(explicit string) Lang {
	v := explicit
	if v == "" {
		v = firstNonEmpty(os.Getenv("LC_ALL"), os.Getenv("LC_MESSAGES"), os.Getenv("LANG"))
	}
	if strings.HasPrefix(strings.ToLower(v), "zh") {
		return ZH
	}
	return EN
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
