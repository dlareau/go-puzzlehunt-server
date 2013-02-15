//= require assets/jquery
//= require assets/ujs
//= require assets/ws

$(function() {
  /* Checkboxes in forms are submitted through a hidden element */
  var boxes = $('input[type=checkbox]');
  $.each(boxes, function(i, el) {
    var hidden = $(el).siblings('input[type=hidden]')
    if (hidden.val() == 'true') {
      $(el).attr('checked', 'checked')
    }
  });
  boxes.change(function() {
    var $me = $(this);
    var hidden = $me.siblings('input[type=hidden]')
    if ($me.is(':checked')) {
      hidden.val('true')
    } else {
      hidden.val('false')
    }
  });

  $(window).on('ajax:error', function(e) {
    e = $(e.target);
    if (e.is('form')) {
      $.get('/admin/auth', function() {
        e.submit();
      })
    }
  });
})
