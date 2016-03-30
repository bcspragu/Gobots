$(function() {
  $.fn.registerAjaxForm = function(callback) {
    if (!this.hasClass('ajax-form')) {
      return;
    }

    this.submit(function(e) {
      var form = this;
      var cb;
      e.preventDefault();
      if (!callback) {
        cb = function(){};
      } else {
        cb = function(resp) {
          callback.call(form, resp);
        }
      }
      var f = $(this);
      var send = formFunction(f);
      send(f.attr('action'), f.serialize(), cb);
      return false;
    });
  };

  $('.login-form').registerAjaxForm(function(resp) {
    if (resp == 'Yes.') {
      window.location.reload();
    } else {
      var notice = $('<div class="alert alert-danger"></div>');
      notice.text("Invalid token");
      $(this).append(notice);
    }
  });

  $('.create-user-form').registerAjaxForm(function(resp) {
    var notice = $('<div class="alert"></div>');
    notice.text(resp);
    $(this).append(notice);
  });

  function formFunction(form) {
    if (form.attr('method') == 'POST') {
      return $.post;
    }
    return $.get;
  }
});
