var outputWS = new WebSocket(config.channel.output + '/output/ws');
outputWS.onopen = function() {
    console.log('[output onopen] connected');
};

outputWS.onmessage = function(e) {
    console.log('[output onmessage]' + e.data);
    var data = JSON.parse(e.data);

    if ('run' === data.cmd) {
        $('#output').text($('#output').text() + data.output);
    } else if ('build' === data.cmd) {
        $('#output').text(data.output);

        if (0 !== data.output.length) { // 说明编译有错误输出
            return;
        }
    } else if ('go get' === data.cmd) {
        $('#output').text($('#output').text() + data.output);
    }

    if ('build' === data.cmd) {
        if ('run' === data.nextCmd) {
            var request = {
                "executable": data.executable
            };

            $.ajax({
                type: 'POST',
                url: '/run',
                data: JSON.stringify(request),
                dataType: "json",
                beforeSend: function(data) {
                    $('#output').text('');
                },
                success: function(data) {

                }
            });
        }
    }
};
outputWS.onclose = function(e) {
    console.log('[output onclose] disconnected (' + e.code + ')');
    delete outputWS;
};
outputWS.onerror = function(e) {
    console.log('[output onerror] ' + e);
};

var wide = {
    curNode: undefined,
    curEditor: undefined,
    _initLayout: function() {
        var mainH = $(window).height() - $(".menu").height() - $(".footer").height() - 2;
        $(".content, .ztree").height(mainH);

        $(".edit-panel").height(mainH - $(".output").height());
    },
    init: function() {
        this._initLayout();

        $("body").bind("mousedown", function(event) {
            if (!(event.target.id === "dirRMenu" || $(event.target).closest("#dirRMenu").length > 0)) {
                $("#dirRMenu").hide();
            }

            if (!(event.target.id === "fileRMenu" || $(event.target).closest("#fileRMenu").length > 0)) {
                $("#fileRMenu").hide();
            }
        });
    },
    save: function() {
        var request = {
            "file": wide.curNode.path,
            "code": wide.curEditor.getValue()
        };
        $.ajax({
            type: 'POST',
            url: '/file/save',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                console.log(data);
            }
        });
    },
    run: function() {
        var request = {
            "file": wide.curNode.path,
            "code": wide.curEditor.getValue()
        };

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('#output').text('');
            },
            success: function(data) {
            }
        });
    },
    goget: function() {
        var request = {
            "file": wide.curNode.path
        };

        $.ajax({
            type: 'POST',
            url: '/go/get',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function(data) {
                $('#output').text('');
            },
            success: function(data) {
            }
        });
    },
    fmt: function() {
        var request = {
            "file": wide.curNode.path,
            "code": wide.curEditor.getValue(),
            "cursorLine": wide.curEditor.getCursor().line,
            "cursorCh": wide.curEditor.getCursor().ch
        };
		
		// TODO: HTML/XML/JSON 格式化处理
		
        $.ajax({
            type: 'POST',
            url: '/go/fmt',
            data: JSON.stringify(request),
            dataType: "json",
            success: function(data) {
                if (data.succ) {
                    wide.curEditor.setValue(data.code);
                }
            }
        });
    }
};

$(document).ready(function() {
    wide.init();
    tree.init();
});