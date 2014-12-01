/* 
 * Copyright (c) 2014, B3log
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

var menu = {
    init: function () {
        this.subMenu();
        this._initPreference();
        this._initAbout();

        // 点击子菜单后消失
        $(".frame li").click(function () {
            $(this).closest(".frame").hide();
            $(".menu > ul > li > a, .menu > ul> li > span").removeClass("selected");
        });
    },
    _initAbout: function () {
        $("#dialogAbout").load('/about', function () {
            $("#dialogAbout").dialog({
                "modal": true,
                "height": 460,
                "width": 800,
                "title": config.label.about,
                "hideFooter": true,
                "afterOpen": function () {
                    $.ajax({
                        url: "http://rhythm.b3log.org/version/wide/latest",
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
        $(".menu > ul > li > a, .menu > ul> li > span").click(function () {
            var $it = $(this);
            $it.next().show();
            $(".menu > ul > li > a, .menu > ul> li > span").removeClass("selected");
            $(this).addClass("selected");

            $(".menu > ul > li > a, .menu > ul> li > span").unbind();

            $(".menu > ul > li > a, .menu > ul> li > span").mouseover(function () {
                $(".frame").hide();
                $(this).next().show();
                $(".menu > ul > li > a, .menu > ul> li > span").removeClass("selected");
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
            var path = tree.fileTree.getNodeByTId(editors.data[i].id).path;
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
            success: function (data) {
                if (data.succ) {
                    window.location.href = "/login";
                }
            }
        });
    },
    openAbout: function () {
        $("#dialogAbout").dialog("open");
    },
    goget: function () {
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
            url: '/go/get',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                bottomGroup.resetOutput();
            },
            success: function (data) {
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
            beforeSend: function (data) {
                bottomGroup.resetOutput();
            },
            success: function (data) {
            }
        });
    },
    // 测试.
    test: function () {
        menu.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.test").hasClass("disabled")) {
            return false;
        }

        var request = newWideRequest();
        request.file = currentPath;

        $.ajax({
            type: 'POST',
            url: '/go/test',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                bottomGroup.resetOutput();
            },
            success: function (data) {
            }
        });
    },
    // Build & Run.
    run: function () {
        menu.saveAllFiles();

        var currentPath = editors.getCurrentPath();
        if (!currentPath) {
            return false;
        }

        if ($(".menu li.run").hasClass("disabled")) {
            return false;
        }

        if ($(".toolbars .ico-stop").length === 1) {
            wide.stop();
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
            beforeSend: function (data) {
                bottomGroup.resetOutput();
            },
            success: function (data) {
                $(".toolbars .ico-buildrun").addClass("ico-stop")
                        .removeClass("ico-buildrun").attr("title", config.label.stop);
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
        request.nextCmd = ""; // 只构建，无下一步操作

        $.ajax({
            type: 'POST',
            url: '/build',
            data: JSON.stringify(request),
            dataType: "json",
            beforeSend: function (data) {
                bottomGroup.resetOutput();
            },
            success: function (data) {
            }
        });
    },
    _initPreference: function () {
        $("#dialogPreference").load('/preference', function () {
            $("#localeSelect").on('change', function () {
                var $dialogPreference = $("#dialogPreference"),
                        $input = $dialogPreference.find("input[name=locale]")

                $input.val(this.value);
            });

            $("#themeSelect").on('change', function () {
                var $dialogPreference = $("#dialogPreference"),
                        $input = $dialogPreference.find("input[name=theme]")

                $input.val(this.value);
            });

            $("#editorThemeSelect").on('change', function () {
                var $dialogPreference = $("#dialogPreference"),
                        $input = $dialogPreference.find("input[name=editorTheme]")

                $input.val(this.value);
            });

            $("#goFmtSelect").on('change', function () {
                var $dialogPreference = $("#dialogPreference"),
                        $input = $dialogPreference.find("input[name=goFmt]")

                $input.val(this.value);
            });

            $("#dialogPreference input").keyup(function () {
                var isChange = false;
                $("#dialogPreference input").each(function () {
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

            $("#dialogPreference select").on("change", function () {
                var isChange = false;
                $("#dialogPreference input").each(function () {
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
                "height": 460,
                "width": 800,
                "title": config.label.perference,
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
                            $goFmt = $dialogPreference.find("input[name=goFmt]"),
                            $workspace = $dialogPreference.find("input[name=workspace]"),
                            $password = $dialogPreference.find("input[name=password]"),
                            $locale = $dialogPreference.find("input[name=locale]"),
                            $theme = $dialogPreference.find("input[name=theme]"),
                            $editorFontFamily = $dialogPreference.find("input[name=editorFontFamily]"),
                            $editorFontSize = $dialogPreference.find("input[name=editorFontSize]"),
                            $editorLineHeight = $dialogPreference.find("input[name=editorLineHeight]"),
                            $editorTheme = $dialogPreference.find("input[name=editorTheme]");
                            $editorTabSize = $dialogPreference.find("input[name=editorTabSize]");

                    $.extend(request, {
                        "fontFamily": $fontFamily.val(),
                        "fontSize": $fontSize.val(),
                        "goFmt": $goFmt.val(),
                        "workspace": $workspace.val(),
                        "password": $password.val(),
                        "locale": $locale.val(),
                        "theme": $theme.val(),
                        "editorFontFamily": $editorFontFamily.val(),
                        "editorFontSize": $editorFontSize.val(),
                        "editorLineHeight": $editorLineHeight.val(),
                        "editorTheme": $editorTheme.val(),
                        "editorTabSize": $editorTabSize.val()
                    });

                    $.ajax({
                        type: 'POST',
                        url: '/preference',
                        data: JSON.stringify(request),
                        success: function (data, textStatus, jqXHR) {
                            if (!data.succ) {
                                return false;
                            }

                            $fontFamily.data("value", $fontFamily.val());
                            $fontSize.data("value", $fontSize.val());
                            $goFmt.data("value", $goFmt.val());
                            $workspace.data("value", $workspace.val());
                            $password.data("value", $password.val());
                            $locale.data("value", $locale.val());
                            $theme.data("value", $theme.val());
                            $editorFontFamily.data("value", $editorFontFamily.val());
                            $editorFontSize.data("value", $editorFontSize.val());
                            $editorLineHeight.data("value", $editorLineHeight.val());
                            $editorTheme.data("value", $editorTheme.val());
                            $editorTabSize.data("value", $editorTabSize.val());

                            var $okBtn = $("#dialogPreference").closest(".dialog-main").find(".dialog-footer > button:eq(0)");
                            $okBtn.prop("disabled", true);
                        }
                    });
                }
            });

            new Tabs({
                id: ".preference"
            });
        });
    },
};