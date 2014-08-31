var tree = {
    fileTree: {},
    newFile: function() {
        $("#dirRMenu ul").hide();
        var name = prompt("Name", "");
        if (!name) {
            return false;
        }

        var request = {
            path: wide.curNode.path + '/' + name,
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
                tree.fileTree.addNodes(wide.curNode, [{
                        "name": name
                    }]);
            }
        });
    },
    newDir: function() {
        $("#dirRMenu ul").hide();
        var name = prompt("Name", "");
        if (!name) {
            return false;
        }

        var request = {
            path: wide.curNode.path + '/' + name,
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
                // TODO: 换成我们风格的 class
                tree.fileTree.addNodes(wide.curNode, [{
                        "name": name,
                        "iconSkin": "ico_close "
                    }]);
            }
        });
    },
    removeIt: function() {
        $("#dirRMenu ul").hide();
        $("#fileRMenu ul").hide();

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
                                    if ("f" === treeNode.type) { // 如果右击了文件
                                        $("#fileRMenu ul").show();
                                        fileRMenu.css({"top": event.clientY + "px", "left": event.clientX + "px", "display": "block"});
                                    } else { // 右击了目录
                                        $("#dirRMenu ul").show();
                                        dirRMenu.css({"top": event.clientY + "px", "left": event.clientX + "px", "display": "block"});
                                    }
                                }
                            },
                            onClick: function(event, treeId, treeNode, clickFlag) {
                                wide.curNode = treeNode;
                                if ("f" === treeNode.type) { // 如果单击了文件
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
                                                return false;
                                            }
                                            editors.newEditor(data);
                                        }
                                    });
                                }
                            }
                        }
                    };
                    tree.fileTree = $.fn.zTree.init($("#files"), setting, data.root.children);
                    // tree.fileTree.expandAll(true);
                }
            }
        });
    }
};