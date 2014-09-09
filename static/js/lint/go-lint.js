(function(mod) {
    mod(CodeMirror);
})(function(CodeMirror) {
    "use strict";

    CodeMirror.registerHelper("lint", "go", function(text) {
        var found = [];

        found.push({from: CodeMirror.Pos(1, 1),
            to: CodeMirror.Pos(1, 10),
            message: "test commpile err"});

        return found;
    });

});
