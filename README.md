<p align = "center">
<img alt="Wide" src="https://user-images.githubusercontent.com/873584/57901570-5355ba00-7898-11e9-96ca-45b75b1d70db.png">
<br><br>
一款基于 Web 的 Go 语言 IDE
<br><br>
<a title="Build Status" target="_blank" href="https://travis-ci.org/b3log/wide"><img src="https://img.shields.io/travis/b3log/wide.svg?style=flat-square"></a>
<a title="Go Report Card" target="_blank" href="https://goreportcard.com/report/github.com/b3log/wide"><img src="https://goreportcard.com/badge/github.com/b3log/wide?style=flat-square"></a>
<a title="Coverage Status" target="_blank" href="https://coveralls.io/repos/github/b3log/wide/badge.svg?branch=master"><img src="https://img.shields.io/coveralls/github/b3log/wide.svg?style=flat-square&color=CC9933"></a>
<a title="Code Size" target="_blank" href="https://github.com/b3log/wide"><img src="https://img.shields.io/github/languages/code-size/b3log/wide.svg?style=flat-square"></a>
<a title="Apache License" target="_blank" href="https://github.com/b3log/wide/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-apache2-orange.svg?style=flat-square"></a>
<br>
<a title="Releases" target="_blank" href="https://github.com/b3log/wide/releases"><img src="https://img.shields.io/github/release/b3log/wide.svg?style=flat-square"></a>
<a title="Release Date" target="_blank" href="https://github.com/b3log/wide/releases"><img src="https://img.shields.io/github/release-date/b3log/wide.svg?style=flat-square&color=99CCFF"></a>
<a title="GitHub Commits" target="_blank" href="https://github.com/b3log/wide/commits/master"><img src="https://img.shields.io/github/commit-activity/m/b3log/wide.svg?style=flat-square"></a>
<a title="Last Commit" target="_blank" href="https://github.com/b3log/wide/commits/master"><img src="https://img.shields.io/github/last-commit/b3log/wide.svg?style=flat-square&color=FF9900"></a>
<a title="GitHub Pull Requests" target="_blank" href="https://github.com/b3log/wide/pulls"><img src="https://img.shields.io/github/issues-pr-closed/b3log/wide.svg?style=flat-square&color=FF9966"></a>
<a title="Hits" target="_blank" href="https://github.com/b3log/hits"><img src="https://hits.b3log.org/b3log/wide.svg"></a>
<br><br>
<a title="GitHub Watchers" target="_blank" href="https://github.com/b3log/wide/watchers"><img src="https://img.shields.io/github/watchers/b3log/wide.svg?label=Watchers&style=social"></a>&nbsp;&nbsp;
<a title="GitHub Stars" target="_blank" href="https://github.com/b3log/wide/stargazers"><img src="https://img.shields.io/github/stars/b3log/wide.svg?label=Stars&style=social"></a>&nbsp;&nbsp;
<a title="GitHub Forks" target="_blank" href="https://github.com/b3log/wide/network/members"><img src="https://img.shields.io/github/forks/b3log/wide.svg?label=Forks&style=social"></a>&nbsp;&nbsp;
<a title="Author GitHub Followers" target="_blank" href="https://github.com/88250"><img src="https://img.shields.io/github/followers/88250.svg?label=Followers&style=social"></a>
</p>

## 简介

Wide 是一款基于 **W**eb 的 Go 语言 **IDE**。

## 动机

目前较为流行的 Go IDE 都有一些缺陷或遗憾：

  * 文本编辑器类（vim/emacs/sublime/Atom 等）：对于新手门槛太高，搭建复杂
  * 插件类（goclipse、IDEA 等）：需要原 IDE 支持，不够专业
  * LiteIDE 界面不够 modern、goland 收费
  * **缺少网络分享、嵌入网站可运行功能**

另外，Go IDE 很少，用 Go 本身开发的 IDE 更是没有，这是一次很好的尝试。关于产品定位的讨论请看[这里](https://hacpai.com/article/1438407961481)。

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

## 开源项目推荐

* 如果你需要搭建一个个人博客系统，可以考虑使用 [Solo](https://github.com/b3log/solo)
* 如果你需要搭建一个多用户博客平台，可以考虑使用 [Pipe](https://github.com/b3log/pipe)
* 如果你需要搭建一个社区平台，可以考虑使用 [Sym](https://github.com/b3log/symphony)
* 欢迎加入我们的小众开源社区，详情请看[这里](https://hacpai.com/article/1463025124998)
