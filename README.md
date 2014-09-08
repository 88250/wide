# Wide #

## Intro ##
A simple <b>W</b>eb **IDE** for golang.

## Motivation ##

* There are a few of GO IDEs, and no one developed by Go itself, this is a nice try
* Web based IDE:
  * Developer needs a browser only
  * Cross-platform, even on mobile devices
  * For the geeks
* Team IDE:
  * Safe and reliable: the project source code stored on the server in real time, the developer's machine crashes without losing any source code 
  * Unified environment: server unified development environment configuration, the developer machine without any additional configuration 
  * Out of the box: 5 minutes to setup a server then open browser to develop, debug
  * Version Control: each developer has its own source code repository, easy sync with the trunk 
* Currently more popular Go IDE has some defects or regrets: 
  * Text editor (vim/emacs/sublime/Atom, etc.): For the Go newbie is too complex 
  * Plug-in (goclipse, etc.): the need for the original IDE support, not professional
  * LiteIDE: only run one process at the same time; no modern user interface 
  * No team development experience 
* A try for commercial open source: a version customized for an enterprise, coreesponding to its development flow  
## Features ##

* Code Highlight
* Autocomplete
* Format
* Run & Debug
* Multiplayer
* Navigation & Jump
* Shell
* Git integration
* Web development

## Demos ##
* [20140823](http://b3log.org/wide/demo/20140823.html)

## Setup ##

1. Downloads source
2. Compiles wide with `go build` 
3. Configures `conf/wide.json`
4. Runs the executable `wide` or `wide.exe`

## License ##

Copyright (c) 2014, B3log Team (http://b3log.org)

Licensed under the [Apache License 2.0](https://github.com/b3log/wide/blob/master/LICENSE).

## Credits ##

* [golang](http://golang.org)
* [CodeMirror](https://github.com/marijnh/CodeMirror)
* [zTree](https://github.com/zTree/zTree_v3) 
* [gocode](https://github.com/nsf/gocode)
* [Gorilla](https://github.com/gorilla)
