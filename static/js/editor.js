var editors = {
    data: [],
    init: function() {
        editors._initAutocomplete();
        editors.tabs = new Tabs({
            id: ".edit-panel",
            clickAfter: function(id) {
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
            },
            removeAfter: function(id, nextId) {
                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === id) {
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
                    $(".ico-fullscreen").hide();
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


        $(".edit-header .tabs").on("dblclick", "div", function() {
            editors.fullscreen();
        });
    },
    fullscreen: function() {
        wide.curEditor.setOption("fullScreen", true);
    },
    _initAutocomplete: function() {
        CodeMirror.registerHelper("hint", "go", function(editor) {
            var word = /[\w$]+/;

            var cur = editor.getCursor(), curLine = editor.getLine(cur.line);

            var start = cur.ch, end = start;
            while (end < curLine.length && word.test(curLine.charAt(end))) {
                ++end;
            }
            while (start && word.test(curLine.charAt(start - 1))) {
                --start;
            }
            var request = {
                code: editor.getValue(),
                cursorLine: cur.line,
                cursorCh: cur.ch
            };

            var autocompleteHints = [];

            $.ajax({
                async: false, // 同步执行
                type: 'POST',
                url: '/autocomplete',
                data: JSON.stringify(request),
                dataType: "json",
                success: function(data) {
                    var autocompleteArray = data[1];

                    if (autocompleteArray) {
                        for (var i = 0; i < autocompleteArray.length; i++) {
                            autocompleteHints[i] = autocompleteArray[i].name;
                        }
                    }
                }
            });

            return {list: autocompleteHints, from: CodeMirror.Pos(cur.line, start), to: CodeMirror.Pos(cur.line, end)};
        });

        CodeMirror.commands.autocompleteAfterDot = function(cm) {
            setTimeout(function() {
                if (!cm.state.completionActive) {
                    cm.showHint({hint: CodeMirror.hint.go, completeSingle: false});
                }
            }, 50);

            return CodeMirror.Pass;
        };

        CodeMirror.commands.autocompleteAnyWord = function(cm) {
            cm.showHint({hint: CodeMirror.hint.auto});
        };

        CodeMirror.commands.gotoLine = function(cm) {
            var line = prompt("Go To Line: ", "0");

            cm.setCursor(CodeMirror.Pos(line - 1, 0));
        };

        CodeMirror.commands.doNothing = function(cm) {
        };

        CodeMirror.commands.jumpToDecl = function(cm) {
            var cur = wide.curEditor.getCursor();

            var request = {
                file: wide.curNode.path,
                code: wide.curEditor.getValue(),
                cursorLine: cur.line,
                cursorCh: cur.ch
            };

            $.ajax({
                type: 'POST',
                url: '/finddecl',
                data: JSON.stringify(request),
                dataType: "json",
                success: function(data) {
                    if (!data.succ) {
                        return;
                    }

                    var request = {
                        path: data.path
                    };

                    $.ajax({
                        type: 'POST',
                        url: '/file',
                        data: JSON.stringify(request),
                        dataType: "json",
                        success: function(data) {
                            if (!data.succ) {
                                alert(data.msg);

                                return false;
                            }

                            // FIXME: V, 这个可能不在文件树里，但是也需要打开一个编辑器
                            // 打开一个新编辑器并定位到跳转的行列
                            var line = data.cursorLine;
                            var ch = data.cursorCh;
                            editors.newEditor(data);
                        }
                    });
                }
            });
        };
    },
    newEditor: function(data) {
        $(".ico-fullscreen").show();
        var id = wide.curNode.tId;
        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            if (editors.data[i].id === id) {
                editors.tabs.setCurrent(id);
                wide.curEditor = editors.data[i].editor;
                return false;
            }
        }

        editors.tabs.add({
            id: id,
            title: '<span title="' + wide.curNode.path + '"><span class="' 
                    + wide.curNode.iconSkin + 'ico"></span>' + wide.curNode.name + '</span>',
            content: '<textarea id="editor' + id + '"></textarea>'
        });

        rulers = [];
        rulers.push({color: "#ccc", column: 120, lineStyle: "dashed"});

        var editor = CodeMirror.fromTextArea(document.getElementById("editor" + id), {
            lineNumbers: true,
            autoCloseBrackets: true,
            matchBrackets: true,
            highlightSelectionMatches: {showToken: /\w/},
            rulers: rulers,
            styleActiveLine: true,
            theme: 'lesser-dark',
            indentUnit: 4,
            foldGutter: true,
            extraKeys: {
                "Ctrl-\\": "autocompleteAnyWord",
                ".": "autocompleteAfterDot",
                "Esc": function(cm) {
                    if (cm.getOption("fullScreen")) {
                        cm.setOption("fullScreen", false);
                    }
                },
                "F11": function(cm) {
                    cm.setOption("fullScreen", !cm.getOption("fullScreen"));
                },
                "Ctrl-G": "gotoLine",
                "Ctrl-E": "deleteLine",
                "Ctrl-D": "doNothing", // 取消默认的 deleteLine
                "Ctrl-B": "jumpToDecl"
            }
        });

        editor.on('cursorActivity', function(cm) {
            var cursor = cm.getCursor();

            $("#footer-cursor").text('|   ' + (cursor.line + 1) + ':' + (cursor.ch + 1) + '   |');
            // TODO: 关闭 tab 的时候要重置
            // TODO: 保存当前编辑器光标位置，切换 tab 的时候要设置回来
        });

        editor.setSize('100%', $(".edit-panel").height() - $(".edit-header").height());
        editor.setValue(data.content);
        editor.setOption("mode", data.mode);

        editor.setOption("gutters", ["CodeMirror-lint-markers", "CodeMirror-foldgutter"]);

        if ("text/x-go" === data.mode || "application/json" === data.mode) {
            editor.setOption("lint", true);
        }

        if ("application/xml" === data.mode || "text/html" === data.mode) {
            editor.setOption("autoCloseTags", true);
        }

        wide.curEditor = editor;
        editors.data.push({
            "editor": editor,
            "id": id
        });
    }
};

editors.init();