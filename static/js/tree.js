var tree = {
    getTIdByPath: function (path) {
       var nodes = tree.fileTree.transformToArray(tree.fileTree.getNodes());
        for (var i = 0, ii = nodes.length; i < ii; i++) {
            if (nodes[i].path === path) {
                return nodes[i].tId;
            }
        }
        
        return undefined;
    },
    fileTree: undefined,
    _isParents: function(tId, parentTId) {
        var node = tree.fileTree.getNodeByTId(tId);
        if (!node || !node.parentTId) {
            return false;
        } else {
            if (node.parentTId === parentTId) {
                return true;
            } else {
                return tree._isParents(node.parentTId, parentTId);
            }
        }
    },
    newFile: function() {
        $("#dirRMenu").hide();
        var name = prompt("Name", "");
        if (!name) {
            return false;
        }

        var request = {
            path: wide.curNode.path + '\\' + name,
            fileType: "f"
        };
        $.ajax({
            type: 'POST',
            url: '/file/new',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                if (!data.succ) {
                    return false;
                }

                var suffix = name.split(".")[1],
                        iconSkin = "ico-ztree-other ";
                switch (suffix) {
                    case "html", "htm":
                        iconSkin = "ico-ztree-html ";
                        break;
                    case "go":
                        iconSkin = "ico-ztree-go ";
                        break;
                    case "css":
                        iconSkin = "ico-ztree-css ";
                        break;
                    case "txt":
                        iconSkin = "ico-ztree-text ";
                        break;
                    case "sql":
                        iconSkin = "ico-ztree-sql ";
                        break;
                    case "properties":
                        iconSkin = "ico-ztree-pro ";
                        break;
                    case "md":
                        iconSkin = "ico-ztree-md ";
                        break;
                    case "md":
                        iconSkin = "ico-ztree-md ";
                        break;
                    case "js", "json":
                        iconSkin = "ico-ztree-js ";
                        break;
                    case "xml":
                        iconSkin = "ico-ztree-xml ";
                        break;
                    case "jpg", "jpeg", "bmp", "gif", "png", "svg", "ico":
                        iconSkin = "ico-ztree-img ";
                        break;
                }

                tree.fileTree.addNodes(wide.curNode, [{
                        "name": name,
                        "iconSkin": iconSkin,
                        "path": request.path
                    }]);
            }
        });
    },
    newDir: function() {
        $("#dirRMenu").hide();
        var name = prompt("Name", "");
        if (!name) {
            return false;
        }

        var request = {
            path: wide.curNode.path + '\\' + name,
            fileType: "d"
        };
        $.ajax({
            type: 'POST',
            url: '/file/new',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                if (!data.succ) {
                    return false;
                }
                tree.fileTree.addNodes(wide.curNode, [{
                        "name": name,
                        "iconSkin": "ico-ztree-dir ",
                        "path": request.path
                    }]);
            }
        });
    },
    removeIt: function() {
        $("#dirRMenu").hide();
        $("#fileRMenu").hide();

        if (!confirm("Remove it?")) {
            return;
        }
        var request = {
            path: wide.curNode.path
        };
        $.ajax({
            type: 'POST',
            url: '/file/remove',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                if (!data.succ) {
                    return false;
                }

                if ("ico-ztree-dir " !== wide.curNode.iconSkin) {
                    // 是文件的话，查看 editor 中是否被打开，如打开则移除
                    for (var i = 0, ii = editors.data.length; i < ii; i++) {
                        if (editors.data[i].id === wide.curNode.tId) {
                            $(".edit-header .tabs > div[data-index=" + wide.curNode.tId + "]").find(".ico-close").click();
                            break;
                        }
                    }
                } else {
                    for (var i = 0, ii = editors.data.length; i < ii; i++) {
                        if (tree._isParents(editors.data[i].id, wide.curNode.tId)) {
                            $(".edit-header .tabs > div[data-index=" + editors.data[i].id + "]").find(".ico-close").click();
                            i--;
                            ii--;
                        }
                    }
                }

                tree.fileTree.removeNode(wide.curNode);
            }
        });
    },
    init: function() {
        $.ajax({
            type: 'GET',
            url: '/files',
            dataType: "json",
            success: function(data) {
                if (data.succ) {
                    var dirRMenu = $("#dirRMenu");
                    var fileRMenu = $("#fileRMenu");
                    var setting = {
                        view: {
                            selectedMulti: false
                        },
                        callback: {
                            onRightClick: function(event, treeId, treeNode) {
                                if (treeNode) {
                                    wide.curNode = treeNode;
                                    if ("ico-ztree-dir " !== treeNode.iconSkin) { // 如果右击了文件
                                        $("#fileRMenu ul").show();
                                        fileRMenu.css({
                                            "top": event.clientY - 10 + "px",
                                            "left": event.clientX + "px",
                                            "display": "block"
                                        });
                                    } else { // 右击了目录
                                        $("#dirRMenu ul").show();
                                        dirRMenu.css({
                                            "top": event.clientY - 10 + "px",
                                            "left": event.clientX + "px",
                                            "display": "block"
                                        });
                                    }
                                }
                            },
                            onClick: function(event, treeId, treeNode, clickFlag) {
                                tree._onClick(treeNode);
                            }
                        }
                    };
                    tree.fileTree = $.fn.zTree.init($("#files"), setting, data.root.children);
                }
            }
        });
    },
    _onClick: function(treeNode) {
        if (wide.curNode) {
            for (var i = 0, ii = editors.data.length; i < ii; i++) {
                // 该节点文件已经打开
                if (editors.data[i].id === treeNode.tId) {
                    editors.tabs.setCurrent(treeNode.tId);
                    wide.curNode = treeNode;
                    wide.curEditor = editors.data[i].editor;
                    return false;
                }
            }
        }

        wide.curNode = treeNode;

        if ("ico-ztree-dir " !== treeNode.iconSkin) { // 如果单击了文件
            var request = {
                path: treeNode.path
            };
            $.ajax({
                type: 'POST',
                url: '/file',
                data: JSON.stringify(request),
                dataType: "json",
                success: function(data) {
                    if (!data.succ) {
                        alert(data.msg);

                        return false;
                    }

                    if ("img" === data.mode) { // 是图片文件的话新建 tab 打开
                        // 最好是开 tab，但这个最终取决于浏览器设置
                        var w = window.open(data.path);

                        return false;
                    }

                    editors.newEditor(data);
                }
            });
        }
    }
};