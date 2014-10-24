var Tabs = function (obj) {
    obj._$tabsPanel = $(obj.id + " > .tabs-panel");
    obj._$tabs = $(obj.id + " > .tabs");
    obj._stack = [];

    this.obj = obj;
    this.obj.STACKSIZE = 64;

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

        $tabs.append('<div data-index="' + data.id + '">'
                + data.title + '<span class="ico-close font-ico"></span></div>');
        $tabsPanel.append('<div data-index="' + data.id + '">' + data.content
                + '</div>');

        this.setCurrent(data.id);

        if (typeof data.after === 'function') {
            data.after();
        }
    },
    del: function (id) {
        var $tabsPanel = this.obj._$tabsPanel,
                $tabs = this.obj._$tabs,
                stack = this.obj._stack,
                prevId = null;
        $tabs.children("div[data-index='" + id + "']").remove();
        $tabsPanel.children("div[data-index='" + id + "']").remove();

        // 移除堆栈中该 id
        for (var i = 0; i < stack.length; i++) {
            if (id === stack[i]) {
                stack.splice(i, 1);
            }
        }
        
        prevId = stack[stack.length - 1];

        if (typeof this.obj.removeAfter === 'function') {
            this.obj.removeAfter(id, prevId);
        }

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

        // tab 顺序入栈，如栈满则清除
        var stack = this.obj._stack;
        if (stack.length === this.obj.STACKSIZE) {
            stack.splice(0, 1);
        }
        if (stack[stack.length - 1] !== id) {
            this.obj._stack.push(id);
        }

        $tabs.children("div").removeClass("current");
        $tabsPanel.children("div").hide();

        $tabs.children("div[data-index='" + id + "']").addClass("current");
        $tabsPanel.children("div[data-index='" + id + "']").show();
    }
});