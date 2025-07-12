window.onload = function () {
  var conn;
  var msg = document.getElementById("msg");
  var log = document.getElementById("log");
  const scrollButton = document.getElementById("scrollToBottom");

  function checkScroll() {
    const isAtBottom = log.scrollHeight - log.clientHeight <= log.scrollTop + 1;
    scrollButton.classList.toggle("show", !isAtBottom);
  }

  function scrollToBottom() {
    log.scrollTop = log.scrollHeight;
    scrollButton.classList.remove("show");
  }

  scrollButton.addEventListener("click", scrollToBottom);
  log.addEventListener("scroll", checkScroll);

  function createMessageElement(parsed) {
    var messageDiv = document.createElement("div");
    messageDiv.className = "message";

    var contentDiv = document.createElement("div");
    contentDiv.className = "message-content";
    contentDiv.textContent = parsed.content;

    var timeSpan = document.createElement("span");
    timeSpan.className = "timestamp";

    if (parsed.createdAt) {
      var date = new Date(parsed.createdAt);
      timeSpan.textContent = date.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      });

      var day = String(date.getDate()).padStart(2, "0");
      var month = String(date.getMonth() + 1).padStart(2, "0");
      var year = date.getFullYear();
      var hours = String(date.getHours()).padStart(2, "0");
      var minutes = String(date.getMinutes()).padStart(2, "0");
      var seconds = String(date.getSeconds()).padStart(2, "0");

      var fullTime = `${day}.${month}.${year} ${hours}:${minutes}:${seconds}`;
      timeSpan.setAttribute("data-fulltime", fullTime);
    } else {
      timeSpan.textContent = "--:--";
      timeSpan.setAttribute("data-fulltime", "Неизвестное время");
    }

    messageDiv.appendChild(contentDiv);
    messageDiv.appendChild(timeSpan);

    return messageDiv;
  }

  function appendLog(element) {
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(element);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }

  document.getElementById("form").onsubmit = function (e) {
    e.preventDefault();
    if (!conn || !msg.value.trim()) return false;

    const messageObject = {
      content: msg.value.trim(),
    };

    conn.send(JSON.stringify(messageObject));
    msg.value = "";
    return false;
  };

  if (window["WebSocket"]) {
    conn = new WebSocket("ws://" + document.location.host + "/ws");

    conn.onclose = function () {
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
        const rawMessage = messages[i].trim();
        if (!rawMessage) continue;

        try {
          const parsed = JSON.parse(rawMessage);
          const messageElement = createMessageElement(parsed);
          appendLog(messageElement);
        } catch (e) {
          console.error("Message parse error:", e);
          var fallbackDiv = document.createElement("div");
          fallbackDiv.className = "message";
          fallbackDiv.textContent = rawMessage;
          appendLog(fallbackDiv);
        }
      }
    };
  } else {
    var item = document.createElement("div");
    item.textContent = "Your browser does not support WebSockets.";
    appendLog(item);
  }
};
