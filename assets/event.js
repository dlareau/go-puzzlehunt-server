var listen = function(url, cb) {
  var es = new EventSource(url);
  es.onmessage = cb;
};
