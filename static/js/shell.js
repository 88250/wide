var shell = {
    _shellWS: undefined,
    _initWS: function () {
        shell.shellWS = new ReconnectingWebSocket(config.channel.shell + '/shell/ws?sid=' + config.wideSessionId);
        shell.shellWS.onopen = function () {
            console.log('[shell onopen] connected');
        };
        shell.shellWS.onmessage = function (e) {
            console.log('[shell onmessage]' + e.data);
            var data = JSON.parse(e.data);
            if ('init-shell' !== data.cmd) {
                $('#shellOutput').val(data.output);
            }
        };
        shell.shellWS.onclose = function (e) {
            console.log('[shell onclose] disconnected (' + e.code + ')');
        };
        shell.shellWS.onerror = function (e) {
            console.log('[shell onerror] ' + e);
        };
    },
    init: function () {
        this._initWS();
        
        $('#shellInput').keydown(function (event) {
            if (13 === event.which) {
                var input = {
                    cmd: $('#shellInput').val()
                };
                shell.shellWS.send(JSON.stringify(input));
                $('#shellInput').val('');
            }
        });
    }
};

$(document).ready(function () {
    shell.init();
});