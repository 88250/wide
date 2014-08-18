var tree = {
    newFile: function() {
        $("#dirRMenu ul").hide();
        var name = prompt("Name", "");
        var request = {
            path: wide.curFile + '/' + name,
            fileType: "f"
        };
        $.ajax({
            type: 'POST',
            url: '/file/new',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                if (data.succ) {
                    console.log(data);
                }
            }
        });
    },
    newDir: function() {
        $("#dirRMenu ul").hide();
        var name = prompt("Name", "");
        var request = {
            path: wide.curFile + '/' + name,
            fileType: "d"
        };
        $.ajax({
            type: 'POST',
            url: '/file/new',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                if (data.succ) {
                    console.log(data);
                }
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
            path: wide.curFile + '/' + name
        };
        $.ajax({
            type: 'POST',
            url: '/file/remove',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                if (data.succ) {
                    console.log(data);
                }
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
                                    wide.curFile = treeNode.path;
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
                                wide.curFile = treeNode.path;
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
                                            if (data.succ) {
                                                editor.setValue(data.content);
                                                editor.setOption("mode", data.mode);
                                            }
                                        }
                                    });
                                }
                            }
                        }
                    };
                    fileTree = $.fn.zTree.init($("#files"), setting, data.root.children);
                    fileTree.expandAll(true);
                }
            }
        });
    }
};