var editors = {
    data: [],
    tabs: {},
    close: function () {
        $(".edit-panel .tabs > div[data-index=" + $(".edit-panel .frame").data("index") + "]").find(".ico-close").click();
    },
    closeOther: function () {
        var currentIndex = $(".edit-panel .frame").data("index");

        // 设置全部关闭标识
        var removeData = [];
        $(".edit-panel .tabs > div").each(function (i) {
            if (currentIndex !== $(this).data("index")) {
                removeData.push($(this).data("index"));
            }
        });
        if (removeData.length === 0) {
            return false;
        }
        var firstIndex = removeData.splice(0, 1);
        $("#dialogCloseEditor").data("removeData", removeData);
        // 开始关闭
        $(".edit-panel .tabs > div[data-index=" + firstIndex + "]").find(".ico-close").click();
    },
    _removeAllMarker: function () {
        var removeData = $("#dialogCloseEditor").data("removeData");
        if (removeData && removeData.length > 0) {
            var removeIndex = removeData.splice(0, 1);
            $("#dialogCloseEditor").data("removeData", removeData);
            $(".edit-panel .tabs > div[data-index=" + removeIndex + "] .ico-close").click();
        }
    },
    _initClose: function () {
        // 关闭、关闭其他、关闭所有
        $(".edit-panel").on("mousedown", '.tabs > div', function (event) {
            event.stopPropagation();

            if (event.button === 0) { // 左键
                $(".edit-panel .frame").hide();
                return false;
            }

            // event.button === 2 右键
            var left = event.screenX;
            if ($(".side").css("left") === "auto" || $(".side").css("left") === "0px") {
                left = event.screenX - $(".side").width();
            }
            $(".edit-panel .frame").show().css({
                "left": left + "px",
                "top": "21px"
            }).data('index', $(this).data("index"));
            return false;
        });
    },
    init: function () {
        $("#dialogCloseEditor").dialog({
            "modal": true,
            "height": 66,
            "width": 260,
            "title": config.label.tip,
            "hideFooter": true,
            "afterOpen": function (fileName) {
                $("#dialogCloseEditor > div:eq(0)").html(config.label.file
                        + ' <b>' + fileName + '</b>. ' + config.label.confirm_save + '?');
                $("#dialogCloseEditor button:eq(0)").focus();
            },
            "afterInit": function () {
                $("#dialogCloseEditor button.save").click(function () {
                    var i = $("#dialogCloseEditor").data("index");
                    wide.fmt(tree.fileTree.getNodeByTId(editors.data[i].id).path, editors.data[i].editor);
                    editors.tabs.del(editors.data[i].id);
                    $("#dialogCloseEditor").dialog("close");

                    editors._removeAllMarker();
                });

                $("#dialogCloseEditor button.discard").click(function () {
                    $("#dialogCloseEditor").dialog("close");

                    editors._removeAllMarker();
                });

                $("#dialogCloseEditor button.cancel").click(function () {
                    var i = $("#dialogCloseEditor").data("index");
                    editors.tabs.del(editors.data[i].id);
                    $("#dialogCloseEditor").dialog("close");

                    editors._removeAllMarker();
                });
            }
        });

        editors.tabs = new Tabs({
            id: ".edit-panel",
            clickAfter: function (id) {
                if (id === 'startPage') {
                    return false;
                }

                // set tree node selected
                var node = tree.fileTree.getNodeByTId(id);
                tree.fileTree.selectNode(node);
                wide.curNode = node;

                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === id) {
                        wide.curEditor = editors.data[i].editor;
                        break;
                    }
                }

                wide.curEditor.focus();
            },
            removeBefore: function (id) {
                if (id === 'startPage') { // 当前关闭的 tab 是起始页
                    editors._removeAllMarker();
                    return true;
                }

                // 移除编辑器
                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === id) {
                        if (editors.data[i].editor.doc.isClean()) {
                            editors._removeAllMarker();
                            return true;
                        } else {
                            $("#dialogCloseEditor").dialog("open", $(".edit-panel .tabs > div[data-index="
                                    + editors.data[i].id + "] > span:eq(0)").text());
                            $("#dialogCloseEditor").data("index", i);
                            return false;
                        }

                        break;
                    }
                }
            },
            removeAfter: function (id, nextId) {
                if ($(".edit-panel .tabs > div").length === 0) {
                    // 全部 tab 都关闭时才 disables 菜单中“全部关闭”的按钮
                    menu.disabled(['close-all']);
                }

                if (id === 'startPage') { // 当前关闭的 tab 是起始页
                    return false;
                }

                // 移除编辑器
                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === id) {
                        editors.data.splice(i, 1);
                        break;
                    }
                }

                if (editors.data.length === 0) { // 起始页可能存在，所以用编辑器数据判断
                    menu.disabled(['save-all', 'build', 'run', 'go-test', 'go-get', 'go-install']);
                    $(".toolbars").hide();
                }

                if (!nextId) {
                    // 不存在打开的编辑器
                    // remove selected tree node
                    tree.fileTree.cancelSelectedNode();
                    wide.curNode = undefined;
                    wide.curEditor = undefined;
                    return false;
                }

                if (nextId === editors.tabs.getCurrentId()) {
                    // 关闭的不是当前编辑器
                    return false;
                }

                // set tree node selected
                var node = tree.fileTree.getNodeByTId(nextId);
                tree.fileTree.selectNode(node);
                wide.curNode = node;

                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === nextId) {
                        wide.curEditor = editors.data[i].editor;
                        break;
                    }
                }
            }
        });

        $(".edit-panel .tabs").on("dblclick", function () {
            if ($(".toolbars .ico-max").length === 1) {
                windows.maxEditor();
            } else {
                windows.restoreEditor();
            }
        });

        this._initCodeMirrorHotKeys();
        this.openStartPage();
        this._initClose();
    },
    openStartPage: function () {
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

        editors.tabs.add({
            id: "startPage",
            title: '<span title="' + config.label.start_page + '">' + config.label.start_page + '</span>',
            content: '<div id="startPage"></div>',
            after: function () {
                $("#startPage").load('/start?sid=' + config.wideSessionId);
                $.ajax({
                    url: "http://symphony.b3log.org/apis/articles?tags=wide,golang&p=1&size=30",
                    type: "GET",
                    dataType: "jsonp",
                    jsonp: "callback",
                    success: function (data, textStatus) {
                        var articles = data.articles;
                        if (0 === articles.length) {
                            return;
                        }

                        // 按 size = 30 取，但只保留最多 10 篇
                        var length = articles.length;
                        if (length > 10) {
                            length = 10;
                        }

                        var listHTML = "<ul><li class='title'>" + config.label.community + "</li>";
                        for (var i = 0; i < length; i++) {
                            var article = articles[i];
                            listHTML += "<li>"
                                    + "<a target='_blank' href='http://symphony.b3log.org"
                                    + article.articlePermalink + "'>"
                                    + article.articleTitle + "</a>&nbsp; <span class='date'>"
                                    + dateFormat(article.articleCreateTime, 'yyyy-MM-dd hh:mm');
                            +"</span></li>";
                        }

                        $("#startPage .news").html(listHTML + "</ul>");
                    }
                });
            }
        });
    },
    getCurrentId: function () {
        var currentId = editors.tabs.getCurrentId();
        if (currentId === 'startPage') {
            currentId = null;
        }
        return currentId;
    },
    getCurrentPath: function () {
        var currentPath = $(".edit-panel .tabs .current span:eq(0)").attr("title");
        if (currentPath === config.label.start_page) {
            currentPath = null;
        }
        return currentPath;
    },
    _initCodeMirrorHotKeys: function () {
        CodeMirror.registerHelper("hint", "go", function (editor) {
            var word = /[\w$]+/;

            var cur = editor.getCursor(), curLine = editor.getLine(cur.line);

            var start = cur.ch, end = start;
            while (end < curLine.length && word.test(curLine.charAt(end))) {
                ++end;
            }
            while (start && word.test(curLine.charAt(start - 1))) {
                --start;
            }

            var request = newWideRequest();
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = editor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            var autocompleteHints = [];

            $.ajax({
                async: false, // 同步执行
                type: 'POST',
                url: '/autocomplete',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (data) {
                    var autocompleteArray = data[1];

                    if (autocompleteArray) {
                        for (var i = 0; i < autocompleteArray.length; i++) {
                            var displayText = '';

                            switch (autocompleteArray[i].class) {
                                case "type":
                                case "const":
                                case "var":
                                case "package":
                                    displayText = '<span class="fn-clear">'// + autocompleteArray[i].class 
                                            + '<b class="fn-left">' + autocompleteArray[i].name + '</b>    '
                                            + autocompleteArray[i].type + '</span>';

                                    break;
                                case "func":
                                    displayText = '<span>'// + autocompleteArray[i].class 
                                            + '<b>' + autocompleteArray[i].name + '</b>'
                                            + autocompleteArray[i].type.substring(4) + '</span>';

                                    break;
                                default:
                                    console.warn("Can't handle autocomplete [" + autocompleteArray[i].class + "]");

                                    break;
                            }

                            autocompleteHints[i] = {
                                displayText: displayText,
                                text: autocompleteArray[i].name
                            };
                        }
                    }

                    // 清除未保存状态
                    editor.doc.markClean();
                    $(".edit-panel .tabs > div.current > span").removeClass("changed");
                }
            });

            return {list: autocompleteHints, from: CodeMirror.Pos(cur.line, start), to: CodeMirror.Pos(cur.line, end)};
        });

        CodeMirror.commands.autocompleteAfterDot = function (cm) {
            setTimeout(function () {
                if (!cm.state.completionActive) {
                    cm.showHint({hint: CodeMirror.hint.go, completeSingle: false});
                }
            }, 50);

            return CodeMirror.Pass;
        };

        CodeMirror.commands.autocompleteAnyWord = function (cm) {
            cm.showHint({hint: CodeMirror.hint.auto});
        };

        CodeMirror.commands.gotoLine = function (cm) {
            $("#dialogGoLinePrompt").dialog("open");
        };

        // 用于覆盖 cm 默认绑定的某些快捷键功能.
        CodeMirror.commands.doNothing = function (cm) {
        };

        CodeMirror.commands.exprInfo = function (cm) {
            var cur = wide.curEditor.getCursor();

            var request = newWideRequest();
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = wide.curEditor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            $.ajax({
                type: 'POST',
                url: '/exprinfo',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (data) {
                    if (!data.succ) {
                        return;
                    }
                    var position = wide.curEditor.cursorCoords();
                    $("body").append('<div style="top:'
                            + (position.top + 15) + 'px;left:' + position.left
                            + 'px" class="edit-exprinfo">' + data.info + '</div>');
                }
            });
        };

        CodeMirror.commands.jumpToDecl = function (cm) {
            var cur = wide.curEditor.getCursor();

            var request = newWideRequest();
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = wide.curEditor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            $.ajax({
                type: 'POST',
                url: '/find/decl',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (data) {
                    if (!data.succ) {
                        return;
                    }

                    var cursorLine = data.cursorLine;
                    var cursorCh = data.cursorCh;

                    var request = newWideRequest();
                    request.path = data.path;

                    $.ajax({
                        type: 'POST',
                        url: '/file',
                        data: JSON.stringify(request),
                        dataType: "json",
                        success: function (data) {
                            if (!data.succ) {
                                $("#dialogAlert").dialog("open", data.msg);

                                return false;
                            }

                            var tId = tree.getTIdByPath(data.path);
                            wide.curNode = tree.fileTree.getNodeByTId(tId);
                            tree.fileTree.selectNode(wide.curNode);

                            data.cursorLine = cursorLine;
                            data.cursorCh = cursorCh;
                            editors.newEditor(data);
                        }
                    });
                }
            });
        };

        CodeMirror.commands.findUsages = function (cm) {
            var cur = wide.curEditor.getCursor();

            var request = newWideRequest();
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = wide.curEditor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            $.ajax({
                type: 'POST',
                url: '/find/usages',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (data) {
                    if (!data.succ) {
                        return;
                    }

                    editors.appendSearch(data.founds, 'usages', '');
                }
            });
        };
    },
    appendSearch: function (data, type, key) {
        var searcHTML = '<ul>';

        for (var i = 0, ii = data.length; i < ii; i++) {
            var contents = data[i].contents[0],
                    index = contents.indexOf(key);
            contents = contents.substring(0, index)
                    + '<b>' + key + '</b>'
                    + contents.substring(index + key.length);

            searcHTML += '<li title="' + data[i].path + '">'
                    + contents + "&nbsp;&nbsp;&nbsp;&nbsp;<span class='path'>" + data[i].path
                    + '<i class="position" data-line="'
                    + data[i].line + '" data-ch="' + data[i].ch + '"> (' + data[i].line + ':'
                    + data[i].ch + ')</i></span></li>';
        }
        searcHTML += '</ul>';

        var $search = $('.bottom-window-group .search'),
                title = config.label.find_usages;
        if (type === "founds") {
            title = config.label.search_text;
        }
        if ($search.find("ul").length === 0) {
            bottomGroup.searchTab = new Tabs({
                id: ".bottom-window-group .search",
                removeAfter: function (id, prevId) {
                    if ($search.find("ul").length === 1) {
                        $search.find(".tabs").hide();
                    }
                }
            });

            $search.on("click", "li", function () {
                $search.find("li").removeClass("selected");
                $(this).addClass("selected");
            });

            $search.on("dblclick", "li", function () {
                var $it = $(this),
                        tId = tree.getTIdByPath($it.attr("title"));
                tree.openFile(tree.fileTree.getNodeByTId(tId));
                tree.fileTree.selectNode(wide.curNode);

                var line = $it.find(".position").data("line") - 1;
                var cursor = CodeMirror.Pos(line, $it.find(".position").data("ch") - 1);


                var editor = wide.curEditor;
                editor.setCursor(cursor);

                var half = Math.floor(editor.getScrollInfo().clientHeight / editor.defaultTextHeight() / 2);
                var cursorCoords = editor.cursorCoords({line: cursor.line - half, ch: 0}, "local");
                editor.scrollTo(0, cursorCoords.top);

                wide.curEditor.focus();
            });

            $search.find(".tabs-panel > div").append(searcHTML);

            $search.find(".tabs .first").text(title);
        } else {
            $search.find(".tabs").show();
            bottomGroup.searchTab.add({
                "id": "search" + (new Date()).getTime(),
                "title": title,
                "content": searcHTML
            });
        }

        // focus
        bottomGroup.tabs.setCurrent("search");
        windows.flowBottom();
        $(".bottom-window-group .search").focus();
    },
    // 新建一个编辑器 Tab，如果已经存在 Tab 则切换到该 Tab.
    newEditor: function (data) {
        $(".toolbars").show();
        var id = wide.curNode.tId;

        var cursor = CodeMirror.Pos(0, 0);
        if (data.cursorLine && data.cursorCh) {
            cursor = CodeMirror.Pos(data.cursorLine - 1, data.cursorCh - 1);
        }

        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            if (editors.data[i].id === id) {
                editors.tabs.setCurrent(id);
                wide.curEditor = editors.data[i].editor;
                var editor = wide.curEditor;

                editor.setCursor(cursor);

                var half = Math.floor(editor.getScrollInfo().clientHeight / editor.defaultTextHeight() / 2);
                var cursorCoords = editor.cursorCoords({line: cursor.line - half, ch: 0}, "local");
                editor.scrollTo(0, cursorCoords.top);

                editor.focus();

                return false;
            }
        }

        editors.tabs.add({
            id: id,
            title: '<span title="' + wide.curNode.path + '"><span class="'
                    + wide.curNode.iconSkin + 'ico"></span>' + wide.curNode.name + '</span>',
            content: '<textarea id="editor' + id + '"></textarea>'
        });

        menu.undisabled(['save-all', 'close-all', 'build', 'run', 'go-test', 'go-get', 'go-install']);

        var rulers = [];
        rulers.push({color: "#ccc", column: 120, lineStyle: "dashed"});

        var textArea = document.getElementById("editor" + id);
        textArea.value = data.content;

        var editor = CodeMirror.fromTextArea(textArea, {
            lineNumbers: true,
            autofocus: true,
            autoCloseBrackets: true,
            matchBrackets: true,
            highlightSelectionMatches: {showToken: /\w/},
            rulers: rulers,
            styleActiveLine: true,
            theme: 'wide',
            indentUnit: 4,
            foldGutter: true,
            path: data.path,
            extraKeys: {
                "Ctrl-\\": "autocompleteAnyWord",
                ".": "autocompleteAfterDot",
                "Ctrl-I": "exprInfo",
                "Ctrl-L": "gotoLine",
                "Ctrl-E": "deleteLine",
                "Ctrl-D": "doNothing", // 取消默认的 deleteLine
                "Ctrl-B": "jumpToDecl",
                "Ctrl-S": function () {
                    wide.saveFile();
                },
                "Shift-Ctrl-S": function () {
                    wide.saveAllFiles();
                },
                "Shift-Alt-F": function () {
                    var currentPath = editors.getCurrentPath();
                    if (!currentPath) {
                        return false;
                    }
                    wide.fmt(currentPath, wide.curEditor);
                },
                "Alt-F7": "findUsages",
                "Shift-Alt-Enter": function () {
                    if (windows.isMaxEditor) {
                        windows.restoreEditor();
                    } else {
                        windows.maxEditor();
                    }
                },
                "Shift-Ctrl-Up": function (cm) {
                    var content = '',
                            selectoion = cm.listSelections()[0];

                    var from = selectoion.anchor,
                            to = selectoion.head;
                    if (from.line > to.line) {
                        from = selectoion.head;
                        to = selectoion.anchor;
                    }

                    for (var i = from.line, max = to.line; i <= max; i++) {
                        content += '\n' + cm.getLine(i);
                    }

                    cm.replaceRange(content, CodeMirror.Pos(to.line));

                    cm.setSelection(CodeMirror.Pos(to.line, to.ch),
                            CodeMirror.Pos(from.line, from.ch));
                },
                "Shift-Ctrl-Down": function (cm) {
                    var content = '',
                            selectoion = cm.listSelections()[0];

                    var from = selectoion.anchor,
                            to = selectoion.head;
                    if (from.line > to.line) {
                        from = selectoion.head;
                        to = selectoion.anchor;
                    }

                    for (var i = from.line, max = to.line; i <= max; i++) {
                        content += '\n' + cm.getLine(i);
                    }

                    cm.replaceRange(content, CodeMirror.Pos(to.line));
                    var offset = to.line - from.line + 1;

                    cm.setSelection(CodeMirror.Pos(to.line + offset, to.ch),
                            CodeMirror.Pos(from.line + offset, from.ch));
                },
                "Shift-Alt-Up": function (cm) {
                    var selectoion = cm.listSelections()[0];

                    var from = selectoion.anchor,
                            to = selectoion.head;
                    if (from.line > to.line) {
                        from = selectoion.head;
                        to = selectoion.anchor;
                    }

                    if (from.line === 0) {
                        return false;
                    }

                    cm.replaceRange('\n' + cm.getLine(from.line - 1), CodeMirror.Pos(to.line));
                    if (from.line === 1) {
                        cm.replaceRange('', CodeMirror.Pos(0, 0),
                                CodeMirror.Pos(1, 0));
                    } else {
                        cm.replaceRange('', CodeMirror.Pos(from.line - 2, cm.getLine(from.line - 2).length),
                                CodeMirror.Pos(from.line - 1, cm.getLine(from.line - 1).length));
                    }

                    cm.setSelection(CodeMirror.Pos(from.line - 1, from.ch),
                            CodeMirror.Pos(to.line - 1, to.ch));
                },
                "Shift-Alt-Down": function (cm) {
                    var selectoion = cm.listSelections()[0];

                    var from = selectoion.anchor,
                            to = selectoion.head;
                    if (from.line > to.line) {
                        from = selectoion.head;
                        to = selectoion.anchor;
                    }

                    if (to.line === cm.lastLine()) {
                        return false;
                    }

                    cm.replaceRange('\n' + cm.getLine(to.line + 1), CodeMirror.Pos(from.line - 1));
                    cm.replaceRange('', CodeMirror.Pos(to.line + 1, cm.getLine(to.line + 1).length),
                            CodeMirror.Pos(to.line + 2, cm.getLine(to.line + 2).length));

                    var offset = to.line - from.line + 1;
                    cm.setSelection(CodeMirror.Pos(to.line + offset, to.ch),
                            CodeMirror.Pos(from.line + offset, from.ch));
                }
            }
        });

        editor.on('cursorActivity', function (cm) {
            $(".edit-exprinfo").remove();
            var cursor = cm.getCursor();

            $(".footer .cursor").text('|   ' + (cursor.line + 1) + ':' + (cursor.ch + 1) + '   |');
            // TODO: 关闭 tab 的时候要重置
        });

        editor.on('focus', function (cm) {
            windows.clearFloat();
        });

        editor.on('blur', function (cm) {
            $(".edit-exprinfo").remove();
        });

        editor.on('changes', function (cm) {
            if (cm.doc.isClean()) {
                // 没有修改过
                $(".edit-panel .tabs > div").each(function () {
                    var $span = $(this).find("span:eq(0)");
                    if ($span.attr("title") === cm.options.path) {
                        $span.removeClass("changed");
                    }
                });
            } else {
                // 修改过
                $(".edit-panel .tabs > div").each(function () {
                    var $span = $(this).find("span:eq(0)");
                    if ($span.attr("title") === cm.options.path) {
                        $span.addClass("changed");
                    }
                });
            }
        });

        editor.setSize('100%', $(".edit-panel").height() - $(".edit-panel .tabs").height());
        editor.setOption("mode", data.mode);
        editor.setOption("gutters", ["CodeMirror-lint-markers", "CodeMirror-foldgutter"]);

        if ("text/x-go" === data.mode || "application/json" === data.mode) {
            editor.setOption("lint", true);
        }

        if ("application/xml" === data.mode || "text/html" === data.mode) {
            editor.setOption("autoCloseTags", true);
        }

        editor.setCursor(cursor);

        var half = Math.floor(editor.getScrollInfo().clientHeight / editor.defaultTextHeight() / 2);
        var cursorCoords = editor.cursorCoords({line: cursor.line - half, ch: 0}, "local");
        editor.scrollTo(0, cursorCoords.top);

        wide.curEditor = editor;
        editors.data.push({
            "editor": editor,
            "id": id
        });
    }
};