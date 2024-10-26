$('#navbar-content').on('show.bs.collapse', () => {
    $('.banner').css('margin-top', '0');
});
$('#navbar-content').on('hide.bs.collapse', () => {
    $('.banner').css('margin-top', '-106px');
});
