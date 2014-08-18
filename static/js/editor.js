(function(mod) {
    mod(CodeMirror);
})(function(CodeMirror) {
    "use strict";

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
});

CodeMirror.commands.autocomplete = function(cm) {
    cm.showHint({hint: CodeMirror.hint.go});
};

var editor = CodeMirror.fromTextArea(document.getElementById('code'), {
    lineNumbers: true,
    theme: 'lesser-dark',
    extraKeys: {
        "Ctrl-\\": "autocomplete"
    }
});

editor.setSize('100%', 450);

editor.addKeyMap({
});

editor.on('keyup', function(cm, event) {

});