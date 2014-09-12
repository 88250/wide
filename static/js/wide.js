var outputWS = new WebSocket(config.channel.output + '/output/ws');
outputWS.onopen = function() {
    console.log('[output onopen] connected');
};

outputWS.onmessage = function(e) {
    console.log('[output onmessage]' + e.data);
    var data = JSON.parse(e.data);

    if (goLintFound) {
        goLintFound = [];
    }

    if ('run' === data.cmd) {
        $('#output').text($('#output').text() + data.output);
    } else if ('build' === data.cmd || 'go install' === data.cmd) {
        $('#output').text(data.output);

        if (0 !== data.output.length) { // 说明编译有错误输出            
            for (var i = 0; i < data.lints.length; i++) {
                var lint = data.lints[i];

                goLintFound.push({from: CodeMirror.Pos(lint.lineNo, 0),
                    to: CodeMirror.Pos(lint.lineNo, 0),
                    message: lint.msg, severity: lint.severity});
            }
        }

        // 触发一次 gutter lint
        CodeMirror.signal(wide.curEditor, "change", wide.curEditor);
    } else if ('go get' === data.cmd || 'go install' === data.cmd) {
        $('#output').text($('#output').text() + data.output);
    }

    if ('build' === data.cmd) {
        if ('run' === data.nextCmd) {
            var request = {
                executable: data.executable
            };

            $.ajax({
                type: 'POST',
                url: '/run',
                data: JSON.stringify(request),
                dataType: "json",
                beforeSend: function(data) {
                    $('#output').text('');
                },
                success: function(data) {

                }
            });
        }
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
    _initLayout: function() {
        var mainH = $(window).height() - $(".menu").height() - $(".footer").height() - 2;
        $(".content, .ztree").height(mainH);

        $(".edit-panel").height(mainH - $(".output").height());
    },
    init: function() {
        this._initLayout();

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
        var request = {
            file: wide.curNode.path,
            code: wide.curEditor.getValue()
        };
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
        // TODO: save all
    },
    closeFile: function() {
        // TODO: save all
    },
    closeAllFiles: function() {
        // TODO: save all
    },
    exit: function() {
        // TODO: exit
    },
    run: function() {
        var request = {
            file: wide.curNode.path,
            code: wide.curEditor.getValue()
        };

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('#output').text('');
            },
            success: function(data) {
            }
        });
    },
    goget: function() {
        var request = {
            file: wide.curNode.path
        };

        $.ajax({
            type: 'POST',
            url: '/go/get',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('#output').text('');
            },
            success: function(data) {
            }
        });
    },
    goinstall: function() {
        var request = {
            file: wide.curNode.path,
            code: wide.curEditor.getValue()
        };

        $.ajax({
            type: 'POST',
            url: '/go/install',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('#output').text('');
            },
            success: function(data) {
            }
        });
    },
    fmt: function() {
        var path = wide.curNode.path;
        var mode = wide.curNode.mode;

        var request = {
            file: path,
            code: wide.curEditor.getValue(),
            cursorLine: wide.curEditor.getCursor().line,
            cursorCh: wide.curEditor.getCursor().ch
        };

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
                // TODO: XML/JSON 格式化处理
                break;
        }


    }
};

$(document).ready(function() {
    wide.init();
    tree.init();
    menu.init();
});