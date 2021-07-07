# 介绍

httpBruter是集网页title获取、指纹探测、CDN探测、VHOST探测、网页截图、IP反查等功能于一体的工具，主要用于目标信息收集工作。

# 参数说明

```
  -binary string
        设置chrome浏览器的默认路径
  -cdn
        开启CDN探测
  -database
        存入远程mongo数据库
  -debug
        开启错误输出
  -finger
        开启CMS指纹探测
  -headless
        使用Headless访问
  -iL string
        输入文件名
  -match string
        使用正则匹配过滤结果
  -no-fallback
        如果IP地址是443端口，再次访问其80端口
  -no-redirects
        设置不跟随重定向
  -oJ string
        设置输出json文件的路径
  -oT string
        设置输出普通文件的路径
  -retries int
        设置重试次数，默认0
  -reverse
        开启反查IP功能
  -reverse-option string
        选择反查IP的接口：'ipchaxun'或'webscan' (default "all")
  -screenshots
        开启截图(保存在当前目录的'screenshots'文件夹中)
  -screenshots-fullpage
        设置全屏截图
  -screenshots-path string
        设置截图存储路径
  -stats
        显示实时统计信息
  -threads int
        设置并发数，如果开启了headless或截图，建议将并发数设置到20以内 (default 50)
  -vhost
        开启VHOST探测
```

# 常见问题

1. 没有返回任何的内容，只是输出统计信息

    解决方法：可以使用参数`-debug`来查看具体原因。如果是请求失败，默认情况下错误信息不会输出到控制台中，所以也就只能看到统计信息。具体请求失败的原因参照输出的错误信息来修复。

2. 一直卡在统计信息很久，没有进行新的请求

    解决办法：使用`Ctrl+C`终止程序，问题原因还没弄清楚。