/*
 * Copyright (c) 2014-2015, b3log.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * @file frontend tool.
 * 
 * @author <a href="mailto:liliyuan@fangstar.net">Liyuan Li</a>
 * @version 0.1.0.0, Dec 15, 2015 
 */
var gulp = require("gulp");
var concat = require('gulp-concat');
var minifyCSS = require('gulp-minify-css');
var uglify = require('gulp-uglify');
var sourcemaps = require("gulp-sourcemaps");

gulp.task('cc', function () {
    // css
    var cssLibs = ['./static/js/lib/jquery-layout/layout-default-latest.css',
        './static/js/lib/codemirror-5.1/codemirror.css',
        './static/js/lib/codemirror-5.1/addon/hint/show-hint.css',
        './static/js/lib/codemirror-5.1/addon/lint/lint.css',
        './static/js/lib/codemirror-5.1/addon/fold/foldgutter.css',
        './static/js/lib/codemirror-5.1/addon/dialog/dialog.css',
        './static/js/overwrite/codemirror/theme/*.css'];
    gulp.src(cssLibs)
            .pipe(minifyCSS())
            .pipe(concat('lib.min.css'))
            .pipe(gulp.dest('./static/css/'));

    gulp.src('./static/js/lib/ztree/zTreeStyle.css')
            .pipe(minifyCSS())
            .pipe(concat('zTreeStyle.min.css'))
            .pipe(gulp.dest('./static/js/lib/ztree/'));

    var cssWide = ['./static/css/dialog.css',
        './static/css/base.css',
        './static/css/wide.css',
        './static/css/side.css',
        './static/css/start.css',
        './static/css/about.css'
    ];

    gulp.src(cssWide)
            .pipe(minifyCSS())
            .pipe(concat('wide.min.css'))
            .pipe(gulp.dest('./static/css/'));


    // js
    var jsLibs = ['./static/js/lib/jquery-2.1.1.min.js',
        './static/js/lib/jquery-ui.min.js',
        './static/js/lib/jquery-layout/jquery.layout-latest.js',
        './static/js/lib/reconnecting-websocket.js',
        './static/js/lib/Autolinker.min.js',
        './static/js/lib/emmet.js',
        './static/js/lib/js-beautify-1.5.4/beautify.js',
        './static/js/lib/js-beautify-1.5.4/beautify-html.js',
        './static/js/lib/js-beautify-1.5.4/beautify-css.js',
        './static/js/lib/jquery-file-upload-9.8.0/vendor/jquery.ui.widget.js',
        './static/js/lib/jquery-file-upload-9.8.0/jquery.iframe-transport.js',
        './static/js/lib/jquery-file-upload-9.8.0/jquery.fileupload.js',
        './static/js/lib/codemirror-5.1/codemirror.min.js',
        './static/js/lib/codemirror-5.1/addon/lint/lint.js',
        './static/js/lib/codemirror-5.1/addon/lint/json-lint.js',
        './static/js/lib/codemirror-5.1/addon/selection/active-line.js',
        './static/js/lib/codemirror-5.1/addon/selection/active-line.js',
        './static/js/overwrite/codemirror/addon/hint/show-hint.js',
        './static/js/lib/codemirror-5.1/addon/hint/anyword-hint.js',
        './static/js/lib/codemirror-5.1/addon/display/rulers.js',
        './static/js/lib/codemirror-5.1/addon/edit/closebrackets.js',
        './static/js/lib/codemirror-5.1/addon/edit/matchbrackets.js',
        './static/js/lib/codemirror-5.1/addon/edit/closetag.js',
        './static/js/lib/codemirror-5.1/addon/search/searchcursor.js',
        './static/js/lib/codemirror-5.1/addon/search/search.js',
        './static/js/lib/codemirror-5.1/addon/dialog/dialog.js',
        './static/js/lib/codemirror-5.1/addon/search/match-highlighter.js',
        './static/js/lib/codemirror-5.1/addon/fold/foldcode.js',
        './static/js/lib/codemirror-5.1/addon/fold/foldgutter.js',
        './static/js/lib/codemirror-5.1/addon/fold/brace-fold.js',
        './static/js/lib/codemirror-5.1/addon/fold/xml-fold.js',
        './static/js/lib/codemirror-5.1/addon/fold/markdown-fold.js',
        './static/js/lib/codemirror-5.1/addon/fold/comment-fold.js',
        './static/js/lib/codemirror-5.1/addon/fold/mode/loadmode.js',
        './static/js/lib/codemirror-5.1/addon/fold/comment/comment.js',
        './static/js/lib/codemirror-5.1/mode/meta.js',
        './static/js/lib/codemirror-5.1/mode/go/go.js',
        './static/js/lib/codemirror-5.1/mode/clike/clike.js',
        './static/js/lib/codemirror-5.1/mode/xml/xml.js',
        './static/js/lib/codemirror-5.1/mode/htmlmixed/htmlmixed.js',
        './static/js/lib/codemirror-5.1/mode/javascript/javascript.js',
        './static/js/lib/codemirror-5.1/mode/markdown/markdown.js',
        './static/js/lib/codemirror-5.1/mode/css/css.js',
        './static/js/lib/codemirror-5.1/mode/shell/shell.js',
        './static/js/lib/codemirror-5.1/mode/sql/sql.js',
        './static/js/lib/codemirror-5.1/keymap/vim.js',
        './static/js/lib/lint/json-lint.js',
        './static/js/lib/lint/go-lint.js'];
    gulp.src(jsLibs)
            .pipe(uglify())
            .pipe(concat('lib.min.js'))
            .pipe(gulp.dest('./static/js/'));

    var jsWide = ['./static/js/tabs.js',
        './static/js/tabs.js',
        './static/js/dialog.js',
        './static/js/editors.js',
        './static/js/notification.js',
        './static/js/tree.js',
        './static/js/wide.js',
        './static/js/session.js',
        './static/js/menu.js',
        './static/js/windows.js',
        './static/js/hotkeys.js',
        './static/js/bottomGroup.js'
    ];
    gulp.src(jsWide)
            .pipe(sourcemaps.init())
            .pipe(uglify())
            .pipe(concat('wide.min.js'))
            .pipe(sourcemaps.write("."))
            .pipe(gulp.dest('./static/js/'));
});