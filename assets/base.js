//= require jquery
//= require ujs
//= require ws

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
