/*
 * Copyright (c) 2014-2016, b3log.org & hacpai.com
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

/*
 * @file playground.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @author <a href="http://88250.b3log.org">Liang Ding</a>
 * @version 1.0.0.1, Dec 8, 2015
 */
var playground = {
    autocompleteMutex: false,
    editor: undefined,
    pid: undefined,
    _resize: function () {
        if (config.disqus) {
            return false;
        }
        if (config.embed) {
            $("#editorDiv").parent().height($(window).height() - 33 - $(".footer").height());
            playground.editor.setSize("auto", ($("#editorDiv").parent().height() * 0.7) + "px");
        } else {
            $("#editor, #output").height($(window).height() - 60 - $(".footer").height());
            playground.editor.setSize("auto", $("#editor").height() + "px");
        }
    },
    _initShare: function () {
        $("#dialogShare").dialog({
            "modal": true,
            "title": config.label.share,
            "hideFooter": true
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

            if (playground.autocompleteMutex && editor.state.completionActive) {
                console.log(1);
                return;
            }

            playground.autocompleteMutex = true;

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

            setTimeout(function () {
                playground.autocompleteMutex = false;
            }, 20);

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
            autoCloseBrackets: true,
            matchBrackets: true,
            highlightSelectionMatches: {showToken: /\w/},
            rulers: [{color: "#ccc", column: 80, lineStyle: "dashed"}],
            styleActiveLine: true,
            theme: "wide",
            tabSize: 4,
            indentUnit: 4,
            indentWithTabs: true,
            foldGutter: true,
            cursorHeight: 1,
            viewportMargin: 500,
            extraKeys: {
                "Ctrl-\\": "autocompleteAnyWord",
                ".": "autocompleteAfterDot"
            }
        });

        playground.editor.setOption("gutters", ["CodeMirror-lint-markers", "CodeMirror-foldgutter"]);

        $(window).resize(function () {
            playground._resize();
        });

        if (config.embed) {
            playground.editor.setSize("auto", ($("#editorDiv").parent().height() * 0.7) + "px");
        } else {
            playground.editor.setSize("auto", $("#editor").height() + "px");
        }

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

        playground.editor.on('keydown', function (cm, evt) {
            if (evt.altKey || evt.ctrlKey || evt.shiftKey) {
                return;
            }

            var k = evt.which;

            if (k < 48) {
                return;
            }

            // hit [0-9]

            if (k > 57 && k < 65) {
                return;
            }

            // hit [a-z]

            if (k > 90) {
                return;
            }

            if (config.autocomplete) {
                if (0.5 <= Math.random()) {
                    CodeMirror.commands.autocompleteAfterDot(cm);
                }
            }
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
            success: function (result) {
                var data = result.data;
                
                playground.editor.setValue(data.code);

                if (!result.succ) {
                    console.log(data);
                    return;
                }

                var url = window.location.protocol + "//" + window.location.host + '/playground/' + data.fileName;

                var request = newWideRequest();
                request.url = url;
                $.ajax({
                    type: 'POST',
                    url: config.context + '/playground/short-url',
                    data: JSON.stringify(request),
                    dataType: "json",
                    success: function (result) {
                        if (!result.succ) {
                            console.log(result);
                            return;
                        }

                        var html = '<div class="fn-clear"><label>' + config.label.url
                                + config.label.colon + '</label><a href="'
                                + url + '" target="_blank">' + url + "</a><br/>";
                        html += '<label>' + config.label.short_url + config.label.colon
                                + '</label><a href="' + result.data + '" target="_blank">'
                                + result.data + '</a><br/>';
                        html += '<label>' + config.label.embeded + config.label.colon
                                + '</label><br/><textarea rows="5" style="width:100%" readonly><iframe style="border:1px solid" src="'
                                + url + '?embed=true" width="99%" height="600"></iframe></textarea>';
                        html += '</div>';

                        $("#dialogShare").html(html);
                        $("#dialogShare").dialog("open");
                    }});
            }
        });
    },
    disqus: function () {
        var url = window.location.href;
        if (url.indexOf("?") >= 0) {
            if (url.indexOf("disqus=") >= 0) {
                url = url.replace("disqus=false", "disqus=true");
                console.log(url);
            } else {
                url += "&disqus=true";
            }
        } else {
            url += "?disqus=true";
        }

        window.location.href = url;
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
            success: function (result) {
                var data = result.data;
                
                playground.editor.setValue(data.code);
                playground.editor.setCursor(cursor);

                if (!result.succ) {
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
                    success: function (result) {
                        console.log(result);
                        
                        var data = result.data;

                        $("#output").val(data.output);

                        if (!result.succ) {
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
                            success: function (result) {
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
            success: function (result) {
                playground.editor.setValue(result.data.code);
                playground.editor.setCursor(cursor);
            }
        });
    }
};

$(document).ready(function () {
    playground.init();
});
            