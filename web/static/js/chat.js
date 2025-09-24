window.onload = async function () {
  const msg = document.getElementById("msg");
  const log = document.getElementById("log");
  const form = document.getElementById("form");
  const scrollButton = document.getElementById("scrollToBottom");
  const loginModal = document.getElementById("login-modal");
  const loginForm = document.getElementById("login-form");
  const loginError = document.getElementById("login-error");
  const loginPassword = document.getElementById("login-password");
  const loginToggleIcon = loginModal.querySelector(".password-toggle");
  const logoutButton = document.getElementById("logout-button");
  const groupIcon = document.getElementById("group-icon");
  const onlineUsersPanel = document.getElementById("online-users-panel");
  const onlineUsersList = document.getElementById("online-users-list");
  const closePanelButton = document.getElementById("close-panel-button");
  const registerModal = document.getElementById("register-modal");
  const registerForm = document.getElementById("register-form");
  const registerError = document.getElementById("register-error");
  const registerPassword = document.getElementById("register-password");
  const registerToggleIcon = registerModal.querySelector(".password-toggle");
  const registerButton = document.getElementById("register-button");
  const loginButton = document.getElementById("login-button");
  const toggleToRegister = document.getElementById("toggle-to-register");
  const toggleToLogin = document.getElementById("toggle-to-login");
  const isMobile =
    /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
      navigator.userAgent
    ) ||
    "ontouchstart" in window ||
    navigator.maxTouchPoints > 0 ||
    navigator.msMaxTouchPoints > 0;

  let conn;
  let isPanelVisible = false;
  let lastRenderedDay = null;
  let editingMessageId = null;

  loginToggleIcon.addEventListener("click", () => {
    const isHidden = loginPassword.type === "password";
    loginPassword.type = isHidden ? "text" : "password";
    loginToggleIcon.src = isHidden ? "/img/eye-visible.svg" : "/img/eye-hidden.svg";
  });

  registerToggleIcon.addEventListener("click", () => {
    const isHidden = registerPassword.type === "password";
    registerPassword.type = isHidden ? "text" : "password";
    registerToggleIcon.src = isHidden ? "/img/eye-visible.svg" : "/img/eye-hidden.svg";
  });

  toggleToRegister.addEventListener("click", () => {
    loginModal.style.display = "none";
    registerModal.style.display = "flex";
  });

  toggleToLogin.addEventListener("click", () => {
    registerModal.style.display = "none";
    loginModal.style.display = "flex";
  });

  registerButton.addEventListener("click", () => {
    registerModal.style.display = "flex";
  });

  loginButton.addEventListener("click", () => {
    loginModal.style.display = "flex";
  });

  function validateEmail(email) {
    const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    return regex.test(email);
  }

  function validateUsername(username) {
    return username.length >= 4 && username.length <= 32;
  }

  function validatePassword(password) {
    return password.length >= 8 && password.length <= 32;
  }

  function validateMessage(message) {
    return message.length > 0 && message.length <= 1024;
  }

  function updateScrollButtonPosition() {
    const formHeight = form.offsetHeight;
    scrollButton.style.bottom = (formHeight + 10) + 'px';
  }

  function adjustTextareaHeight() {
    this.style.height = 'auto';
    const scrollHeight = this.scrollHeight;
    const maxHeight = 180;
    const singleLineHeight = 50;

    if (scrollHeight > maxHeight) {
      this.style.height = maxHeight + 'px';
      this.style.overflowY = 'auto';
    } else {
      this.style.height = scrollHeight + 'px';
      this.style.overflowY = 'hidden';
    }

    const isExpanded = scrollHeight > singleLineHeight;
    if (isExpanded) {
      form.classList.add('expanded');
    } else {
      form.classList.remove('expanded');
    }

    updateScrollButtonPosition();
  }

  async function checkToken() {
    const token = localStorage.getItem("accessToken");
    if (!token) {
      loginModal.style.display = "none";
      registerModal.style.display = "none";
      msg.disabled = true;
      logoutButton.style.display = "none";
      groupIcon.style.display = "none";
      registerButton.style.display = "inline-block";
      loginButton.style.display = "inline-block";
      return false;
    }

    try {
      const response = await fetch("/api/auth/verify", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
      });

      if (response.ok) {
        loginModal.style.display = "none";
        registerModal.style.display = "none";
        msg.disabled = false;
        logoutButton.style.display = "inline-block";
        groupIcon.style.display = "inline-block";
        registerButton.style.display = "none";
        loginButton.style.display = "none";
        return true;
      } else {
        const refreshed = await refreshAccessToken();
        if (refreshed) return await checkToken();
        loginModal.style.display = "none";
        registerModal.style.display = "none";
        msg.disabled = true;
        logoutButton.style.display = "none";
        groupIcon.style.display = "none";
        registerButton.style.display = "inline-block";
        loginButton.style.display = "inline-block";
        return false;
      }
    } catch {
      loginModal.style.display = "none";
      registerModal.style.display = "none";
      msg.disabled = true;
      logoutButton.style.display = "none";
      groupIcon.style.display = "none";
      registerButton.style.display = "inline-block";
      loginButton.style.display = "inline-block";
      return false;
    }
  }

  async function refreshAccessToken() {
    const token = localStorage.getItem("accessToken");
    if (!token) return false;

    try {
      const response = await fetch("/api/auth/refresh", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        credentials: "include",
      });

      if (!response.ok) throw new Error("Refresh failed");

      const data = await response.json();
      localStorage.setItem("accessToken", data.token);
      return true;
    } catch {
      localStorage.removeItem("accessToken");
      return false;
    }
  }

  async function init() {
    const ok = await checkToken();
    if (ok) startWebSocket();
    adjustTextareaHeight.call(msg);
    updateScrollButtonPosition();
  }

  await init();

  loginForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    loginError.textContent = "";

    const email = document.getElementById("login-email").value.trim();
    const password = loginPassword.value;

    if (!validateEmail(email)) {
      loginError.textContent = "Введите корректный email";
      return;
    }

    if (!validatePassword(password)) {
      loginError.textContent = "Пароль должен быть от 8 до 32 символов";
      return;
    }

    try {
      const resp = await fetch("/api/auth/sign-in", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (resp.ok) {
        const data = await resp.json();
        localStorage.setItem("accessToken", data.token);
        loginModal.style.display = "none";
        msg.disabled = false;
        registerButton.style.display = "none";
        loginButton.style.display = "none";
        groupIcon.style.display = "inline-block";
        logoutButton.style.display = "inline-block";
        loginForm.reset();
        loginPassword.type = "password";
        loginToggleIcon.src = "/img/eye-hidden.svg";
        startWebSocket();
      } else if (resp.status === 401) {
        loginError.textContent = "Неверный email или пароль";
      } else {
        loginError.textContent = "Ошибка входа, попробуйте позже";
      }
    } catch {
      loginError.textContent = "Ошибка сети";
    }
  });

  registerForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    registerError.textContent = "";

    const email = document.getElementById("register-email").value.trim();
    const username = document.getElementById("register-username").value.trim();
    const password = registerPassword.value;

    if (!validateEmail(email)) {
      registerError.textContent = "Введите корректный email";
      return;
    }

    if (!validateUsername(username)) {
      registerError.textContent = "Имя пользователя должно быть от 4 до 32 символов";
      return;
    }

    if (!validatePassword(password)) {
      registerError.textContent = "Пароль должен быть от 8 до 32 символов";
      return;
    }

    try {
      const resp = await fetch("/api/auth/sign-up", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, username }),
      });

      if (resp.ok) {
        const loginResp = await fetch("/api/auth/sign-in", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        });

        if (loginResp.ok) {
          const data = await loginResp.json();
          localStorage.setItem("accessToken", data.token);
          registerModal.style.display = "none";
          msg.disabled = false;
          registerButton.style.display = "none";
          loginButton.style.display = "none";
          groupIcon.style.display = "inline-block";
          logoutButton.style.display = "inline-block";
          registerForm.reset();
          registerPassword.type = "password";
          registerToggleIcon.src = "/img/eye-hidden.svg";
          startWebSocket();
        } else {
          registerError.textContent = "Ошибка входа после регистрации";
        }
      } else {
        registerError.textContent = "Ошибка регистрации. Попробуйте позже";
      }
    } catch {
      registerError.textContent = "Ошибка сети";
    }
  });

  logoutButton.addEventListener("click", async () => {
    const token = localStorage.getItem("accessToken");
    if (!token) return;

    try {
      const response = await fetch("/api/auth/sign-out", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
      });

      if (!response.ok) {
        const errorResponse = await response.json();
        if (response.status === 401 && errorResponse.err === "token has expired") {
          const refreshed = await refreshAccessToken();
          if (refreshed) return logoutButton.click();
        }
        throw new Error(errorResponse.message || "Ошибка при выходе");
      }
    } catch {}

    localStorage.removeItem("accessToken");
    if (conn) conn.close();
    logoutButton.style.display = "none";
    groupIcon.style.display = "none";
    registerButton.style.display = "inline-block";
    loginButton.style.display = "inline-block";
    loginModal.style.display = "none";
    registerModal.style.display = "none";
    msg.disabled = true;
    log.innerHTML = "";
    onlineUsersPanel.classList.remove("visible");
    isPanelVisible = false;
  });

  function checkScroll() {
    const isAtBottom = Math.abs(log.scrollHeight - log.clientHeight - log.scrollTop) < 2;
    scrollButton.classList.toggle("show", !isAtBottom);
  }

  function scrollToBottom() {
    log.scrollTop = log.scrollHeight;
    scrollButton.classList.remove("show");
  }

  scrollButton.addEventListener("click", scrollToBottom);
  log.addEventListener("scroll", checkScroll);

  function startOfDay(d) {
    const x = new Date(d);
    x.setHours(0,0,0,0);
    return x;
  }

  function createDaySeparator(createdAt) {
    const date = new Date(createdAt);
    const options = { day: "numeric", month: "long" };
    const formatted = date.toLocaleDateString("en-US", options);

    const div = document.createElement("div");
    div.className = "day-separator";
    div.textContent = formatted;
    div.setAttribute("data-day", startOfDay(date).toISOString());
    return div;
  }

  function resetLastRenderedDay() {
    lastRenderedDay = null;
  }

  function getCurrentUserId() {
    const token = localStorage.getItem("accessToken");
    if (!token) return null;
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.sub;
  }

  function createMessageElement(parsed) {
    const messageDiv = document.createElement("div");
    messageDiv.className = "message";

    if (parsed.id) messageDiv.id = `msg-${parsed.id}`;

    if (parsed.userId === getCurrentUserId()) {
      messageDiv.classList.add("own");
    }

    const headerDiv = document.createElement("div");
    headerDiv.className = "message-header";

    const usernameSpan = document.createElement("span");
    usernameSpan.className = "username";
    usernameSpan.textContent = parsed.username || "Аноним";

    const timeSpan = document.createElement("span");
    timeSpan.className = "timestamp";

    const ts = parsed.createdAt;
    const updatedTs = parsed.updatedAt;
    if (ts) {
      const date = new Date(ts);
      const shortTime = date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
      timeSpan.textContent = shortTime;

      const isUpdated = updatedTs && new Date(updatedTs) > date;
      if (isUpdated) {
        const editedSpan = document.createElement("span");
        editedSpan.className = "edited";
        editedSpan.textContent = " (изменено)";
        timeSpan.appendChild(editedSpan);
      }

      const fullCreated = date.toLocaleString("ru-RU").replace(",", "");
      let dataFulltime = fullCreated;
      if (isUpdated) {
        const updateDate = new Date(updatedTs);
        const fullUpdated = updateDate.toLocaleString("ru-RU").replace(",", "");
        dataFulltime += ". Изменено: " + fullUpdated;
      }
      timeSpan.setAttribute("data-fulltime", dataFulltime);
    } else {
      timeSpan.textContent = "--:--";
      timeSpan.setAttribute("data-fulltime", "Неизвестное время");
    }

    headerDiv.appendChild(usernameSpan);
    headerDiv.appendChild(timeSpan);

    if (parsed.userId === getCurrentUserId()) {
      const optionsBtn = document.createElement("button");
      optionsBtn.className = "options-btn";
      const img = document.createElement("img");
      img.src = "/img/options.svg";
      img.alt = "Опции сообщения";
      optionsBtn.appendChild(img);

      const optionsMenu = document.createElement("div");
      optionsMenu.className = "options-menu";
      optionsMenu.style.display = "none";

      const editBtn = document.createElement("button");
      const editImg = document.createElement("img");
      editImg.src = "/img/update.svg";
      editImg.alt = "Редактировать";
      editBtn.appendChild(editImg);
      editBtn.appendChild(document.createTextNode("Редактировать"));
      editBtn.onclick = () => {
        optionsMenu.style.display = "none";
        const contentDiv = messageDiv.querySelector('.message-content');
        msg.value = contentDiv.textContent;
        editingMessageId = parsed.id;
        msg.focus();
      };

      const deleteBtn = document.createElement("button");
      const deleteImg = document.createElement("img");
      deleteImg.src = "/img/delete.svg";
      deleteImg.alt = "Удалить";
      deleteBtn.appendChild(deleteImg);
      deleteBtn.appendChild(document.createTextNode("Удалить"));
      deleteBtn.onclick = () => {
        if (!conn) return;
        conn.send(JSON.stringify({ type: "delete", messageId: parsed.id }));
        optionsMenu.style.display = "none";
      };

      optionsMenu.appendChild(editBtn);
      optionsMenu.appendChild(deleteBtn);

      optionsBtn.onclick = (e) => {
        e.stopPropagation();
        optionsMenu.style.display = optionsMenu.style.display === "none" ? "flex" : "none";
      };

      document.addEventListener("click", () => {
        optionsMenu.style.display = "none";
      });

      headerDiv.appendChild(optionsBtn);
      headerDiv.appendChild(optionsMenu);
    }

    const contentDiv = document.createElement("div");
    contentDiv.className = "message-content";
    contentDiv.style.whiteSpace = "pre-line";
    contentDiv.textContent = parsed.content;

    messageDiv.appendChild(headerDiv);
    messageDiv.appendChild(contentDiv);

    return messageDiv;
  }

  function appendLog(element) {
    const doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(element);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }

  function appendMessageWithSeparator(parsed) {
    const createdAt = parsed.createdAt || new Date().toISOString();
    const msgDayIso = startOfDay(new Date(createdAt)).toISOString();

    if (lastRenderedDay !== msgDayIso) {
      const sep = createDaySeparator(createdAt);
      appendLog(sep);
      lastRenderedDay = msgDayIso;
    }

    const el = createMessageElement(parsed);
    appendLog(el);
  }

  function scrollLogToBottom() {
    log.scrollTop = log.scrollHeight - log.clientHeight;
  }

  function sendMessage() {
    if (!conn) return;

    const text = msg.value.trim();
    if (!validateMessage(text)) {
      alert("Сообщение должно содержать от 1 до 1024 символов");
      return;
    }

    if (editingMessageId) {
      const msgEl = document.getElementById(`msg-${editingMessageId}`);
      const contentDiv = msgEl.querySelector('.message-content');
      if (text === contentDiv.textContent) {
        editingMessageId = null;
        msg.value = "";
        adjustTextareaHeight.call(msg);
        return;
      }
      conn.send(JSON.stringify({ type: "update", messageId: editingMessageId, content: text }));
      editingMessageId = null;
    } else {
      conn.send(JSON.stringify({ type: "message", content: text }));
    }
    msg.value = "";
    adjustTextareaHeight.call(msg);
  }

  document.getElementById("form").onsubmit = function (e) {
    e.preventDefault();
    sendMessage();
    return false;
  };

  msg.addEventListener('keydown', function(e) {
    if (!isMobile && e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  });

  msg.addEventListener('input', function() {
    const wasAtBottom = log.scrollTop + log.clientHeight >= log.scrollHeight - 1;
    adjustTextareaHeight.call(this);
    
    if (wasAtBottom) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
    checkScroll();
  });

  function scheduleTokenRefresh() {
    const token = localStorage.getItem("accessToken");
    if (!token) return;

    const payload = JSON.parse(atob(token.split(".")[1]));
    const expiry = payload.exp * 1000;
    const now = Date.now();
    const refreshTime = expiry - 30 * 1000;
    const timeout = refreshTime - now;

    if (timeout > 0) {
      setTimeout(async () => {
        const refreshed = await refreshAccessToken();
        if (refreshed) {
          const newToken = localStorage.getItem("accessToken");
          conn.send(JSON.stringify({ type: "auth_refresh", token: newToken }));
          scheduleTokenRefresh();
        } else {
          conn.close();
        }
      }, timeout);
    }
  }

  function updateOnlineUsers(users) {
    onlineUsersList.innerHTML = "";
    if (!Array.isArray(users) || users.length === 0) {
      console.warn("No valid users array provided:", users);
      const li = document.createElement("li");
      li.textContent = "Нет пользователей онлайн";
      onlineUsersList.appendChild(li);
      return;
    }
    users.forEach(user => {
      const li = document.createElement("li");
      li.innerHTML = `<span class="status-indicator"></span>${user}`;
      onlineUsersList.appendChild(li);
    });
  }

  function togglePanel() {
    isPanelVisible = !isPanelVisible;
    onlineUsersPanel.classList.toggle("visible", isPanelVisible);
    groupIcon.classList.toggle("active", isPanelVisible);
  }

  groupIcon.addEventListener("click", togglePanel);
  closePanelButton.addEventListener("click", togglePanel);

  document.addEventListener("click", (e) => {
    if (
      isPanelVisible &&
      !onlineUsersPanel.contains(e.target) &&
      !groupIcon.contains(e.target) &&
      !closePanelButton.contains(e.target)
    ) {
      togglePanel();
    }
  });
  
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && isPanelVisible) {
      togglePanel();
    }
  });

  function startWebSocket() {
    const token = localStorage.getItem("accessToken");
    if (!token) return;

    if (conn) {
      console.log("Closing existing WebSocket connection, readyState:", conn.readyState);
      conn.close();
      conn = null;
    }

    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsHost = window.location.hostname;
    const wsPort = "8080";
    const wsUrl = `${wsProtocol}//${wsHost}:${wsPort}/ws?accessToken=${encodeURIComponent(token)}`;
    
    conn = new WebSocket(wsUrl);

    conn.onopen = function () {
      console.log("WebSocket opened");
      conn.send(JSON.stringify({ type: "auth", token }));
      scheduleTokenRefresh();
    };

    conn.onclose = function (event) {
      console.log("WebSocket closed, code:", event.code, "reason:", event.reason, "wasClean:", event.wasClean);
      appendLog(document.createElement("div"));
      onlineUsersPanel.classList.remove("visible");
      isPanelVisible = false;
      updateOnlineUsers([]);
    };

    conn.onerror = function (evt) {
      console.error("WebSocket error:", evt);
    };

    conn.onmessage = function (evt) {
      const messages = evt.data.split("\n");
      let shouldScrollToBottom = false;

      for (let i = 0; i < messages.length; i++) {
        const rawMessage = messages[i].trim();
        if (!rawMessage) continue;

        try {
          const parsed = JSON.parse(rawMessage);

          if (parsed.type === "auth_ok") {
            console.log("Auth successful, requesting history");
            resetLastRenderedDay();
            conn.send(JSON.stringify({ type: "history" }));
            continue;
          }

          if (parsed.type === "online_users") {
            console.log("Received online_users:", parsed.users);
            updateOnlineUsers(parsed.users || []);
            continue;
          }

          if (parsed.type === "delete") {
            console.log("Received a message deletion:", parsed.messageId);
            const msgEl = document.getElementById(`msg-${parsed.messageId}`);
            if (msgEl) {
              const currentScrollTop = log.scrollTop;
              const msgHeight = msgEl.offsetHeight;
              const msgRect = msgEl.getBoundingClientRect();
              const logRect = log.getBoundingClientRect();
              const isAboveViewport = msgRect.bottom < logRect.top;

              let prevSibling = msgEl.previousElementSibling;

              msgEl.remove();

              let prevSeparator = null;
              let current = prevSibling;
              while (current) {
                if (current.classList.contains("day-separator")) {
                  prevSeparator = current;
                  break;
                }
                current = current.previousElementSibling;
              }

              if (prevSeparator) {
                let hasMessagesInGroup = false;
                let nextElem = prevSeparator.nextElementSibling;
                while (nextElem && !nextElem.classList.contains("day-separator")) {
                  if (nextElem.classList.contains("message")) {
                    hasMessagesInGroup = true;
                    break;
                  }
                  nextElem = nextElem.nextElementSibling;
                }

                if (!hasMessagesInGroup) {
                  const sepHeight = prevSeparator.offsetHeight;
                  prevSeparator.remove();
                  const lastSep = log.querySelector('.day-separator:last-of-type');
                  if (lastSep) {
                    lastRenderedDay = lastSep.getAttribute('data-day');
                  } else {
                    lastRenderedDay = null;
                  }

                  if (isAboveViewport) {
                    log.scrollTop = currentScrollTop - msgHeight - sepHeight;
                  }
                }
              }

              if (isAboveViewport) {
                log.scrollTop = currentScrollTop - msgHeight;
              } else {
                log.scrollTop = currentScrollTop;
              }
            }
            continue;
          }

          if (parsed.type === "update") {
            console.log("Received a message update:", parsed.messageId);
            const updated = parsed.Message;
            const msgEl = document.getElementById(`msg-${updated.id}`);
            if (msgEl) {
              const contentDiv = msgEl.querySelector('.message-content');
              contentDiv.textContent = updated.content;

              const timeSpan = msgEl.querySelector('.timestamp');
              const date = new Date(updated.createdAt);
              const shortTime = date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
              timeSpan.textContent = shortTime;

              const isUpdated = updated.updatedAt && new Date(updated.updatedAt) > date;
              if (isUpdated) {
                const editedSpan = document.createElement("span");
                editedSpan.className = "edited";
                editedSpan.textContent = " (изменено)";
                timeSpan.appendChild(editedSpan);
              }

              const fullCreated = date.toLocaleString("ru-RU").replace(",", "");
              let dataFulltime = fullCreated;
              if (isUpdated) {
                const updateDate = new Date(updated.updatedAt);
                const fullUpdated = updateDate.toLocaleString("ru-RU").replace(",", "");
                dataFulltime += ". Изменено: " + fullUpdated;
              }
              timeSpan.setAttribute("data-fulltime", dataFulltime);
            }
            continue;
          }

          appendMessageWithSeparator(parsed);
          shouldScrollToBottom = true;
        } catch (e) {
          console.error("Failed to parse message:", rawMessage, e);
          const fallbackDiv = document.createElement("div");
          fallbackDiv.className = "message";
          fallbackDiv.textContent = rawMessage;
          appendLog(fallbackDiv);
          shouldScrollToBottom = true;
        }
      }

      if (shouldScrollToBottom) {
        scrollLogToBottom();
      }
    };
  }

  window.addEventListener('resize', () => {
    updateScrollButtonPosition();
    const wasAtBottom = log.scrollTop + log.clientHeight >= log.scrollHeight - 1;
    if (wasAtBottom) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
    checkScroll();
  });
};