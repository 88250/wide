var editors = {
    data: [],
    tabs: {},
    init: function () {
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
            removeAfter: function (id, nextId) {
                if (id === 'startPage') {
                    return false;
                }

                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === id) {
                        wide.fmt(tree.fileTree.getNodeByTId(editors.data[i].id).path, editors.data[i].editor);
                        editors.data.splice(i, 1);
                        break;
                    }
                }

                if (!nextId) {
                    // 不存在打开的编辑器
                    // remove selected tree node
                    tree.fileTree.cancelSelectedNode();
                    wide.curNode = undefined;

                    wide.curEditor = undefined;

                    menu.disabled(['save-all', 'close-all', 'run', 'go-get', 'go-install']);
                    $(".toolbars").hide();
                    return false;
                }

                if (nextId === editors.tabs.getCurrentId()) {
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
        this.openStartPage()
    },
    openStartPage: function () {
        editors.tabs.add({
            id: "startPage",
            title: '<span title="' + config.label.initialise + '">' + config.label.initialise + '</span>',
            content: '<div id="startPage"></div>',
            after: function () {
                $("#startPage").load('/start');
                $.ajax({
                    url: "http://symphony.b3log.org/apis/articles?tags=wide,golang&p=1&size=30",
                    type: "GET",
                    dataType: "jsonp",
                    jsonp: "callback",
                    error: function () {
                        $("#startPage").html("Loading B3log Announcement failed :-(");
                    },
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

                        var listHTML = "<ul>";
                        for (var i = 0; i < length; i++) {
                            var article = articles[i];
                            var articleLiHtml = "<li>"
                                    + "<a target='_blank' href='http://symphony.b3log.org" + article.articlePermalink + "'>"
                                    + article.articleTitle + "</a>&nbsp; <span class='date'>" + new Date(article.articleCreateTime);
                            +"</span></li>"
                            listHTML += articleLiHtml;
                        }
                        listHTML += "</ul>";

                        $("#startPage .news").html(listHTML);
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
        if (currentPath === config.label.initialise) {
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
                title = config.label.usages;
        if (type === "founds") {
            title = config.label.search_text;
        }
        if ($search.find("ul").length === 0) {
            wide.searchTab = new Tabs({
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

                var cursor = CodeMirror.Pos($it.find(".position").data("line") - 1, $it.find(".position").data("ch") - 1);
                wide.curEditor.setCursor(cursor);
                wide.curEditor.focus();
            });

            $search.find(".tabs-panel > div").append(searcHTML);

            $search.find(".tabs .first").text(title);
        } else {
            $search.find(".tabs").show();
            wide.searchTab.add({
                "id": "search" + (new Date()).getTime(),
                "title": title,
                "content": searcHTML
            });
        }

        // focus
        wide.bottomWindowTab.setCurrent("search");
        windows.flowBottom();
        $(".bottom-window-group .search").focus();
    },
    // 新建一个编辑器 Tab，如果已经存在 Tab 则切换到该 Tab.
    newEditor: function (data) {
        $(".toolbars").show();
        var id = wide.curNode.tId;

        // 光标位置
        var cursor = CodeMirror.Pos(0, 0);
        if (data.cursorLine && data.cursorCh) {
            cursor = CodeMirror.Pos(data.cursorLine - 1, data.cursorCh - 1);
        }

        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            if (editors.data[i].id === id) {
                editors.tabs.setCurrent(id);
                wide.curEditor = editors.data[i].editor;
                wide.curEditor.setCursor(cursor);
                wide.curEditor.focus();

                return false;
            }
        }

        editors.tabs.add({
            id: id,
            title: '<span title="' + wide.curNode.path + '"><span class="'
                    + wide.curNode.iconSkin + 'ico"></span>' + wide.curNode.name + '</span>',
            content: '<textarea id="editor' + id + '"></textarea>'
        });

        menu.undisabled(['save-all', 'close-all', 'run', 'go-get', 'go-install']);

        var rulers = [];
        rulers.push({color: "#ccc", column: 120, lineStyle: "dashed"});

        var editor = CodeMirror.fromTextArea(document.getElementById("editor" + id), {
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
            extraKeys: {
                "Ctrl-\\": "autocompleteAnyWord",
                ".": "autocompleteAfterDot",
                "Ctrl-I": "exprInfo",
                "Ctrl-G": "gotoLine",
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
                "Alt-F7": "findUsages"
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

        editor.setSize('100%', $(".edit-panel").height() - $(".edit-panel .tabs").height());
        editor.setValue(data.content);
        editor.setOption("mode", data.mode);
        editor.setOption("gutters", ["CodeMirror-lint-markers", "CodeMirror-foldgutter"]);

        if ("text/x-go" === data.mode || "application/json" === data.mode) {
            editor.setOption("lint", true);
        }

        if ("application/xml" === data.mode || "text/html" === data.mode) {
            editor.setOption("autoCloseTags", true);
        }

        editor.setCursor(cursor);

        wide.curEditor = editor;
        editors.data.push({
            "editor": editor,
            "id": id
        });
    }
};