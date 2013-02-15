//= require assets/jquery
//= require assets/ujs
//= require assets/ws

$(function() {
  $(window).on('ajax:error', function(e) {
    e = $(e.target);
    if (e.is('form')) {
      $.get('/teams/auth', function() {
        e.submit();
      })
    }
  });
})
