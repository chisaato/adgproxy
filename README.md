# adgproxy 插件

使用 AdGuard 相关组件实现 DNS 客户端的 CoreDNS 插件.

配置文件

```text
adgproxy {
    # 上游,可以写多个
    upstream https://1.1.1.1/dns-query
    # 引导服务器,可以写多个
    bootstrap https://223.5.5.5/dns-query
    # 模式,负载均衡或并行
    mode load_balance
    # mode parallel
    # 是否允许不安全的 TLS,你不会想用 true 的
    # 只有声明 true 的时候才会开启 insecure 只写一个 insecure 也当做 false
    # insecure false
    # geosite 规则文件,留空就用运行目录下的 geosite.dat
    geosite geosite.dat
    # geoip 规则文件,留空就用运行目录下的 geoip.dat
    geoip geoip.dat
    # 规则,自上而下处理,不写的默认 PASS (即交给后续处理)
    # PROXY 表示转发
    rule GEOSITE,google,PROXY
    rule GEOSITE,facebook,PROXY
    # PASS 表示不处理,交给后续
    rule GEOSITE,bilibili,PASS
}
```
