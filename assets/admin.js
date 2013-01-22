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

  /* On the queue page, open a text box for submitting an email */
  $(document).delegate('.needs-response', 'click', function() {
    $(this).siblings('form').show();
    return false;
  });
  $(document).delegate('.canned-response', 'click', function() {
    $(this).siblings('form').submit();
    return false;
  });
})
