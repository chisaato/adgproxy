package adgproxy

import (
	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"strings"
	"time"
)

var pluginName = "adgproxy"

// init registers this plugin.
func init() { plugin.Register(pluginName, setup) }

type configFromFile struct {
	Upstreams  []string
	Bootstraps []string
	Insecure   bool
}

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	a := &ADGProxy{
		ConfigFromFile: &configFromFile{
			Insecure: false,
		},
	}
	err := parseConfig(c, a)
	if err != nil {
		return plugin.Error(pluginName, err)
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

func parseConfig(c *caddy.Controller, a *ADGProxy) error {
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
				a.ConfigFromFile.Upstreams = append(a.ConfigFromFile.Upstreams, upsteam)
			case "bootstrap":
				log.Debug("解析到 bootstrap 节")
				bootstrap := c.RemainingArgs()[0]
				a.ConfigFromFile.Bootstraps = append(a.ConfigFromFile.Bootstraps, bootstrap)
			case "insecure":
				log.Debug("解析到 insecure 节")
				// 没有参数的话就是 true
				if len(c.RemainingArgs()) == 0 {
					a.ConfigFromFile.Insecure = true
				}
				// 或者参数为一个且为 true
				if len(c.RemainingArgs()) == 1 && strings.ToLower(c.RemainingArgs()[0]) == "true" {
					a.ConfigFromFile.Insecure = true
				}
				// 其他情况都是 false 也不需要写了
			}
		}
	}
	return toADGUpstream(a)
}

// toADGUpstream 解析配置并转换为 dnsproxy 需要的上游
func toADGUpstream(a *ADGProxy) error {
	// 读取来自文件的 Config 并组装 dnsproxy 配置
	opts := &upstream.Options{}
	// 如果 bootstrap 为空则添加一个默认的
	if len(a.ConfigFromFile.Bootstraps) == 0 {
		a.ConfigFromFile.Bootstraps = append(a.ConfigFromFile.Bootstraps, "https://223.5.5.5/dns-query")
	}
	opts.Bootstrap = a.ConfigFromFile.Bootstraps
	// 暂定 5 秒
	opts.Timeout = time.Second * 5
	opts.InsecureSkipVerify = a.ConfigFromFile.Insecure
	for i, up := range a.ConfigFromFile.Upstreams {
		u, err := upstream.AddressToUpstream(up, opts)
		if err != nil {
			log.Errorf("解析上游 %d: %s 失败", i, err)
			return err
		}
		a.Upstreams = append(a.Upstreams, u)
		log.Infof("上游 %d 为: %s", i, up)
	}
	return nil
}
