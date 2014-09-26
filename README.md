# Wide [![Build Status](https://drone.io/github.com/b3log/wide/status.png)](https://drone.io/github.com/b3log/wide/latest)

## Intro

A <b>W</b>eb <b>IDE</b> IDE for Teams using Golang.

## Motivation

 * **Team** IDE:
   * Safe and reliable: the project source code stored on the server in real time, the developer's machine crashes without losing any source code 
   * Unified environment: server unified development environment configuration, the developer machine without any additional configuration 
   * Out of the box: 5 minutes to setup a server then open browser to develop, debug
   * Version Control: each developer has its own source code repository, easy sync with the trunk 
 * **Web based** IDE:
   * Developer needs a browser only
   * Cross-platform, even on mobile devices
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


 1. Browser sends code assist request
 2. Handler gets user workspace of the request with HTTP session
 3. Server executes ````gocode````/````ide_stub````<br/>
    3.1 Sets environment variables (e.g. ${GOPATH})<br/>
    3.2 ````gocode```` with ````lib-path```` parameter

## Documents

 * [用户指南](http://88250.gitbooks.io/wide-user-guide/zh-cn/index.html)
 * [开发指南](http://88250.gitbooks.io/wide-dev-guide/zh-cn/index.html)

## Demos

 * 20140913, png ![](http://b3log.org/wide/demo/20140913.png)

### Olds

 * [20140910, png](http://b3log.org/wide/demo/20140910.png)
 * [20140823, swf](http://b3log.org/wide/demo/20140823.html)

## Setup from sources

 1. Downloads source
 2. Gets dependencies with 
    * `go get -u`
    * `go get -u github.com/88250/ide_stub`
    * `go get -u github.com/nsf/gocode`
 3. Compiles wide with `go build` 
 4. Configures `conf/wide.json`
 5. Runs the executable `wide` or `wide.exe`

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
