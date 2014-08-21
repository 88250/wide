var editors = {
    data: [],
    init: function() {
        editors._initAutocomplete();
        editors._initTabs();
    },
    _initAutocomplete: function() {
        CodeMirror.registerHelper("hint", "go", function(editor, options) {
            var cur = editor.getCursor();

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

                    for (var i = 0; i < autocompleteArray.length; i++) {
                        autocompleteHints[i] = autocompleteArray[i].name;
                    }
                }
            });

            return {list: autocompleteHints, from: cur, to: cur};
        });

        CodeMirror.commands.autocomplete = function(cm) {
            cm.showHint({hint: CodeMirror.hint.go});
        };
    },
    _initTabs: function() {
        var $tabsPanel = $(".edit-panel .tabs-panel"),
                $tabs = $(".edit-panel .tabs");

        $tabs.on("click", "span", function() {
            var $it = $(this);
            if ($it.hasClass("current")) {
                return false;
            }

            var id = $it.data("id");

            $tabs.children("span").removeClass("current");
            $tabsPanel.children("div").hide();

            $it.addClass("current");
            $("#editor" + id).parent().show();

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
        });
    },
    _selectTab: function(id, editor) {
        var $tabsPanel = $(".edit-panel .tabs-panel"),
                $tabs = $(".edit-panel .tabs");

        var $currentTab = $tabs.children(".current");
        if ($currentTab.data("id") === id) {
            return false;
        }

        $tabs.children("span").removeClass("current");
        $tabsPanel.children("div").hide();

        $tabs.children("span[data-id='" + id + "']").addClass("current");
        $("#editor" + id).parent().show();
        wide.curEditor = editor;
    },
    newEditor: function(data) {
        var id = wide.curNode.tId;
        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            if (editors.data[i].id === id) {
                editors._selectTab(id, editors.data[i].editor);
                return false;
            }
        }

        var $tabsPanel = $(".edit-panel .tabs-panel"),
                $tabs = $(".edit-panel .tabs");

        $tabs.children("span").removeClass("current");
        $tabsPanel.children("div").hide();

        $tabsPanel.append('<div><textarea id="editor' + id + '" name="code"></textarea></div>');
        $tabs.append('<span data-id="' + id + '" class="current">' + wide.curNode.name + '</span>');

        var editor = CodeMirror.fromTextArea(document.getElementById("editor" + id), {
            lineNumbers: true,
            theme: 'lesser-dark',
			indentUnit: 4,
            extraKeys: {
                "Ctrl-\\": "autocomplete"
            }
        });
        editor.setSize('100%', 450);
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