package adgproxy

import (
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/coredns/coredns/core/dnsserver"
	log2 "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/coredns/caddy"
)

// TestSetup tests the various things that should be parsed by setup.
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `adgproxy`)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}

	c = caddy.NewTestController("dns", `adgproxy more`)
	if err := setup(c); err == nil {
		t.Fatalf("Expected errors, but got: %v", err)
	}
}

func TestReadBlock(t *testing.T) {
	c := caddy.NewTestController(
		"dns",
		`
debug
adgproxy {
    upstream https://1.1.1.1/dns-query
    upstream https://1.0.0.1/dns-query
    bootstrap https://223.5.5.5/dns-query
    bootstrap https://119.29.29.29/dns-query
    mode load_balance
    insecure true
    geosite geosite.dat
    rule GEOSITE,bilibili,PASS
    rule GEOSITE,google,PROXY
    rule GEOSITE,facebook,PROXY
}
`,
	)
	a := &ADGProxy{
		ConfigFromFile: &configFromFile{
			Insecure: false,
		},
	}
	log2.D.Set()
	log.Debug("现在解析配置")
	config := dnsserver.GetConfig(c)
	config.Debug = true
	err := parseConfig(c, a)
	if err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}
	//for _, up := range a.ConfigFromFile.Upstreams {
	//	t.Log(up)
	//}
	// 要确认的是, upstream 有 2 个
	assert.Equal(t, 2, len(a.ConfigFromFile.Upstreams))
	// Bootstrap 有 2 个
	assert.Equal(t, 2, len(a.ConfigFromFile.Bootstraps))
	// 模式为负载均衡
	assert.Equal(t, proxy.UpstreamModeLoadBalance, a.ConfigFromFile.UpstreamMode)
	// 不允许不安全的 TLS
	assert.Equal(t, true, a.ConfigFromFile.Insecure)
	// 包括规则有 3 个
	assert.Equal(t, 3, len(a.ConfigFromFile.Rules))

}
func TestOutputConfigTokens(t *testing.T) {
	c := caddy.NewTestController(
		"dns",
		`
adgproxy {
    upstream https://1.1.1.1/dns-query
    upstream https://1.0.0.1/dns-query
    bootstrap https://223.5.5.5/dns-query
    bootstrap https://119.29.29.29/dns-query
    mode load_balance
    insecure true
    geosite geosite.dat
    rule GEOSITE,bilibili,PASS
    rule GEOSITE,google,PROXY
    rule GEOSITE,facebook,PROXY
}}
}
`,
	)
	// 调用 Next 输出每一个 token
	for c.Next() {
		for c.NextBlock() {
			t.Log(c.Val())
		}
	}
}
