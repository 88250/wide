# [Wide](https://github.com/b3log/wide) [![Build Status](https://img.shields.io/travis/b3log/wide.svg?style=flat)](https://travis-ci.org/b3log/wide) [![Coverage Status](https://img.shields.io/coveralls/b3log/wide.svg?style=flat)](https://coveralls.io/r/b3log/wide) [![Apache License](http://img.shields.io/badge/license-apache2-orange.svg?style=flat)](http://www.apache.org/licenses/LICENSE-2.0) [![API Documentation](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](http://godoc.org/github.com/b3log/wide) [![Download](http://img.shields.io/badge/download-~3.1K-red.svg?style=flat)](http://pan.baidu.com/s/1dD3XwOT)

_Have a [try](http://wide.b3log.org/signup) first, then [download](http://pan.baidu.com/s/1dD3XwOT) and setup it on your local area network, enjoy yourself!_

先试试我们搭建好的[在线服务](http://wide.b3log.org/signup)，你可以在这里[下载](http://pan.baidu.com/s/1dD3XwOT)并在本地环境运行，然后邀请小伙伴们来玩吧！

> * 关于 Wide 的产品定位，请看[这里](http://hacpai.com/article/1438407961481)，并欢迎参与讨论~
> * 加入[**黑客派**](http://hacpai.com/register)，与其他程序员、设计师共同成长！

## Introduction

A <b>W</b>eb-based <b>IDE</b> for Teams using Go programming language/Golang.

![Hello, 世界](https://cloud.githubusercontent.com/assets/873584/4606377/d0ca3c2a-521b-11e4-912c-d955ab05850b.png)

## Authors

[Daniel](https://github.com/88250) and [Vanessa](https://github.com/Vanessa219) are the main authors of Wide, [here](https://github.com/b3log/wide/graphs/contributors) are all contributors.

Wide 的主要作者是 [Daniel](https://github.com/88250) 与 [Vanessa](https://github.com/Vanessa219)，所有贡献者可以在[这里](https://github.com/b3log/wide/graphs/contributors)看到。

## Motivation

* **Team** IDE:
  * _Safe and reliable_: the project source code stored on the server in real time, the developer's machine crashes without losing any source code 
  * _Unified environment_: server unified development environment configuration, the developer machine without any additional configuration 
  * _Out of the box_: 5 minutes to setup a server then open browser to develop, debug
  * _Version Control_: each developer has its own source code repository, easy sync with the trunk 
* **Web-based** IDE:
  * Developer needs a browser only
  * Cross-platform, even on mobile devices
  * Easy to extend
  * Easy to integrate with other systems
  * For the geeks
* A try for commercial-open source: versions customized for enterprises, close to their development work flows respectively
* Currently more popular Go IDE has some defects or regrets: 
  * Text editor (vim/emacs/sublime/Atom, etc.): For the Go newbie is too complex 
  * Plug-in (goclipse, etc.): the need for the original IDE support, not professional
  * LiteIDE: no modern user interface :p
  * No team development experience 
* There are a few of GO IDEs, and no one developed by Go itself, this is a nice try

## Features

* [X] Code Highlight, Folding: Go/HTML/JavaScript/Markdown etc.
* [X] Autocomplete: Go/HTML etc.
* [X] Format: Go/HTML/JSON etc.
* [X] Build & Run
* [X] Multiplayer: a real team development experience
* [X] Navigation, Jump to declaration, Find usages, File search etc.
* [X] Shell: run command on the server
* [X] Web development: HTML/JS/CSS editor with [Emmet](http://emmet.io) integrated
* [X] Go tool: go get/install/fmt etc.
* [X] File Import & Export
* [X] Themes: editor and UI adjust, respectively
* [X] Cross-Compilation
* [ ] Debug
* [ ] Git integration: git command on the web

## Screenshots

* **Overview**
  
  ![Overview](https://cloud.githubusercontent.com/assets/873584/5450620/1d51831e-8543-11e4-930b-670871902425.png)
* **Goto File**
  
  ![Goto File](https://cloud.githubusercontent.com/assets/873584/5450616/1d495da6-8543-11e4-9285-f9d9c60779ac.png)
* **Autocomplete**
  
  ![Autocomplete](https://cloud.githubusercontent.com/assets/873584/5450619/1d4d5712-8543-11e4-8fe4-35dbc8348a6e.png)
* **Theme**
  
  ![4](https://cloud.githubusercontent.com/assets/873584/5450617/1d4c0826-8543-11e4-8b86-f79a4e41550a.png)
* **Show Expression Info**
  
  ![Show Expression Info](https://cloud.githubusercontent.com/assets/873584/5450618/1d4cd9f4-8543-11e4-950f-121bd3ff4a39.png)
* **Build Error Info**
  
  ![Build Error Info](https://cloud.githubusercontent.com/assets/873584/5450632/3e51cccc-8543-11e4-8ca8-8d2427aa16b8.png)
* **Git Clone**

  ![Git Clone](https://cloud.githubusercontent.com/assets/873584/6545235/2284f230-c5b7-11e4-985e-7e04367921b1.png)
* **Cross-Compilation**
  
  ![Cross-Compilation](https://cloud.githubusercontent.com/assets/873584/10130037/226d75fc-65f7-11e5-94e4-25ee579ca175.png)

## Architecture 

### Build & Run

![Build & Run](https://cloud.githubusercontent.com/assets/873584/4389219/3642bc62-43f3-11e4-8d1f-06d7aaf22784.png)

 * A browser tab corresponds to a Wide session
 * Execution output push via WebSocket

Flow: 
 1. Browser sends ````Build```` request
 2. Server executes ````go build```` command via ````os/exec````<br/>
    2.1. Generates a executable file
 3. Browser sends ````Run```` request
 4. Server executes the file via ````os/exec````<br/>
    4.1. A running process<br/>
    4.2. Execution output push via WebSocket channel
 5. Browser renders with callback function ````ws.onmessage````

### Code Assist

![Code Assist](https://cloud.githubusercontent.com/assets/873584/4399135/3b80c21c-4463-11e4-8e94-7f7e8d12a4df.png)

 * Autocompletion
 * Find Usages/Jump To Declaration/etc.

Flow: 
 1. Browser sends code assist request
 2. Handler gets user workspace of the request with HTTP session
 3. Server executes ````gocode````/````ide_stub(gotools)````<br/>
    3.1 Sets environment variables (e.g. ${GOPATH})<br/>
    3.2 ````gocode```` with ````lib-path```` parameter

## Documents

* [用户指南](https://www.gitbook.com/book/88250/wide-user-guide)
* [开发指南](https://www.gitbook.com/book/88250/wide-dev-guide)

## Setup

### Download Binary

We have provided OS-specific executable binary as follows: 

* linux-amd64/386
* windows-amd64/386
* darwin-amd64/386

Download [HERE](http://pan.baidu.com/s/1dD3XwOT)!

### Build Wide for yourself

1. [Download](https://github.com/b3log/wide/archive/master.zip) source or by `git clone https://github.com/b3log/wide`
2. Get dependencies with 
   * `go get`
   * `go get github.com/visualfc/gotools github.com/nsf/gocode github.com/bradfitz/goimports`
3. Compile wide with `go build` 

### Docker

1. Get image: `sudo docker pull 88250/wide:latest`
2. Run: `sudo docker run -p 127.0.0.1:7070:7070 88250/wide:latest ./wide -docker=true -channel=ws://127.0.0.1:7070`
3. Open browser: http://127.0.0.1:7070

## Known Issues

* [Shell is not available on Windows](https://github.com/b3log/wide/issues/32)
* [Rename directory](https://github.com/b3log/wide/issues/251)

## Terms

* This software is open sourced under the Apache License 2.0
* You can not get rid of the "Powered by [B3log](http://b3log.org)" from any page, even which you made
* If you want to use this software for commercial purpose, please mail to support@liuyun.io for a commercial license request
* Copyright &copy; b3log.org, all rights reserved

## Credits

Wide is made possible by the following open source projects.

* [golang](http://golang.org)
* [CodeMirror](https://github.com/marijnh/CodeMirror)
* [zTree](https://github.com/zTree/zTree_v3) 
* [LiteIDE](https://github.com/visualfc/liteide)
* [gocode](https://github.com/nsf/gocode)
* [Gorilla](https://github.com/gorilla)
* [Docker](https://docker.com)

----

<img src="https://cloud.githubusercontent.com/assets/873584/4606328/4e848b96-5219-11e4-8db1-fa12774b57b4.png" width="256px" />
