# 迅雷離線

迅雷離線API、命令行工具以及後臺工具

[![Go Report Card](https://goreportcard.com/badge/github.com/zyxar/xunlei)](https://goreportcard.com/report/github.com/zyxar/xunlei)
[![Build Status](https://travis-ci.org/zyxar/xunlei.png?branch=master)](https://travis-ci.org/zyxar/xunlei)

rewritten based on [xltask](https://github.com/zyxar/xltask)

Inspired by
- [task.js](http://cloud.vip.xunlei.com/190/js/task.js?269)
- [xunlei-lixian](https://github.com/iambus/xunlei-lixian)

## Package Status

- [x] protocol
- [x] lx

## xunlei/protocol

[![GoDoc](https://godoc.org/github.com/zyxar/xunlei/protocol?status.svg)](https://godoc.org/github.com/zyxar/xunlei/protocol)

## xunlei/`lx -d`

後臺工具

```bash
$ lx -d #start web server and daemonize
$ lx -loop #start web server
```

- [x] 接受 `RESTful` 調用
- [ ] 接受 `JSON-RPC` 調用

## xunlei/static

Simple web front end, build upon:

  - jQuery
  - Bootstrap

## xunlei/`lx`

命令行工具

支持以 `aria2c` 爲（後臺）下載工具，或者自行定製

![](http://farm4.staticflickr.com/3697/10421561225_aa3ea3f4e5_c.jpg)
![](http://farm6.staticflickr.com/5530/10461504605_8dc2b2737b_c.jpg)

## TODO

- [x] really daemonize `lx(-d)`;
- [x] start `aria2` standalone, and submit tasks to it via RPC;
- [ ] Node.js frontend/client;
- [ ] VOD API and play demo;

## LICENSE

MPL v2
