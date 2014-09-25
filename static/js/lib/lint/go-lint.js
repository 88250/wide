goLintFound = [];

(function (mod) {
    mod(CodeMirror);
})(function (CodeMirror) {
    "use strict";
    
    CodeMirror.registerHelper("lint", "go", function (text) {
        return goLintFound;
    });
});
