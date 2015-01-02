/*
 * Copyright (c) 2015, B3log
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

var session = {
    init: function () {
        this._initWS();

        // 定时（30 秒）保存会话内容.
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

            request.currentFile = currentFile; // 当前编辑器
            request.fileTree = fileTree; // 文件树展开状态
            request.files = filse; // 编辑器打开状态

            $.ajax({
                type: 'POST',
                url: config.context + '/session/save',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (data) {
                }
            });
        }, 30000);
    },
    restore: function () {
        if (!config.latestSessionContent) {
            return;
        }

        var fileTree = config.latestSessionContent.FileTree,
                files = config.latestSessionContent.Files,
                currentFile = config.latestSessionContent.CurrentFile,
                id = "",
                nodesToOpen = [];

        var nodes = tree.fileTree.transformToArray(tree.fileTree.getNodes());

        for (var i = 0, ii = nodes.length; i < ii; i++) {
            // expand tree
            for (var j = 0, jj = fileTree.length; j < jj; j++) {
                if (nodes[i].path === fileTree[j]) {
                    // 当父节点都展开时，才展开该节点
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
                        // 设置状态
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
                id = nodes[i].tId;

                // FIXME: 上面的展开是异步进行的，所以执行到这里的时候可能还没有展开完，导致定位不了可视区域
                tree.fileTree.selectNode(nodes[i]);
                wide.curNode = nodes[i];
            }
        }

        // 处理编辑器打开顺序
        for (var m = 0, mm = files.length; m < mm; m++) {
            for (var n = 0, nn = nodesToOpen.length; n < nn; n++) {
                if (nodesToOpen[n].path === files[m]) {
                    tree.openFile(nodesToOpen[n]);
                    break;
                }
            }
        }

        // 设置当前编辑器
        editors.tabs.setCurrent(id);
        for (var c = 0, max = editors.data.length; c < max; c++) {
            if (id === editors.data[c].id) {
                wide.curEditor = editors.data[c].editor;
                break;
            }
        }
    },
    _initWS: function () {
        // 用于保持会话，如果该通道断开，则服务器端会销毁会话状态，回收相关资源.
        var sessionWS = new ReconnectingWebSocket(config.channel + '/session/ws?sid=' + config.wideSessionId);

        sessionWS.onopen = function () {
            console.log('[session onopen] connected');

            var dateFormat = function (time, fmt) {
                var date = new Date(time);
                var dateObj = {
                    "M+": date.getMonth() + 1, //月份 
                    "d+": date.getDate(), //日 
                    "h+": date.getHours(), //小时 
                    "m+": date.getMinutes(), //分 
                    "s+": date.getSeconds(), //秒 
                    "q+": Math.floor((date.getMonth() + 3) / 3), //季度 
                    "S": date.getMilliseconds() //毫秒 
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
            console.log('[session onmessage]' + e.data);
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
            console.log('[session onerror] ' + JSON.parse(e));
        };
    }
};