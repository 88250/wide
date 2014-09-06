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

        CodeMirror.commands.autocompleteRightPart = function(cm) {
            setTimeout(function() {
                var cur = cm.getCursor();
                var curLine = cm.getLine(cur.line);
                var curChar = curLine.charAt(cur.ch - 1);

                replacement = '';

                switch (curChar) {
                    case '(':
                        replacement = ')';
                        break;
                    case '[':
                        replacement = ']';
                        break;
                    case '{':
                        replacement = '}';
                        break;
                    default: // " or '
                        replacement = curChar;
                        break;
                }

                cm.replaceRange(replacement, CodeMirror.Pos(cur.line, cur.ch));
                cm.setCursor(CodeMirror.Pos(cur.line, cur.ch));
            }, 50);

            return CodeMirror.Pass;
        };

        CodeMirror.commands.gotoLine = function(cm) {
            var line = prompt("Go To Line: ", "0");

            cm.setCursor(CodeMirror.Pos(line - 1, 0));
        };

        CodeMirror.commands.doNothing = function(cm) {
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
            title: '<span title="' + wide.curNode.path + '">' + wide.curNode.name + '</span>',
            content: '<textarea id="editor' + id + '"></textarea>'
        });

        rulers = [];
        rulers.push({color: "#ccc", column: 120, lineStyle: "dashed"});

        var editor = CodeMirror.fromTextArea(document.getElementById("editor" + id), {
            lineNumbers: true,
            highlightSelectionMatches: {showToken: /\w/},
            rulers: rulers,
            styleActiveLine: true,
            theme: 'lesser-dark',
            indentUnit: 4,
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
                "'('": "autocompleteRightPart",
                "'['": "autocompleteRightPart",
                "'{'": "autocompleteRightPart",
                "'\"'": "autocompleteRightPart",
                "'''": "autocompleteRightPart",
                "Ctrl-G": "gotoLine",
                "Ctrl-E": "deleteLine",
                "Ctrl-D": "doNothing" // 取消默认的 deleteLine
            }
        });
        editor.setSize('100%', $(".edit-panel").height() - $(".edit-header").height());
        editor.setValue(data.content);
        editor.setOption("mode", data.mode);
                
        editor.setOption("gutters", ["CodeMirror-lint-markers"]);
        
        if ("application/json" === data.mode) {
            editor.setOption("lint", true);
        }

        wide.curEditor = editor;
        editors.data.push({
            "editor": editor,
            "id": id
        });
    }
};

editors.init();