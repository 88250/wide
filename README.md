# [Wide](https://github.com/b3log/wide) [![Build Status](https://img.shields.io/travis/b3log/wide.svg?style=flat)](https://travis-ci.org/b3log/wide) [![Go Report Card](https://goreportcard.com/badge/github.com/b3log/wide)](https://goreportcard.com/report/github.com/b3log/wide) [![Coverage Status](https://img.shields.io/coveralls/b3log/wide.svg?style=flat)](https://coveralls.io/r/b3log/wide) [![Apache License](https://img.shields.io/badge/license-apache2-orange.svg?style=flat)](https://www.apache.org/licenses/LICENSE-2.0) [![API Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/b3log/wide) [![Download](https://img.shields.io/badge/download-~4.3K-red.svg?style=flat)](https://pan.baidu.com/s/1dD3XwOT)

先试试我们搭建好的[在线服务](https://wide.b3log.org/signup)，你可以在这里[下载](https://pan.baidu.com/s/1dD3XwOT)并在本地环境运行，然后邀请小伙伴们来玩吧！

## 简介

Wide 是一个基于 **W**eb 的 Go 语言 **IDE**。

![](https://cloud.githubusercontent.com/assets/873584/4606377/d0ca3c2a-521b-11e4-912c-d955ab05850b.png)

## 动机

目前较为流行的 Go IDE 都有一些缺陷或遗憾：
  * 文本编辑器类（vim/emacs/sublime/Atom 等）：对于新手门槛太高，搭建复杂
  * 插件类（goclipse、IDEA 等）：需要原 IDE 支持，不够专业
  * LiteIDE 界面不够 modern、goland 收费
  * **缺少网络分享、嵌入网站可运行功能**

另外，Go IDE 很少，用 Go 本身开发的 IDE 更是没有，这是一个很好的尝试。关于产品定位的讨论请看[这里](https://hacpai.com/article/1438407961481)。

## 特性

基于 Web 的 IDE：

* 只需要浏览器就能进行开发、运行
* 跨平台，甚至在移动设备上
* 易进行功能扩展
* 易与其他系统集成
* 极客体验
  
核心功能：

* 代码高亮、折叠：Go/HTML/JavaScript/Markdown 等
* 自动完成：Go/HTML 等
* 编译检查：编辑器提示编译错误
* 格式化：Go/HTML/JSON 等
* 运行：支持同时运行多个程序
* 代码导航：跳转到声明，查找使用，文件搜索等
* Web 开发：前端（HTML/JS/CSS）开发支持
* go tool：go get/install/fmt 等
* 项目文件导出
* UI/编辑器多主题
* 支持交叉编译

## 界面

### 主界面
  
![Overview](https://cloud.githubusercontent.com/assets/873584/5450620/1d51831e-8543-11e4-930b-670871902425.png)

### 跳转到文件
  
![Goto File](https://cloud.githubusercontent.com/assets/873584/5450616/1d495da6-8543-11e4-9285-f9d9c60779ac.png)

### 自动完成
  
![Autocomplete](https://cloud.githubusercontent.com/assets/873584/5450619/1d4d5712-8543-11e4-8fe4-35dbc8348a6e.png)

### 主题 

![Theme](https://cloud.githubusercontent.com/assets/873584/5450617/1d4c0826-8543-11e4-8b86-f79a4e41550a.png)

### 查看表达式
  
![Show Expression Info](https://cloud.githubusercontent.com/assets/873584/5450618/1d4cd9f4-8543-11e4-950f-121bd3ff4a39.png)

### 构建报错提示
  
![Build Error Info](https://cloud.githubusercontent.com/assets/873584/5450632/3e51cccc-8543-11e4-8ca8-8d2427aa16b8.png)

### 交叉编译

![Cross-Compilation](https://cloud.githubusercontent.com/assets/873584/10130037/226d75fc-65f7-11e5-94e4-25ee579ca175.png)

### Playground

![Playground](https://cloud.githubusercontent.com/assets/873584/21209772/449ecfd2-c2b1-11e6-9aa6-a83477d9f269.gif)
  
## 架构 

### 构建与运行

![Build & Run](https://cloud.githubusercontent.com/assets/873584/4389219/3642bc62-43f3-11e4-8d1f-06d7aaf22784.png)

* 一个浏览器 tab 对应一个 Wide 会话
* 通过 WebSocket 进行程序执行输出推送

1. 客户端浏览器发送 ````Build```` 请求
2. 服务器使用 ````os/exec```` 执行 ````go build```` 命令<br/>
   2.1. 生成可执行文件
3. 客户端浏览器发送 ````Run```` 请求
4. 服务器使用 ````os/exec```` 执行文件<br/>
   4.1. 生成进程<br/>
   4.2. 运行结果输出到 WebSocket 通道
5. 客户端浏览器监听 ````ws.onmessage```` 到消息后做展现

### 代码辅助

![](https://cloud.githubusercontent.com/assets/873584/4399135/3b80c21c-4463-11e4-8e94-7f7e8d12a4df.png)

* 自动完成
* 查找使用

1. 浏览器客户端发送代码辅助请求
2. Handler 根据请求对应的 HTTP 会话获取用户工作空间
3. 执行 `gocode`/`ide_stub(gotools)` 命令<br/>
   3.1 设置环境变量（${GOPATH} 为用户工作空间路径）<br/>
   3.2 `gocode` 命令需要设置参数 `lib-path`

## 文档

* [用户指南](https://hacpai.com/article/1538873544275)
* [开发指南](https://hacpai.com/article/1538876422995)

## 社区

* [讨论区](https://hacpai.com/tag/wide)
* [报告问题](https://github.com/b3log/wide/issues/new/choose)

## 授权

Wide 使用 [Apache License, Version 2](https://www.apache.org/licenses/LICENSE-2.0) 作为开源协议，请务必遵循该开源协议相关约定。

## 鸣谢

* [golang](https://golang.org)
* [CodeMirror](https://github.com/marijnh/CodeMirror)
* [zTree](https://github.com/zTree/zTree_v3) 
* [LiteIDE](https://github.com/visualfc/liteide)
* [gocode](https://github.com/nsf/gocode)
* [Gorilla](https://github.com/gorilla)
* [Docker](https://docker.com)

----

<img src="https://cloud.githubusercontent.com/assets/873584/4606328/4e848b96-5219-11e4-8db1-fa12774b57b4.png" width="256px" />
