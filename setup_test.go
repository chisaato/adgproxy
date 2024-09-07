package adgproxy

import (
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
adgproxy {
	upstream https://dns.google/dns-query
	upstream https://1.1.1.1/dns-query
	//bootstrap https://223.5.5.5/dns-query
	insecure
}
`,
	)
	a := &ADGProxy{
		ConfigFromFile: &configFromFile{
			Insecure: false,
		},
	}
	err := parseConfig(c, a)
	if err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}
	for _, up := range a.ConfigFromFile.Upstreams {
		t.Log(up)
	}
}
