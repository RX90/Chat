window.onload = function () {
  var conn;
  var msg = document.getElementById("msg");
  var log = document.getElementById("log");

  function appendLog(item) {
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }

  document.getElementById("form").onsubmit = function () {
    if (!conn || !msg.value) {
      return false;
    }

    const messageObject = {
      content: msg.value.trim(),
    };

    console.log("OUTGOING JSON:\n" + JSON.stringify(messageObject, null, 2));

    conn.send(JSON.stringify(messageObject));
    msg.value = "";
    return false;
  };

  if (window["WebSocket"]) {
    const token = localStorage.getItem("accessToken");
    conn = new WebSocket(
      "ws://" +
        document.location.host +
        "/ws?accessToken=" +
        encodeURIComponent(token)
    );
    conn.onclose = function (evt) {
      var item = document.createElement("div");
      item.innerHTML = "<b>Connection closed. Reload page</b>";
      appendLog(item);
    };
    conn.onerror = function (evt) {
      console.error("WebSocket error:", evt);
    };
    conn.onmessage = function (evt) {
      var messages = evt.data.split("\n");

      for (var i = 0; i < messages.length; i++) {
        const rawMessage = messages[i];

        try {
          const parsed = JSON.parse(rawMessage);
          console.log("INCOMING JSON:\n" + JSON.stringify(parsed, null, 2));
        } catch (e) {
          console.log("NON-JSON MESSAGE:", rawMessage);
        }

        var item = document.createElement("div");
        item.innerText = messages[i];
        appendLog(item);
      }
    };
  } else {
    var item = document.createElement("div");
    item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
    appendLog(item);
  }
};
