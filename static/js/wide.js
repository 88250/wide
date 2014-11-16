/* 
 * Copyright (c) 2014, B3log
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

var wide = {
    curNode: undefined,
    curEditor: undefined,
    curProcessId: undefined, // 当前正在运行的进程 id（pid）
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
                            $("#dialogRemoveConfirm").dialog("close");
                            bottomGroup.tabs.setCurrent("notification");
                            windows.flowBottom();
                            $(".bottom-window-group .notification").focus();
                            return false;
                        }

                        $("#dialogRemoveConfirm").dialog("close");
                        tree.fileTree.removeNode(wide.curNode);

                        if (!tree.isDir()) {
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

        $("#dialogRenamePrompt").dialog({
            "modal": true,
            "height": 52,
            "width": 260,
            "title": config.label.rename,
            "okText": config.label.rename,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                var index = wide.curNode.name.lastIndexOf(".");
                $("#dialogRenamePrompt > input").val(wide.curNode.name.substring(0, index)).focus();
                // TODO: 全选
                $("#dialogRenamePrompt").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
            },
            "ok": function () {
                var name = $("#dialogRenamePrompt > input").val(),
                        request = newWideRequest();

                request.oldPath = wide.curNode.path;

                var pathIndex = wide.curNode.path.lastIndexOf(config.pathSeparator),
                        nameIndex = wide.curNode.name.lastIndexOf("."),
                        ext = wide.curNode.name.substring(nameIndex, wide.curNode.name.length);
                request.newPath = wide.curNode.path.substring(0, pathIndex) + config.pathSeparator
                        + name + ext;

                $.ajax({
                    type: 'POST',
                    url: '/file/rename',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (!data.succ) {
                            $("#dialogRenamePrompt").dialog("close");
                            bottomGroup.tabs.setCurrent("notification");
                            windows.flowBottom();
                            $(".bottom-window-group .notification").focus();
                            return false;
                        }

                        $("#dialogRenamePrompt").dialog("close");

                        // update tree node
                        wide.curNode.name = name + ext;
                        wide.curNode.title = request.newPath;
                        wide.curNode.path = request.newPath;
                        tree.fileTree.updateNode(wide.curNode);

                        // update open editor tab name
                        var $currentSpan = $(".edit-panel .tabs > div[data-index=" + wide.curNode.tId + "] > span:eq(0)");
                        $currentSpan.attr("title", request.newPath);
                        $currentSpan.html($currentSpan.find("span").html() + wide.curNode.name);
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
                            $("#dialogNewFilePrompt").dialog("close");
                            bottomGroup.tabs.setCurrent("notification");
                            windows.flowBottom();
                            $(".bottom-window-group .notification").focus();
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
                                "mode": data.mode,
                                "removable": true,
                                "creatable": true
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
                            $("#dialogNewDirPrompt").dialog("close");
                            bottomGroup.tabs.setCurrent("notification");
                            windows.flowBottom();
                            $(".bottom-window-group .notification").focus();
                            return false;
                        }

                        $("#dialogNewDirPrompt").dialog("close");

                        tree.fileTree.addNodes(wide.curNode, [{
                                "name": name,
                                "iconSkin": "ico-ztree-dir ",
                                "path": request.path,
                                "removable": true,
                                "creatable": true
                            }]);
                    }
                });
            }
        });

        $("#dialogGoFilePrompt").dialog({
            "modal": true,
            "height": 300,
            "width": 660,
            "title": config.label.goto_file,
            "okText": config.label.go,
            "cancelText": config.label.cancel,
            "afterInit": function () {
                hotkeys.bindList($("#dialogGoFilePrompt > input"), $("#dialogGoFilePrompt > .list"), function ($selected) {
                    var tId = tree.getTIdByPath($selected.text());
                    tree.openFile(tree.fileTree.getNodeByTId(tId));
                    $("#dialogGoFilePrompt").dialog("close");
                });

                $("#dialogGoFilePrompt > input").keydown(function () {
                    var name = $("#dialogGoFilePrompt > input").val();

                    var request = newWideRequest();
                    request.path = '';
                    request.name = '*' + name + '*';
                    if (wide.curNode) {
                        request.path = wide.curNode.path;
                    }

                    $.ajax({
                        type: 'POST',
                        url: '/file/find/name',
                        data: JSON.stringify(request),
                        dataType: "json",
                        success: function (data) {
                            if (!data.succ) {
                                return;
                            }

                            var goFileHTML = '';
                            for (var i = 0, max = data.founds.length; i < max; i++) {
                                if (i === 0) {
                                    goFileHTML += '<li class="selected">' + data.founds[i].path + '</li>';
                                } else {
                                    goFileHTML += '<li>' + data.founds[i].path + '</li>';
                                }
                            }

                            $("#dialogGoFilePrompt > ul").html(goFileHTML);
                        }
                    });
                });
            },
            "afterOpen": function () {
                $("#dialogGoFilePrompt > input").val('').focus();
                $("#dialogGoFilePrompt").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
                $("#dialogGoFilePrompt .list").html('').data("index", 0);
            },
            "ok": function () {
                var tId = tree.getTIdByPath($("#dialogGoFilePrompt .selected").text());
                tree.openFile(tree.fileTree.getNodeByTId(tId));
                $("#dialogGoFilePrompt").dialog("close");
            }
        });

        $("#dialogGoLinePrompt").dialog({
            "modal": true,
            "height": 52,
            "width": 260,
            "title": config.label.goto_line,
            "okText": config.label.go,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogGoLinePrompt > input").val('').focus();
                $("#dialogGoLinePrompt").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
            },
            "ok": function () {
                var line = parseInt($("#dialogGoLinePrompt > input").val()) - 1;
                $("#dialogGoLinePrompt").dialog("close");

                var editor = wide.curEditor;
                var cursor = editor.getCursor();

                editor.setCursor(CodeMirror.Pos(line, cursor.ch));

                var half = Math.floor(editor.getScrollInfo().clientHeight / editor.defaultTextHeight() / 2);
                var cursorCoords = editor.cursorCoords({line: line - half, ch: cursor.ch}, "local");
                editor.scrollTo(0, cursorCoords.top);

                editor.focus();
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
                "width": 800,
                "title": config.label.about,
                "hideFooter": true,
                "afterOpen": function () {
                    $.ajax({
                        url: "http://rhythm.b3log.org/version/wide/latest",
                        type: "GET",
                        dataType: "jsonp",
                        jsonp: "callback",
                        success: function (data, textStatus) {
                            if ($("#dialogAbout .version").text() === data.wideVersion) {
                                $(".upgrade").text(config.label.uptodate);
                            } else {
                                $(".upgrade").html(config.label.new_version_available + config.label.colon
                                        + "<a href='" + data.wideDownload
                                        + "' target='_blank'>" + data.wideVersion + "</a>");
                            }
                        }
                    });
                }
            });
        });
    },
    openPreference: function () {
        $("#dialogPreference").dialog("open");
    },
    _initPreference: function () {
        $("#dialogPreference").load('/preference', function () {
            $("#dialogPreference").dialog({
                "modal": true,
                "height": 460,
                "width": 800,
                "title": config.label.perference,
                "ok": function () {
                    var request = newWideRequest();
                    request.executable = data.executable;

                    $.ajax({
                        type: 'POST',
                        url: '/preference',
                        data: JSON.stringify(request),
                        success: function (data, textStatus, jqXHR) {
                            
                        }
                    });
                }
            });

            new Tabs({
                id: ".preference"
            });
        });
    },
    _initLayout: function () {
        var mainH = $(window).height() - $(".menu").height() - $(".footer").height(),
                bottomH = Math.floor(mainH * 0.3);
        // 减小初始化界面抖动
        $(".content").height(mainH).css("position", "relative");
        $(".side .tabs-panel").height(mainH - 20);

        $(".bottom-window-group > .tabs-panel > div > div").height(bottomH - 20);
    },
    _initWS: function () {
        var outputWS = new ReconnectingWebSocket(config.channel.output + '/output/ws?sid=' + config.wideSessionId);
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
                    dataType: "json"
                });
            }

            switch (data.cmd) {
                case 'run': // 正在运行
                    bottomGroup.fillOutput($('.bottom-window-group .output > div').html() + data.output);
                    wide.curProcessId = data.pid;

                    break;
                case 'run-done': // 运行结束  
                    wide.curProcessId = undefined;
                    // 运行结束后修改 [构建&运行] 图标状态为可用状态
                    $(".toolbars .ico-stop").removeClass("ico-stop")
                            .addClass("ico-buildrun").attr("title", config.label.build_n_run);

                    break;
                case 'start-build':
                case 'start-test':
                case 'start-install':
                case 'start-get':
                    bottomGroup.fillOutput(data.output);

                    break;
                case 'go test':
                case 'go install':
                case 'go get':
                    bottomGroup.fillOutput($('.bottom-window-group .output > div').html() + data.output);

                    break;
                case 'build':
                    bottomGroup.fillOutput($('.bottom-window-group .output > div').html() + data.output);

                    if (data.lints) { // 说明编译有错误输出            
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

                    break;
            }
        };
        outputWS.onclose = function (e) {
            console.log('[output onclose] disconnected (' + e.code + ')');
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
                $(".menu > ul > li > a, .menu > ul> li > span").unbind("mouseover").removeClass("selected");
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

        this._initPreference();

        $(window).resize(function () {
            wide._initLayout();
            var editorDatas = editors.data,
                    height = $(".edit-panel").height() - $(".edit-panel .tabs").height();
            for (var i = 0, ii = editorDatas.length; i < ii; i++) {
                editorDatas[i].editor.setSize("100%", height);
            }

        });
    },
    _save: function (path, editor) {
        if (!path) {
            return false;
        }

        var request = newWideRequest();
        request.file = path;
        request.code = editor.getValue();

        $.ajax({
            type: 'POST',
            url: '/file/save',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                // reset the save state

                editor.doc.markClean();
                $(".edit-panel .tabs > div").each(function () {
                    var $span = $(this).find("span:eq(0)");
                    if ($span.attr("title") === path) {
                        $span.removeClass("changed");
                    }
                });
            }
        });
    },
    saveFile: function () {
        var path = editors.getCurrentPath();
        if (!path) {
            return false;
        }

        var editor = wide.curEditor;
        if (editor.doc.isClean()) { // no modification
            return false;
        }

        if ("text/x-go" === editor.getOption("mode")) {
            wide.gofmt(path, wide.curEditor); // go fmt will save

            return;
        }

        wide._save(path, wide.curEditor);
    },
    saveAllFiles: function () {
        if ($(".menu li.save-all").hasClass("disabled")) {
            return false;
        }
        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            var path = tree.fileTree.getNodeByTId(editors.data[i].id).path;
            var editor = editors.data[i].editor;

            if ("text/x-go" === editor.getOption("mode")) {
                wide.fmt(path, editor);
            } else {
                wide._save(path, editor);
            }
        }
    },
    closeAllFiles: function () {
        if ($(".menu li.close-all").hasClass("disabled")) {
            return false;
        }

        // 设置全部关闭标识
        var removeData = [];
        $(".edit-panel .tabs > div").each(function (i) {
            if (i !== 0) {
                removeData.push($(this).data("index"));
            }
        });
        $("#dialogCloseEditor").data("removeData", removeData);
        // 开始关闭
        $(".edit-panel .tabs .ico-close:eq(0)").click();
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

        if (!wide.curProcessId) {
            return false;
        }

        var request = newWideRequest();
        request.pid = wide.curProcessId;

        $.ajax({
            type: 'POST',
            url: '/stop',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                $(".toolbars .ico-stop").removeClass("ico-stop")
                        .addClass("ico-buildrun").attr("title", config.label.build_n_run);
            }
        });
    },
    // 构建.
    build: function () {
        wide.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.build").hasClass("disabled")) {
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
                bottomGroup.resetOutput();
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
                bottomGroup.resetOutput();
            },
            success: function (data) {
                $(".toolbars .ico-buildrun").addClass("ico-stop")
                        .removeClass("ico-buildrun").attr("title", config.label.stop);
            }
        });
    },
    // 测试.
    test: function () {
        wide.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.test").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;

        $.ajax({
            type: 'POST',
            url: '/go/test',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                bottomGroup.resetOutput();
            },
            success: function (data) {
            }
        });
    },
    goget: function () {
        wide.saveAllFiles();

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
                bottomGroup.resetOutput();
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

        $.ajax({
            type: 'POST',
            url: '/go/install',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                bottomGroup.resetOutput();
            },
            success: function (data) {
            }
        });
    },
    gofmt: function (path, editor) {
        var cursor = editor.getCursor();
        var scrollInfo = editor.getScrollInfo();

        var request = newWideRequest();
        request.file = path;
        request.code = editor.getValue();
        request.cursorLine = cursor.line;
        request.cursorCh = cursor.ch;

        $.ajax({
            async: false, // sync
            type: 'POST',
            url: '/go/fmt',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                if (data.succ) {
                    editor.setValue(data.code);
                    editor.setCursor(cursor);
                    editor.scrollTo(null, scrollInfo.top);

                    wide._save(path, editor);
                }
            }
        });
    },
    fmt: function (path, editor) {
        var mode = editor.getOption("mode");

        var cursor = editor.getCursor();
        var scrollInfo = editor.getScrollInfo();

        var request = newWideRequest();
        request.file = path;
        request.code = editor.getValue();
        request.cursorLine = cursor.line;
        request.cursorCh = cursor.ch;

        var formatted = null;

        switch (mode) {
            case "text/x-go":
                $.ajax({
                    async: false, // sync
                    type: 'POST',
                    url: '/go/fmt',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        if (data.succ) {
                            formatted = data.code;
                        }
                    }
                });

                break;
            case "text/html":
                formatted = html_beautify(editor.getValue());
                break;
            case "text/javascript":
            case "application/json":
                formatted = js_beautify(editor.getValue());
                break;
            case "text/css":
                formatted = css_beautify(editor.getValue());
                break;
            default :
                break;
        }

        if (formatted) {
            editor.setValue(formatted);
            editor.setCursor(cursor);
            editor.scrollTo(null, scrollInfo.top);

            wide._save(path, editor);
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
    bottomGroup.init();
});