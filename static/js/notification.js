var notificationWS = new WebSocket(config.channel.shell + '/notification/ws');
notificationWS.onopen = function() {
    console.log('[notification onopen] connected');
};
notificationWS.onmessage = function(e) {
    console.log('[notification onmessage]' + e.data);
    var data = JSON.parse(e.data);
    if ('init-notification' !== data.cmd) {
        $('#notification').val(data.output);
    }
};
notificationWS.onclose = function(e) {
    console.log('[notification onclose] disconnected (' + e.code + ')');
    delete notificationWS;
};
notificationWS.onerror = function(e) {
    console.log('[notification onerror] ' + e);
};

var notification = {
    init: function() {
       
    }
};

$(document).ready(function() {
    notification.init();
});