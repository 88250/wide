var shellWS = new WebSocket(config.channel.shell + '/shell/ws?sid=' + config.wideSessionId);
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

var shell = {
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
    }
};

$(document).ready(function() {
    shell.init();
});