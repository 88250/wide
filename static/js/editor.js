var editors = {
    data: [],
    init: function() {
        editors._initAutocomplete();
        editors._initTabs();
    },
    _initAutocomplete: function() {
        CodeMirror.registerHelper("hint", "go", function(editor, options) {
			var word = /[\w$]+/
			
            var cur = editor.getCursor(), curLine = editor.getLine(cur.line);
			
			var start = cur.ch, end = start;
		    while (end < curLine.length && word.test(curLine.charAt(end))) ++end;
		    while (start && word.test(curLine.charAt(start - 1))) --start;
		    var curWord = start != end && curLine.slice(start, end);
			
			console.log(curWord)

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
		
		
		CodeMirror.registerHelper("hint", "anyword", function(editor, options) {
			var WORD = /[\w$]+/, RANGE = 500;
			
           	var word = options && options.word || WORD;
		    var range = options && options.range || RANGE;
		    var cur = editor.getCursor(), curLine = editor.getLine(cur.line);
		    var start = cur.ch, end = start;
		    while (end < curLine.length && word.test(curLine.charAt(end))) ++end;
		    while (start && word.test(curLine.charAt(start - 1))) --start;
		    var curWord = start != end && curLine.slice(start, end);
		
		    var list = [], seen = {};
		    var re = new RegExp(word.source, "g");
		    for (var dir = -1; dir <= 1; dir += 2) {
		      var line = cur.line, endLine = Math.min(Math.max(line + dir * range, editor.firstLine()), editor.lastLine()) + dir;
		      for (; line != endLine; line += dir) {
		        var text = editor.getLine(line), m;
		        while (m = re.exec(text)) {
		          if (line == cur.line && m[0] === curWord) continue;
		          if ((!curWord || m[0].lastIndexOf(curWord, 0) == 0) && !Object.prototype.hasOwnProperty.call(seen, m[0])) {
		            seen[m[0]] = true;
		            list.push(m[0]);
		          }
		        }
		      }
		    }
			
			console.log(list);
			
		    return {list: list, from: CodeMirror.Pos(cur.line, start), to: CodeMirror.Pos(cur.line, end)};
        });

        CodeMirror.commands.anyword = function(cm) {
            cm.showHint({hint: CodeMirror.hint.anyword, completeSingle: false});
        };
		
		CodeMirror.commands.autocompleteAfterDot = function(cm) {
			var cur = cm.getCursor();
			console.log(cm.getRange(CodeMirror.Pos(cur.line, cur.ch - 1), cur));
			
			var tok = cm.getTokenAt(cm.getCursor());
			console.log(tok)
			
			
			setTimeout(function() {
          		if (!cm.state.completionActive) {
            		cm.showHint({hint: CodeMirror.hint.go, completeSingle: false});
				}
        	}, 100)
			
			return CodeMirror.Pass;
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
                "Ctrl-\\": "anyword",
				".": "autocompleteAfterDot"
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