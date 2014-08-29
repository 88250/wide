var outputWS = new WebSocket(config.channel.output + '/output/ws');
outputWS.onopen = function() {
    console.log('[output onopen] connected');
};

outputWS.onmessage = function(e) {
    console.log('[output onmessage]' + e.data);
    var data = JSON.parse(e.data);

    if ('run' === data.cmd) {
        $('#output').val($('#output').val() + data.output);
    } else if ('build' === data.cmd) {
        $('#output').val(data.output);

        if (0 !== data.output.length) { // 说明编译有错误输出
            return;
        }
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
                    $('#output').val('');
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

var shellWS = new WebSocket(config.channel.shell + '/shell/ws');
shellWS.onopen = function() {
    console.log('[shell onopen] connected');
};
shellWS.onmessage = function(e) {
    console.log('[shell onmessage]' + e.data);
    var data = JSON.parse(e.data);
    if ('init-shell' !== data.cmd) {
        $('#shellOutput').val(data.output);
    }
};
shellWS.onclose = function(e) {
    console.log('[shell onclose] disconnected (' + e.code + ')');
    delete shellWS;
};
shellWS.onerror = function(e) {
    console.log('[shell onerror] ' + e);
};

var wide = {
    curNode: undefined,
    curEditor: undefined,
    init: function() {
        $('#shellInput').keydown(function(event) {
            if (13 === event.which) {
                var input = {
                    cmd: $('#shellInput').val()
                };
                shellWS.send(JSON.stringify(input));
                $('#shellInput').val('');
            }
        });

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
                $('#output').val('');
            },
            success: function(data) {
                executable = data.executable;

                if (data.succ) {

                }
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
        $.ajax({
            type: 'POST',
            url: '/fmt',
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