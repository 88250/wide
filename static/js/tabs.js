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
 * @file tabs.js
 *
 * @author <a href="http://vanessa.b3log.org">Liyuan Li</a>
 * @version 1.0.0.1, Dec 8, 2015
 */
var Tabs = function (obj) {
    obj._$tabsPanel = $(obj.id + " > .tabs-panel");
    obj._$tabs = $(obj.id + " > .tabs");
    obj._stack = [];

    this.obj = obj;
    this.obj.STACKSIZE = 64;

    this._init(obj);

    // DOM 元素存在时，应顺序入栈
    var _it = this;
    $(obj.id + " > .tabs > div").each(function () {
        var id = $(this).data("index");
        if (obj._stack.length === _it.obj.STACKSIZE) {
            obj._stack.splice(0, 1);
        }
        if (obj._stack[obj._stack.length - 1] !== id) {
            _it.obj._stack.push(id);
        }
    });

};

$.extend(Tabs.prototype, {
    _init: function (obj) {
        var _that = this;

        obj._$tabs.on("click", "div", function (event) {
            if ($(this).hasClass('current')) {
                return false;
            }

            var id = $(this).data("index");
            _that.setCurrent(id);
            if (typeof (obj.clickAfter) === "function") {
                obj.clickAfter(id);
            }
        });

        obj._$tabs.on("click", ".ico-close", function (event) {
            var id = $(this).parent().data("index"),
                    isRemove = true;

            if (typeof obj.removeBefore === 'function') {
                isRemove = obj.removeBefore(id);
            }

            if (isRemove) {
                _that.del(id);
            }
            event.stopPropagation();
        });
    },
    _hasId: function (id) {
        var $tabs = this.obj._$tabs;
        if ($tabs.find('div[data-index="' + id + '"]').length === 0) {
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
                + data.title + ' <span class="ico-close font-ico"></span></div>');
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
                i--;
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
    setCurrent: function (path) {
        if (!path) {
            return false;
        }

        var $tabsPanel = this.obj._$tabsPanel,
                $tabs = this.obj._$tabs;

        var $currentTab = $tabs.children(".current");
        if ($currentTab.data("index") === path) {
            return false;
        }

        // tab 顺序入栈，如栈满则清除
        var stack = this.obj._stack;
        if (stack.length === this.obj.STACKSIZE) {
            stack.splice(0, 1);
        }
        if (stack[stack.length - 1] !== path) {
            this.obj._stack.push(path);
        }

        $tabs.children("div").removeClass("current");
        $tabsPanel.children("div").hide();

        $tabs.children('div[data-index="' + path + '"]').addClass("current");
        $tabsPanel.children('div[data-index="' + path + '"]').show();

        if (typeof this.obj.setAfter === 'function') {
            this.obj.setAfter();
        }

        var id = this.getCurrentId();
        if ("startPage" === id) {
            return;
        }

        // set tree node selected
        var tId = tree.getTIdByPath(id);
        var node = tree.fileTree.getNodeByTId(tId);
        tree.fileTree.selectNode(node);
        wide.curNode = node;

        for (var i = 0, ii = editors.data.length; i < ii; i++) {
            if (editors.data[i].id === id) {
                wide.curEditor = editors.data[i].editor;
                break;
            }
        }

        if (wide.curEditor) {
            var cursor = wide.curEditor.getCursor();
            wide.curEditor.setCursor(cursor);
            wide.curEditor.focus();
            wide.refreshOutline();

            $(".footer .cursor").text('|   ' + (cursor.line + 1) + ':' + (cursor.ch + 1) + '   |');
        }
    }
});