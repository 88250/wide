var hotkeys = {
    defaultKeyMap: {
        // Ctrl+0 焦点切换到当前编辑器   
        goEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 48
        },
        // Ctrl+1 焦点切换到文件树
        goFileTree: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 49
        },
        // Ctrl+4 焦点切换到输出窗口   
        goOutPut: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 52
        },
        // Ctrl+5 焦点切换到搜索窗口   
        goSearch: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 53
        },
        // Ctrl+6 焦点切换到通知窗口   
        goNotification: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 54
        },
        // Ctrl+Q 关闭当前编辑器   
        closeCurEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 81
        },
        // Ctrl+D 窗口组切换   
        changeEditor: {
            ctrlKey: true,
            altKey: false,
            shiftKey: false,
            which: 68
        },
        // F6 构建并运行
        buildRun: {
            ctrlKey: false,
            altKey: false,
            shiftKey: false,
            which: 117
        }
    },
    _bindFileTree: function () {
        $("#files").keydown(function (event) {
            event.preventDefault();

            switch (event.which) {
                case 46: // 删除
                    tree.removeIt();
                    break;
                case 13: // 回车
                    if (!wide.curNode) {
                        return false;
                    }

                    if (wide.curNode.iconSkin === "ico-ztree-dir ") { // 选中节点是目录
                        // 不做任何处理
                        return false;
                    }

                    // 模拟点击：打开文件
                    tree._onClick(wide.curNode);

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
                        if (preNode && preNode.iconSkin === "ico-ztree-dir "
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
                        if (wide.curNode.iconSkin === "ico-ztree-dir " && wide.curNode.open) {
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

                    if (wide.curNode.iconSkin !== "ico-ztree-dir " || !wide.curNode.open) {
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

                    if (wide.curNode.iconSkin !== "ico-ztree-dir " || wide.curNode.open) {
                        return false;
                    }

                    tree.fileTree.expandNode(wide.curNode, true, false, true);
                    $("#files").focus();

                    break;
            }
        });
    },
    init: function () {
        this._bindFileTree();

        var hotKeys = this.defaultKeyMap;
        $(document).keydown(function (event) {
            if (event.ctrlKey === hotKeys.goEditor.ctrlKey
                    && event.which === hotKeys.goEditor.which) {  // Ctrl+0 焦点切换到当前编辑器
                if (wide.curEditor) {
                    wide.curEditor.focus();
                }
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goFileTree.ctrlKey
                    && event.which === hotKeys.goFileTree.which) { // Ctrl+1 焦点切换到文件树
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

            if (event.ctrlKey === hotKeys.goOutPut.ctrlKey
                    && event.which === hotKeys.goOutPut.which) { // Ctrl+4 焦点切换到输出窗口   
                wide.bottomWindowTab.setCurrent("output");

                windows.flowBottom();
                $(".bottom-window-group .output").focus();
                event.preventDefault();

                return;
            }
            if (event.ctrlKey === hotKeys.goSearch.ctrlKey
                    && event.which === hotKeys.goSearch.which) { // Ctrl+5 焦点切换到搜索窗口  
                wide.bottomWindowTab.setCurrent("search");
                windows.flowBottom();
                $(".bottom-window-group .search").focus();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goNotification.ctrlKey
                    && event.which === hotKeys.goNotification.which) { // Ctrl+6 焦点切换到通知窗口          
                wide.bottomWindowTab.setCurrent("notification");
                windows.flowBottom();
                $(".bottom-window-group .notification").focus();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.closeCurEditor.ctrlKey
                    && event.which === hotKeys.closeCurEditor.which) {  // Ctrl+Q 关闭当前编辑器   
                if (editors.tabs.getCurrentId()) {
                    editors.tabs.del(editors.tabs.getCurrentId());
                }
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.changeEditor.ctrlKey
                    && event.which === hotKeys.changeEditor.which) { // Ctrl+D 窗口组切换
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
                    wide.bottomWindowTab.setCurrent(nextId);
                    $(".bottom-window-group ." + nextId).focus();

                    event.preventDefault();

                    return false;
                }

                if (editors.data.length > 1) {
                    var nextId = "";
                    for (var i = 0, ii = editors.data.length; i < ii; i++) {
                        if (editors.tabs.getCurrentId() === editors.data[i].id) {
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

                    editors.tabs.setCurrent(nextId);
                    wide.curNode = tree.fileTree.getNodeByTId(nextId);
                    tree.fileTree.selectNode(wide.curNode);

                    wide.curEditor.focus();
                }

                event.preventDefault();

                return false;
            }

            if (event.which === hotKeys.buildRun.which) { // F6 构建并运行
                wide.run();
                event.preventDefault();

                return;
            }
        });
    }
};