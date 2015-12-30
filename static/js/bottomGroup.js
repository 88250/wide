/*
 * Copyright (c) 2014-2016, b3log.org & hacpai.com
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

/*
 * @file bottomGroup.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @author <a href="http://88250.b3log.org">Liang Ding</a>
 * @version 1.1.0.1, Dec 8, 2015
 */
var bottomGroup = {
    tabs: undefined,
    searchTab: undefined,
    init: function () {
        this._initTabs();
        this._initFrame();

        $('.bottom-window-group .output').click(function () {
            $(this).focus();
        });

        $('.bottom-window-group .output').on('click', '.path', function (event) {
            var $path = $(this),
                    tId = tree.getTIdByPath($path.data("path"));
            tree.openFile(tree.fileTree.getNodeByTId(tId),
                    CodeMirror.Pos($path.data("line") - 1, $path.data("column") - 1));
            event.preventDefault();
            return false;
        });
    },
    _initFrame: function () {
        $(".bottom-window-group .output").parent().mouseup(function (event) {
            event.stopPropagation();

            if (event.button === 0) { // 左键
                $(".bottom-window-group .frame").hide();
                return;
            }

            // event.button === 2 右键
            var left = event.screenX,
                    $it = $(this);
            if ($(".side").css("left") === "auto" || $(".side").css("left") === "0px") {
                left = event.screenX - $(".side").width();
            }
            $(".bottom-window-group .frame").show().css({
                "left": left + "px",
                "top": (event.offsetY + event.target.offsetTop - $it.scrollTop() - 10) + "px"
            });
            return;
        });
    },
    clear: function (id) {
        $('.bottom-window-group .' + id + ' > div').text('');
    },
    resetOutput: function () {
        this.clear('output');
        bottomGroup.tabs.setCurrent("output");
        windows.flowBottom();
    },
    _initTabs: function () {
        this.tabs = new Tabs({
            id: ".bottom-window-group",
            clickAfter: function (id) {
                this._$tabsPanel.find("." + id).focus();
            }
        });
    },
    fillOutput: function (data) {
        var $output = $('.bottom-window-group .output');

        data = data.replace(/\n/g, '<br/>');

        if (-1 !== data.indexOf("<br/>")) {
            data = Autolinker.link(data);
        }

        $output.find("div").html(data);
        $output.parent().scrollTop($output[0].scrollHeight);
    }
};
