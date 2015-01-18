/*
 * Copyright (c) 2014-2015, b3log.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

var notification = {
    init: function () {
        $(".notification-count").click(function () {
            bottomGroup.tabs.setCurrent("notification");
            $(".bottom-window-group .notification").focus();
            $(this).hide();
        });

        this._initWS();
    },
    _initWS: function () {
        var notificationWS = new ReconnectingWebSocket(config.channel + '/notification/ws?sid=' + config.wideSessionId);

        notificationWS.onopen = function () {
            console.log('[notification onopen] connected');
        };

        notificationWS.onmessage = function (e) {
            var data = JSON.parse(e.data),
                    $notification = $('.bottom-window-group .notification > table'),
                    notificationHTML = '';
            
            if (data.cmd && "init-notification" === data.cmd) {
                console.log('[notification onmessage]' + e.data);
                
                return;
            }

            notificationHTML += '<tr><td class="severity">' + data.severity
                    + '</td><td class="message">' + data.message
                    + '</td><td class="type">' + data.type + '</td></tr>';
            $notification.append(notificationHTML);

            $(".notification-count").show();
        };

        notificationWS.onclose = function (e) {
            console.log('[notification onclose] disconnected (' + e.code + ')');
        };

        notificationWS.onerror = function (e) {
            console.log('[notification onerror] ' + JSON.parse(e));
        };
    }
};