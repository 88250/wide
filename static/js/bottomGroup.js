var bottomGroup = {
    tabs: undefined,
    searchTab: undefined,
    init: function () {
        this._initTabs();
        this._initFrame();
    },
    _initFrame: function () {
        $(".bottom-window-group .output").mousedown(function (event) {
            event.stopPropagation();

            if (event.button === 0) { // 左键
                $(".bottom-window-group .frame").hide();
                return false;
            }

            // event.button === 2 右键
            var left = event.screenX;
            if ($(".side").css("left") === "auto" || $(".side").css("left") === "0px") {
                left = event.screenX - $(".side").width();
            }
            $(".bottom-window-group .frame").show().css({
                "left": left + "px",
                "top": (event.offsetY + 20) + "px"
            });
            return false;
        });
    },
    clear: function () {
        $('.bottom-window-group .output > div').text('');
    },
    clearOutput: function () {
        $('.bottom-window-group .output > div').text('');
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
        $output.find("div").html(data.replace(/\n/g, '<br/>'));
        $output.parent().scrollTop($output[0].scrollHeight);
    }
};