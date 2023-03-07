package adgproxy

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"strings"
)

var pluginName = "adgproxy"

// init registers this plugin.
func init() { plugin.Register(pluginName, setup) }

type configFromFile struct {
	Upstreams  []string
	Bootstraps []string
	Insecure   bool
}

var ConfigFromFile *configFromFile

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	ConfigFromFile = &configFromFile{
		Insecure: false,
	}
	parseConfig(c)
	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	a := &ADGProxy{
		ConfigFromFile: ConfigFromFile,
	}
	dnsserver.GetConfig(c).AddPlugin(
		func(next plugin.Handler) plugin.Handler {
			a.Next = next
			return a
		},
	)
	c.OnStartup(
		func() error {
			return a.OnStartup()
		},
	)

	// All OK, return a nil error.
	return nil
}

func parseConfig(c *caddy.Controller) {
	for c.Next() {
		log.Debug("进入下一个token")
		for c.NextBlock() {
			//println(c.Val())
			log.Debug("解析到 " + c.Val() + " 节")
			selector := strings.ToLower(c.Val())

			switch selector {
			case "upstream":
				log.Debug("解析到 upstream 节")
				upsteam := c.RemainingArgs()[0]
				ConfigFromFile.Upstreams = append(ConfigFromFile.Upstreams, upsteam)
			case "bootstrap":
				log.Debug("解析到 bootstrap 节")
				bootstrap := c.RemainingArgs()[0]
				ConfigFromFile.Bootstraps = append(ConfigFromFile.Bootstraps, bootstrap)
			case "insecure":
				log.Debug("解析到 insecure 节")
				// 没有参数的话就是 true
				if len(c.RemainingArgs()) == 0 {
					ConfigFromFile.Insecure = true
				}
				// 或者参数为一个且为 true
				if len(c.RemainingArgs()) == 1 && strings.ToLower(c.RemainingArgs()[0]) == "true" {
					ConfigFromFile.Insecure = true
				}
				// 其他情况都是 false 也不需要写了
			}
		}
	}
}
