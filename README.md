# Wide [![Build Status](https://drone.io/github.com/b3log/wide/status.png)](https://drone.io/github.com/b3log/wide/latest)

## Intro

A <b>W</b>eb-based <b>IDE</b> for Teams using Golang.

<img src="https://cloud.githubusercontent.com/assets/873584/4606377/d0ca3c2a-521b-11e4-912c-d955ab05850b.png" width="100%" />

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

* Code Highlight, Folding: Go/HTML/JavaScript/Markdown etc.
* Autocomplete: Go/HTML etc.
* Format: Go/HTML/JSON etc.
* Run & Debug: run/debug multiple processes at the same time
* Multiplayer: a real team development experience
* Navigation, Jump to declaration, Find usages, File search etc.
* Shell: run command on the server
* Git integration: git command on the web
* Web development: Frontend devlopment (HTML/JS/CSS) all in one
* Go tool: go get/install/fmt etc.

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
 * Find Usages

Flow: 
 1. Browser sends code assist request
 2. Handler gets user workspace of the request with HTTP session
 3. Server executes ````gocode````/````ide_stub````<br/>
    3.1 Sets environment variables (e.g. ${GOPATH})<br/>
    3.2 ````gocode```` with ````lib-path```` parameter

## Documents

* [用户指南](http://88250.gitbooks.io/wide-user-guide)
* [开发指南](http://88250.gitbooks.io/wide-dev-guide)

## Demos

* 20141024-1.0.0, png ![](http://b3log.org/wide/demo/20141024.png)

### Olds
* [20140927, png](http://b3log.org/wide/demo/20140927.png)
* [20140913, png](http://b3log.org/wide/demo/20140913.png)
* [20140910, png](http://b3log.org/wide/demo/20140910.png)
* [20140823, swf](http://b3log.org/wide/demo/20140823.html)

## Setup

### Download Binary

We have provided OS-specific executable binary as follows: 

* linux-amd64/386
* windows-amd64/386
* darwin-amd64/386

Download [HERE](http://pan.baidu.com/s/1dD3XwOT)!

### Build Wide for yourself

1. [Download](https://github.com/b3log/wide/archive/master.zip) source or by `git clone`
2. Get dependencies with 
   * `go get -u`
   * `go get -u github.com/88250/ide_stub`
   * `go get -u github.com/nsf/gocode`
3. Compile wide with `go build` 

## Known Issues

* [Shell is not available on Windows](https://github.com/b3log/wide/issues/32)

## License

Copyright (c) 2014, B3log Team (http://b3log.org)

Licensed under the [Apache License 2.0](https://github.com/b3log/wide/blob/master/LICENSE).

## Credits

* [golang](http://golang.org)
* [CodeMirror](https://github.com/marijnh/CodeMirror)
* [zTree](https://github.com/zTree/zTree_v3) 
* [LiteIDE](https://github.com/visualfc/liteide)
* [gocode](https://github.com/nsf/gocode)
* [Gorilla](https://github.com/gorilla)
* [GoBuild](http://gobuild.io)

----

<img src="https://cloud.githubusercontent.com/assets/873584/4606328/4e848b96-5219-11e4-8db1-fa12774b57b4.png" width="256px" />
</center>
