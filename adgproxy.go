// Package example is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package adgproxy

import (
	"context"
	"time"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin(pluginName)

// ADGProxy is an example plugin to show how to write a plugin.
type ADGProxy struct {
	Next plugin.Handler
	// Upstreams 记录 dnsproxy 所用的上游结构体
	Upstreams []upstream.Upstream
	// ConfigFromFile 记录从配置文件解析出来的结果
	ConfigFromFile *configFromFile
	// Options 记录 dnsproxy 的配置
	Options *upstream.Options
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (a ADGProxy) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// This function could be simpler. I.a. just fmt.Println("example") here, but we want to show
	// a slightly more complex example as to make this more interesting.
	// Here we wrap the dns.ResponseWriter in a new ResponseWriter and call the next plugin, when the
	// answer comes back, it will print "example".

	// Debug log that we've have seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received response")

	// Wrap.
	pw := NewResponsePrinter(w)

	// Export metric with the server label set to the current server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	// Call next plugin (if any).
	return plugin.NextOrFailure(a.Name(), a.Next, ctx, pw, r)
}

// Name implements the Handler interface.
func (a ADGProxy) Name() string { return pluginName }

func (a *ADGProxy) OnStartup() error {
	// 读取来自文件的 Config 并组装 dnsproxy 配置
	opts := &upstream.Options{}
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

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("example")
	return r.ResponseWriter.WriteMsg(res)
}
