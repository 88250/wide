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

var shell = {
    _shellWS: undefined,
    _initWS: function () {
        shell.shellWS = new ReconnectingWebSocket(config.channel + '/shell/ws?sid=' + config.wideSessionId);
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