package adgproxy

import (
	"errors"
)

var ErrNoUpstreams = errors.New("没有配置上游 DNS")
var ErrGeoSiteFileNotFound = errors.New("geosite 文件不存在")
var ErrGeoIPFileNotFound = errors.New("geoip 文件不存在")
