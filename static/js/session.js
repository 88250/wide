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
                url: '/session/save',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (data) {
                }
            });
        }, 30000);
    },
    restore: function () {
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
                    tree.fileTree.expandNode(nodes[i], true, false, false);
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

        editors.tabs.setCurrent(id);
    },
    _initWS: function () {
        // 用于保持会话，如果该通道断开，则服务器端会销毁会话状态，回收相关资源.
        var sessionWS = new ReconnectingWebSocket(config.channel.session + '/session/ws?sid=' + config.wideSessionId);

        sessionWS.onopen = function () {
            console.log('[session onopen] connected');
        };

        sessionWS.onmessage = function (e) {
            console.log('[session onmessage]' + e.data);
            var data = JSON.parse(e.data);

        };
        sessionWS.onclose = function (e) {
            console.log('[session onclose] disconnected (' + e.code + ')');
            delete sessionWS;
        };
        sessionWS.onerror = function (e) {
            console.log('[session onerror] ' + JSON.parse(e));
        };
    }
};