# goconfig-center的apollo组件

完成对goconfig-center的扩展，使goconfig-center支持apollo配置中心

对接方式：apollo

配置
```toml
[[config_center.drivers]]
enable = true
driver = "apollo" // 二选一
prefix = "" // 可以为空
appId = "" // appid，required
endpoint = "" // apollo地址，required
namespace = "" // 命名空间，多个已逗号分割，required
cluster = "" // 集群
accessKey = "" // 密钥
slb = false // 是否启用slb
ip = "" // 是否指定ip
```
