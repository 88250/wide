var hotkeys = {
    defaultKeyMap: {
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
        // F6 构建并运行
        buildRun: {
            ctrlKey: false,
            altKey: false,
            shiftKey: false,
            which: 117
        }
    },
    _bindFileTree: function() {
        // TODO: 滚动处理
        $("#files").keydown(function(event) {
            switch (event.which) {
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
                            // 当前节点的上一个节点是目录且打开时
                            node = preNode.children[preNode.children.length - 1];
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
                        if (wide.curNode && wide.curNode.isLastNode && !wide.curNode.open
                                && wide.curNode.getParentNode().isLastNode) {
                            // 当前节点为最底部的节点
                            return false;
                        }

                        node = wide.curNode.getNextNode();
                        if (wide.curNode.iconSkin === "ico-ztree-dir " && wide.curNode.open) {
                            // 当前节点是目录且打开时
                            node = wide.curNode.children[0];
                        }

                        if (wide.curNode.isLastNode && wide.curNode.level !== 0
                                && wide.curNode.getParentNode().getNextNode()) {
                            // 当前节点为最后一个节点，但其父节点还有下一个节点
                            node = wide.curNode.getParentNode().getNextNode();
                        }
                    }

                    wide.curNode = node;
                    tree.fileTree.selectNode(node);
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
    init: function() {
        this._bindFileTree();

        var hotKeys = this.defaultKeyMap;
        $(document).keydown(function(event) {
            if (event.ctrlKey === hotKeys.goFileTree.ctrlKey
                    && event.which === hotKeys.goFileTree.which) { // Ctrl+1 焦点切换到文件树
                // 有些元素需设置 tabindex 为 -1 时才可以 focus
                $("#files").focus();
                event.preventDefault();

                return;
            }

            if (event.ctrlKey === hotKeys.goOutPut.ctrlKey
                    && event.which === hotKeys.goOutPut.which) { // Ctrl+4 焦点切换到输出窗口                
                $("#output").focus();
                event.preventDefault();

                return;
            }

            if (event.which === hotKeys.buildRun.which) { // F6 构建并运行
                wide.run();
                event.preventDefault();

                return;
            }
        });
    }
};