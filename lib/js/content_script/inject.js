var background = (function () {
  let tmp = {};
  /*  */
  chrome.runtime.onMessage.addListener(function (request) {
    for (let id in tmp) {
      if (tmp[id] && (typeof tmp[id] === "function")) {
        if (request.path === "background-to-page") {
          if (request.method === id) {
            tmp[id](request.data);
          }
        }
      }
    }
  });
  /*  */
  return {
    "receive": function (id, callback) {
      tmp[id] = callback;
    },
    "send": function (id, data) {
      chrome.runtime.sendMessage({
        "method": id, 
        "data": data,
        "path": "page-to-background"
      }, function () {
        return chrome.runtime.lastError;
      });
    }
  }
})();

const ikey = "webgl-defender-sandboxed-frame";

if (document.documentElement.getAttribute(ikey) === null) {
  parent.postMessage(ikey, '*');
  window.top.postMessage(ikey, '*');
} else {
  document.documentElement.removeAttribute(ikey);
}

window.addEventListener("message", function (e) {
  if (e.data && e.data === "webgl-defender-alert") {
    e.preventDefault();
    e.stopPropagation();
    /*  */
    background.send("fingerprint", {
      "host": document.location.host
    });
  }
}, false);