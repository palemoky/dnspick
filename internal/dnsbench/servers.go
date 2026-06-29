// Package dnsbench provides a concurrent benchmarking engine for DNS servers:
// concurrent queries, connection reuse, result aggregation and scoring. It
// contains no command-line or terminal presentation logic.
package dnsbench

// Protocol identifies the DNS transport used by a server.
type Protocol string

// Supported protocols.
const (
	UDP Protocol = "udp"
	DOT Protocol = "dot"
	DOH Protocol = "doh"
)

// Server describes a DNS server to be tested.
type Server struct {
	Name     string
	Address  string
	Protocol Protocol
	IsSystem bool // whether this is the detected system default DNS
}

// Domain categories. These are stable internal keys; use CategoryLabel for
// localized display text.
const (
	CategoryDomestic = "domestic"
	CategoryForeign  = "foreign"
	CategoryCustom   = "custom"
)



// Domain is a test domain with its category.
type Domain struct {
	Name, Category string
}

// DefaultServers is the built-in list of default DNS servers.
var DefaultServers = []Server{
	// {Name: "AliDNS 1 (UDP)", Address: "223.5.5.5", Protocol: UDP},
	// {Name: "AliDNS 2 (UDP)", Address: "223.6.6.6", Protocol: UDP},
	// {Name: "BaiduDNS (UDP)", Address: "180.76.76.76", Protocol: UDP},
	// {Name: "DNSPod 1 (UDP)", Address: "119.28.28.28", Protocol: UDP},
	// {Name: "DNSPod 2 (UDP)", Address: "119.29.29.29", Protocol: UDP},
	// {Name: "114DNS 1 (UDP)", Address: "114.114.114.114", Protocol: UDP},
	// {Name: "114DNS 2 (UDP)", Address: "114.114.115.115", Protocol: UDP},
	// {Name: "Bytedance 1 (UDP)", Address: "180.184.1.1", Protocol: UDP},
	// {Name: "Bytedance 2 (UDP)", Address: "180.184.2.2", Protocol: UDP},
	// {Name: "OneDNS 1 (UDP)", Address: "117.50.10.10", Protocol: UDP},
	// {Name: "OneDNS 2 (UDP)", Address: "52.80.52.52", Protocol: UDP},
	{Name: "Google 1 (UDP)", Address: "8.8.8.8", Protocol: UDP},
	// {Name: "Google 2 (UDP)", Address: "8.8.4.4", Protocol: UDP},
	{Name: "Cloudflare 1 (UDP)", Address: "1.1.1.1", Protocol: UDP},
	// {Name: "Cloudflare 2 (UDP)", Address: "1.0.0.1", Protocol: UDP},
	{Name: "OpenDNS 1 (UDP)", Address: "208.67.222.222", Protocol: UDP},
	// {Name: "OpenDNS 2 (UDP)", Address: "208.67.220.220", Protocol: UDP},
	// {Name: "Quad9 (UDP)", Address: "9.9.9.9", Protocol: UDP},

	// {Name: "AliDNS (DoT)", Address: "dns.alidns.com", Protocol: DOT},
	// {Name: "DNSPod (DoT)", Address: "dot.pub", Protocol: DOT},
	// {Name: "Google (DoT)", Address: "dns.google", Protocol: DOT},
	// {Name: "Cloudflare (DoT)", Address: "one.one.one.one", Protocol: DOT},
	// {Name: "Quad9 (DoT)", Address: "dns.quad9.net", Protocol: DOT},

	// All DoH servers use the RFC 8484 standard /dns-query endpoint (wire-format, application/dns-message).
	// {Name: "AliDNS (DoH)", Address: "https://dns.alidns.com/dns-query", Protocol: DOH},
	// {Name: "DNSPod (DoH)", Address: "https://doh.pub/dns-query", Protocol: DOH},
	// {Name: "Cloudflare (DoH)", Address: "https://cloudflare-dns.com/dns-query", Protocol: DOH},
	// {Name: "Google (DoH)", Address: "https://dns.google/dns-query", Protocol: DOH},
	// {Name: "Quad9 (DoH)", Address: "https://dns.quad9.net/dns-query", Protocol: DOH},

	{Name: "Tokyo Japan", Address: "210.224.86.126", Protocol: UDP},
	{Name: "Hong Kong 1", Address: "203.80.96.10", Protocol: UDP},
	{Name: "Hong Kong 2", Address: "203.80.96.9", Protocol: UDP},
	{Name: "Hong Kong 3", Address: "202.67.240.222", Protocol: UDP},
	{Name: "Hong Kong 4", Address: "202.67.240.221", Protocol: UDP},
}

// DefaultDomains is the built-in list of test domains (a balanced selection per category, deduplicated across same-company domains).
var DefaultDomains = []Domain{
	// {Name: "baidu.com", Category: CategoryDomestic},
	// {Name: "qq.com", Category: CategoryDomestic},
	// {Name: "taobao.com", Category: CategoryDomestic},
	// {Name: "jd.com", Category: CategoryDomestic},
	{Name: "bilibili.com", Category: CategoryDomestic},
	// {Name: "douyin.com", Category: CategoryDomestic},
	// {Name: "weibo.com", Category: CategoryDomestic},
	// {Name: "163.com", Category: CategoryDomestic},
	// {Name: "zhihu.com", Category: CategoryDomestic},
	// {Name: "aliyun.com", Category: CategoryDomestic},

	// {Name: "google.com", Category: CategoryForeign},
	// {Name: "youtube.com", Category: CategoryForeign},
	{Name: "github.com", Category: CategoryForeign},
	// {Name: "facebook.com", Category: CategoryForeign},
	// {Name: "x.com", Category: CategoryForeign},
	// {Name: "apple.com", Category: CategoryForeign},
	// {Name: "chatgpt.com", Category: CategoryForeign},
	// {Name: "bing.com", Category: CategoryForeign},
	// {Name: "tiktok.com", Category: CategoryForeign},
	// {Name: "cloudflare.com", Category: CategoryForeign},
}
