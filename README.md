# 迅雷離線

迅雷離線API、命令行工具以及後臺工具

[![Build Status](https://travis-ci.org/matzoe/xunlei.png?branch=master)](https://travis-ci.org/matzoe/xunlei)

rewritten based on [xltask](https://github.com/zyxar/xltask)

Inspired by [task.js](http://cloud.vip.xunlei.com/190/js/task.js?269), and [xunlei-lixian](https://github.com/iambus/xunlei-lixian)

## Package Status

- api:  **✏**
- lxc:  **✏**
- <del>lxd:  **✍**</del> *replaced by `lxc -d`*

## xunlei/api

見 [API 文檔](https://godoc.org/github.com/matzoe/xunlei/api)

## xunlei/`lxc -d`

後臺工具

```bash
lxc -d #start web server and daemonize
lxc -loop #start web server
```

- [x] 接受 `RESTful` 調用
- [ ] 接受 `JSON-RPC` 調用

## xunlei/static

Simple web front end, build upon:

  - jQuery
  - Bootstrap

## xunlei/lxc

命令行工具

支持以 `aria2c` 爲下載工具，或者自行定製

![](http://farm4.staticflickr.com/3697/10421561225_aa3ea3f4e5_c.jpg)
![](http://farm6.staticflickr.com/5530/10461504605_8dc2b2737b_c.jpg)

## TODO

- [x] really daemonize `lxc(-d)`;
- [ ] Node.js frontend/client;
- [ ] start `aria2` standalone, and submit tasks to it via RPC;
- [ ] VOD API and play demo;

## LICENSE

MPL v2