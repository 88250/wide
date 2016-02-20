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
 * @file editor.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @author <a href="http://88250.b3log.org">Liang Ding</a>
 * @version 1.1.1.0, Jan 12, 2016
 */
var editors = {
    autocompleteMutex: false,
    data: [],
    tabs: {},
    getEditorByPath: function (path) {
        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            if (editors.data[i].editor.options.path === path) {
                return editors.data[i].editor;
            }
        }
    },
    close: function () {
        $('.edit-panel .tabs > div[data-index="' + $('.edit-panel .frame').data('index') + ']').find('.ico-close').click();
    },
    closeOther: function () {
        var currentIndex = $(".edit-panel .frame").data("index");

        // 设置全部关闭标识
        var removeData = [];
        $(".edit-panel .tabs > div").each(function (i) {
            if (currentIndex !== $(this).data("index")) {
                removeData.push($(this).data("index"));
            }
        });
        if (removeData.length === 0) {
            return false;
        }
        var firstIndex = removeData.splice(0, 1);
        $("#dialogCloseEditor").data("removeData", removeData);
        // 开始关闭
        $('.edit-panel .tabs > div[data-index="' + firstIndex + '"]').find(".ico-close").click();
    },
    _removeAllMarker: function () {
        var removeData = $("#dialogCloseEditor").data("removeData");
        if (removeData && removeData.length > 0) {
            var removeIndex = removeData.splice(0, 1);
            $("#dialogCloseEditor").data("removeData", removeData);
            $('.edit-panel .tabs > div[data-index="' + removeIndex + '"] .ico-close').click();
        }
        if (wide.curEditor) {
            wide.curEditor.focus();
        }
    },
    _initClose: function () {
        new ZeroClipboard($("#copyFilePath"));

        // 关闭、关闭其他、关闭所有
        $(".edit-panel").on("mouseup", '.tabs > div', function (event) {
            event.stopPropagation();

            if (event.button === 0) { // 左键
                $(".edit-panel .frame").hide();
                return false;
            }

            // event.button === 2 右键
            var left = event.screenX;
            if ($(".side").css("left") === "auto" || $(".side").css("left") === "0px") {
                left = event.screenX - $(".side").width();
            }
            $(".edit-panel .frame").show().css({
                "left": left + "px",
                "top": "21px"
            }).data('index', $(this).data("index"));

            $("#copyFilePath").attr('data-clipboard-text', $(this).find("span:eq(0)").attr("title"));
            return false;
        });
    },
    init: function () {
        $("#dialogCloseEditor").dialog({
            "modal": true,
            "height": 90,
            "width": 260,
            "title": config.label.tip,
            "hideFooter": true,
            "afterOpen": function (fileName) {
                $("#dialogCloseEditor > div:eq(0)").html(config.label.file
                        + ' <b>' + fileName + '</b>. ' + config.label.confirm_save + '?');
                $("#dialogCloseEditor button:eq(0)").focus();
            },
            "afterInit": function () {
                $("#dialogCloseEditor button.save").click(function () {
                    var i = $("#dialogCloseEditor").data("index");
                    wide.fmt(editors.data[i].id, editors.data[i].editor);
                    editors.tabs.del(editors.data[i].id);
                    $("#dialogCloseEditor").dialog("close");
                    editors._removeAllMarker();
                });

                $("#dialogCloseEditor button.discard").click(function () {
                    var i = $("#dialogCloseEditor").data("index");
                    editors.tabs.del(editors.data[i].id);
                    $("#dialogCloseEditor").dialog("close");
                    editors._removeAllMarker();
                });

                $("#dialogCloseEditor button.cancel").click(function (event) {
                    $("#dialogCloseEditor").dialog("close");
                    editors._removeAllMarker();
                });
            }
        });

        editors.tabs = new Tabs({
            id: ".edit-panel",
            setAfter: function () {
                if (wide.curEditor) {
                    wide.curEditor.focus();
                }
            },
            clickAfter: function (id) {
                if (id === 'startPage') {
                    wide.curEditor = undefined;
                    $(".footer .cursor").text('');
                    wide.refreshOutline();
                    
                    return false;
                }
            },
            removeBefore: function (id) {
                if (id === 'startPage') { // 当前关闭的 tab 是起始页
                    editors._removeAllMarker();
                    return true;
                }

                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === id) {
                        if (editors.data[i].editor.doc.isClean()) {
                            editors._removeAllMarker();
                            return true;
                        } else {
                            $("#dialogCloseEditor").dialog("open", $('.edit-panel .tabs > div[data-index="'
                                    + editors.data[i].id + '"] > span:eq(0)').text());
                            $("#dialogCloseEditor").data("index", i);
                            return false;
                        }

                        break;
                    }
                }
            },
            removeAfter: function (id, nextId) {
                if ($(".edit-panel .tabs > div").length === 0) {
                    // 全部 tab 都关闭时才 disables 菜单中“全部关闭”的按钮
                    menu.disabled(['close-all']);
                }

                // 移除编辑器
                for (var i = 0, ii = editors.data.length; i < ii; i++) {
                    if (editors.data[i].id === id) {
                        editors.data.splice(i, 1);
                        break;
                    }
                }

                if (editors.data.length === 0) { // 起始页可能存在，所以用编辑器数据判断
                    menu.disabled(['save-all', 'build', 'run', 'go-test', 'go-vet', 'go-get', 'go-install',
                        'find', 'find-next', 'find-previous', 'replace', 'replace-all',
                        'format', 'autocomplete', 'jump-to-decl', 'expr-info', 'find-usages', 'toggle-comment',
                        'edit']);

                    // remove selected tree node
                    tree.fileTree.cancelSelectedNode();
                    wide.curNode = undefined;
                    wide.curEditor = undefined;
                    wide.refreshOutline();
                    $(".footer .cursor").text('');
                    return false;
                }

                if (!nextId) {
                    // 编辑器区域不存在打开的 Tab
                    // remove selected tree node
                    tree.fileTree.cancelSelectedNode();
                    wide.curNode = undefined;
                    wide.curEditor = undefined;
                    wide.refreshOutline();
                    $(".footer .cursor").text('');
                    return false;
                }

                if (nextId === editors.tabs.getCurrentId()) {
                    // 关闭的不是当前编辑器
                    return false;
                }
            }
        });

        this._initCodeMirrorHotKeys();
        this.openStartPage();
        this._initClose();
    },
    openStartPage: function () {
        wide.curEditor = undefined;
        wide.refreshOutline();
        $(".footer .cursor").text('');

        var dateFormat = function (time, fmt) {
            var date = new Date(time);
            var dateObj = {
                "M+": date.getMonth() + 1, //月份 
                "d+": date.getDate(), //日 
                "h+": date.getHours(), //小时 
                "m+": date.getMinutes(), //分 
                "s+": date.getSeconds(), //秒 
                "q+": Math.floor((date.getMonth() + 3) / 3), //季度 
                "S": date.getMilliseconds() //毫秒 
            };
            if (/(y+)/.test(fmt))
                fmt = fmt.replace(RegExp.$1, (date.getFullYear() + "").substr(4 - RegExp.$1.length));
            for (var k in dateObj)
                if (new RegExp("(" + k + ")").test(fmt)) {
                    fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1)
                            ? (dateObj[k]) : (("00" + dateObj[k]).substr(("" + dateObj[k]).length)));
                }
            return fmt;
        };

        editors.tabs.add({
            id: "startPage",
            title: '<span title="' + config.label.start_page
                    + '"><span class="ico-start font-ico"></span> ' + config.label.start_page + '</span>',
            content: '<div id="startPage"></div>',
            after: function () {
                $("#startPage").load(config.context + '/start?sid=' + config.wideSessionId);
                $.ajax({
                    url: "https://hacpai.com/apis/articles?tags=wide,golang&p=1&size=20",
                    type: "GET",
                    dataType: "jsonp",
                    jsonp: "callback",
                    success: function (data, textStatus) {
                        var articles = data.articles;
                        if (0 === articles.length) {
                            return;
                        }

                        // 按 size = 20 取，但只保留最多 9 篇
                        var length = articles.length;
                        if (length > 9) {
                            length = 9;
                        }

                        var listHTML = "<ul><li class='title'>" + config.label.community + "</li>";
                        for (var i = 0; i < length; i++) {
                            var article = articles[i];
                            listHTML += "<li>"
                                    + "<a target='_blank' href='"
                                    + article.articlePermalink + "'>"
                                    + article.articleTitle + "</a>&nbsp; <span class='date'>"
                                    + dateFormat(article.articleCreateTime, 'yyyy-MM-dd');
                            +"</span></li>";
                        }

                        $("#startPage .news").html(listHTML + "</ul>");
                    }
                });
            }
        });
    },
    getCurrentId: function () {
        var ret = editors.tabs.getCurrentId();
        if (ret === 'startPage') {
            ret = null;
        }
        
        return ret;
    },
    getCurrentPath: function () {
        var currentPath = $(".edit-panel .tabs .current span:eq(0)").attr("title");
        if (currentPath === config.label.start_page) {
            currentPath = null;
        }
        return currentPath;
    },
    _initCodeMirrorHotKeys: function () {
        CodeMirror.registerHelper("hint", "go", function (editor) {
            editor = wide.curEditor; // 使用当前编辑器覆盖实参，因为异步调用的原因，实参不一定正确
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
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = editor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            var autocompleteHints = [];

            if (editors.autocompleteMutex && editor.state.completionActive) {
                return;
            }

            editors.autocompleteMutex = true;

            $.ajax({
                async: false, // 同步执行
                type: 'POST',
                url: config.context + '/autocomplete',
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

                    editor.doc.markClean();
                    $(".edit-panel .tabs .current > span:eq(0)").removeClass("changed");
                }
            });

            setTimeout(function () {
                editors.autocompleteMutex = false;
            }, 20);

            return {list: autocompleteHints, from: CodeMirror.Pos(cur.line, start), to: CodeMirror.Pos(cur.line, end)};
        });

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

        CodeMirror.commands.autocompleteAnyWord = function (cm) {
            cm.showHint({hint: CodeMirror.hint.auto});
        };

        CodeMirror.commands.gotoLine = function (cm) {
            $("#dialogGoLinePrompt").dialog("open");
        };

        // 用于覆盖 cm 默认绑定的某些快捷键功能.
        CodeMirror.commands.doNothing = function (cm) {
        };

        CodeMirror.commands.exprInfo = function (cm) {
            var cur = wide.curEditor.getCursor();

            var request = newWideRequest();
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = wide.curEditor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            $.ajax({
                type: 'POST',
                url: config.context + '/exprinfo',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (result) {
                    if (!result.succ) {
                        return;
                    }
                    
                    var position = wide.curEditor.cursorCoords();
                    $("body").append('<div style="top:'
                            + (position.top + 15) + 'px;left:' + position.left
                            + 'px" class="edit-exprinfo">' + result.data + '</div>');
                }
            });
        };

        CodeMirror.commands.copyLinesDown = function (cm) {
            var content = '',
                    selectoion = cm.listSelections()[0];

            var from = selectoion.anchor,
                    to = selectoion.head;
            if (from.line > to.line) {
                from = selectoion.head;
                to = selectoion.anchor;
            }

            for (var i = from.line, max = to.line; i <= max; i++) {
                if (to.ch !== 0 || i !== max) { // 下一行选中为0时，不应添加内容
                    content += '\n' + cm.getLine(i);
                }
            }
            // 下一行选中为0时，应添加到上一行末
            var replaceToLine = to.line;
            if (to.ch === 0) {
                replaceToLine = to.line - 1;
            }
            cm.replaceRange(content, CodeMirror.Pos(replaceToLine));

            var offset = replaceToLine - from.line + 1;
            cm.setSelection(CodeMirror.Pos(from.line + offset, from.ch),
                    CodeMirror.Pos(to.line + offset, to.ch));
        };

        CodeMirror.commands.copyLinesUp = function (cm) {
            var content = '',
                    selectoion = cm.listSelections()[0];

            var from = selectoion.anchor,
                    to = selectoion.head;
            if (from.line > to.line) {
                from = selectoion.head;
                to = selectoion.anchor;
            }

            for (var i = from.line, max = to.line; i <= max; i++) {
                if (to.ch !== 0 || i !== max) { // 下一行选中为0时，不应添加内容
                    content += '\n' + cm.getLine(i);
                }
            }

            // 下一行选中为0时，应添加到上一行末
            var replaceToLine = to.line;
            if (to.ch === 0) {
                replaceToLine = to.line - 1;
            }
            cm.replaceRange(content, CodeMirror.Pos(replaceToLine));

            cm.setSelection(CodeMirror.Pos(from.line, from.ch),
                    CodeMirror.Pos(to.line, to.ch));
        };

        CodeMirror.commands.moveLinesUp = function (cm) {
            var selectoion = cm.listSelections()[0];

            var from = selectoion.anchor,
                    to = selectoion.head;
            if (from.line > to.line) {
                from = selectoion.head;
                to = selectoion.anchor;
            }

            if (from.line === 0) {
                return false;
            }
            // 下一行选中为0时，应添加到上一行末
            var replaceToLine = to.line;
            if (to.ch === 0) {
                replaceToLine = to.line - 1;
            }
            cm.replaceRange('\n' + cm.getLine(from.line - 1), CodeMirror.Pos(replaceToLine));
            if (from.line === 1) {
                // 移除第一行的换行
                cm.replaceRange('', CodeMirror.Pos(0, 0),
                        CodeMirror.Pos(1, 0));
            } else {
                cm.replaceRange('', CodeMirror.Pos(from.line - 2, cm.getLine(from.line - 2).length),
                        CodeMirror.Pos(from.line - 1, cm.getLine(from.line - 1).length));
            }

            cm.setSelection(CodeMirror.Pos(from.line - 1, from.ch),
                    CodeMirror.Pos(to.line - 1, to.ch));
        };

        CodeMirror.commands.moveLinesDown = function (cm) {
            var selectoion = cm.listSelections()[0];

            var from = selectoion.anchor,
                    to = selectoion.head;
            if (from.line > to.line) {
                from = selectoion.head;
                to = selectoion.anchor;
            }

            if (to.line === cm.lastLine()) {
                return false;
            }

            // 下一行选中为0时，应添加到上一行末
            var replaceToLine = to.line;
            if (to.ch === 0) {
                replaceToLine = to.line - 1;
            }
            // 把选中的下一行添加到选中区域的上一行
            if (from.line === 0) {
                cm.replaceRange(cm.getLine(replaceToLine + 1) + '\n', CodeMirror.Pos(0, 0));
            } else {
                cm.replaceRange('\n' + cm.getLine(replaceToLine + 1), CodeMirror.Pos(from.line - 1));
            }
            // 删除选中的下一行
            cm.replaceRange('', CodeMirror.Pos(replaceToLine + 1, cm.getLine(replaceToLine + 1).length),
                    CodeMirror.Pos(replaceToLine + 2, cm.getLine(replaceToLine + 2).length));

            cm.setSelection(CodeMirror.Pos(from.line + 1, from.ch),
                    CodeMirror.Pos(to.line + 1, to.ch));
        };

        CodeMirror.commands.jumpToDecl = function (cm) {
            var cur = wide.curEditor.getCursor();

            var request = newWideRequest();
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = wide.curEditor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            $.ajax({
                type: 'POST',
                url: config.context + '/find/decl',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (result) {
                    if (!result.succ) {
                        return;
                    }
                    
                    var data = result.data;

                    var tId = tree.getTIdByPath(data.path);
                    wide.curNode = tree.fileTree.getNodeByTId(tId);
                    tree.fileTree.selectNode(wide.curNode);

                    tree.openFile(wide.curNode, CodeMirror.Pos(data.cursorLine - 1, data.cursorCh - 1));
                }
            });
        };

        CodeMirror.commands.findUsages = function (cm) {
            var cur = wide.curEditor.getCursor();

            var request = newWideRequest();
            request.path = $(".edit-panel .tabs .current > span:eq(0)").attr("title");
            request.code = wide.curEditor.getValue();
            request.cursorLine = cur.line;
            request.cursorCh = cur.ch;

            $.ajax({
                type: 'POST',
                url: config.context + '/find/usages',
                data: JSON.stringify(request),
                dataType: "json",
                success: function (result) {
                    if (!result.succ) {
                        return;
                    }

                    editors.appendSearch(result.data, 'usages', '');
                }
            });
        };

        CodeMirror.commands.selectIdentifier = function (cm) {
            var cur = cm.getCursor();
            var word = cm.findWordAt(cur);
            cm.extendSelection(word.anchor, word.head);
        };
    },
    appendSearch: function (data, type, key) {
        var searcHTML = '<ul class="list">',
                key = key.toLowerCase();

        for (var i = 0, ii = data.length; i < ii; i++) {
            var contents = '',
                    lowerCaseContents = data[i].contents[0].toLowerCase(),
                    matches = lowerCaseContents.split(key),
                    startIndex = 0,
                    endIndex = 0;
            for (var j = 0, max = matches.length; j < max; j++) {
                startIndex = endIndex + matches[j].length;
                endIndex = startIndex + key.length;
                var keyWord = data[i].contents[0].substring(startIndex, endIndex);
                if (keyWord !== '') {
                    keyWord = '<b>' + keyWord + '</b>';
                }
                contents += data[i].contents[0].substring(startIndex - matches[j].length, startIndex) + keyWord;
            }

            searcHTML += '<li title="' + data[i].path + '">'
                    + contents + "&nbsp;&nbsp;&nbsp;&nbsp;<span class='ft-small'>" + data[i].path
                    + '<i class="position" data-line="'
                    + data[i].line + '" data-ch="' + data[i].ch + '"> (' + data[i].line + ':'
                    + data[i].ch + ')</i></span></li>';
        }

        if (data.length === 0) {
            searcHTML += '<li>' + config.label.search_no_match + '</li>';
        }
        searcHTML += '</ul>';

        var $search = $('.bottom-window-group .search'),
                title = config.label.find_usages;
        if (type === "founds") {
            title = config.label.search_text;
        }
        if ($search.find("ul").length === 0) {
            bottomGroup.searchTab = new Tabs({
                id: ".bottom-window-group .search",
                removeAfter: function (id, prevId) {
                    if ($search.find("ul").length === 1) {
                        $search.find(".tabs").hide();
                    }
                }
            });

            $search.on("click", "li", function () {
                $search.find("li").removeClass("selected");
                $(this).addClass("selected");
            });

            $search.on("dblclick", "li", function () {
                var $it = $(this),
                        tId = tree.getTIdByPath($it.attr("title"));
                tree.openFile(tree.fileTree.getNodeByTId(tId));
                tree.fileTree.selectNode(wide.curNode);

                var line = $it.find(".position").data("line") - 1;
                var cursor = CodeMirror.Pos(line, $it.find(".position").data("ch") - 1);


                var editor = wide.curEditor;
                editor.setCursor(cursor);

                var half = Math.floor(editor.getScrollInfo().clientHeight / editor.defaultTextHeight() / 2);
                var cursorCoords = editor.cursorCoords({line: cursor.line - half, ch: 0}, "local");
                editor.scrollTo(0, cursorCoords.top);

                wide.curEditor.focus();
            });

            $search.find(".tabs-panel > div").append(searcHTML);

            $search.find(".tabs .first").text(title);
        } else {
            $search.find(".tabs").show();
            bottomGroup.searchTab.add({
                "id": "search" + (new Date()).getTime(),
                "title": title,
                "content": searcHTML
            });
        }

        // focus
        bottomGroup.tabs.setCurrent("search");
        windows.flowBottom();
        $(".bottom-window-group .search").focus();
    },
    // 新建一个编辑器 Tab，如果已经存在 Tab 则切换到该 Tab.
    newEditor: function (data, cursor) {
        var id = wide.curNode.id;

        editors.tabs.add({
            id: id,
            title: '<span title="' + wide.curNode.path + '"><span class="'
                    + wide.curNode.iconSkin + 'ico"></span>' + wide.curNode.name + '</span>',
            content: '<textarea id="editor' + id + '"></textarea>'
        });

        menu.undisabled(['save-all', 'close-all', 'build', 'run', 'go-test', 'go-vet', 'go-get', 'go-install',
            'find', 'find-next', 'find-previous', 'replace', 'replace-all',
            'format', 'autocomplete', 'jump-to-decl', 'expr-info', 'find-usages', 'toggle-comment',
            'edit']);

        var textArea = document.getElementById("editor" + id);
        textArea.value = data.content;

        var editor = CodeMirror.fromTextArea(textArea, {
            lineNumbers: true,
            autofocus: true,
            autoCloseBrackets: true,
            matchBrackets: true,
            highlightSelectionMatches: {showToken: /\w/},
            rulers: [{color: "#ccc", column: 120, lineStyle: "dashed"}],
            styleActiveLine: true,
            theme: config.editorTheme,
            tabSize: config.editorTabSize,
            indentUnit: 4,
            indentWithTabs: true,
            foldGutter: true,
            cursorHeight: 1,
            path: data.path,
            readOnly: wide.curNode.isGOAPI,
            profile: 'xhtml', // define Emmet output profile
            extraKeys: {
                "Ctrl-\\": "autocompleteAnyWord",
                ".": "autocompleteAfterDot",
                "Ctrl-/": 'toggleComment',
                "Ctrl-I": "exprInfo",
                "Ctrl-L": "gotoLine",
                "Ctrl-E": "deleteLine",
                "Ctrl-D": "doNothing", // 取消默认的 deleteLine
                "Ctrl-B": "jumpToDecl",
                "Ctrl-S": function () {
                    wide.saveFile();
                },
                "Shift-Ctrl-S": function () {
                    menu.saveAllFiles();
                },
                "Shift-Alt-F": function () {
                    var currentPath = editors.getCurrentPath();
                    if (!currentPath) {
                        return false;
                    }
                    wide.fmt(currentPath, wide.curEditor);
                },
                "Alt-F7": "findUsages",
                "Shift-Alt-Enter": function () {
                    if (windows.isMaxEditor) {
                        windows.restoreEditor();
                    } else {
                        windows.maxEditor();
                    }
                },
                "Shift-Ctrl-Up": "copyLinesUp",
                "Shift-Ctrl-Down": "copyLinesDown",
                "Shift-Alt-Up": "moveLinesUp",
                "Shift-Alt-Down": "moveLinesDown",
                "Shift-Alt-J": "selectIdentifier"
            }
        });

        if ("text/html" === data.mode) {
            emmetCodeMirror(editor);
        }

        editor.on('cursorActivity', function (cm) {
            $(".edit-exprinfo").remove();
            var cursor = cm.getCursor();

            $(".footer .cursor").text('|   ' + (cursor.line + 1) + ':' + (cursor.ch + 1) + '   |');
        });

        editor.on('blur', function (cm) {
            $(".edit-exprinfo").remove();
        });

        editor.on('changes', function (cm) {
            if (cm.doc.isClean()) { // no modification
                $(".edit-panel .tabs > div").each(function () {
                    var $span = $(this).find("span:eq(0)");
                    if ($span.attr("title") === cm.options.path) {
                        $span.removeClass("changed");
                    }
                });

                return;
            }

            // changed

            $(".edit-panel .tabs > div").each(function () {
                var $span = $(this).find("span:eq(0)");
                if ($span.attr("title") === cm.options.path) {
                    $span.addClass("changed");
                }
            });
        });

        editor.on('keydown', function (cm, evt) {
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

        editor.setSize('100%', $(".edit-panel").height() - $(".edit-panel .tabs").height());
        editor.setOption("mode", data.mode);
        editor.setOption("gutters", ["CodeMirror-lint-markers", "CodeMirror-foldgutter"]);

        if ("wide" !== config.keymap) {
            editor.setOption("keyMap", config.keymap);
        }

        if ("text/x-go" === data.mode || "application/json" === data.mode) {
            editor.setOption("lint", true);
        }

        if ("application/xml" === data.mode || "text/html" === data.mode) {
            editor.setOption("autoCloseTags", true);
        }

        wide.curEditor = editor;
        editors.data.push({
            "editor": editor,
            "id": id
        });

        $(".footer .cursor").text('|   ' + (cursor.line + 1) + ':' + (cursor.ch + 1) + '   |');

        var half = Math.floor(wide.curEditor.getScrollInfo().clientHeight / wide.curEditor.defaultTextHeight() / 2);
        var cursorCoords = wide.curEditor.cursorCoords({line: cursor.line - half, ch: 0}, "local");
        wide.curEditor.scrollTo(0, cursorCoords.top);

        editor.setCursor(cursor);
        editor.focus();
    }
};