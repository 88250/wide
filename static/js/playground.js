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

var playground = {
    editor: undefined,
    pid: undefined,
    _resize: function () {
        $("#editor, #output").height($(window).height() - 57 - $(".footer").height());
        playground.editor.setSize("auto", $("#editor").height() + "px");
    },
    _initShare: function () {
        $("#dialogShare").dialog({
            "modal": true,
            "height": 460,
            "width": 800,
            "title": config.label.share,
            "hideFooter": true,
            "afterOpen": function () {
            }
        });
    },
    _initWideShare: function () {
        $(".share-panel .font-ico").click(function () {
            var key = $(this).attr('class').split('-')[2];
            var url = "https://wide.b3log.org", pic = 'https://wide.b3log.org/static/images/wide-logo.png';
            var urls = {};
            urls.email = "mailto:?subject=" + $('title').text()
                    + "&body=" + $('meta[name=description]').attr('content') + ' ' + url;

            var twitterShare = encodeURIComponent($('meta[name=description]').attr('content') + " " + url + " #golang");
            urls.twitter = "https://twitter.com/intent/tweet?status=" + twitterShare;

            urls.facebook = "https://www.facebook.com/sharer/sharer.php?u=" + url;
            urls.googleplus = "https://plus.google.com/share?url=" + url;

            var title = encodeURIComponent($('title').text() + '. \n' + $('meta[name=description]').attr('content')
                    + " #golang#");
            urls.weibo = "http://v.t.sina.com.cn/share/share.php?title=" + title + "&url=" + url + "&pic=" + pic;
            urls.tencent = "http://share.v.t.qq.com/index.php?c=share&a=index&title=" + title +
                    "&url=" + url + "&pic=" + pic;

            window.open(urls[key], "_blank", "top=100,left=200,width=648,height=618");

            $(".menu .share-panel").hide();
        });
    },
    init: function () {
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
            request.code = editor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            var autocompleteHints = [];

            $.ajax({
                async: false, // 同步执行
                type: 'POST',
                url: config.context + '/playground/autocomplete',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (data) {
                    var autocompleteArray = data[1];

                    if (autocompleteArray) {
                        for (var i = 0; i < autocompleteArray.length; i++) {
                            var displayText = '',
                                    text = autocompleteArray[i].name;

                            switch (autocompleteArray[i].class) {
                                case "type":
                                    displayText = '<span class="fn-clear"><span class="ico-type ico"></span>'// + autocompleteArray[i].class 
                                            + '<b>' + autocompleteArray[i].name + '</b>    '
                                            + autocompleteArray[i].type + '</span>';
                                    break;
                                case "const":
                                    displayText = '<span class="fn-clear"><span class="ico-const ico"></span>'// + autocompleteArray[i].class 
                                            + '<b>' + autocompleteArray[i].name + '</b>    '
                                            + autocompleteArray[i].type + '</span>';
                                    break;
                                case "var":
                                    displayText = '<span class="fn-clear"><span class="ico-var ico"></span>'// + autocompleteArray[i].class 
                                            + '<b>' + autocompleteArray[i].name + '</b>    '
                                            + autocompleteArray[i].type + '</span>';
                                    break;
                                case "package":
                                    displayText = '<span class="fn-clear"><span class="ico-package ico"></span>'// + autocompleteArray[i].class 
                                            + '<b>' + autocompleteArray[i].name + '</b>    '
                                            + autocompleteArray[i].type + '</span>';
                                    break;
                                case "func":
                                    displayText = '<span><span class="ico-func ico"></span>'// + autocompleteArray[i].class 
                                            + '<b>' + autocompleteArray[i].name + '</b>'
                                            + autocompleteArray[i].type.substring(4) + '</span>';
                                    text += '()';
                                    break;
                                default:
                                    console.warn("Can't handle autocomplete [" + autocompleteArray[i].class + "]");
                                    break;
                            }

                            autocompleteHints[i] = {
                                displayText: displayText,
                                text: text
                            };
                        }
                    }
                }
            });

            return {list: autocompleteHints, from: CodeMirror.Pos(cur.line, start), to: CodeMirror.Pos(cur.line, end)};
        });

        CodeMirror.commands.autocompleteAnyWord = function (cm) {
            cm.showHint({hint: CodeMirror.hint.auto});
        };

        CodeMirror.commands.autocompleteAfterDot = function (cm) {
            var mode = cm.getMode();
            if (mode && "go" !== mode.name) {
                return CodeMirror.Pass;
            }

            var token = cm.getTokenAt(cm.getCursor());

            if ("comment" === token.type || "string" === token.type) {
                return CodeMirror.Pass;
            }

            setTimeout(function () {
                if (!cm.state.completionActive) {
                    cm.showHint({hint: CodeMirror.hint.go, completeSingle: false});
                }
            }, 50);

            return CodeMirror.Pass;
        };

        playground.editor = CodeMirror.fromTextArea($("#editor")[0], {
            lineNumbers: true,
            autofocus: true,
            autoCloseBrackets: true,
            matchBrackets: true,
            highlightSelectionMatches: {showToken: /\w/},
            rulers: [{color: "#ccc", column: 80, lineStyle: "dashed"}],
            styleActiveLine: true,
            theme: "wide",
            tabSize: 4,
            indentUnit: 4,
            foldGutter: true,
            cursorHeight: 1,
            viewportMargin: 500,
            extraKeys: {
                "Ctrl-\\": "autocompleteAnyWord",
                ".": "autocompleteAfterDot"
            }
        });

        $("#editorDiv").show();

        $(window).resize(function () {
            playground._resize();
        });

        var hovered = false;
        $(".menu .ico-share").hover(function () {
            $(".menu .share-panel").show();
            hovered = true;
        }, function () {
            if (!hovered) {
                $(".menu .share-panel").hide();
            }

            hovered = false;
            setTimeout(function () {
                if (!hovered) {
                    $(".menu .share-panel").hide();
                }
            }, 500);
        });

        $(".menu .share-panel").hover(function () {
            $(".menu .share-panel").show();
            hovered = true;
        }, function () {
            $(".menu .share-panel").hide();
            hovered = false;
        });

        playground.editor.on('changes', function (cm) {
            $("#url").html("");
        });

        this._initWS();
        this._resize();
        this._initWideShare();
        this._initShare();
        menu._initAbout();
    },
    _initWS: function () {
        // Used for session retention, server will release all resources of the session if this channel closed
        var sessionWS = new ReconnectingWebSocket(config.channel + '/session/ws?sid=' + config.wideSessionId);

        sessionWS.onopen = function () {
            console.log('[session onopen] connected');
        };

        sessionWS.onmessage = function (e) {
            console.log('[session onmessage]' + e.data);
        };
        sessionWS.onclose = function (e) {
            console.log('[session onclose] disconnected (' + e.code + ')');
        };
        sessionWS.onerror = function (e) {
            console.log('[session onerror] ' + JSON.parse(e));
        };

        var playgroundWS = new ReconnectingWebSocket(config.channel + '/playground/ws?sid=' + config.wideSessionId);

        playgroundWS.onopen = function () {
            console.log('[playground onopen] connected');
        };

        playgroundWS.onmessage = function (e) {
            console.log('[playground onmessage]' + e.data);

            var data = JSON.parse(e.data);

            if ("init-playground" === data.cmd) {
                return;
            }

            playground.pid = data.pid;

            var val = $("#output").val();
            $("#output").val(val + data.output);
        };
        playgroundWS.onclose = function (e) {
            console.log('[playground onclose] disconnected (' + e.code + ')');
        };
        playgroundWS.onerror = function (e) {
            console.log('[playground onerror] ' + JSON.parse(e));
        };
    },
    share: function () {
        if (!playground.editor) {
            return;
        }

        var code = playground.editor.getValue();

        var request = newWideRequest();
        request.code = code;

        $.ajax({
            type: 'POST',
            url: config.context + '/playground/save',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                playground.editor.setValue(data.code);

                if (!data.succ) {
                    return;
                }

                var url = window.location.protocol + "//" + window.location.host + '/playground/' + data.fileName;
                var html = 'URL: <a href="' + url + '" target="_blank">' + url + "</a><br/>";
                html += "Embeded: xxxx";

                $("#dialogShare").html(html);
                $("#dialogShare").dialog("open");
            }
        });
    },
    stop: function () {
        if (!playground.editor) {
            return;
        }

        var cursor = playground.editor.getCursor();
        playground.editor.focus();

        playground.editor.setCursor(cursor);

        if (!playground.pid) {
            return;
        }

        var request = newWideRequest();
        request.pid = playground.pid;

        $.ajax({
            type: 'POST',
            url: config.context + '/playground/stop',
            data: JSON.stringify(request),
            dataType: "json"
        });
    },
    run: function () {
        if (!playground.editor) {
            return;
        }

        var cursor = playground.editor.getCursor();
        playground.editor.focus();

        var code = playground.editor.getValue();

        // Step 1. save & format code
        var request = newWideRequest();
        request.code = code;

        $("#output").val("");

        $.ajax({
            type: 'POST',
            url: config.context + '/playground/save',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                // console.log(data);
                playground.editor.setValue(data.code);
                playground.editor.setCursor(cursor);

                if (!data.succ) {
                    return;
                }

                // Step 2. compile code
                var request = newWideRequest();
                request.fileName = data.fileName;

                $.ajax({
                    type: 'POST',
                    url: config.context + '/playground/build',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (data) {
                        // console.log(data);

                        $("#output").val(data.output);

                        if (!data.succ) {
                            return;
                        }

                        // Step 3. run the executable binary and handle its output
                        var request = newWideRequest();
                        request.executable = data.executable;

                        $.ajax({
                            type: 'POST',
                            url: config.context + '/playground/run',
                            data: JSON.stringify(request),
                            dataType: "json",
                            success: function (data) {
                                // console.log(data);
                            }
                        });
                    }
                });
            }
        });
    },
    format: function () {
        if (!playground.editor) {
            return;
        }

        var cursor = playground.editor.getCursor();
        playground.editor.focus();

        var code = playground.editor.getValue();

        var request = newWideRequest();
        request.code = code;

        $.ajax({
            type: 'POST',
            url: config.context + '/playground/save',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (data) {
                playground.editor.setValue(data.code);
                playground.editor.setCursor(cursor);
            }
        });
    }
};

$(document).ready(function () {
    playground.init();
});
            