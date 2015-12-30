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
 * @file hotkeys.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @author <a href="http://88250.b3log.org">Liang Ding</a>
 * @version 1.0.0.2, Dec 15, 2015
 */
var hotkeys = {
    defaultKeyMap: {
        // Ctrl-0
        goEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 48,
            fun: function () {
                if (wide.curEditor) {
                    wide.curEditor.focus();
                }
            }
        },
        // Ctrl-1
        goFileTree: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 49,
            fun: function () {
                // 有些元素需设置 tabindex 为 -1 时才可以 focus
                if (windows.outerLayout.west.state.isClosed) {
                    windows.outerLayout.slideOpen('west');
                }
                $("#files").focus();
            }
        },
        // Ctrl-2
        goOutline: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 50,
            fun: function () {
                if (windows.innerLayout.east.state.isClosed) {
                    windows.innerLayout.slideOpen('east');
                }

                $("#outline").focus();
            }
        },
        // Ctrl-4
        goOutput: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 52,
            fun: function () {
                bottomGroup.tabs.setCurrent("output");
                windows.flowBottom();
                $(".bottom-window-group .output").focus();
            }
        },
        // Ctrl-5
        goSearch: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 53,
            fun: function () {
                bottomGroup.tabs.setCurrent("search");
                windows.flowBottom();
                $(".bottom-window-group .search").focus();
            }
        },
        // Ctrl-6
        goNotification: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 54,
            fun: function () {
                bottomGroup.tabs.setCurrent("notification");
                windows.flowBottom();
                $(".bottom-window-group .notification").focus();
            }
        },
        // Alt-C
        clearWindow: {
            ctrlKey: false,
            altKey: true,
            shiftKey: false,
            which: 67
        },
        // Ctrl-D 窗口组切换   
        changeEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 68
        },
        // Ctrl-F search  
        search: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 70
        },
        // Ctrl-Q close current editor   
        closeCurEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 81
        },
        // Ctrl-R
        rename: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 82
        },
        // Shift-Alt-O 跳转到文件
        goFile: {
            ctrlKey: false,
            altKey: true,
            shiftKey: true,
            which: 79
        },
        // F5 Build
        build: {
            ctrlKey: false,
            altKey: false,
            shiftKey: false,
            which: 116
        },
        // F6 Build & Run
        buildRun: {
            ctrlKey: false,
            altKey: false,
            shiftKey: false,
            which: 117
        }
    },
    bindList: function ($source, $list, enterFun) {
        $list.data("index", 0);
        $source.keydown(function (event) {
            var index = $list.data("index"),
                    count = $list.find("li").length;

            if (count === 0) {
                return true;
            }

            if (event.which === 38) {   // up
                index--;
                if (index < 0) {
                    index = count - 1;
                }
            }

            if (event.which === 40) {   // down
                index++;
                if (index > count - 1) {
                    index = 0;
                }
            }

            var $selected = $list.find("li:eq(" + index + ")");

            if (event.which === 13) {   // enter
                enterFun($selected);
            }

            $list.find("li").removeClass("selected");
            $list.data("index", index);
            $selected.addClass("selected");

            if (index === 0) {
                $list.scrollTop(0);
            } else {
                if ($selected[0].offsetTop + $list.scrollTop() > $list.height()) {
                    if (event.which === 40) {
                        $list.scrollTop($list.scrollTop() + $selected.height());
                    } else {
                        $list.scrollTop($selected[0].offsetTop);
                    }
                } else {
                    $list.scrollTop(0);
                }
            }

            // 阻止上下键改变光标位置
            if (event.which === 38 || event.which === 40 || event.which === 13) {
                return false;
            }
        });
    },
    _bindOutput: function () {
        $(".bottom-window-group .output").keydown(function (event) {
            var hotKeys = hotkeys.defaultKeyMap;
            if (event.altKey === hotKeys.clearWindow.altKey
                    && event.which === hotKeys.clearWindow.which) {  // Alt-C clear output
                bottomGroup.clear('output');

                event.preventDefault();

                return;
            }
        });
    },
    _bindFileTree: function () {
        $("#files").keydown(function (event) {
            event.preventDefault();

            var hotKeys = hotkeys.defaultKeyMap;
            if (event.ctrlKey === hotKeys.search.ctrlKey
                    && event.which === hotKeys.search.which) {  // Ctrl-F 搜索
                $("#dialogSearchForm").dialog("open");
                return;
            }

            if (event.ctrlKey === hotKeys.rename.ctrlKey
                    && event.which === hotKeys.rename.which) {  // Ctrl-R 重命名
                if (wide.curNode.removable) {
                    $("#dialogRenamePrompt").dialog("open");
                }
                return;
            }

            switch (event.which) {
                case 46: // delete
                    tree.removeIt();
                    break;
                case 13: // enter
                    if (!wide.curNode) {
                        return false;
                    }

                    if (tree.isDir()) {
                        if (wide.curNode.open) {
                            return false;
                        }

                        tree.fileTree.expandNode(wide.curNode, true, false, true);
                        $("#files").focus();

                        break;
                    }

                    tree.openFile(wide.curNode);

                    break;
                case 38: // up
                    var node = {};

                    if (!wide.curNode) { // select the first one if no node been selected
                        node = tree.fileTree.getNodeByTId("files_1");
                    } else {
                        if (wide.curNode && wide.curNode.isFirstNode && wide.curNode.level === 0) {
                            return false;
                        }

                        node = wide.curNode.getPreNode();
                        if (wide.curNode.isFirstNode && wide.curNode.getParentNode()) {
                            node = wide.curNode.getParentNode();
                        }

                        var preNode = wide.curNode.getPreNode();
                        if (preNode && tree.isDir() && preNode.open) {
                            node = tree.getCurrentNodeLastNode(preNode);
                        }
                    }

                    wide.curNode = node;
                    tree.fileTree.selectNode(node);
                    $("#files").focus();
                    break;
                case 40: // down
                    var node = {};

                    if (!wide.curNode) { // select the first one if no node been selected                        
                        node = tree.fileTree.getNodeByTId("files_1");
                    } else {
                        if (wide.curNode && tree.isBottomNode(wide.curNode)) {
                            return false;
                        }

                        node = wide.curNode.getNextNode();
                        if (tree.isDir() && wide.curNode.open) {
                            node = wide.curNode.children[0];
                        }

                        var nextShowNode = tree.getNextShowNode(wide.curNode);
                        if (wide.curNode.isLastNode && wide.curNode.level !== 0 && !wide.curNode.open && nextShowNode) {
                            node = nextShowNode;
                        }
                    }

                    if (node) {
                        wide.curNode = node;
                        tree.fileTree.selectNode(node);
                    }

                    $("#files").focus();
                    break;
                case 37: // left
                    if (!wide.curNode) {
                        wide.curNode = tree.fileTree.getNodeByTId("files_1");
                        tree.fileTree.selectNode(wide.curNode);
                        $("#files").focus();
                        return false;
                    }

                    if (!tree.isDir() || !wide.curNode.open) {
                        return false;
                    }

                    tree.fileTree.expandNode(wide.curNode, false, false, true);
                    $("#files").focus();
                    break;
                case 39: // right
                    if (!wide.curNode) {
                        wide.curNode = tree.fileTree.getNodeByTId("files_1");
                        tree.fileTree.selectNode(wide.curNode);
                        $("#files").focus();
                        return false;
                    }

                    if (!tree.isDir() || wide.curNode.open) {
                        return false;
                    }

                    tree.fileTree.expandNode(wide.curNode, true, false, true);
                    $("#files").focus();

                    break;
                case 116: // F5
                    if (!wide.curNode || !tree.isDir()) {
                        return false;
                    }

                    tree.refresh(wide.curNode);

                    break;
            }
        });
    },
    _bindDocument: function () {
        var hotKeys = this.defaultKeyMap;
        $(document).keydown(function (event) {
            if (event.ctrlKey === hotKeys.goEditor.ctrlKey
                    && event.which === hotKeys.goEditor.which) {  // Ctrl-0 焦点切换到当前编辑器
                hotKeys.goEditor.fun();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goFileTree.ctrlKey
                    && event.which === hotKeys.goFileTree.which) { // Ctrl-1 焦点切换到文件树
                hotKeys.goFileTree.fun();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goOutline.ctrlKey
                    && event.which === hotKeys.goOutline.which) { // Ctrl-2 焦点切换到大纲
                hotKeys.goOutline.fun();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goOutput.ctrlKey
                    && event.which === hotKeys.goOutput.which) { // Ctrl-4 焦点切换到输出窗口   
                hotKeys.goOutput.fun();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goSearch.ctrlKey
                    && event.which === hotKeys.goSearch.which) { // Ctrl-5 焦点切换到搜索窗口  
                hotKeys.goSearch.fun();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goNotification.ctrlKey
                    && event.which === hotKeys.goNotification.which) { // Ctrl-6 焦点切换到通知窗口  
                hotKeys.goNotification.fun();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.closeCurEditor.ctrlKey
                    && event.which === hotKeys.closeCurEditor.which) {  // Ctrl-Q 关闭当前编辑器   
                $(".edit-panel .tabs > div.current").find(".ico-close").click();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.changeEditor.ctrlKey
                    && event.which === hotKeys.changeEditor.which) { // Ctrl-D 窗口组切换
                if (document.activeElement.className === "notification"
                        || document.activeElement.className === "output"
                        || document.activeElement.className === "search") {
                    // 焦点在底部窗口组时，对底部进行切换
                    var tabs = ["output", "search", "notification"],
                            nextPath = "";
                    for (var i = 0, ii = tabs.length; i < ii; i++) {
                        if (bottomGroup.tabs.getCurrentId() === tabs[i]) {
                            if (i < ii - 1) {
                                nextPath = tabs[i + 1];
                            } else {
                                nextPath = tabs[0];
                            }
                            break;
                        }
                    }
                    bottomGroup.tabs.setCurrent(nextPath);
                    $(".bottom-window-group ." + nextPath).focus();

                    event.preventDefault();

                    return false;
                }

                if (editors.data.length > 1) {
                    var nextPath = "";
                    for (var i = 0, ii = editors.data.length; i < ii; i++) {
                        var currentId = editors.getCurrentId();
                        if (currentId) {
                            if (currentId === editors.data[i].id) {
                                if (i < ii - 1) {
                                    nextPath = editors.data[i + 1].id;
                                    wide.curEditor = editors.data[i + 1].editor;
                                } else {
                                    nextPath = editors.data[0].id;
                                    wide.curEditor = editors.data[0].editor;
                                }
                                break;
                            }
                        }
                    }

                    editors.tabs.setCurrent(nextPath);
                    var nextTId = tree.getTIdByPath(nextPath);
                    wide.curNode = tree.fileTree.getNodeByTId(nextTId);

                    tree.fileTree.selectNode(wide.curNode);
                    wide.refreshOutline();
                    var cursor = wide.curEditor.getCursor();
                    $(".footer .cursor").text('|   ' + (cursor.line + 1) + ':' + (cursor.ch + 1) + '   |');
                    wide.curEditor.focus();
                }

                event.preventDefault();

                return false;
            }

            if (event.which === hotKeys.build.which) { // F5 Build
                menu.build();
                event.preventDefault();

                return;
            }

            if (event.which === hotKeys.buildRun.which) { // F6 Build & Run
                menu.run();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goFile.ctrlKey
                    && event.altKey === hotKeys.goFile.altKey
                    && event.shiftKey === hotKeys.goFile.shiftKey
                    && event.which === hotKeys.goFile.which) { // Shift-Alt-O 跳转到文件
                $("#dialogGoFilePrompt").dialog("open");
            }
        });
    },
    init: function () {
        this._bindFileTree();
        this._bindOutput();
        this._bindDocument();
    }
};