$(function() {
  $(window).on('ajax:error', function(e) {
    e = $(e.target);
    if (e.is('form')) {
      $.get('/map', function() {
        e.submit();
      })
    }
  });
})
