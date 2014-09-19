// 用于保持会话，如果该通道断开，则服务器端会销毁会话状态，回收相关资源.
var sessionWS = new WebSocket(config.channel.session + '/session/ws?sid=' + config.wideSessionId);
sessionWS.onopen = function () {
    console.log('[session onopen] connected');
};

sessionWS.onmessage = function (e) {
    console.log('[session onmessage]' + e.data);
    var data = JSON.parse(e.data);

};
sessionWS.onclose = function (e) {
    console.log('[session onclose] disconnected (' + e.code + ')');
    delete sessionWS;
};
sessionWS.onerror = function (e) {
    console.log('[session onerror] ' + JSON.parse(e));
};

