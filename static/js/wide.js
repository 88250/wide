/*
 * Copyright (c) 2014-2015, b3log.org
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
    curProcessId: undefined, // curent running process id (pid)
    refreshOutline: function () {
        if (!wide.curEditor ||
                (wide.curEditor && wide.curEditor.doc.getMode().name !== "go")) {
            $("#outline").html('');
            return false;
        }

        var request = newWideRequest();
        request.code = wide.curEditor.getValue();

        $.ajax({
            type: 'POST',
            async: false,
            url: config.context + '/outline',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                if (!data.succ) {
                    return;
                }

                var outlineHTML = '<ul class="list">',
                        decls = ['constDecls', 'varDecls', 'funcDecls',
                            'structDecls', 'interfaceDecls', 'typeDecls'];

                for (var i = 0, max = decls.length; i < max; i++) {
                    var key = decls[i];
                    for (var j = 0, maxj = data[key].length; j < maxj; j++) {
                        var obj = data[key][j];
                        outlineHTML += '<li data-ch="' + obj.Ch + '" data-line="'
                                + obj.Line + '"><span class="ico ico-'
                                + key.replace('Decls', '') + '"></span> ' + obj.Name + '</li>';
                    }
                }
                $("#outline").html(outlineHTML + '</ul>');

                $("#outline li").dblclick(function () {
                    var $it = $(this),
                            cursor = CodeMirror.Pos($it.data('line'), $it.data("ch"));

                    var editor = wide.curEditor;
                    editor.setCursor(cursor);

                    var half = Math.floor(editor.getScrollInfo().clientHeight / editor.defaultTextHeight() / 2);
                    var cursorCoords = editor.cursorCoords({line: cursor.line - half, ch: 0}, "local");
                    editor.scrollTo(0, cursorCoords.top);

                    editor.focus();
                });
            }
        });
    },
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
            "height": 36,
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
            "height": 36,
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
                    url: config.context + '/file/remove',
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
                                if (tree.isParents(editors.data[i].id, wide.curNode.tId)) {
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
                    url: config.context + '/file/new',
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
                        var iconSkin = wide.getClassBySuffix(name.split(".")[1]);

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
                    url: config.context + '/file/new',
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
                                "creatable": true,
                                "isParent": true
                            }]);
                    }
                });
            }
        });

        $("#dialogGoFilePrompt").dialog({
            "modal": true,
            "height": 320,
            "width": 660,
            "title": config.label.goto_file,
            "okText": config.label.go,
            "cancelText": config.label.cancel,
            "afterInit": function () {
                $("#dialogGoFilePrompt").on("dblclick", "li", function () {
                    var tId = tree.getTIdByPath($(this).find(".ft-small").text());
                    tree.openFile(tree.fileTree.getNodeByTId(tId));
                    tree.fileTree.selectNode(wide.curNode);
                    $("#dialogGoFilePrompt").dialog("close");
                    wide.curEditor.focus();
                });

                $("#dialogGoFilePrompt").on("click", "li", function () {
                    var $list = $("#dialogGoFilePrompt > .list");
                    $list.find("li").removeClass("selected");
                    $list.data("index", $(this).data("index"));
                    $(this).addClass("selected");
                });

                hotkeys.bindList($("#dialogGoFilePrompt > input"), $("#dialogGoFilePrompt > .list"), function ($selected) {
                    var tId = tree.getTIdByPath($selected.find(".ft-small").text());
                    tree.openFile(tree.fileTree.getNodeByTId(tId));
                    tree.fileTree.selectNode(wide.curNode);
                    $("#dialogGoFilePrompt").dialog("close");
                    wide.curEditor.focus();
                });

                $("#dialogGoFilePrompt > input").bind("input", function () {
                    var name = $("#dialogGoFilePrompt > input").val();

                    var request = newWideRequest();
                    request.path = '';
                    request.name = '*' + name + '*';
                    if (wide.curNode) {
                        request.path = wide.curNode.path;
                    }

                    $.ajax({
                        type: 'POST',
                        url: config.context + '/file/find/name',
                        data: JSON.stringify(request),
                        dataType: "json",
                        success: function (data) {
                            if (!data.succ) {
                                return;
                            }

                            var goFileHTML = '';
                            for (var i = 0, max = data.founds.length; i < max; i++) {
                                var path = data.founds[i].path,
                                        name = path.substr(path.lastIndexOf(config.pathSeparator) + 1),
                                        icoSkin = wide.getClassBySuffix(name.split(".")[1]);
                                if (i === 0) {
                                    goFileHTML += '<li data-index="' + i + '" class="selected" title="'
                                            + path + '"><span class="'
                                            + icoSkin + 'ico"></span>'
                                            + name + '&nbsp;&nbsp;&nbsp;&nbsp;<span class="ft-small">'
                                            + path + '</span></li>';
                                } else {
                                    goFileHTML += '<li data-index="' + i + '" title="'
                                            + path + '"><span class="' + icoSkin + 'ico"></span>'
                                            + name + '&nbsp;&nbsp;&nbsp;&nbsp;<span class="ft-small">'
                                            + path + '</span></li>';
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
                var tId = tree.getTIdByPath($("#dialogGoFilePrompt .selected .ft-small").text());
                tree.openFile(tree.fileTree.getNodeByTId(tId));
                tree.fileTree.selectNode(wide.curNode);
                $("#dialogGoFilePrompt").dialog("close");
                wide.curEditor.focus();
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

        $("#dialogGitClonePrompt").dialog({
            "modal": true,
            "height": 52,
            "width": 360,
            "title": config.label.git_clone,
            "okText": config.label.confirm,
            "cancelText": config.label.cancel,
            "afterOpen": function () {
                $("#dialogGitClonePrompt > input").val('').focus();
                $("#dialogGitClonePrompt").closest(".dialog-main").find(".dialog-footer > button:eq(0)").prop("disabled", true);
            },
            "ok": function () {
                $("#dialogGitClonePrompt").dialog("close");
                
                var request = newWideRequest();
                request.path = wide.curNode.path;
                request.repository = $("#dialogGitClonePrompt > input").val();

                $.ajax({
                    type: 'POST',
                    url: config.context + '/git/clone',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {

                    }
                });
            }
        });
    },
    _initLayout: function () {
        var mainH = $(window).height() - $(".menu").height() - $(".footer").height() - 2,
                bottomH = Math.floor(mainH * 0.3);
        // 减小初始化界面抖动
        $(".content").height(mainH).css("position", "relative");
        $(".side .tabs-panel").height(mainH - 20);

        var $bottomGroup = $(".bottom-window-group");
        if ($bottomGroup.hasClass("bottom-window-group-max")) {
            $(".bottom-window-group > .tabs-panel > div > div").height(mainH - $bottomGroup.children(".tabs").height());
        } else {
            $(".bottom-window-group > .tabs-panel > div > div").height(bottomH - $bottomGroup.children(".tabs").height());
        }

        if ($(".side-right").hasClass("side-right-max")) {
            $(".side-right > .tabs-panel > div").height(mainH - $bottomGroup.children(".tabs").height());
        } else {
            $(".side-right > .tabs-panel > div").height($('.side-right').height() - $bottomGroup.children(".tabs").height());
        }

        $("#startPage").height($('.side-right').height() - $bottomGroup.children(".tabs").height() - 100);
    },
    _initWS: function () {
        var outputWS = new ReconnectingWebSocket(config.channel + '/output/ws?sid=' + config.wideSessionId);
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
                    url: config.context + '/run',
                    data: JSON.stringify(request),
                    dataType: "json"
                });
            }

            switch (data.cmd) {
                case 'run':
                    var content = $('.bottom-window-group .output > div').html();
                    if (!wide.curProcessId || '' === content) {
                        bottomGroup.fillOutput(content + '<pre>' + data.output + '</pre>');
                    } else {
                        bottomGroup.fillOutput(content.replace(/<\/pre>$/g, data.output + '</pre>'));
                    }

                    wide.curProcessId = data.pid;

                    break;
                case 'run-done':
                    bottomGroup.fillOutput($('.bottom-window-group .output > div').html().replace(/<\/pre>$/g, data.output + '</pre>'));

                    wide.curProcessId = undefined;
                    $("#buildRun").removeClass("ico-stop")
                            .addClass("ico-buildrun").attr("title", config.label.build_n_run);

                    break;
                case 'start-build':
                case 'start-test':
                case 'start-vet':
                case 'start-install':
                case 'start-get':
                case 'start-git_clone':
                    bottomGroup.fillOutput(data.output);

                    break;
                case 'go test':
                case 'go vet':
                case 'go install':
                case 'go get':
                    bottomGroup.fillOutput($('.bottom-window-group .output > div').html() + data.output);

                    break;
                case 'git clone':
                    bottomGroup.fillOutput($('.bottom-window-group .output > div').html() + data.output);
                    tree.fileTree.reAsyncChildNodes(wide.curNode, "refresh", false);

                    break;
                case 'build':
                    bottomGroup.fillOutput($('.bottom-window-group .output > div').html() + data.output);

                    if (data.lints) { // has build error
                        var files = {};

                        for (var i = 0; i < data.lints.length; i++) {
                            var lint = data.lints[i];

                            goLintFound.push({from: CodeMirror.Pos(lint.lineNo, 0),
                                to: CodeMirror.Pos(lint.lineNo, 0),
                                message: lint.msg, severity: lint.severity});

                            files[lint.file] = lint.file;
                        }

                        $("#buildRun").removeClass("ico-stop")
                                .addClass("ico-buildrun").attr("title", config.label.build_n_run);

                        // trigger gutter lint
                        for (var path in files) {
                            var editor = editors.getEditorByPath(path);
                            CodeMirror.signal(editor, "change", editor);
                        }
                    }

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
        $("body").bind("mouseup", function (event) {
            // MAC 右键文件树失效
            if (event.which === 3) {
                return false;
            }

            $(".frame").hide();

            if (!($(event.target).closest(".frame").length === 1 || event.target.className === "frame")) {
                $(".menu > ul > li").unbind().removeClass("selected");
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
            url: config.context + '/file/save',
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

            // build the file at once
            var request = newWideRequest();
            request.file = path;
            request.code = editor.getValue();
            request.nextCmd = ""; // build only, no following operation
            $.ajax({
                type: 'POST',
                url: config.context + '/build',
                data: JSON.stringify(request),
                dataType: "json",
                beforeSend: function (data) {
                    bottomGroup.resetOutput();
                },
                success: function (data) {
                }
            });

            // refresh outline
            wide.refreshOutline();

            return;
        }

        wide._save(path, wide.curEditor);
    },
    stop: function () {
        if ($("#buildRun").hasClass("ico-buildrun")) {
            menu.run();
            return false;
        }

        if (!wide.curProcessId) {
            return false;
        }

        var request = newWideRequest();
        request.pid = wide.curProcessId;

        $.ajax({
            type: 'POST',
            url: config.context + '/stop',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                $("#buildRun").removeClass("ico-stop")
                        .addClass("ico-buildrun").attr("title", config.label.build_n_run);
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
            url: config.context + '/go/fmt',
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
                    url: config.context + '/go/fmt',
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
    getClassBySuffix: function (suffix) {
        var iconSkin = "ico-ztree-other ";
        switch (suffix) {
            case "html":
            case "htm":
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
            case "jpg":
            case "jpeg":
            case "bmp":
            case "gif":
            case "png":
            case "svg":
            case "ico":
                iconSkin = "ico-ztree-img ";
                break;
        }

        return iconSkin;
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
