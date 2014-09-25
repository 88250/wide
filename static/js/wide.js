var wide = {
    curNode: undefined,
    curEditor: undefined,
    curProcessId: undefined, // 当前正在运行的进程 id（pid）
    bottomWindowTab: undefined,
    _initDialog: function () {
        $("#dialogAlert").dialog({
            "height": 26,
            "width": 260,
            "title": config.label.tip,
            "hiddenOk": true,
            "cancelText": config.label.confirm,
            "afterOpen": function (msg) {
                $("#dialogAlert").html(msg);
            }
        });

        $("#dialogRemoveConfirm").dialog({
            "height": 26,
            "width": 260,
            "title": config.label.delete,
            "okText": config.label.delete,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogRemoveConfirm > b").html('"' + wide.curNode.name + '"');
            },
            "ok": function () {
                var request = newWideRequest();
                request.path = wide.curNode.path;

                $.ajax({
                    type: 'POST',
                    url: '/file/remove',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (!data.succ) {
                            return false;
                        }

                        $("#dialogRemoveConfirm").dialog("close");
                        tree.fileTree.removeNode(wide.curNode);

                        if ("ico-ztree-dir " !== wide.curNode.iconSkin) {
                            // 是文件的话，查看 editor 中是否被打开，如打开则移除
                            for (var i = 0, ii = editors.data.length; i < ii; i++) {
                                if (editors.data[i].id === wide.curNode.tId) {
                                    $(".edit-header .tabs > div[data-index=" + wide.curNode.tId + "]").find(".ico-close").click();
                                    break;
                                }
                            }
                        } else {
                            for (var i = 0, ii = editors.data.length; i < ii; i++) {
                                if (tree._isParents(editors.data[i].id, wide.curNode.tId)) {
                                    $(".edit-header .tabs > div[data-index=" + editors.data[i].id + "]").find(".ico-close").click();
                                    i--;
                                    ii--;
                                }
                            }
                        }
                    }
                });
            }
        });

        $(".dialog-prompt > input").keydown(function (event) {
            $(".dialog-prompt > .tip").text('');

            if (event.which === 13) {
                $(this).closest(".dialog-main").find(".dialog-footer > button:eq(0)").click();
            }
        });

        $("#dialogNewFilePrompt").dialog({
            "height": 52,
            "width": 260,
            "title": config.label.create_file,
            "okText": config.label.create,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogNewFilePrompt > input").val('').focus();
            },
            "ok": function () {
                var request = newWideRequest(),
                        name = $("#dialogNewFilePrompt > input").val();
                if ($.trim(name) === "") {
                    $("#dialogNewFilePrompt > .tip").text(config.label.input_no_empty);
                    return false;
                }

                request.path = wide.curNode.path + '\\' + name;
                request.fileType = "f";

                $.ajax({
                    type: 'POST',
                    url: '/file/new',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (!data.succ) {
                            return false;
                        }
                        $("#dialogNewFilePrompt").dialog("close");
                        var suffix = name.split(".")[1],
                                iconSkin = "ico-ztree-other ";
                        switch (suffix) {
                            case "html", "htm":
                                iconSkin = "ico-ztree-html ";
                                break;
                            case "go":
                                iconSkin = "ico-ztree-go ";
                                break;
                            case "css":
                                iconSkin = "ico-ztree-css ";
                                break;
                            case "txt":
                                iconSkin = "ico-ztree-text ";
                                break;
                            case "sql":
                                iconSkin = "ico-ztree-sql ";
                                break;
                            case "properties":
                                iconSkin = "ico-ztree-pro ";
                                break;
                            case "md":
                                iconSkin = "ico-ztree-md ";
                                break;
                            case "js", "json":
                                iconSkin = "ico-ztree-js ";
                                break;
                            case "xml":
                                iconSkin = "ico-ztree-xml ";
                                break;
                            case "jpg", "jpeg", "bmp", "gif", "png", "svg", "ico":
                                iconSkin = "ico-ztree-img ";
                                break;
                        }

                        tree.fileTree.addNodes(wide.curNode, [{
                                "name": name,
                                "iconSkin": iconSkin,
                                "path": request.path,
                                "mode": data.mode
                            }]);
                    }
                });
            }
        });

        $("#dialogNewDirPrompt").dialog({
            "height": 52,
            "width": 260,
            "title": config.label.create_dir,
            "okText": config.label.create,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogNewDirPrompt > input").val('').focus();
            },
            "ok": function () {
                var name = $("#dialogNewDirPrompt > input").val(),
                        request = newWideRequest();
                if ($.trim(name) === "") {
                    $("#dialogNewDirPrompt > .tip").text(config.label.input_no_empty);
                    return false;
                }
                request.path = wide.curNode.path + '\\' + name;
                request.fileType = "d";

                $.ajax({
                    type: 'POST',
                    url: '/file/new',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (!data.succ) {
                            return false;
                        }

                        $("#dialogNewDirPrompt").dialog("close");

                        tree.fileTree.addNodes(wide.curNode, [{
                                "name": name,
                                "iconSkin": "ico-ztree-dir ",
                                "path": request.path
                            }]);
                    }
                });
            }
        });

        $("#dialogGoLinePrompt").dialog({
            "height": 52,
            "width": 260,
            "title": config.label.goto_line,
            "okText": config.label.goto,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogGoLinePrompt > input").val('').focus();
            },
            "ok": function () {
                var line = parseInt($("#dialogGoLinePrompt > input").val());
                if ($.trim(line) === "") {
                    $("#dialogGoLinePrompt > .tip").text(config.label.input_no_empty);
                    return false;
                }
                $("#dialogGoLinePrompt").dialog("close");
                wide.curEditor.setCursor(CodeMirror.Pos(line - 1, 0));
                wide.curEditor.focus();
            }
        });
    },
    _initLayout: function () {
        var mainH = $(window).height() - $(".menu").height() - $(".footer").height() - 2,
                bottomH = Math.floor(mainH * 0.3);
        $(".content").height(mainH);
        $(".side .tabs-panel").height(mainH - 20);

        $(".bottom-window-group .output, notification").height(bottomH - 24);
        $(".bottom-window-group .notification, .bottom-window-group .search").height(bottomH - 20);
    },
    _initBottomWindowGroup: function () {
        this.bottomWindowTab = new Tabs({
            id: ".bottom-window-group",
            clickAfter: function (id) {
                this._$tabsPanel.find("." + id).focus();
            }
        });
    },
    _initWS: function () {
        var outputWS = new WebSocket(config.channel.output + '/output/ws?sid=' + config.wideSessionId);
        outputWS.onopen = function () {
            console.log('[output onopen] connected');
        };

        outputWS.onmessage = function (e) {
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
                    beforeSend: function (data) {
                        $('.bottom-window-group .output').text('');
                    },
                    success: function (data) {

                    }
                });
            }

            if ('run' === data.cmd) { // 正在运行
                $('.bottom-window-group .output').text($('.bottom-window-group .output').text() + data.output);

                wide.curProcessId = data.pid;
            } else if ('run-done' === data.cmd) { // 运行结束                
                wide.curProcessId = undefined;
                // 运行结束后修改 [构建&运行] 图标状态为可用状态
                $(".toolbars .ico-stop").removeClass("ico-stop")
                        .addClass("ico-buildrun").attr("title", config.label.build_n_run);
            } else if ('build' === data.cmd || 'go install' === data.cmd) {
                $('.bottom-window-group .output').text(data.output);

                if (0 !== data.output.length) { // 说明编译有错误输出            
                    for (var i = 0; i < data.lints.length; i++) {
                        var lint = data.lints[i];

                        goLintFound.push({from: CodeMirror.Pos(lint.lineNo, 0),
                            to: CodeMirror.Pos(lint.lineNo, 0),
                            message: lint.msg, severity: lint.severity});
                    }

                    $(".toolbars .ico-stop").removeClass("ico-stop")
                            .addClass("ico-buildrun").attr("title", config.label.build_n_run);
                }

                // 触发一次 gutter lint
                CodeMirror.signal(wide.curEditor, "change", wide.curEditor);
            } else if ('go get' === data.cmd || 'go install' === data.cmd) {
                $('.bottom-window-group .output').text($('.bottom-window-group .output').text() + data.output);
            }
        };
        outputWS.onclose = function (e) {
            console.log('[output onclose] disconnected (' + e.code + ')');
            delete outputWS;
        };
        outputWS.onerror = function (e) {
            console.log('[output onerror] ' + e);
        };
    },
    _initFullscreen: function () {
        $(".bottom-window-group .tabs").dblclick(function () {
            var $it = $(".bottom-window-group");
            if ($it.hasClass("bottom-window-group-fullscreen")) {
                $(".bottom-window-group").removeClass("bottom-window-group-fullscreen");

                var bottomH = $(".bottom-window-group").height();

                $(".bottom-window-group .output, notification").height(bottomH - 24);
                $(".bottom-window-group .notification, .bottom-window-group .search").height(bottomH - 20);
            } else {
                var bottomH = $(".content, .ztree").height();
                $(".bottom-window-group .output, notification").height(bottomH - 23);
                $(".bottom-window-group .notification, .bottom-window-group .search").height(bottomH - 19);

                $(".bottom-window-group").addClass("bottom-window-group-fullscreen");
            }
        });

        $(".side .tabs").dblclick(function () {
            var $it = $(".side");
            if ($it.hasClass("side-fullscreen")) {
                $it.removeClass("side-fullscreen");
            } else {
                $it.addClass("side-fullscreen");
            }
        });
    },
    init: function () {
        this._initFullscreen();

        this._initWS();

        this._initBottomWindowGroup();

        $("body").bind("mousedown", function (event) {
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

        this._initDialog();

        this._initLayout();
    },
    _save: function () {
        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");
        request.code = wide.curEditor.getValue();

        $.ajax({
            type: 'POST',
            url: '/file/save',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
            }
        });
    },
    saveFile: function () {
        // 格式化后会对文件进行保存
        this.fmt();
    },
    saveAllFiles: function () {
        // TODO: save all open files
        for (var i = 0, ii = editors.data.length; i < ii; i++) {

        }
        console.log("TODO: save all files");
    },
    closeFile: function () {
        // TODO: close file
    },
    closeAllFiles: function () {
        // TODO: close all files
    },
    exit: function () {
        // TODO: exit
    },
    stop: function () {
        if ($(".toolbars .ico-buildrun").length === 1) {
            wide.run();
            return false;
        }

        var request = newWideRequest();
        request.pid = wide.curProcessId;

        $.ajax({
            type: 'POST',
            url: '/stop',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                // $('.bottom-window-group .output').text('');
            },
            success: function (data) {
                $(".toolbars .ico-stop").removeClass("ico-stop")
                        .addClass("ico-buildrun").attr("title", config.label.build_n_run);
            }
        });
    },
    // 构建 & 运行.
    run: function () {
        if ($(".toolbars .ico-stop").length === 1) {
            wide.stop();
            return false;
        }

        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");
        request.code = wide.curEditor.getValue();

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                $('.bottom-window-group .output').text('');
            },
            success: function (data) {
                $(".toolbars .ico-buildrun").addClass("ico-stop")
                        .removeClass("ico-buildrun").attr("title", config.label.stop);
            }
        });
    },
    goget: function () {
        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");

        $.ajax({
            type: 'POST',
            url: '/go/get',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                $('.bottom-window-group .output').text('');
            },
            success: function (data) {
            }
        });
    },
    goinstall: function () {
        var request = newWideRequest();
        request.file = $(".edit-header .current span:eq(0)").attr("title");
        request.code = wide.curEditor.getValue();

        $.ajax({
            type: 'POST',
            url: '/go/install',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                $('.bottom-window-group .output').text('');
            },
            success: function (data) {
            }
        });
    },
    fmt: function () {
        var path = $(".edit-header .current span:eq(0)").attr("title");
        var mode = wide.curEditor.getOption("mode");

        var request = newWideRequest();
        request.file = path;
        request.code = wide.curEditor.getValue();
        request.cursorLine = wide.curEditor.getCursor().line;
        request.cursorCh = wide.curEditor.getCursor().ch;

        switch (mode) {
            case "text/x-go": // 会保存文件
                $.ajax({
                    type: 'POST',
                    url: '/go/fmt',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (data.succ) {
                            wide.curEditor.setValue(data.code);
                        }
                    }
                });

                break;
            case "text/html": // 会保存文件
                $.ajax({
                    type: 'POST',
                    url: '/html/fmt',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
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

                    wide._save();
                } catch (e) {
                    delete e;
                }

                break;
            default :
                // TODO: XML 格式化处理
                // 所有文件格式化后都需要进行保存
                wide._save();
                break;
        }
    }
};

$(document).ready(function () {
    wide.init();
    tree.init();
    menu.init();
    hotkeys.init();
    notification.init();
    session.init();
});