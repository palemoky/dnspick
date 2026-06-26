<div align="center">

# DNS Pick

[![Go Cross-Platform Build](https://github.com/palemoky/dnspick/actions/workflows/ci.yml/badge.svg)](https://github.com/palemoky/dnspick/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/palemoky/dnspick)](https://goreportcard.com/report/github.com/palemoky/dnspick)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Pick the DNS that fits you

**English** | [简体中文](README.zh-CN.md)

</div>

`dnspick` (**DNS** + **pick**) is a cross-platform command-line tool. It concurrently benchmarks a set of popular and custom DNS servers (over UDP, DoT, DoH and DoH3) by repeatedly querying a list of common domains, then scores them intelligently from **average latency** and **success rate**. It also folds in the default DNS you are currently using for comparison and gives a recommendation at the end.

---

## Features

*   **Cross-platform**: Runs on Windows, macOS, Linux, Raspberry Pi (ARM/ARM64) and other mainstream platforms.
*   **Multi-protocol**: Tests plain UDP DNS, DNS-over-TLS (DoT), DNS-over-HTTPS (DoH, RFC 8484 wire-format) and DNS-over-HTTP/3 (DoH3, over QUIC) side by side.
*   **More accurate measurement**: Reuses one connection per server and bounds concurrency to avoid requests contending with each other and distorting latency; DoT/DoH are warmed up before timing.
*   **Composite score**: More than just a speed test! Combines **average latency** and **success rate** into a single score (see [Scoring](#-how-is-the-composite-score-computed)) and recommends the Top 3.
*   **Detects your current DNS**: Automatically probes the system default DNS you are using (ISP/router), includes it in the comparison and gives an optimization suggestion.
*   **Live categorized progress**: Domestic/foreign domains are laid out side by side, with per-domain progress updated in real time.
*   **Bilingual UI**: Follows the system `LANG` by default, or switch manually with `--lang en|zh`.
*   **Highly customizable**: Custom test domains, query count, timeout, concurrency and more.

---

## Demo

**While testing:**

```
dnspick: benchmarking 33 DNS servers against 20 domains...

Progress: 45% (594/1320)
┌──────────────┬────────┬────────────────┬────────┐
│   Domestic   │ Status │    Foreign     │ Status │
├──────────────┼────────┼────────────────┼────────┤
│ baidu.com    │ ✔      │ google.com     │ ✔      │
│ qq.com       │ ✔      │ youtube.com    │ 60%    │
│ taobao.com   │ 60%    │ github.com     │ -      │
│ jd.com       │ -      │ facebook.com   │ -      │
│ ...          │        │ ...            │        │
└──────────────┴────────┴────────────────┴────────┘
```

**When finished:**

```
--- Benchmark Results ---
┌────┬───────────────────────┬─────────────────┬─────────────┬────────────────┬────────┐
│ #  │      DNS Server       │     Address     │ Avg Latency │  Success Rate  │ Score  │
├────┼───────────────────────┼─────────────────┼─────────────┼────────────────┼────────┤
│ 1  │ Freenom 2 (UDP)       │ 80.80.81.81     │ 6.16ms      │ 100.0% (40/40) │ 162.33 │
│ 2  │ Cloudflare 2 (UDP)    │ 1.0.0.1         │ 6.227ms     │ 100.0% (40/40) │ 160.58 │
│ 3  │ DNSPod 1 (UDP)        │ 119.28.28.28    │ 6.337ms     │ 100.0% (40/40) │ 157.80 │
│ 6  │ AliDNS 1 (UDP)        │ 223.5.5.5       │ 6.517ms     │ 100.0% (40/40) │ 153.44 │
│ 7  │ Current default DNS (current) │ 192.168.50.2 │ 6.518ms │ 100.0% (40/40) │ 153.41 │
│ .. │ ...                   │ ...             │ ...         │ ...            │ ...    │
└────┴───────────────────────┴─────────────────┴─────────────┴────────────────┴────────┘

--- Top 3 Recommendations ---
#1: Freenom 2 (UDP) (80.80.81.81)
    Score: 162.33, avg latency: 6.16ms, success rate: 100.0%
#2: Cloudflare 2 (UDP) (1.0.0.1)
    Score: 160.58, avg latency: 6.227ms, success rate: 100.0%
#3: DNSPod 1 (UDP) (119.28.28.28)
    Score: 157.80, avg latency: 6.337ms, success rate: 100.0%

✅ Current default DNS (192.168.50.2) is good enough (ranked #7, only 358µs slower); no change needed.
```

---

## Installation

### One-line install (recommended)

The script auto-detects your OS and CPU architecture, downloads the matching build and installs it onto your `PATH`.

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/palemoky/dnspick/main/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/palemoky/dnspick/main/install.ps1 | iex
```

> 💡 macOS users don't need to deal with Gatekeeper manually — the script clears the quarantine flag automatically.

### Manual install

You can also download a prebuilt binary for your OS directly from the [GitHub Releases](https://github.com/palemoky/dnspick/releases) page.

1.  Go to the latest Release page.
2.  Download the archive matching your OS and CPU architecture (e.g. `dnspick-windows-amd64.zip`).
3.  Extract it and use it directly in your terminal.

For convenience, move the extracted executable into a directory on your system `PATH` (e.g. `/usr/local/bin` or `C:\Windows\System32`).

### Updating

Update in place to the latest release at any time:

```bash
dnspick update
```

dnspick also checks for a newer release in the background while it runs. On an interactive terminal it prints a notice and updates itself in place automatically when one is found; when the output is piped or in CI it just prints a one-line hint instead (so scripted runs are never self-modified). The check is non-blocking, only happens on release builds, and is silently skipped when offline.

### macOS says "damaged / move to Trash"

Because this tool is not notarized through Apple's paid program, macOS Gatekeeper blocks programs downloaded from the internet and shows messages like "cannot verify the developer" or "damaged and should be moved to the Trash". This is expected and the **file is not actually damaged**. Use any one of the following:

```bash
# Option 1: clear the download quarantine flag (recommended, one line)
xattr -dr com.apple.quarantine ./dnspick
```

- **Option 2 (GUI):** In Finder, right-click the file → choose "Open" → confirm "Open" again in the dialog.
- **Option 3 (newer macOS):** Double-click to run once (it will be blocked), then go to **System Settings → Privacy & Security** and click "Open Anyway" at the bottom.

After that it runs normally, with no further prompts.

---

## Usage

Just run it to start testing:
```bash
dnspick
```

**Custom parameters:**

You can customize the test behavior via command-line flags.

```bash
# See all available flags
dnspick --help

# Example: query each domain 5 times against a custom domain list
dnspick -q 5 -d "google.com,github.com,youtube.com"

# Example: set the per-query timeout to 1s and reduce concurrency to 8
dnspick -t 1s -c 8

# Example: force the Chinese UI
dnspick --lang zh

# Example: emit machine-readable JSON (e.g. pipe into jq)
dnspick --json | jq '.recommendation.top'
```

| Flag              | Short | Default       | Description                                                                 |
| ----------------- | ----- | ------------- | --------------------------------------------------------------------------- |
| `--domains`       | `-d`  | built-in list | Comma-separated custom domains to test (deduplicated); falls back to the built-in domestic/foreign list |
| `--servers`       | `-s`  | built-in list | Comma-separated custom DNS servers to test; protocol is inferred from the scheme (`1.1.1.1` → UDP, `tls://host` → DoT, `https://host/dns-query` → DoH, `h3://host/dns-query` → DoH3) |
| `--queries`       | `-q`  | `3`           | Number of queries per domain                                                |
| `--timeout`       | `-t`  | `2s`          | Timeout per query                                                           |
| `--concurrency`   | `-c`  | `16`          | Maximum number of servers tested concurrently                               |
| `--no-system-dns` |       | `false`       | Do not detect or test the current system default DNS                        |
| `--lang`          |       | `$LANG`       | UI language: `en` or `zh` (defaults to the system `LANG` environment variable) |
| `--json`          |       | `false`       | Output machine-readable JSON to stdout (suppresses the progress UI)         |

---

## JSON output (for automation)

Pass `--json` to get a single JSON document on **stdout**, suitable for scripts, crawlers and CI. Status messages are written to **stderr** and the live progress UI is suppressed, so stdout stays a clean, pipe-friendly JSON stream (`dnspick --json | jq ...`).

```jsonc
{
  "schema": 2,                  // output schema version; bumped on breaking changes
  "queries_per_domain": 3,
  "servers_tested": 28,
  "domains_tested": 20,
  "results": [                  // every server, sorted by score (best first)
    {
      "rank": 1,                // 1-based rank in this list
      "name": "Quad9 (UDP)",
      "address": "9.9.9.9",
      "protocol": "udp",        // udp | dot | doh | doh3
      "is_system": false,       // true for your detected system default DNS
      "avg_latency_ms": 4.235,  // average latency in milliseconds
      "success_rate": 1.0,      // 0.0–1.0
      "successes": 60,
      "total": 60,
      "score": 236.10           // composite score (see below)
    }
  ],
  "recommendation": {
    "top": [                    // up to 3 reliable picks, with their overall rank
      { "rank": 1, "name": "Quad9 (UDP)", "address": "9.9.9.9", "protocol": "udp" }
    ],
    "system_dns": {             // omitted when --no-system-dns or none detected
      "name": "Current default DNS",
      "address": "192.168.50.2",
      "rank": 5,
      "verdict": "good_enough", // best | good_enough | switch | all_failed
      "should_switch": false,   // actionable boolean derived from verdict
      "is_internal_dns": true   // private (RFC 1918/4193) or loopback resolver
    }
  }
}
```

**Field notes:**

| Field | Description |
| --- | --- |
| `schema` | Version of the JSON structure. Guard on it in your consumers; it is bumped on any backward-incompatible change. |
| `results[]` | All tested servers, sorted by `score` descending. |
| `avg_latency_ms` | Average successful-query latency in milliseconds (microsecond precision). |
| `success_rate` | Fraction of successful queries, `0.0`–`1.0`. |
| `recommendation.top[]` | Up to 3 servers with a success rate above 98%, in ranked order. Empty when none qualify. |
| `recommendation.system_dns.verdict` | Stable enum: `best` (already optimal), `good_enough` (keep it), `switch` (a clearly better server exists), `all_failed` (every query failed). |
| `recommendation.system_dns.is_internal_dns` | `true` when the system DNS is a private (RFC 1918/4193) or loopback resolver; switching to an external DNS may break internal hostname resolution. |
| `protocol` | Transport for the server: `udp`, `dot` (DNS-over-TLS), `doh` (DNS-over-HTTPS) or `doh3` (DNS-over-HTTP/3). DoT addresses are shown as `tls://host` in the text report; DoH3 keeps its real `https://` endpoint (HTTP/3 is negotiated underneath) and is identified by the `(DoH3)` server name. |
| `recommendation.system_dns.should_switch` | Convenience boolean: `true` when `verdict` is `switch` or `all_failed`. |

---

## How is the composite score computed?

The "Score" column in the results is a **relative score — higher is better** — used to weigh "fast" and "stable" together in a single table. The formula is:

```
Score = (1 / avg latency in seconds) × success rate²
```

It has two parts:

- **Latency term `1 / avg latency`**: The lower the latency, the higher the score, scaled inversely. For example, an average latency of `5ms` (0.005s) scores `200`, `10ms` scores `100`, and `20ms` scores `50` — halve the latency and the score doubles.
- **Reliability weight `success rate²`**: The success rate is penalized by **squaring**, making it very sensitive to packet loss/timeouts. A `100%` success rate multiplies by `1.0`, `90%` leaves only `0.81`, and `50%` is cut all the way down to `0.25`. This prevents a "very low latency but occasionally failing" server from ranking near the top.

A few examples:

| Avg Latency | Success Rate | Score   | Notes                                   |
|---|---|---|---|
| 5ms | 100% | `200.0` | Fast and stable — the best              |
| 10ms | 100% | `100.0` | Stable but twice as slow                |
| 5ms | 90% | `162.0` | Very fast, but docked for occasional failures |
| 5ms | 50% | `50.0`  | Fast but highly unstable — score plunges |

> Note: The score is only meaningful for **ranking within a single run**; absolute values from different networks or different times are not comparable. At equal latency, DoT/DoH are usually slightly lower than UDP because of the encrypted handshake — this is expected.

Finally, the tool draws a conclusion based on the ranking: if your current default DNS is already the best, or only a few milliseconds slower than #1 (gap < 15ms), it reports "no change needed"; it only recommends switching when the current DNS is clearly slower or clearly less reliable.
