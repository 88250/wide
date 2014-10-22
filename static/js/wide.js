var wide = {
    curNode: undefined,
    curEditor: undefined,
    curProcessId: undefined, // 当前正在运行的进程 id（pid）
    bottomWindowTab: undefined,
    searchTab: undefined,
    _initDialog: function () {
        $(".dialog-prompt > input").keyup(function (event) {
            var $okBtn = $(this).closest(".dialog-main").find(".dialog-footer > button:eq(0)");
            if (event.which === 13 && !$okBtn.prop("disabled")) {
                $okBtn.click();
            }

            if ($.trim($(this).val()) === "") {
                $okBtn.prop("disabled", true);
            } else {
                $okBtn.prop("disabled", false);
            }
        });


        $("#dialogAlert").dialog({
            "modal": true,
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
            "modal": true,
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
                                    $(".edit-panel .tabs > div[data-index=" + wide.curNode.tId + "]").find(".ico-close").click();
                                    break;
                                }
                            }
                        } else {
                            for (var i = 0, ii = editors.data.length; i < ii; i++) {
                                if (tree._isParents(editors.data[i].id, wide.curNode.tId)) {
                                    $(".edit-panel .tabs > div[data-index=" + editors.data[i].id + "]").find(".ico-close").click();
                                    i--;
                                    ii--;
                                }
                            }
                        }
                    }
                });
            }
        });

        $("#dialogNewFilePrompt").dialog({
            "modal": true,
            "height": 52,
            "width": 260,
            "title": config.label.create_file,
            "okText": config.label.create,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogNewFilePrompt > input").val('').focus();
                $("#dialogNewFilePrompt").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
            },
            "ok": function () {
                var request = newWideRequest(),
                        name = $("#dialogNewFilePrompt > input").val();

                request.path = wide.curNode.path + config.pathSeparator + name;
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
            "modal": true,
            "height": 52,
            "width": 260,
            "title": config.label.create_dir,
            "okText": config.label.create,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogNewDirPrompt > input").val('').focus();
                $("#dialogNewDirPrompt").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
            },
            "ok": function () {
                var name = $("#dialogNewDirPrompt > input").val(),
                        request = newWideRequest();

                request.path = wide.curNode.path + config.pathSeparator + name;
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
            "modal": true,
            "height": 52,
            "width": 260,
            "title": config.label.goto_line,
            "okText": config.label.goto,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogGoLinePrompt > input").val('').focus();
                $("#dialogGoLinePrompt").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
            },
            "ok": function () {
                var line = parseInt($("#dialogGoLinePrompt > input").val());
                $("#dialogGoLinePrompt").dialog("close");
                wide.curEditor.setCursor(CodeMirror.Pos(line - 1, 0));
                wide.curEditor.focus();
            }
        });

        $("#dialogSearchForm > input:eq(0)").keyup(function (event) {
            var $okBtn = $(this).closest(".dialog-main").find(".dialog-footer > button:eq(0)");
            if (event.which === 13 && !$okBtn.prop("disabled")) {
                $okBtn.click();
            }

            if ($.trim($(this).val()) === "") {
                $okBtn.prop("disabled", true);
            } else {
                $okBtn.prop("disabled", false);
            }
        });

        $("#dialogSearchForm > input:eq(1)").keyup(function (event) {
            var $okBtn = $(this).closest(".dialog-main").find(".dialog-footer > button:eq(0)");
            if (event.which === 13 && !$okBtn.prop("disabled")) {
                $okBtn.click();
            }
        });

        $("#dialogSearchForm").dialog({
            "modal": true,
            "height": 62,
            "width": 260,
            "title": config.label.search,
            "okText": config.label.search,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogSearchForm > input:eq(0)").val('').focus();
                $("#dialogSearchForm > input:eq(1)").val('');
                $("#dialogSearchForm").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
            },
            "ok": function () {
                var request = newWideRequest();
                request.dir = wide.curNode.path;
                request.text = $("#dialogSearchForm > input:eq(0)").val();
                request.extension = $("#dialogSearchForm > input:eq(1)").val();

                $.ajax({
                    type: 'POST',
                    url: '/file/search/text',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (!data.succ) {
                            return;
                        }

                        $("#dialogSearchForm").dialog("close");
                        editors.appendSearch(data.founds, 'founds', request.text);
                    }
                });
            }
        });

        $("#dialogAbout").load('/about', function () {
            $("#dialogAbout").dialog({
                "modal": true,
                "height": 460,
                "width": 860,
                "title": config.label.about,
                "hideFooter": true
            });
        });
    },
    _initLayout: function () {
        var mainH = $(window).height() - $(".menu").height() - $(".footer").height() - 1,
                bottomH = Math.floor(mainH * 0.3);
        $(".content").height(mainH);
        $(".side .tabs-panel").height(mainH - 20);

        $(".bottom-window-group .output").height(bottomH - 27);
        $(".bottom-window-group > .tabs-panel > div > div").height(bottomH - 20);
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

            // TODO: 重构成 switch-case
            
            if ('run' === data.cmd) { // 正在运行
                wide.fillOutput($('.bottom-window-group .output').text() + data.output);
                wide.curProcessId = data.pid;
            } else if ('run-done' === data.cmd) { // 运行结束                
                wide.curProcessId = undefined;
                // 运行结束后修改 [构建&运行] 图标状态为可用状态
                $(".toolbars .ico-stop").removeClass("ico-stop")
                        .addClass("ico-buildrun").attr("title", config.label.build_n_run);
            } else if ('build' === data.cmd || 'go install' === data.cmd) {
                wide.fillOutput(data.output);

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
                wide.fillOutput($('.bottom-window-group .output').text() + data.output);
            } else if ('pre-build' === data.cmd) {
                wide.fillOutput(data.output);
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
    _initFooter: function () {
        $(".footer .cursor").dblclick(function () {
            $("#dialogGoLinePrompt").dialog("open");
        });
    },
    init: function () {
        this._initFooter();

        this._initWS();

        this._initBottomWindowGroup();

        // 点击隐藏弹出层
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

        // 刷新提示
        window.onbeforeunload = function () {
            if (editors.data.length > 0) {
                return config.label.confirm_save;
            }
        };

        // 禁止鼠标右键菜单
        document.oncontextmenu = function () {
            return false;
        };

        this._initDialog();

        this._initLayout();
    },
    _save: function () {
        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;
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
        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        // 格式化后会对文件进行保存
        this.fmt(currentPath, wide.curEditor);
    },
    saveAllFiles: function () {
        if ($(".menu li.save-all").hasClass("disabled")) {
            return false;
        }
        
        // TODO: 只保存未保存过的文件

        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            this.fmt(tree.fileTree.getNodeByTId(editors.data[i].id).path, editors.data[i].editor);
        }
    },
    closeAllFiles: function () {
        if ($(".menu li.close-all").hasClass("disabled")) {
            return false;
        }
        this.saveAllFiles();
        editors.data = [];
        tree.fileTree.cancelSelectedNode();
        wide.curNode = undefined;
        wide.curEditor = undefined;
        $(".toolbars").hide();

        $(".edit-panel .tabs, .edit-panel .tabs-panel").html('');
        menu.disabled(['save-all', 'close-all', 'run', 'go-get', 'go-install']);
    },
    exit: function () {
        var request = newWideRequest();

        $.ajax({
            type: 'POST',
            url: '/logout',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                if (data.succ) {
                    window.location.href = "/login";
                }
            }
        });
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
    fillOutput: function (data) {
        var $output = $('.bottom-window-group .output');
        $output.text(data);
        $output.scrollTop($output[0].scrollHeight);
    },
    // 构建.
    build: function () {
        wide.saveAllFiles();
        
        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;
        request.code = wide.curEditor.getValue();
        request.nextCmd = ""; // 只构建，无下一步操作

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                $('.bottom-window-group .output').text('');
            },
            success: function (data) {
            }
        });
    },
    // 构建并运行.
    run: function () {
        wide.saveAllFiles();
        
        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.run").hasClass("disabled")) {
            return false;
        }

        if ($(".toolbars .ico-stop").length === 1) {
            wide.stop();
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;
        request.code = wide.curEditor.getValue();
        request.nextCmd = "run";

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
        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.go-get").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;

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
        wide.saveAllFiles();
        
        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.go-install").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;
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
    fmt: function (path, curEditor) {
        var mode = curEditor.getOption("mode");

        var cursor = curEditor.getCursor();
        var scrollInfo = curEditor.getScrollInfo();

        var request = newWideRequest();
        request.file = path;
        request.code = curEditor.getValue();
        request.cursorLine = cursor.line;
        request.cursorCh = cursor.ch;

        switch (mode) {
            case "text/x-go": // 会保存文件
                $.ajax({
                    type: 'POST',
                    url: '/go/fmt',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (data.succ) {
                            curEditor.setValue(data.code);
                            curEditor.setCursor(cursor);
                            curEditor.scrollTo(null, scrollInfo.top);
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
                            curEditor.setValue(data.code);
                            curEditor.setCursor(cursor);
                            curEditor.scrollTo(null, scrollInfo.top);
                        }
                    }
                });

                break;
            case "application/json":
                try {
                    // 在客户端浏览器中进行 JSON 格式化
                    var json = JSON.parse(curEditor.getValue());
                    curEditor.setValue(JSON.stringify(json, "", "    "));
                    curEditor.setCursor(cursor);
                    curEditor.scrollTo(null, scrollInfo.top);

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
    },
    openAbout: function () {
        $("#dialogAbout").dialog("open");
    }
};

$(document).ready(function () {
    wide.init();
    tree.init();
    menu.init();
    hotkeys.init();
    notification.init();
    session.init();
    editors.init();
    windows.init();
});