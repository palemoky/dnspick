<div align="center">

# DNS Pick

[![Go Cross-Platform Build](https://github.com/palemoky/dnspick/actions/workflows/ci.yml/badge.svg)](https://github.com/palemoky/dnspick/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/palemoky/dnspick)](https://goreportcard.com/report/github.com/palemoky/dnspick)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

选出适合你的 DNS

**简体中文** | [English](README.md)

</div>

`dnspick`（**DNS** + **pick**）是一个跨平台命令行工具：它并发基准测试一批主流及自定义 DNS 服务器（涵盖 UDP、DoT、DoH、DoH3），对一组常用的国内/国外域名反复查询，结合**平均延迟**与**成功率**智能评分；同时把你当前正在用的默认 DNS 一并纳入对比，最后给出建议。

---

## 功能特性

*   **跨平台支持**: 完美运行于 Windows, macOS, Linux, Raspberry Pi (ARM/ARM64) 等主流平台。
*   **多协议支持**: 同时测试传统 UDP DNS、DNS-over-TLS (DoT)、DNS-over-HTTPS (DoH，RFC 8484 标准 wire-format) 和 DNS-over-HTTP/3 (DoH3，基于 QUIC)。
*   **测量更准**: 每台服务器复用连接、限制并发，避免大量请求互相争抢导致延迟失真；DoT/DoH 预热后再计时。
*   **综合评分**: 不只是测速！结合**平均延迟**与**成功率**给出综合评分（详见[评分说明](#-综合评分是怎么算的)），并推荐 Top 3。
*   **检测当前 DNS**: 自动探测你正在使用的系统默认 DNS（运营商/路由器）一并参与对比，并给出优化建议。
*   **实时分类进度**: 国内/国外域名并列成表，逐域名实时显示进度。
*   **中英双语界面**: 默认跟随系统 `LANG`，也可用 `--lang en|zh` 手动切换。
*   **高度可定制**: 可自定义测试域名、查询次数、超时、并发数等。

---

## 运行演示

**测试中**：

```
DNS 选优工具: 开始对 33 个 DNS 服务器、20 个域名进行综合基准测试...

测试进度: 45% (594/1320)
┌──────────────┬──────┬────────────────┬──────┐
│     国内      │ 状态 │       国外      │  状态 │
├──────────────┼──────┼────────────────┼──────┤
│ baidu.com    │ ✔    │ google.com     │ ✔    │
│ qq.com       │ ✔    │ youtube.com    │ 60%  │
│ taobao.com   │ 60%  │ github.com     │ -    │
│ jd.com       │ -    │ facebook.com   │ -    │
│ ...          │      │ ...            │      │
└──────────────┴──────┴────────────────┴──────┘
```

**测试完成**：

```
--- 综合测试结果 ---
┌────┬───────────────────────┬─────────────────┬───────────┬────────────────┬──────────┐
│ #  │      DNS 服务器        │      地址        │ 平均延迟   │      成功率     │  综合评分 │
├────┼───────────────────────┼─────────────────┼───────────┼────────────────┼──────────┤
│ 1  │ Freenom 2 (UDP)       │ 80.80.81.81     │ 6.16ms    │ 100.0% (40/40) │ 162.33   │
│ 2  │ Cloudflare 2 (UDP)    │ 1.0.0.1         │ 6.227ms   │ 100.0% (40/40) │ 160.58   │
│ 3  │ DNSPod 1 (UDP)        │ 119.28.28.28    │ 6.337ms   │ 100.0% (40/40) │ 157.80   │
│ 6  │ AliDNS 1 (UDP)        │ 223.5.5.5       │ 6.517ms   │ 100.0% (40/40) │ 153.44   │
│ 7  │ 当前默认 DNS (当前)     │ 192.168.50.2    │ 6.518ms   │ 100.0% (40/40) │ 153.41   │
│ .. │ ...                   │ ...             │ ...       │ ...            │ ...      │
└────┴───────────────────────┴─────────────────┴───────────┴────────────────┴──────────┘

--- 最佳 DNS 推荐 (Top 3) ---
#1: Freenom 2 (UDP) (80.80.81.81)
    综合评分: 162.33, 平均延迟: 6.16ms, 成功率: 100.0%
#2: Cloudflare 2 (UDP) (1.0.0.1)
    综合评分: 160.58, 平均延迟: 6.227ms, 成功率: 100.0%
#3: DNSPod 1 (UDP) (119.28.28.28)
    综合评分: 157.80, 平均延迟: 6.337ms, 成功率: 100.0%

✅ 当前默认 DNS (192.168.50.2) 已足够好（排名第 7，仅慢 358µs），无需调整。
```

---

## 安装

### 一键安装（推荐）

脚本会自动检测操作系统与 CPU 架构，下载对应版本并安装到 `PATH`。

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/palemoky/dnspick/main/install.sh | sh
```

**Windows（PowerShell）**

```powershell
irm https://raw.githubusercontent.com/palemoky/dnspick/main/install.ps1 | iex
```

> 💡 macOS 用户无需手动处理 Gatekeeper 拦截，脚本会自动解除隔离标记。

### 手动安装

您也可以直接从 [GitHub Releases](https://github.com/palemoky/dnspick/releases) 页面下载适用于您操作系统的预编译版本。

1.  前往最新的 Release 页面。
2.  根据您的操作系统和 CPU 架构下载对应的压缩包（例如 `dnspick-windows-amd64.zip`）。
3.  解压后即可直接在终端中使用。

为了方便使用，建议将解压后的可执行文件移动到您系统的 `PATH` 环境变量所包含的目录中（例如 `/usr/local/bin` 或 `C:\Windows\System32`）。

### 更新

随时原地升级到最新版本：

```bash
dnspick update
```

dnspick 在运行时还会在后台检查是否有新版本。在交互式终端中若发现新版本会打印提示并自动原地更新；输出被管道接管或在 CI 中则只打印一行升级提示（脚本化运行不会被自动修改）。该检查不阻塞主流程，仅在正式发布版本中触发，离线时静默跳过。

### macOS 提示「已损坏 / 移动到垃圾桶」

由于本工具未经过 Apple 付费的公证（notarization），macOS 的 Gatekeeper 会拦截从网络下载的程序，弹出「无法验证开发者」「已损坏，应移到废纸篓」等提示。这是正常现象，**并非文件损坏**。任选一种方式解除即可：

```bash
# 方式一：解除下载隔离标记（推荐，一行搞定）
xattr -dr com.apple.quarantine ./dnspick
```

- **方式二（图形界面）**：在访达中右键点击该文件 → 选择「打开」→ 在弹窗中再次确认「打开」。
- **方式三（较新版本 macOS）**：先双击运行一次（会被拦截），再到 **系统设置 → 隐私与安全性**，点击底部的「仍要打开」。

解除后即可正常使用，后续运行不再提示。

---

## 使用方法

直接运行即可开始测试：
```bash
dnspick
```

**自定义参数：**

可通过命令行参数自定义测试行为。

```bash
# 查看所有可用参数
dnspick --help

# 示例：每个域名查询5次，并使用自定义的域名列表
dnspick -q 5 -d "google.com,github.com,youtube.com"

# 示例：将单次查询超时设为 1 秒，并把并发服务器数降到 8
dnspick -t 1s -c 8

# 示例：强制使用英文界面
dnspick --lang en

# 示例：输出机器可读的 JSON（例如交给 jq 处理）
dnspick --json | jq '.recommendation.top'
```

| 参数              | 简写 | 默认值   | 描述                                                         |
| ----------------- | ---- | -------- | ------------------------------------------------------------ |
| `--domains`       | `-d` | 内置列表 | 自定义测试域名列表，以逗号分隔（自动去重）；不指定则使用内置的国内/国外域名 |
| `--servers`       | `-s` | 内置列表 | 自定义 DNS 服务器列表，以逗号分隔；协议按 scheme 自动推断（`1.1.1.1` → UDP，`tls://host` → DoT，`https://host/dns-query` → DoH，`h3://host/dns-query` → DoH3） |
| `--queries`       | `-q` | `3`      | 每个域名的查询次数                                           |
| `--timeout`       | `-t` | `2s`     | 单次查询超时时间                                             |
| `--concurrency`   | `-c` | `16`     | 同时测试的服务器数量上限                                     |
| `--no-system-dns` |      | `false`  | 不检测、不测试当前系统默认 DNS                               |
| `--lang`          |      | `$LANG`  | 界面语言：`en` 或 `zh`（默认跟随系统 `LANG` 环境变量）       |
| `--json`          |      | `false`  | 以机器可读的 JSON 输出到 stdout（不显示进度界面）            |

---

## JSON 输出（用于自动化）

加上 `--json` 即可在 **stdout** 得到单个 JSON 文档，适合脚本、爬虫和 CI 使用。状态信息写入 **stderr**，并且不显示实时进度界面，因此 stdout 是干净、可直接管道处理的 JSON 流（`dnspick --json | jq ...`）。

```jsonc
{
  "schema": 2,                  // 输出结构版本号；不兼容变更时递增
  "queries_per_domain": 3,
  "servers_tested": 28,
  "domains_tested": 20,
  "results": [                  // 所有服务器，按综合评分从高到低排序
    {
      "rank": 1,                // 在本列表中的排名（从 1 开始）
      "name": "Quad9 (UDP)",
      "address": "9.9.9.9",
      "protocol": "udp",        // udp | dot | doh | doh3
      "is_system": false,       // 为你检测到的系统默认 DNS 时为 true
      "avg_latency_ms": 4.235,  // 平均延迟（毫秒）
      "success_rate": 1.0,      // 0.0–1.0
      "successes": 60,
      "total": 60,
      "score": 236.10           // 综合评分（见下文）
    }
  ],
  "recommendation": {
    "top": [                    // 最多 3 个可靠推荐，附带其总排名
      { "rank": 1, "name": "Quad9 (UDP)", "address": "9.9.9.9", "protocol": "udp" }
    ],
    "system_dns": {             // 使用 --no-system-dns 或未检测到时省略
      "name": "Current default DNS",
      "address": "192.168.50.2",
      "rank": 5,
      "verdict": "good_enough", // best | good_enough | switch | all_failed
      "should_switch": false,   // 由 verdict 推导出的可操作布尔值
      "is_internal_dns": true   // 内网（RFC 1918/4193）或回环解析器
    }
  }
}
```

**字段说明：**

| 字段 | 描述 |
| --- | --- |
| `schema` | JSON 结构版本号。请在消费方据此做判断；任何不兼容变更都会递增该值。 |
| `results[]` | 所有被测服务器，按 `score` 降序排列。 |
| `avg_latency_ms` | 成功查询的平均延迟（毫秒，微秒精度）。 |
| `success_rate` | 查询成功率，范围 `0.0`–`1.0`。 |
| `recommendation.top[]` | 成功率超过 98% 的服务器，最多 3 个，按排名排序。无符合者时为空。 |
| `recommendation.system_dns.verdict` | 稳定枚举：`best`（已最优）、`good_enough`（无需更换）、`switch`（存在明显更优的服务器）、`all_failed`（全部查询失败）。 |
| `recommendation.system_dns.is_internal_dns` | 系统 DNS 为内网（RFC 1918/4193）或回环解析器时为 `true`；此时切换到外部 DNS 可能导致内部域名无法解析。 |
| `protocol` | 服务器的传输协议：`udp`、`dot`（DNS-over-TLS）、`doh`（DNS-over-HTTPS）或 `doh3`（DNS-over-HTTP/3）。文本报告中 DoT 地址显示为 `tls://host`；DoH3 保留真实的 `https://` 地址（HTTP/3 在底层协商），靠"DNS 服务器"列的 `(DoH3)` 区分。 |
| `recommendation.system_dns.should_switch` | 便捷布尔值：当 `verdict` 为 `switch` 或 `all_failed` 时为 `true`。 |

---

## 综合评分是怎么算的？

结果表里的「综合评分」是一个**相对分数，越高越好**，用来在一张表里同时权衡"快"和"稳"。计算公式为：

```
综合评分 = (1 / 平均延迟秒数) × 成功率²
```

它由两部分构成：

- **延迟分 `1 / 平均延迟`**：延迟越低，分数越高，且呈反比放大。例如平均延迟 `5ms`（0.005 秒）得 `200` 分，`10ms` 得 `100` 分，`20ms` 得 `50` 分——延迟减半，分数翻倍。
- **可靠性权重 `成功率²`**：成功率做**平方**惩罚，对丢包/超时非常敏感。`100%` 成功率乘以 `1.0`，`90%` 只剩 `0.81`，`50%` 直接砍到 `0.25`。这样可以避免"延迟很低但偶尔解析失败"的服务器排到前面。

举几个例子：

| 平均延迟 | 成功率 | 综合评分 | 说明 |
|---|---|---|---|
| 5ms | 100% | `200.0` | 又快又稳，最佳 |
| 10ms | 100% | `100.0` | 稳定但慢一倍 |
| 5ms | 90% | `162.0` | 很快，但偶有失败被扣分 |
| 5ms | 50% | `50.0` | 虽快但极不稳定，分数大跌 |

> 说明：评分只用于**同一次测试内部排序**，不同网络环境/不同时间跑出来的绝对分值没有可比性。延迟相同时，DoT/DoH 因含加密握手通常略低于 UDP，属正常现象。

最后工具会结合排名给出结论：若你当前的默认 DNS 已经是最优、或只比第一名慢几毫秒（差距 < 15ms），会提示「无需调整」；只有当它明显更慢或明显更不稳定时，才建议切换。
