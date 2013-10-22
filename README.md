# 迅雷離線

迅雷離線API、命令行工具以及後臺工具

[![Build Status](https://travis-ci.org/matzoe/xunlei.png?branch=master)](https://travis-ci.org/matzoe/xunlei)

原 repo 見[此](https://github.com/zyxar/xltask)

Inspired by [task.js](http://cloud.vip.xunlei.com/190/js/task.js?269), and [xunlei-lixian](https://github.com/iambus/xunlei-lixian)


## xunlei/api

見 [API 文檔](https://godoc.org/github.com/matzoe/xunlei/api)

## xunlei/lxc

命令行工具

支持命令列表：

- `cls`, `clear`
- `ls`
- `ld`
- `le`
- `lc`
- `ll`
- `info`
- `dl`, `download`
- `add`
- `rm`, `delete`
- `purge`
- `readd`
- `pause`
- `restart`
- `rename`, `mv`
- `delay`
- `link`
- `dispatch`
- `version`
- `update`
- `quit`, `exit`

支持以 `aria2c` 爲下載工具，或者自行定製

## xunlei/lxd

後臺工具

- 接受 `RESTful` 調用
- 接受 `JSON-RPC` 調用
