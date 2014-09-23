var session = {
    init: function () {
        this._initWS();

        // 定时（30 秒）保存会话内容.
        setInterval(function () {
            var request = newWideRequest(),
                    filse = [],
                    fileTree = [],
                    currentFile = "";

            editors.tabs.obj._$tabs.find("div").each(function () {
                var $it = $(this);
                if ($it.hasClass("current")) {
                    currentFile = $it.find("span:eq(0)").attr("title");
                }

                filse.push($it.find("span:eq(0)").attr("title"));
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
        }, 5000);
    },
    _initWS: function () {
        // 用于保持会话，如果该通道断开，则服务器端会销毁会话状态，回收相关资源.
        var sessionWS = new WebSocket(config.channel.session + '/session/ws?sid=' + config.wideSessionId);

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