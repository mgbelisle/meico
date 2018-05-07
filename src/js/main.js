$('#navbar-content').on('show.bs.collapse', function () {
    $('.banner').css('margin-top', '0');
});
$('#navbar-content').on('hide.bs.collapse', function () {
    $('.banner').css('margin-top', '-106px');
});
