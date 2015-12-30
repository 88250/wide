/*
 * Copyright (c) 2014-2016, b3log.org & hacpai.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
 * @file session.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @version 1.1.0.1, Dec 8, 2015
 */
var session = {
    init: function () {
        this._initWS();

        var getLayoutState = function (paneState) {
            var state = 'normal';
            if (paneState.isClosed) {
                state = 'min';
            } else if (paneState.size >= $('body').width()) {
                state = 'max';
            }

            return state;
        };

        // save session content every 30 seconds
        setInterval(function () {
            var request = newWideRequest(),
                    filse = [],
                    fileTree = [],
                    currentId = editors.getCurrentId(),
                    currentFile = currentId ? editors.getCurrentPath() : "";

            editors.tabs.obj._$tabs.find("div").each(function () {
                var $it = $(this);
                if ($it.find("span:eq(0)").attr("title") !== config.label.start_page) {
                    filse.push($it.find("span:eq(0)").attr("title"));
                }
            });

            fileTree = tree.getOpenPaths();

            request.currentFile = currentFile; // current editor file
            request.fileTree = fileTree; // file tree expansion state
            request.files = filse; // editor tabs


            request.layout = {
                "side": {
                    "size": windows.outerLayout.west.state.size,
                    "state": getLayoutState(windows.outerLayout.west.state)
                },
                "sideRight": {
                    "size": windows.innerLayout.east.state.size,
                    "state": getLayoutState(windows.innerLayout.east.state)
                },
                "bottom": {
                    "size": windows.innerLayout.south.state.size,
                    "state": getLayoutState(windows.innerLayout.south.state)
                }
            };

            $.ajax({
                type: 'POST',
                url: config.context + '/session/save',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (result) {
                }
            });
        }, 30000);
    },
    restore: function () {
        if (!config.latestSessionContent) {
            return;
        }

        var fileTree = config.latestSessionContent.fileTree,
                files = config.latestSessionContent.files,
                currentFile = config.latestSessionContent.currentFile,
                id = "",
                nodesToOpen = [];

        var nodes = tree.fileTree.transformToArray(tree.fileTree.getNodes());

        for (var i = 0, ii = nodes.length; i < ii; i++) {
            // expand tree
            for (var j = 0, jj = fileTree.length; j < jj; j++) {
                if (nodes[i].path === fileTree[j]) {
                    // expand this node only if its parents are open
                    var parents = tree.getAllParents(tree.fileTree.getNodeByTId(nodes[i].tId)),
                            isOpen = true;
                    for (var l = 0, max = parents.length; l < max; l++) {
                        if (parents[l].open === false) {
                            isOpen = false;
                        }
                    }
                    if (isOpen) {
                        tree.fileTree.expandNode(nodes[i], true, false, true);
                    } else {
                        // flag it is open
                        nodes[i].open = true;
                    }
                    break;
                }
            }

            // open editors
            for (var k = 0, kk = files.length; k < kk; k++) {
                if (nodes[i].path === files[k]) {
                    nodesToOpen.push(nodes[i]);
                    break;
                }
            }

            if (nodes[i].path === currentFile) {
                id = nodes[i].path;

                // FIXME: 上面的展开是异步进行的，所以执行到这里的时候可能还没有展开完，导致定位不了可视区域
                tree.fileTree.selectNode(nodes[i]);
                wide.curNode = nodes[i];
            }
        }

        // handle the open sequence of editors
        for (var m = 0, mm = files.length; m < mm; m++) {
            for (var n = 0, nn = nodesToOpen.length; n < nn; n++) {
                if (nodesToOpen[n].path === files[m]) {
                    tree.openFile(nodesToOpen[n]);
                    break;
                }
            }
        }

        // set the current editor
        editors.tabs.setCurrent(id);
        for (var c = 0, max = editors.data.length; c < max; c++) {
            if (id === editors.data[c].id) {
                wide.curEditor = editors.data[c].editor;
                break;
            }
        }        
    },
    _initWS: function () {
        // Used for session retention, server will release all resources of the session if this channel closed
        var sessionWS = new ReconnectingWebSocket(config.channel + '/session/ws?sid=' + config.wideSessionId);

        sessionWS.onopen = function () {
            console.log('[session onopen] connected');

            var dateFormat = function (time, fmt) {
                var date = new Date(time);
                var dateObj = {
                    "M+": date.getMonth() + 1,
                    "d+": date.getDate(),
                    "h+": date.getHours(),
                    "m+": date.getMinutes(),
                    "s+": date.getSeconds(),
                    "q+": Math.floor((date.getMonth() + 3) / 3),
                    "S": date.getMilliseconds()
                };
                if (/(y+)/.test(fmt))
                    fmt = fmt.replace(RegExp.$1, (date.getFullYear() + "").substr(4 - RegExp.$1.length));
                for (var k in dateObj)
                    if (new RegExp("(" + k + ")").test(fmt)) {
                        fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1)
                                ? (dateObj[k]) : (("00" + dateObj[k]).substr(("" + dateObj[k]).length)));
                    }
                return fmt;
            };

            var data = {type: "Network", severity: "INFO",
                message: "Connected to server [sid=" + config.wideSessionId + "], " + dateFormat(new Date().getTime(), 'yyyy-MM-dd hh:mm:ss')},
            $notification = $('.bottom-window-group .notification > table'),
                    notificationHTML = '';

            notificationHTML += '<tr><td class="severity">' + data.severity
                    + '</td><td class="message">' + data.message
                    + '</td><td class="type">' + data.type + '</td></tr>';
            $notification.append(notificationHTML);
        };

        sessionWS.onmessage = function (e) {
            var data = JSON.parse(e.data);
            switch (data.cmd) {
                case 'create-file':
                    var node = tree.fileTree.getNodeByTId(tree.getTIdByPath(data.dir)),
                            name = data.path.replace(data.dir + '/', ''),
                            mode = CodeMirror.findModeByFileName(name),
                            iconSkin = wide.getClassBySuffix(name.split(".")[1]);

                    if (data.type && data.type === 'f') {
                        tree.fileTree.addNodes(node, [{
                                "id": data.path,
                                "name": name,
                                "iconSkin": iconSkin,
                                "path": data.path,
                                "mode": mode,
                                "removable": true,
                                "creatable": true
                            }]);

                    } else {
                        tree.fileTree.addNodes(node, [{
                                "id": data.path,
                                "name": name,
                                "iconSkin": "ico-ztree-dir ",
                                "path": data.path,
                                "removable": true,
                                "creatable": true,
                                "isParent": true
                            }]);
                    }
                    break;
                case 'remove-file':
                case 'rename-file':
                    var node = tree.fileTree.getNodeByTId(tree.getTIdByPath(data.path));
                    tree.fileTree.removeNode(node);

                    var nodes = tree.fileTree.transformToArray(node);
                    for (var i = 0, ii = nodes.length; i < ii; i++) {
                        editors.tabs.del(nodes[i].path);
                    }

                    break;
            }
        };
        sessionWS.onclose = function (e) {
            console.log('[session onclose] disconnected (' + e.code + ')');

            var data = {type: "Network", severity: "ERROR",
                message: "Disconnected from server, trying to reconnect it [sid=" + config.wideSessionId + "]"},
            $notification = $('.bottom-window-group .notification > table'),
                    notificationHTML = '';

            notificationHTML += '<tr><td class="severity">' + data.severity
                    + '</td><td class="message">' + data.message
                    + '</td><td class="type">' + data.type + '</td></tr>';
            $notification.append(notificationHTML);

            $(".notification-count").show();
        };
        sessionWS.onerror = function (e) {
            console.log('[session onerror]');
        };
    }
};