/*
 * Copyright (c) 2014-present, b3log.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
 * @file dialog.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @version 1.0.0.1, Dec 8, 2015
 */
(function ($) {
    $.fn.extend({
        dialog: {
            version: "0.0.1.7",
            author: "v@b3log.org"
        }
    });

    var dpuuid = new Date().getTime();
    var PROP_NAME = 'dialog';

    var Dialog = function () {
        this._defaults = {
            "styleClass": {
                "background": "dialog-background",
                "panel": "dialog-panel",
                "main": "dialog-main",
                "footer": "dialog-footer",
                "headerMiddle": "dialog-header-middle",
                "headerBg": "dialog-header-bg",
                "closeIcon": "dialog-close-icon",
                "closeIconHover": "dialog-close-icon-hover",
                "title": "dialog-title"
            }
        };
    };

    $.extend(Dialog.prototype, {
        _attach: function (target, settings) {
            if (!target.id) {
                this.uuid++;
                target.id = 'dp' + this.uuid;
            }
            var inst = this._newInst($(target));

            inst.settings = $.extend({}, settings || {});
            $.data(target, PROP_NAME, inst);
            this._init(target);
        },
        /* Create a new instance object. */
        _newInst: function (target) {
            // escape jQuery meta chars
            var id = target[0].id.replace(/([^A-Za-z0-9_])/g, '\\\\$1');
            return {
                id: id
            };
        },
        _getInst: function (target) {
            try {
                return $.data(target, PROP_NAME);
            } catch (err) {
                throw 'Missing instance data for this dialog';
            }
        },
        _destroyDialog: function (target) {
            var inst = $.dialog._getInst(target);
            var id = inst.id;
            $.removeData(target, PROP_NAME);

            $(target).prependTo("#" + id + "Wrap").unwrap();
            $(target).removeAttr("style");

            var styleClass = this._getDefaults($.dialog._defaults, inst.settings, "styleClass");
            $("." + styleClass.background).remove();
            $("#" + id + "Dialog").remove();
        },
        _init: function (target) {
            var inst = this._getInst(target);
            var id = inst.id,
                    settings = inst.settings;
            var windowH = $(window).height(),
                    windowW = $(window).width();
            var styleClass = this._getDefaults($.dialog._defaults, settings, "styleClass"),
                    dialogH = settings.height ? settings.height : parseInt(windowH * 0.6),
                    dialogW = settings.width ? settings.width : parseInt(windowW * 0.6);

            // get settings or default value.
            settings.title = settings.title ? settings.title : "";
            settings.okText = settings.okText ? settings.okText : "Ok";
            settings.cancelText = settings.cancelText ? settings.cancelText : "Cancel";

            // build HTML.
            var footerHTML = "",
                    headerHTML = "<div class='"
                    + styleClass.headerBg + "'><div class='"
                    + styleClass.title + "'>"
                    + settings.title + "</div><a href='javascript:void(0);' class='ico-close font-ico "
                    + styleClass.closeIcon + "'></a></div>";

            // Sets footerHTML.
            if (!settings.hideFooter) {
                if (!settings.hiddenOk) {
                    footerHTML = "<button>" + settings.okText +
                            "</button>";
                }
                footerHTML += "<button>"
                        + settings.cancelText + "</button>";
            }

            var dialogHTML = "<div id='" + id + "Dialog' class='" + styleClass.panel
                    + "' style='width: " + dialogW + "px;' onselectstart='return false;'>" + headerHTML
                    + "<div class='" + styleClass.main + "'><div style='overflow: auto; height: "
                    + dialogH + "px;'></div><div class='" + styleClass.footer + "'>"
                    + footerHTML + "</div></div>";

            var bgHTML = "";
            if (settings.modal && $("." + styleClass.background).length === 0) {
                var bgHeight = windowH < document.documentElement.scrollHeight
                        ? document.documentElement.scrollHeight : windowH;
                bgHTML = "<div style='height:" + bgHeight
                        + "px;' class='" + styleClass.background + "'></div>";
            }

            // Package dialog.
            $("#" + id).wrap("<div id='" + id + "Wrap'></div>");
            var cloneObj = $(target).clone(true);
            $(target).remove();
            $('body').append(bgHTML + dialogHTML);
            $($("#" + id + "Dialog ." + styleClass.main + " div").get(0)).append(cloneObj);
            $(cloneObj).show();

            // Bind event.
            $("#" + id + "Dialog ." + styleClass.closeIcon).bind("click", function () {
                $.dialog._close(id, settings);
            });

            var $buttons = $("#" + id + "Dialog ." + styleClass.footer + " button");
            $($buttons.get(1)).bind("click", function () {
                $.dialog._close(id, settings);
            });

            $($buttons.get(0)).bind("click", function () {
                if (settings.ok === undefined || settings.ok()) {
                    $.dialog._close(id, settings);
                }
            });

            this._bindMove(id, styleClass.headerBg, dialogH, dialogW);

            // esc exit
            $(window).keyup(function (event) {
                if (event.keyCode === 27) {
                    $.dialog._close(id, settings);
                }
            });

            $(window).resize(function () {
                var height = $("body").height() > $(window).height() ? $("body").height() : $(window).height();
                $(".dialog-background").height(height);
            });

            if (typeof settings.afterInit === "function") {
                settings.afterInit();
            }
        },
        _bindMove: function (id, className) {
            $("#" + id + "Dialog ." + className).mousedown(function (event) {
                var _document = document;
                if (!event) {
                    event = window.event;
                }
                var dialog = document.getElementById(id + "Dialog");
                var x = event.clientX - parseInt(dialog.style.left),
                        y = event.clientY - parseInt(dialog.style.top);
                _document.ondragstart = "return false;";
                _document.onselectstart = "return false;";
                _document.onselect = "document.selection.empty();";

                if (this.setCapture) {
                    this.setCapture();
                } else if (window.captureEvents) {
                    window.captureEvents(Event.MOUSEMOVE | Event.MOUSEUP);
                }

                _document.onmousemove = function (event) {
                    if (!event) {
                        event = window.event;
                    }
                    var positionX = event.clientX - x,
                            positionY = event.clientY - y;
                    if (positionX < 0) {
                        positionX = 0;
                    }
                    if (positionX > $(window).width() - $(dialog).width()) {
                        positionX = $(window).width() - $(dialog).width();
                    }
                    if (positionY > $(window).height() - $(dialog).height()) {
                        positionY = $(window).height() - $(dialog).height();
                    }
                    if (positionY < 0) {
                        positionY = 0;
                    }
                    dialog.style.left = positionX + "px";
                    dialog.style.top = positionY + "px";
                };

                _document.onmouseup = function () {
                    if (this.releaseCapture) {
                        this.releaseCapture();
                    } else if (window.captureEvents) {
                        window.captureEvents(Event.MOUSEMOVE | Event.MOUSEUP);
                    }
                    _document.onmousemove = null;
                    _document.onmouseup = null;
                    _document.ondragstart = null;
                    _document.onselectstart = null;
                    _document.onselect = null;
                };
            });
        },
        _close: function (id, settings) {
            if ($("#" + id + "Dialog").css("display") === "none") {
                return;
            }
            if (settings.close === undefined || settings.close()) {
                $("#" + id + "Dialog").hide();
                if (settings.modal) {
                    var styleClass = this._getDefaults($.dialog._defaults, settings, "styleClass");
                    $("." + styleClass.background).hide();
                }
            }
        },
        _closeDialog: function (target) {
            var inst = this._getInst(target);
            var id = inst.id,
                    settings = inst.settings;
            $.dialog._close(id, settings);
        },
        _openDialog: function (target, msg) {
            var inst = this._getInst(target);
            var id = inst.id,
                    settings = inst.settings,
                    top = "", left = "",
                    $dialog = $("#" + id + "Dialog"),
                    windowH = $(window).height(),
                    windowW = $(window).width(),
                    dialogH = settings.height ? settings.height : parseInt(windowH * 0.6),
                    dialogW = settings.width ? settings.width : parseInt(windowW * 0.6);

            // Sets position.
            if (settings.position) {
                top = settings.position.top;
                left = settings.position.left;
            } else {
                // 20(footer) + 23(header)
                top = parseInt((windowH - dialogH - 43) / 2);
                if (top < 0) {
                    top = 0;
                }
                left = parseInt((windowW - dialogW) / 2);
            }
            $dialog.css({
                "top": top + "px",
                "left": left + "px"
            }).show();

            if (settings.modal) {
                var styleClass = this._getDefaults($.dialog._defaults, settings, "styleClass");
                $("." + styleClass.background).show();
            }

            if (typeof settings.afterOpen === "function") {
                settings.afterOpen(msg);
            }

            $("#" + id + "Dialog .dialog-footer button:eq(0)").focus();
        },
        _updateDialog: function (target, data) {
            var inst = this._getInst(target);
            var id = inst.id,
                    settings = inst.settings;
            var styleClass = this._getDefaults($.dialog._defaults, settings, "styleClass");
            $.extend(settings, data);
            var $dialog = $("#" + id + "Dialog");
            if (data.position) {
                $dialog.css({
                    "top": data.position.top,
                    "left": data.position.left
                });
            }

            if (data.width) {
                $dialog.width(data.width + 26);
                $dialog.find("." + styleClass.main + " div")[0].style.width = data.width + "px";
                $dialog.find("." + styleClass.headerBg).width(data.width + 18);
            }

            if (data.height) {
                $dialog.find("." + styleClass.main + " div")[0].style.height = data.height + "px";
            }

            if (data.title) {
                $dialog.find("." + styleClass.title).html(data.title);
            }

            if (data.modal !== undefined) {
                if (data.modal) {
                    $("." + styleClass.background).show();
                } else {
                    $("." + styleClass.background).hide();
                }
            }

            if (data.hideFooter !== undefined) {
                if (data.hideFooter) {
                    $dialog.find("." + styleClass.footer).hide();
                } else {
                    $dialog.find("." + styleClass.footer).show();
                }
            }

        },
        _getDefaults: function (defaults, settings, key) {
            if (key === "styleClass") {
                if (settings.theme === "default" || settings.theme === undefined) {
                    return defaults.styleClass;
                }
                settings.styleClass = {};
                for (var styleName in defaults[key]) {
                    settings.styleClass[styleName] = settings.theme + "-" + defaults.styleClass[styleName];
                }
            } else if (key === "height" || key === "width") {
                if (settings[key] === null || settings[key] === undefined) {
                    return "auto";
                } else {
                    return settings[key] + "px";
                }
            } else {
                if (settings[key] === null || settings[key] === undefined) {
                    return defaults[key];
                }
            }
            return settings[key];
        }
    });

    $.fn.dialog = function (options) {
        var otherArgs = Array.prototype.slice.call(arguments);

        if (typeof options === 'string') {
            otherArgs.shift();
            return $.dialog['_' + options + 'Dialog'].apply($.dialog, [this[0]].concat(otherArgs));
        }
        return this.each(function () {
            $.dialog._attach(this, options);
        });
    };

    $.dialog = new Dialog();

    // Add another global to avoid noConflict issues with inline event handlers
    window['DP_jQuery_' + dpuuid] = $;
})(jQuery);