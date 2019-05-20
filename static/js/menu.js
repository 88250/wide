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
 * @file menu.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @author <a href="http://88250.b3log.org">Liang Ding</a>
 * @version 1.0.1.3, Oct 5, 2018
 */
var menu = {
    init: function () {
        this.subMenu();
        this._initPreference();
        this._initAbout();
        this._initShare();

        // 点击子菜单后消失
        $(".menu .frame li").click(function () {
            $(".menu > ul > li").unbind().removeClass("selected");
            menu.subMenu();
        });
    },
    _initShare: function () {
        $(".menu .ico-share").hover(function () {
            $(".menu .share-panel").show();
        });

        $(".share-panel .font-ico").click(function () {
            var key = $(this).attr('class').split('-')[2];
            var url = "https://wide.b3log.org", pic = 'https://wide.b3log.org/static/images/wide-logo.png';
            var urls = {};
            urls.email = "mailto:?subject=" + $('title').text()
                    + "&body=" + $('meta[name=description]').attr('content') + ' ' + url;

            var twitterShare = encodeURIComponent($('meta[name=description]').attr('content') + " " + url + " #golang");
            urls.twitter = "https://twitter.com/intent/tweet?status=" + twitterShare;

            urls.facebook = "https://www.facebook.com/sharer/sharer.php?u=" + url;
            urls.googleplus = "https://plus.google.com/share?url=" + url;

            var title = encodeURIComponent($('title').text() + '. \n' + $('meta[name=description]').attr('content')
                    + " #golang#");
            urls.weibo = "http://v.t.sina.com.cn/share/share.php?title=" + title + "&url=" + url + "&pic=" + pic;
            urls.qqz = "https://sns.qzone.qq.com/cgi-bin/qzshare/cgi_qzshare_onekey?url=" + url + "&sharesource=qzone&title=" + title+ "&pics=" + pic;

            window.open(urls[key], "_blank", "top=100,left=200,width=648,height=618");
        });
    },
    _initAbout: function () {
        $("#dialogAbout").load('/about', function () {
            $("#dialogAbout").dialog({
                "modal": true,
                "title": config.label.about,
                "hideFooter": true,
                "afterOpen": function () {
                    $.ajax({
                        url: "https://rhythm.b3log.org/version/wide/latest",
                        type: "GET",
                        dataType: "jsonp",
                        jsonp: "callback",
                        success: function (data, textStatus) {
                            if ($("#dialogAbout .version").text() === data.wideVersion) {
                                $(".upgrade").text(config.label.uptodate);
                            } else {
                                $(".upgrade").html(config.label.new_version_available + config.label.colon
                                        + "<a href='" + data.wideDownload
                                        + "' target='_blank'>" + data.wideVersion + "</a>");
                            }
                        }
                    });
                }
            });
        });
    },
    disabled: function (list) {
        for (var i = 0, max = list.length; i < max; i++) {
            $(".menu li." + list[i]).addClass("disabled");
        }
    },
    undisabled: function (list) {
        for (var i = 0, max = list.length; i < max; i++) {
            $(".menu li." + list[i]).removeClass("disabled");
        }
    },
    // 焦点不在菜单上时需点击展开子菜单，否则为鼠标移动展开
    subMenu: function () {
        $(".menu > ul > li").click(function (event) {
            if ($(event.target).closest(".frame").length === 1) {
                return;
            }
            var $it = $(this);
            $it.find('.frame').show();
            $(".menu > ul > li").removeClass("selected");
            $(this).addClass("selected");

            $(".menu > ul > li").unbind();

            $(".menu > ul > li").mouseover(function () {
                if ($(event.target).closest(".frame").length === 1) {
                    return;
                }
                $(".menu .frame").hide();
                $(this).find('.frame').show();
                $(".menu > ul > li").removeClass("selected");
                $(this).addClass("selected");
            });
        });
    },
    openPreference: function () {
        $("#dialogPreference").dialog("open");
    },
    saveAllFiles: function () {
        if ($(".menu li.save-all").hasClass("disabled")) {
            return false;
        }
        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            var path = editors.data[i].id;
            var editor = editors.data[i].editor;

            if ("text/x-go" === editor.getOption("mode")) {
                wide.fmt(path, editor);
            } else {
                wide._save(path, editor);
            }
        }
    },
    closeAllFiles: function () {
        if ($(".menu li.close-all").hasClass("disabled")) {
            return false;
        }

        // 设置全部关闭标识
        var removeData = [];
        $(".edit-panel .tabs > div").each(function (i) {
            if (i !== 0) {
                removeData.push($(this).data("index"));
            }
        });
        $("#dialogCloseEditor").data("removeData", removeData);
        // 开始关闭
        $(".edit-panel .tabs .ico-close:eq(0)").click();
    },
    exit: function () {
        var request = newWideRequest();

        $.ajax({
            type: 'POST',
            url: '/logout',
            data: JSON.stringify(request),
            dataType: "json",
            success: function (result) {
                if (result.succ) {
                    window.location.href = "/login";
                }
            }
        });
    },
    openAbout: function () {
        $("#dialogAbout").dialog("open");
    },
    gomod: function () {
        menu.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.go-get").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;

        $.ajax({
            type: 'POST',
            url: '/go/mod',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function () {
                bottomGroup.resetOutput();
            },
            success: function (result) {
            }
        });
    },
    goinstall: function () {
        menu.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.go-install").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;

        $.ajax({
            type: 'POST',
            url: '/go/install',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function () {
                bottomGroup.resetOutput();
            },
            success: function (result) {
            }
        });
    },
    // go test.
    test: function () {
        menu.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.go-test").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;

        $.ajax({
            type: 'POST',
            url: '/go/test',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function () {
                bottomGroup.resetOutput();
            },
            success: function (result) {
            }
        });
    },
    // go vet.
    govet: function () {
        menu.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.go-vet").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;

        $.ajax({
            type: 'POST',
            url: '/go/vet',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function () {
                bottomGroup.resetOutput();
            },
            success: function (result) {
            }
        });
    },
    // Build & Run.
    run: function () {
        menu.saveAllFiles();

        if ($("#buildRun").hasClass("ico-stop")) {
            wide.stop();
            return false;
        }

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.run").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;
        request.code = wide.curEditor.getValue();
        request.nextCmd = "run";

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function () {
                bottomGroup.resetOutput();

                $("#buildRun").addClass("ico-stop")
                        .removeClass("ico-buildrun").attr("title", config.label.stop);
            },
            success: function (result) {
            }
        });
    },
    // Build.
    build: function () {
        menu.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.build").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;
        request.code = wide.curEditor.getValue();
        request.nextCmd = ""; // build only, no following operation

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function () {
                bottomGroup.resetOutput();
            },
            success: function (result) {
            }
        });
    },
    _initPreference: function () {
        $("#dialogPreference").load('/preference', function () {
            $("#dialogPreference input").keyup(function () {
                var isChange = false,
                        emptys = [],
                        emptysTip = '';
                $("#dialogPreference input").each(function () {
                    var $it = $(this);
                    // data-value 如为数字，则不会和 value 一样转换为 String，再次不使用全等
                    if ($it.val() != $it.data("value")) {
                        isChange = true;
                    }

                    if ($.trim($it.val()) === '') {
                        emptys.push($it);
                    }
                });

                var $okBtn = $("#dialogPreference").closest(".dialog-main").find(".dialog-footer > button:eq(0)");
                if (isChange) {
                    $okBtn.prop("disabled", false);
                } else {
                    $okBtn.prop("disabled", true);
                }

                if (emptys.length === 0) {
                    $("#dialogPreference").find(".tip").html("");
                    $okBtn.prop("disabled", false);
                } else {
                    for (var i = 0, max = emptys.length; i < max; i++) {
                        var tabIndex = emptys[i].closest('div').data("index"),
                                text = $.trim(emptys[i].parent().text());
                        emptysTip += '[' + $('#dialogPreference .tabs > div[data-index="' + tabIndex + '"]').text()
                                + '] -> [' + text.substr(0, text.length - 1)
                                + ']: ' + config.label.no_empty + "<br/>";
                    }
                    $("#dialogPreference").find(".tip").html(emptysTip);
                    $okBtn.prop("disabled", true);
                }
            });

            $("#dialogPreference select").on("change", function () {
                var isChange = false;
                $("#dialogPreference select").each(function () {
                    if ($(this).val() !== $(this).data("value")) {
                        isChange = true;
                    }
                });

                var $okBtn = $("#dialogPreference").closest(".dialog-main").find(".dialog-footer > button:eq(0)");
                if (isChange) {
                    $okBtn.prop("disabled", false);
                } else {
                    $okBtn.prop("disabled", true);
                }
            });

            $("#dialogPreference").dialog({
                "modal": true,
                "height": 280,
                "width": 800,
                "title": config.label.preference,
                "okText": config.label.apply,
                "cancelText": config.label.cancel,
                "afterOpen": function () {
                    var $okBtn = $("#dialogPreference").closest(".dialog-main").find(".dialog-footer > button:eq(0)");
                    $okBtn.prop("disabled", true);
                },
                "ok": function () {
                    var request = newWideRequest(),
                            $dialogPreference = $("#dialogPreference"),
                            $fontFamily = $dialogPreference.find("input[name=fontFamily]"),
                            $fontSize = $dialogPreference.find("input[name=fontSize]"),
                            $goFmt = $dialogPreference.find("select[name=goFmt]"),
                            $GoBuildArgsForLinux = $dialogPreference.find("input[name=GoBuildArgsForLinux]"),
                            $GoBuildArgsForWindows = $dialogPreference.find("input[name=GoBuildArgsForWindows]"),
                            $GoBuildArgsForDarwin = $dialogPreference.find("input[name=GoBuildArgsForDarwin]"),
                            $workspace = $dialogPreference.find("input[name=workspace]"),
                            $password = $dialogPreference.find("input[name=password]"),
                            $email = $dialogPreference.find("input[name=email]"),
                            $locale = $dialogPreference.find("select[name=locale]"),
                            $theme = $dialogPreference.find("select[name=theme]"),
                            $editorFontFamily = $dialogPreference.find("input[name=editorFontFamily]"),
                            $editorFontSize = $dialogPreference.find("input[name=editorFontSize]"),
                            $editorLineHeight = $dialogPreference.find("input[name=editorLineHeight]"),
                            $editorTheme = $dialogPreference.find("select[name=editorTheme]"),
                            $editorTabSize = $dialogPreference.find("input[name=editorTabSize]"),
                            $keymap = $dialogPreference.find("select[name=keymap]");

                    $.extend(request, {
                        "fontFamily": $fontFamily.val(),
                        "fontSize": $fontSize.val(),
                        "goFmt": $goFmt.val(),
                        "GoBuildArgsForLinux": $GoBuildArgsForLinux.val(),
                        "GoBuildArgsForWindows": $GoBuildArgsForWindows.val(),
                        "GoBuildArgsForDarwin": $GoBuildArgsForDarwin.val(),
                        "workspace": $workspace.val(),
                        "password": $password.val(),
                        "locale": $locale.val(),
                        "theme": $theme.val(),
                        "editorFontFamily": $editorFontFamily.val(),
                        "editorFontSize": $editorFontSize.val(),
                        "editorLineHeight": $editorLineHeight.val(),
                        "editorTheme": $editorTheme.val(),
                        "editorTabSize": $editorTabSize.val(),
                        "keymap": $keymap.val()
                    });

                    if (config.keymap !== $keymap.val()) {
                        window.location.reload();
                    }

                    $.ajax({
                        type: 'POST',
                        url: '/preference',
                        data: JSON.stringify(request),
                        success: function (result, textStatus, jqXHR) {
                            if (!result.succ) {
                                return false;
                            }

                            $fontFamily.data("value", $fontFamily.val());
                            $fontSize.data("value", $fontSize.val());
                            $goFmt.data("value", $goFmt.val());
                            $GoBuildArgsForLinux.data("value", $GoBuildArgsForLinux.val());
                            $GoBuildArgsForWindows.data("value", $GoBuildArgsForWindows.val());
                            $GoBuildArgsForDarwin.data("value", $GoBuildArgsForDarwin.val());
                            $workspace.data("value", $workspace.val());
                            $password.data("value", $password.val());
                            $email.data("value", $email.val());
                            $locale.data("value", $locale.val());
                            $theme.data("value", $theme.val());
                            $editorFontFamily.data("value", $editorFontFamily.val());
                            $editorFontSize.data("value", $editorFontSize.val());
                            $editorLineHeight.data("value", $editorLineHeight.val());
                            $editorTheme.data("value", $editorTheme.val());
                            $editorTabSize.data("value", $editorTabSize.val());
                            $keymap.data("value", $keymap.val());

                            // update the config
                            config.keymap = $keymap.val();

                            var $okBtn = $("#dialogPreference").closest(".dialog-main").find(".dialog-footer > button:eq(0)");
                            $okBtn.prop("disabled", true);

                            $("#themesLink").attr("href", '/static/css/themes/' + $theme.val() + '.css');

                            config.editorTheme = $editorTheme.val();
                            for (var i = 0, ii = editors.data.length; i < ii; i++) {
                                editors.data[i].editor.setOption("theme", $editorTheme.val());
                            }
                        }
                    });
                }
            });

            new Tabs({
                id: ".preference"
            });
        });
    }
};
