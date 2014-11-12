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

        // 点击子菜单后消失
        $(".frame li").click(function () {
            $(this).closest(".frame").hide();
            $(".menu > ul > li > a, .menu > ul> li > span").removeClass("selected");
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
    }
};