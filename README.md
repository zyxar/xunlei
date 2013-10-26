# 迅雷離線

迅雷離線API、命令行工具以及後臺工具

[![Build Status](https://travis-ci.org/matzoe/xunlei.png?branch=master)](https://travis-ci.org/matzoe/xunlei)

原 repo 見[此](https://github.com/zyxar/xltask)

Inspired by [task.js](http://cloud.vip.xunlei.com/190/js/task.js?269), and [xunlei-lixian](https://github.com/iambus/xunlei-lixian)

## Package Status

- api:  **✏**
- lxc:  **✏**
- lxd:  **✍**

## xunlei/api

見 [API 文檔](https://godoc.org/github.com/matzoe/xunlei/api)

## xunlei/lxc

命令行工具

支持命令列表：

| Cmds              | Level   | Status|
| ----------------- |:-------:|:-----:|
| `cls`, `clear`    | Utility | **✓** |
| `saveconf`        | Utility |       |
| `loadconf`        | Utility |       |
| `quit`, `exit`    | Utility | **✓** |
| `version`         | Utility | **✓** |
| `help`            | Utility |       |
| `?*`              | Utility |       |
| `relogin`         | Service | **✓** |
| `ison`            | Service | **✓** |
| `ls`              | Service | **✓** |
| `ls`              | Service | **✓** |
| `ld`              | Service | **✓** |
| `le`              | Service | **✓** |
| `lc`              | Service | **✓** |
| `ll`              | Service | **✓** |
| `add`             | Service | **✓** |
| `rm`, `delete`    | Service | **✓** |
| `purge`           | Service | **✓** |
| `readd`           | Service | **✓** |
| `readdexp`        | Service |       |
| `find`            | Service | **✓** |
| `update`          | Service | **✓** |
| `info`            | Task    | **✓** |
| `dl`, `download`  | Task    | **✓** |
| `dt`              | Task    | **✓** |
| `ti`              | Task    | **✓** |
| `pause`           | Task    | **✓** |
| `resume`          | Task    | **✓** |
| `rename`, `mv`    | Task    | **✓** |
| `delay`           | Task    | **✓** |
| `submit`          | Task    |       |
| `play`            | Task    |       |
| `link`            | Task    | **✓** |

支持以 `aria2c` 爲下載工具，或者自行定製

![](http://farm4.staticflickr.com/3697/10421561225_aa3ea3f4e5_c.jpg)
![](http://farm6.staticflickr.com/5530/10461504605_8dc2b2737b_c.jpg)

## xunlei/lxd

後臺工具

- 接受 `RESTful` 調用
- 接受 `JSON-RPC` 調用
