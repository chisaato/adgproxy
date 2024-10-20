package adgproxy

import (
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"os"
	"path"
	"strings"
	"time"
)

var pluginName = "adgproxy"

// init registers this plugin.
func init() { plugin.Register(pluginName, setup) }

type configFromFile struct {

	// Upstreams 上游服务器
	Upstreams []string
	// Bootstraps 引导服务器
	Bootstraps []string
	// Insecure 是否允许不安全的连接
	Insecure bool
	// UpstreamMode 上游模式
	UpstreamMode proxy.UpstreamMode
	// Timeout 连接超时时间
	Timeout time.Duration
	// GeoSite geosite 文件路径
	GeoSite string
	// GeoIP geoip 文件路径
	GeoIP string
	// Rules 规则
	Rules []string
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
		log.Debugf("进入下一个 token 内容是 %s", c.Val())
		for c.NextBlock() {
			//println(c.Val())
			//log.Debug("准备解析 " + c.Val() + " 节")
			selector := strings.ToLower(c.Val())

			switch selector {
			case "upstream":
				log.Debug("解析到 upstream 节")
				upsteam := c.RemainingArgs()[0]
				log.Debugf("得到上游服务器 %s", upsteam)
				a.ConfigFromFile.Upstreams = append(a.ConfigFromFile.Upstreams, upsteam)
				break
			case "bootstrap":
				log.Debug("解析到 bootstrap 节")
				bootstrap := c.RemainingArgs()[0]
				log.Debugf("得到引导服务器 %s", bootstrap)
				a.ConfigFromFile.Bootstraps = append(a.ConfigFromFile.Bootstraps, bootstrap)
				break
			case "insecure":
				log.Debug("解析到 insecure 节")
				// 只有后面跟了个 true 才会是 true
				if len(c.RemainingArgs()) == 1 {
					insecure := c.Val()
					log.Debugf("得到不安全连接 %s", insecure)
					if insecure == "true" {
						a.ConfigFromFile.Insecure = true
					}
				}
				break
			case "mode":
				log.Debug("解析到 mode 节")
				// 必须是 load_balance 或者 parallel
				mode := c.RemainingArgs()[0]
				log.Debugf("得到模式 %s", mode)
				switch mode {
				case "load_balance":
					a.ConfigFromFile.UpstreamMode = proxy.UpstreamModeLoadBalance
				case "parallel":
					a.ConfigFromFile.UpstreamMode = proxy.UpstreamModeParallel
				default:
					log.Debugf("未知的模式 %s 将默认使用 load_balance", mode)
					a.ConfigFromFile.UpstreamMode = proxy.UpstreamModeLoadBalance
				}
				break
			case "timeout":
				log.Debug("解析到 timeout 节")
				// 没有就是 5 秒
				timeout := c.RemainingArgs()[0]
				log.Debugf("得到超时时间 %s", timeout)
				t, err := time.ParseDuration(timeout)
				if err != nil {
					log.Debugf("解析超时时间 %s 失败, 使用默认值 5 秒", timeout)
					t = time.Second * 5
				}
				a.ConfigFromFile.Timeout = t
				break
			case "geosite":
				log.Debug("解析到 geosite 节")
				// 没有参数就用运行目录下的 geosite.dat
				if len(c.RemainingArgs()) == 0 {
					log.Debug("没有参数,默认为 geosite.dat")
					cwd, _ := os.Getwd()
					a.ConfigFromFile.GeoSite = path.Join(cwd, "geosite.dat")
				} else {
					log.Debugf("得到 geosite 文件 %s", c.Val())
					a.ConfigFromFile.GeoSite = c.Val()
				}
				// 检测文件在不在
				if _, err := os.Stat(a.ConfigFromFile.GeoSite); err != nil {
					log.Errorf("geosite 文件 %s 不存在", a.ConfigFromFile.GeoSite)
					return ErrGeoSiteFileNotFound
				}
				break
			case "geoip":
				log.Debug("解析到 geoip 节")
				// 没有参数就用运行目录下的 geoip.dat
				if len(c.RemainingArgs()) == 0 {
					cwd, _ := os.Getwd()
					a.ConfigFromFile.GeoIP = path.Join(cwd, "geoip.dat")
				} else {
					a.ConfigFromFile.GeoIP = c.Val()
				}
				// 检测文件在不在
				if _, err := os.Stat(a.ConfigFromFile.GeoIP); err != nil {
					log.Errorf("geoip 文件 %s 不存在", a.ConfigFromFile.GeoIP)
					return ErrGeoIPFileNotFound
				}
				break

			case "rule":
				log.Debug("解析到 rule 节")
				// 跟前面解析上游一样,塞进去就算完事
				rule := c.RemainingArgs()[0]
				log.Debugf("得到规则 %s", rule)
				a.ConfigFromFile.Rules = append(a.ConfigFromFile.Rules, rule)
			}
		}
	}
	// 校验一下
	// 首先起码你要有上游对不对
	if len(a.ConfigFromFile.Upstreams) == 0 {
		return plugin.Error(pluginName, ErrNoUpstreams)
	}
	// 如果没有引导服务器就添加一个默认的
	if len(a.ConfigFromFile.Bootstraps) == 0 {
		a.ConfigFromFile.Bootstraps = append(a.ConfigFromFile.Bootstraps, "https://223.5.5.5/dns-query")
	}
	// 检测

	return toADGUpstream(a)
}
