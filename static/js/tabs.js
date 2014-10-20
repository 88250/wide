var Tabs = function (obj) {
    obj._$tabsPanel = $(obj.id + " > .tabs-panel");
    obj._$tabs = $(obj.id + " > .tabs");

    this.obj = obj;

    this._init(obj);
};

$.extend(Tabs.prototype, {
    _init: function (obj) {
        var _that = this;

        obj._$tabs.on("click", "div", function (event) {
            var id = $(this).data("index");
            _that.setCurrent(id);
            if (typeof (obj.clickAfter) === "function") {
                obj.clickAfter(id);
            }
        });

        obj._$tabs.on("click", ".ico-close", function (event) {
            var id = $(this).parent().data("index");
            _that.del(id);
            event.stopPropagation();
        });
    },
    _hasId: function (id) {
        var $tabs = this.obj._$tabs;
        if ($tabs.find("div[data-index=" + id + "]").length === 0) {
            return false;
        }
        return true;
    },
    add: function (data) {
        // 添加当前 tab
        if (this.getCurrentId() === data.id) {
            return false;
        }

        // 当前 tab 已经存在
        if (this._hasId(data.id)) {
            this.setCurrent(data.id);
            return false;
        }

        var $tabsPanel = this.obj._$tabsPanel,
                $tabs = this.obj._$tabs;

        this.obj._prevId = $tabs.children("div.current").data("index");

        $tabs.children("div").removeClass("current");
        $tabsPanel.children("div").hide();

        $tabs.append('<div class="current" data-index="' + data.id + '">'
                + data.title + '<span class="ico-close font-ico"></span></div>');
        $tabsPanel.append('<div data-index="' + data.id + '">' + data.content
                + '</div>');
        
        if (typeof data.after === 'function') {
            data.after();
        }
    },
    del: function (id) {
        var $tabsPanel = this.obj._$tabsPanel,
                $tabs = this.obj._$tabs,
                prevId = undefined,
                currentId = $tabs.children(".current").data("index");
        $tabs.children("div[data-index='" + id + "']").remove();
        $tabsPanel.children("div[data-index='" + id + "']").remove();

        if (this.obj._prevId === id) {
            this.obj._prevId = $tabs.children("div:first").data("index");
        }

        if (currentId !== id) {
            prevId = currentId;
        } else {
            prevId = this.obj._prevId;
        }

        this.obj.removeAfter(id, prevId);
        this.setCurrent(prevId);
    },
    getCurrentId: function () {
        var $tabs = this.obj._$tabs;
        return $tabs.children(".current").data("index");
    },
    setCurrent: function (id) {
        if (!id) {
            return false;
        }

        var $tabsPanel = this.obj._$tabsPanel,
                $tabs = this.obj._$tabs;

        var $currentTab = $tabs.children(".current");
        if ($currentTab.data("index") === id) {
            return false;
        }

        this.obj._prevId = $currentTab.data("index");

        $tabs.children("div").removeClass("current");
        $tabsPanel.children("div").hide();

        $tabs.children("div[data-index='" + id + "']").addClass("current");
        $tabsPanel.children("div[data-index='" + id + "']").show();
    }
});