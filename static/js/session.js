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

// 定时（30 秒）保存会话内容.
setTimeout(function () {
    var request = newWideRequest();
    
    // TODO: 会话状态保存
    request.currentFile = "current file"; // 当前编辑器
    request.fileTree = ["1/", "2/"]; // 文件树展开状态
    request.files = ["1.go", "2.go", "3.go"]; // 编辑器打开状态

    $.ajax({
        type: 'POST',
        url: '/session/save',
        data: JSON.stringify(request),
        dataType: "json",
        success: function (data) {
        }
    });
}, 30000);
