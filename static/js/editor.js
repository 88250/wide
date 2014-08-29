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
    },
    _initAutocomplete: function() {
        CodeMirror.registerHelper("hint", "go", function(editor) {
            var word = /[\w$]+/;

            var cur = editor.getCursor(), curLine = editor.getLine(cur.line);

            var start = cur.ch, end = start;
            while (end < curLine.length && word.test(curLine.charAt(end)))
                ++end;
            while (start && word.test(curLine.charAt(start - 1)))
                --start;

            var request = {
                code: editor.getValue(),
                cursorLine: editor.getCursor().line,
                cursorCh: editor.getCursor().ch
            };

            // XXX: 回调有问题，暂时不使用 WS 协议
            //editorWS.send(JSON.stringify(request));

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
    },
    newEditor: function(data) {
        var id = wide.curNode.tId;
        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            if (editors.data[i].id === id) {
                editors.tabs.setCurrent(id);
                wide.curEditor = editor;
                return false;
            }
        }

        editors.tabs.add({
            id: id,
            title: '<span title="' + wide.curNode.path + '">' + wide.curNode.name + '</span>',
            content: '<textarea id="editor' + id + '"></textarea>'
        });

        var editor = CodeMirror.fromTextArea(document.getElementById("editor" + id), {
            lineNumbers: true,
            theme: 'lesser-dark',
            indentUnit: 4,
            extraKeys: {
                "Ctrl-\\": "autocompleteAnyWord",
                ".": "autocompleteAfterDot"
            }
        });
        editor.setSize('100%', 430);
        editor.setValue(data.content);
        editor.setOption("mode", data.mode);

        wide.curEditor = editor;
        editors.data.push({
            "editor": editor,
            "id": id
        });
    },
    removeEditor: function() {

    }
};

editors.init();