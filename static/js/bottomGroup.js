var bottomGroup = {
    tabs: undefined,
    searchTab: undefined,
    init: function () {
        this._initTabs();
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