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
 * @file windows.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @author <a href="http://88250.b3log.org">Liang Ding</a>
 * @version 1.1.2.2, Feb 16, 2016
 */
var windows = {
    isMaxEditor: false,
    outerLayout: {},
    innerLayout: {},
    init: function () {
        if (!config.latestSessionContent) {
            config.latestSessionContent = {
                "fileTree": [],
                "files": [],
                "currentFile": "",
            };
        }
        
        if (!config.latestSessionContent.layout) {
            config.latestSessionContent.layout = {
                "side": {
                    "size": 200,
                    "state": 'normal'
                },
                "sideRight": {
                    "size": 200,
                    "state": 'normal'
                },
                "bottom": {
                    "size": 100,
                    "state": 'normal'
                }
            };
        }

        var layout = config.latestSessionContent.layout;

        this.outerLayout = $('body').layout({
            north__paneSelector: ".menu",
            center__paneSelector: ".content",
            south__paneSelector: ".footer",
            north__size: 22,
            south__size: 19,
            spacing_open: 2,
            north__spacing_open: 0,
            south__spacing_open: 0,
            defaults: {
                fxSpeed_open: 300,
                fxSpeed_close: 100,
                fxSettings_close: {
                    easing: "easeOutQuint"
                },
                fxSettings_open: {
                    easing: "easeInQuint"
                }
            },
            west: {
                size: layout.side.size,
                paneSelector: ".side",
                togglerLength_open: 0,
                togglerLength_closed: 15,
                togglerAlign_closed: "top",
                slideTrigger_open: "mouseover",
                spacing_closed: 15,
                minSize: 100,
                togglerClass: "ico-restore",
                togglerTip_open: config.label.min,
                togglerTip_closed: config.label.restore_side,
                resizerTip: config.label.resize,
                initClosed: (layout.side.state === 'min')
            }
        });

        this.innerLayout = $('div.content').layout({
            spacing_open: 2,
            defaults: {
                fxSpeed_open: 300,
                fxSpeed_close: 100,
                fxSettings_close: {
                    easing: "easeOutQuint"
                },
                fxSettings_open: {
                    easing: "easeInQuint"
                }
            },
            center: {
                paneSelector: ".edit-panel"
            },
            east: {
                size: layout.sideRight.size,
                paneSelector: ".side-right",
                togglerLength_open: 0,
                togglerLength_closed: 15,
                togglerAlign_closed: "top",
                slideTrigger_open: "mouseover",
                spacing_closed: 15,
                minSize: 100,
                togglerClass: "ico-restore",
                togglerTip_open: config.label.min,
                togglerTip_closed: config.label.restore_outline,
                resizerTip: config.label.resize,
                initClosed: (layout.sideRight.state === 'min')
            },
            south: {
                size: layout.bottom.size,
                paneSelector: ".bottom-window-group",
                togglerLength_open: 0,
                togglerLength_closed: 15,
                togglerAlign_closed: "top",
                slideTrigger_open: "mouseover",
                spacing_closed: 16,
                minSize: 100,
                togglerClass: "ico-restore",
                togglerTip_open: config.label.min,
                togglerTip_closed: config.label.restore_bottom,
                resizerTip: config.label.resize,
                initClosed: (layout.bottom.state === 'min'),
                ondrag_end: function (type, pane) {
                    windows.refreshEditor(pane, 'drag');
                },
                onresize_end: function (type, pane) {
                    windows.refreshEditor(pane, 'resize');
                },
                onclose_end: function (type, pane) {
                    windows.refreshEditor(pane, 'close');
                },
                onopen_end: function (type, pane) {
                    windows.refreshEditor(pane, 'open');
                },
                onshow_end: function (type, pane) {
                    windows.refreshEditor(pane, 'show');
                }
            }
        });

        this.outerLayout.addCloseBtn(".side .ico-min", "west");
        this.innerLayout.addCloseBtn(".side-right .ico-min", "east");
        this.innerLayout.addCloseBtn(".bottom-window-group .ico-min", "south");

        if (layout.side.state === 'max') {
            windows.maxSide();
        }
        if (layout.sideRight.state === 'max') {
            windows.maxSideRight();
        }
        if (layout.bottom.state === 'max') {
            windows.maxBottom();
        }

        $(".toolbars .ico-max").click(function () {
            windows.toggleEditor();
        });

        $(".edit-panel .tabs").on("dblclick", function () {
            windows.toggleEditor();
        });

        $(".bottom-window-group .tabs").dblclick(function () {
            var $it = $(".bottom-window-group");
            if ($it.hasClass("bottom-window-group-max")) {
                windows.restoreBottom();
            } else {
                windows.maxBottom($it);
            }
        });

        $(".side .tabs").dblclick(function () {
            var $it = $(".side");
            if ($it.hasClass("side-max")) {
                windows.restoreSide();
            } else {
                windows.restoreSide($it);
            }
        });

        $(".side-right .tabs").dblclick(function () {
            var $it = $(".side-right");
            if ($it.hasClass("side-right-max")) {
                windows.restoreSideRight();
            } else {
                windows.maxSideRight($it);
            }
        });

        $('.bottom-window-group .search').height($('.bottom-window-group .tabs-panel').height());
        $(window).resize(function () {
            windows.refreshEditor($('.bottom-window-group'));
        });

    },
    maxEditor: function () {
        var $it = $(".toolbars .font-ico");
        windows.outerLayout.close('west');
        windows.innerLayout.close('south');
        windows.innerLayout.close('east');
        $it.removeClass('ico-max').addClass('ico-restore').attr('title', config.label.min);
        windows.isMaxEditor = true;
    },
    maxBottom: function ($it) {
        $it.data('height', $it.height()).addClass("bottom-window-group-max").find('.ico-min').hide();
        windows.outerLayout.hide('west');
        windows.innerLayout.hide('east');
        windows.innerLayout.sizePane('south', $('.content').height());
    },
    maxSide: function ($it) {
        $it.data('width', $it.width()).addClass("side-max").find('.ico-min').hide();
        $('.content').hide();
        windows.outerLayout.sizePane('west', $('body').width());
    },
    maxSideRight: function ($it) {
        $it.addClass("side-right-max").data('width', $it.width()).find('.ico-min').hide();
        windows.outerLayout.hide('west');
        windows.innerLayout.hide('south');
        windows.innerLayout.sizePane('east', $('body').width());
    },
    toggleEditor: function () {
        var $it = $(".toolbars .font-ico");
        if ($it.hasClass('ico-restore')) {
            windows.restoreEditor();
        } else {
            windows.maxEditor();
        }
    },
    restoreBottom: function () {
        var $it = $(".bottom-window-group");
        $it.removeClass("bottom-window-group-max").find('.ico-min').show();
        windows.outerLayout.show('west');
        windows.innerLayout.show('east');
        windows.innerLayout.sizePane('south', $it.data('height'));
    },
    restoreSide: function () {
        var $it = $(".side");
        $it.removeClass("side-max").find('.ico-min').show();
        $('.content').show();
        windows.outerLayout.sizePane('west', $it.data('width'));
    },
    restoreSideRight: function () {
        var $it = $(".side-right");
        $it.removeClass("side-right-max").find('.ico-min').show();
        windows.outerLayout.show('west');
        windows.innerLayout.show('south');
        windows.innerLayout.sizePane('east', $it.data('width'));
    },
    restoreEditor: function () {
        windows.outerLayout.open('west');
        windows.innerLayout.open('south');
        windows.innerLayout.open('east');
        windows.isMaxEditor = false;
        $(".toolbars .font-ico").addClass('ico-max').removeClass('ico-restore').attr('title', config.label.max_editor);
    },
    refreshEditor: function (pane, type) {
        var editorDatas = editors.data,
                height = $('.content').height() - pane.height() - 24;
        switch (type) {
            case 'close':
                height = $('.content').height() - 40;
                break;
            default:
                break;
        }
        for (var i = 0, ii = editorDatas.length; i < ii; i++) {
            editorDatas[i].editor.setSize("100%", height);
        }

        $('.bottom-window-group .search').height($('.bottom-window-group .tabs-panel').height());
    },
    flowBottom: function () {
        if (windows.innerLayout.south.state.isClosed) {
            windows.innerLayout.slideOpen('south');
        }
    }
};