var notification = {
    init: function() {
        $(".notification-count").click(function() {
            wide.bottomWindowTab.setCurrent("notification");
            $(".bottom-window-group .notification").focus();
            $(this).hide();
        });

        this._initWS();
    },
    _initWS: function() {
        var notificationWS = new WebSocket(config.channel.shell + '/notification/ws?sid=' + config.wideSessionId);

        notificationWS.onopen = function() {
            console.log('[notification onopen] connected');
        };

        notificationWS.onmessage = function(e) {
            var data = JSON.parse(e.data);
            if ('init-notification' !== data.cmd) {
                $(".notification-count").show();
                $('.bottom-window-group .notification').text(
                        $('.bottom-window-group .notification').text() + data.output);
            }
        };

        notificationWS.onclose = function(e) {
            console.log('[notification onclose] disconnected (' + e.code + ')');
            delete notificationWS;
        };

        notificationWS.onerror = function(e) {
            console.log('[notification onerror] ' + e);
        };
    }
};