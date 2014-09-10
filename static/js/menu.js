var menu = {
    init: function() {
        this.subMenu();
    },
    subMenu: function () {
        $(".menu > ul > li > a, .menu > ul> li > span").click(function() {
            var $it = $(this);
            $it.next().show();

            $(".menu > ul > li > a, .menu > ul> li > span").unbind();

            $(".menu > ul > li > a, .menu > ul> li > span").mouseover(function() {
                $(".frame").hide();
                var $it = $(this);
                $it.next().show();
            });
        });
    }
};