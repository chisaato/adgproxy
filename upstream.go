package adgproxy

import (
	"fmt"
	"github.com/AdguardTeam/dnsproxy/upstream"
	"log/slog"
)

func initBootstrap(
	bootstraps []string,
	opts *upstream.Options,
) (r upstream.Resolver, err error) {
	var resolvers []upstream.Resolver

	for i, b := range bootstraps {
		var ur *upstream.UpstreamResolver
		ur, err = upstream.NewUpstreamResolver(b, opts)
		if err != nil {
			return nil, fmt.Errorf("creating bootstrap resolver at index %d: %w", i, err)
		}

		resolvers = append(resolvers, upstream.NewCachingResolver(ur))
	}

	switch len(resolvers) {
	case 0:
		// 必须写一个上游,所以这个时候报错
		return nil, fmt.Errorf("no bootstraps provided")
		//return upstream.ConsequentResolver{ net.DefaultResolver}, nil
	case 1:
		return resolvers[0], nil
	default:
		return upstream.ParallelResolver(resolvers), nil
	}
}

// toADGUpstream 解析配置并转换为 dnsproxy 需要的上游
func toADGUpstream(a *ADGProxy) error {
	httpVersions := upstream.DefaultHTTPVersions

	bootOpts := &upstream.Options{
		Logger:             slog.Default(),
		HTTPVersions:       httpVersions,
		InsecureSkipVerify: a.ConfigFromFile.Insecure,
		Timeout:            a.ConfigFromFile.Timeout,
	}
	boot, err := initBootstrap(a.ConfigFromFile.Bootstraps, bootOpts)
	if err != nil {
		return fmt.Errorf("initializing bootstrap: %w", err)
	}
	upstreamOpts := &upstream.Options{
		Logger:             slog.Default(),
		HTTPVersions:       httpVersions,
		InsecureSkipVerify: a.ConfigFromFile.Insecure,
		Bootstrap:          boot,
		Timeout:            a.ConfigFromFile.Timeout,
	}
	// 读取来自文件的 Config 并组装 dnsproxy 配置
	for i, up := range a.ConfigFromFile.Upstreams {
		u, err := upstream.AddressToUpstream(up, upstreamOpts)
		if err != nil {
			log.Errorf("解析上游 %d: %s 失败", i, err)
			return err
		}
		a.Upstreams = append(a.Upstreams, u)
		log.Infof("上游 %d 为: %s", i, up)
	}
	return nil
}
