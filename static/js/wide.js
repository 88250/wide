var outputWS = new WebSocket(config.channel.output + '/output/ws?sid=' + config.wideSessionId);
outputWS.onopen = function() {
    console.log('[output onopen] connected');
};

outputWS.onmessage = function(e) {
    console.log('[output onmessage]' + e.data);
    var data = JSON.parse(e.data);

    if (goLintFound) {
        goLintFound = [];
    }

    if ('run' === data.nextCmd) {
        var request = newWideRequest();
        request.executable = data.executable;

        $.ajax({
            type: 'POST',
            url: '/run',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('.bottom-window-group .output').text('');
            },
            success: function(data) {

            }
        });
    }

    if ('run' === data.cmd) { // 正在运行
        $('.bottom-window-group .output').text($('.bottom-window-group .output').text() + data.output);
    } else if ('run-done' === data.cmd) { // 运行结束
        // TODO: 运行结束后修改 [构建&运行] 图标状态为可用状态
    } else if ('build' === data.cmd || 'go install' === data.cmd) {
        $('.bottom-window-group .output').text(data.output);

        if (0 !== data.output.length) { // 说明编译有错误输出            
            for (var i = 0; i < data.lints.length; i++) {
                var lint = data.lints[i];

                goLintFound.push({from: CodeMirror.Pos(lint.lineNo, 0),
                    to: CodeMirror.Pos(lint.lineNo, 0),
                    message: lint.msg, severity: lint.severity});
            }

            // TODO: 修改 [构建&运行] 图标状态为可用状态
        }

        // 触发一次 gutter lint
        CodeMirror.signal(wide.curEditor, "change", wide.curEditor);
    } else if ('go get' === data.cmd || 'go install' === data.cmd) {
        $('.bottom-window-group .output').text($('.bottom-window-group .output').text() + data.output);
    }
};
outputWS.onclose = function(e) {
    console.log('[output onclose] disconnected (' + e.code + ')');
    delete outputWS;
};
outputWS.onerror = function(e) {
    console.log('[output onerror] ' + e);
};

var wide = {
    curNode: undefined,
    curEditor: undefined,
    bottomWindowTab: undefined,
    _initLayout: function() {
        var mainH = $(window).height() - $(".menu").height() - $(".footer").height() - 2;
        $(".content, .ztree").height(mainH);

        $(".edit-panel").height(mainH - $(".bottom-window-group").height());
    },
    _initBottomWindowGroup: function() {
        this.bottomWindowTab = new Tabs({
            id: ".bottom-window-group"
        });
    },
    init: function() {
        this._initLayout();

        this._initBottomWindowGroup();

        $("body").bind("mousedown", function(event) {
            if (!(event.target.id === "dirRMenu" || $(event.target).closest("#dirRMenu").length > 0)) {
                $("#dirRMenu").hide();
            }

            if (!(event.target.id === "fileRMenu" || $(event.target).closest("#fileRMenu").length > 0)) {
                $("#fileRMenu").hide();
            }

            if (!($(event.target).closest(".frame").length > 0 || event.target.className === "frame")) {
                $(".frame").hide();
                $(".menu > ul > li > a, .menu > ul> li > span").unbind("mouseover");
                menu.subMenu();
            }
        });

    },
    saveFile: function() {
        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");
        request.code = wide.curEditor.getValue();

        $.ajax({
            type: 'POST',
            url: '/file/save',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
            }
        });
    },
    saveAllFiles: function() {
        // TODO: save all files
    },
    closeFile: function() {
        // TODO: close file
    },
    closeAllFiles: function() {
        // TODO: close all files
    },
    exit: function() {
        // TODO: exit
    },
    // 构建 & 运行.
    run: function() {
        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");
        request.code = wide.curEditor.getValue();

        // TODO: 修改 [构建&运行] 图标状态为不可用状态

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('.bottom-window-group .output').text('');
            },
            success: function(data) {
            }
        });
    },
    goget: function() {
        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");

        $.ajax({
            type: 'POST',
            url: '/go/get',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('.bottom-window-group .output').text('');
            },
            success: function(data) {
            }
        });
    },
    goinstall: function() {
        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");
        request.code = wide.curEditor.getValue();

        $.ajax({
            type: 'POST',
            url: '/go/install',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('.bottom-window-group .output').text('');
            },
            success: function(data) {
            }
        });
    },
    fmt: function() {
        var path = $(".edit-header .current span:eq(0)").attr("title");
        var mode = wide.curNode.mode;

        var request = newWideRequest();
        request.file = path;
        request.code = wide.curEditor.getValue();
        request.cursorLine = wide.curEditor.getCursor().line;
        request.cursorCh = wide.curEditor.getCursor().ch;

        switch (mode) {
            case "text/x-go":
                $.ajax({
                    type: 'POST',
                    url: '/go/fmt',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function(data) {
                        if (data.succ) {
                            wide.curEditor.setValue(data.code);
                        }
                    }
                });

                break;
            case "text/html":
                $.ajax({
                    type: 'POST',
                    url: '/html/fmt',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function(data) {
                        if (data.succ) {
                            wide.curEditor.setValue(data.code);
                        }
                    }
                });

                break;
            case "application/json":
                try {
                    // 在客户端浏览器中进行 JSON 格式化
                    var json = JSON.parse(wide.curEditor.getValue());
                    wide.curEditor.setValue(JSON.stringify(json, "", "    "));

                    this.save();
                } catch (e) {
                    delete e;
                }

                break;
            default :
                // TODO: XML 格式化处理
                break;
        }
    }
};

$(document).ready(function() {
    wide.init();
    tree.init();
    menu.init();
    hotkeys.init();
});