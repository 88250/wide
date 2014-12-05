/* 
 * Copyright (c) 2014, B3log
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

var hotkeys = {
    defaultKeyMap: {
        // Ctrl-0 焦点切换到当前编辑器   
        goEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 48
        },
        // Ctrl-1 焦点切换到文件树
        goFileTree: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 49
        },
        // Ctrl-4 焦点切换到输出窗口   
        goOutput: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 52
        },
        // Ctrl-5 焦点切换到搜索窗口   
        goSearch: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 53
        },
        // Ctrl-6 焦点切换到通知窗口   
        goNotification: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 54
        },
        // Ctrl-C 清空窗口内容   
        clearWindow: {
            ctrlKey: true,
            altKey: false,
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
        // Ctrl-F 搜索  
        search: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 70
        },
        // Ctrl-Q 关闭当前编辑器   
        closeCurEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 81
        },
        // Ctrl-R 重命名   
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
            event.preventDefault();

            var hotKeys = hotkeys.defaultKeyMap;
            if (event.ctrlKey === hotKeys.clearWindow.ctrlKey
                    && event.which === hotKeys.clearWindow.which) {  // Ctrl-F 搜索
                bottomGroup.clear('output');
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
                case 46: // 删除
                    tree.removeIt();
                    break;
                case 13: // 回车
                    if (!wide.curNode) {
                        return false;
                    }

                    if (tree.isDir()) { // 选中节点是目录
                        // 不做任何处理
                        return false;
                    }

                    // 模拟点击：打开文件
                    tree.openFile(wide.curNode);

                    break;
                case 38: // 上
                    var node = {};

                    if (!wide.curNode) { // 没有选中节点时，默认选中第一个
                        node = tree.fileTree.getNodeByTId("files_1");
                    } else {
                        if (wide.curNode && wide.curNode.isFirstNode && wide.curNode.level === 0) {
                            // 当前节点为顶部第一个节点
                            return false;
                        }

                        node = wide.curNode.getPreNode();
                        if (wide.curNode.isFirstNode && wide.curNode.getParentNode()) {
                            // 当前节点为第一个节点且有父亲
                            node = wide.curNode.getParentNode();
                        }

                        var preNode = wide.curNode.getPreNode();
                        if (preNode && tree.isDir()
                                && preNode.open) {
                            // 当前节点的上一个节点是目录且打开时，获取打开节点中的最后一个节点
                            node = tree.getCurrentNodeLastNode(preNode);
                        }
                    }

                    wide.curNode = node;
                    tree.fileTree.selectNode(node);
                    $("#files").focus();
                    break;
                case 40: // 下
                    var node = {};

                    if (!wide.curNode) { // 没有选中节点时，默认选中第一个
                        node = tree.fileTree.getNodeByTId("files_1");
                    } else {
                        if (wide.curNode && tree.isBottomNode(wide.curNode)) {
                            // 当前节点为最底部的节点
                            return false;
                        }

                        node = wide.curNode.getNextNode();
                        if (tree.isDir() && wide.curNode.open) {
                            // 当前节点是目录且打开时
                            node = wide.curNode.children[0];
                        }

                        var nextShowNode = tree.getNextShowNode(wide.curNode);
                        if (wide.curNode.isLastNode && wide.curNode.level !== 0 && !wide.curNode.open
                                && nextShowNode) {
                            // 当前节点为最后一个叶子节点，但其父或祖先节点还有下一个节点
                            node = nextShowNode;
                        }
                    }

                    if (node) {
                        wide.curNode = node;
                        tree.fileTree.selectNode(node);
                    }

                    $("#files").focus();
                    break;
                case 37: // 左
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
                case 39: // 右
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
            }
        });
    },
    _bindDocument: function () {
        var hotKeys = this.defaultKeyMap;
        $(document).keydown(function (event) {
            if (event.ctrlKey === hotKeys.goEditor.ctrlKey
                    && event.which === hotKeys.goEditor.which) {  // Ctrl-0 焦点切换到当前编辑器
                if (wide.curEditor) {
                    wide.curEditor.focus();
                }
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goFileTree.ctrlKey
                    && event.which === hotKeys.goFileTree.which) { // Ctrl-1 焦点切换到文件树
                // 有些元素需设置 tabindex 为 -1 时才可以 focus
                if ($(".footer .ico-restore:eq(0)").css("display") === "inline") {
                    // 当文件树最小化时
                    $(".side").css({
                        "left": "0"
                    });

                    if ($(".footer .ico-restore:eq(1)").css("display") === "inline") {
                        // 当底部最小化时
                        $(".bottom-window-group").css("top", "100%").hide();
                    }
                }

                $("#files").focus();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goOutput.ctrlKey
                    && event.which === hotKeys.goOutput.which) { // Ctrl-4 焦点切换到输出窗口   
                bottomGroup.tabs.setCurrent("output");

                windows.flowBottom();
                $(".bottom-window-group .output").focus();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goSearch.ctrlKey
                    && event.which === hotKeys.goSearch.which) { // Ctrl-5 焦点切换到搜索窗口  
                bottomGroup.tabs.setCurrent("search");
                windows.flowBottom();
                $(".bottom-window-group .search").focus();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goNotification.ctrlKey
                    && event.which === hotKeys.goNotification.which) { // Ctrl-6 焦点切换到通知窗口          
                bottomGroup.tabs.setCurrent("notification");
                windows.flowBottom();
                $(".bottom-window-group .notification").focus();
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
                            nextId = "";
                    for (var i = 0, ii = tabs.length; i < ii; i++) {
                        if (document.activeElement.className === tabs[i]) {
                            if (i < ii - 1) {
                                nextId = tabs[i + 1];
                            } else {
                                nextId = tabs[0];
                            }
                            break;
                        }
                    }
                    bottomGroup.tabs.setCurrent(nextId);
                    $(".bottom-window-group ." + nextId).focus();

                    event.preventDefault();

                    return false;
                }

                if (editors.data.length > 1) {
                    var nextId = "";
                    for (var i = 0, ii = editors.data.length; i < ii; i++) {
                        var currentId = editors.getCurrentId();
                        if (currentId) {
                            if (currentId === editors.data[i].id) {
                                if (i < ii - 1) {
                                    nextId = editors.data[i + 1].id;
                                    wide.curEditor = editors.data[i + 1].editor;
                                } else {
                                    nextId = editors.data[0].id;
                                    wide.curEditor = editors.data[0].editor;
                                }
                                break;
                            }
                        }
                    }

                    editors.tabs.setCurrent(nextId);
                    wide.curNode = tree.fileTree.getNodeByTId(nextId);
                    tree.fileTree.selectNode(wide.curNode);

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