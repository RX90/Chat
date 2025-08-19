window.onload = async function () {
  const msg = document.getElementById("msg");
  const log = document.getElementById("log");
  const scrollButton = document.getElementById("scrollToBottom");
  const signinModal = document.getElementById("signin-modal");
  const signinForm = document.getElementById("signin-form");
  const signinError = document.getElementById("signin-error");
  const toggleQuestion = document.getElementById("toggle-question");
  const toggleLink = document.getElementById("toggle-link");
  const usernameGroup = document.getElementById("username-group");
  const formTitle = document.getElementById("form-title");
  const formSubtitle = document.getElementById("form-subtitle");
  const submitButton = document.getElementById("submit-button");
  const logoutButton = document.getElementById("logout-button");
  const groupIcon = document.getElementById("group-icon");
  const onlineUsersPanel = document.getElementById("online-users-panel");
  const onlineUsersList = document.getElementById("online-users-list");
  const closePanelButton = document.getElementById("close-panel-button");

  let conn;
  let isSignUpMode = false;
  let isPanelVisible = false;
  let lastRenderedDay = null;
  let editingMessageId = null;

  function updateFormMode() {
    usernameGroup.style.display = isSignUpMode ? "block" : "none";
    formTitle.textContent = isSignUpMode ? "Создание аккаунта" : "Добро пожаловать";
    formSubtitle.textContent = isSignUpMode
      ? "Заполните поля, чтобы зарегистрироваться"
      : "Войдите в свой аккаунт, чтобы продолжить";
    submitButton.textContent = isSignUpMode ? "Создать аккаунт" : "Войти";
    toggleQuestion.textContent = isSignUpMode
      ? "Уже есть аккаунт?"
      : "Ещё нет аккаунта?";
    toggleLink.textContent = isSignUpMode ? "Войти" : "Создать";
    signinError.textContent = "";
    signinForm.reset();
  }

  updateFormMode();

  toggleLink.addEventListener("click", () => {
    isSignUpMode = !isSignUpMode;
    updateFormMode();
  });

  function validateEmail(email) {
    return /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(email);
  }

  function validateUsername(username) {
    return username.length >= 4 && username.length <= 32;
  }

  function validatePassword(password) {
    return password.length >= 8 && password.length <= 32;
  }

  function validateMessage(message) {
    return message.length > 0 && message.length <= 255;
  }

  async function checkToken() {
    const token = localStorage.getItem("accessToken");
    if (!token) {
      signinModal.style.display = "flex";
      msg.disabled = true;
      logoutButton.style.display = "none";
      groupIcon.style.display = "none";
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
        signinModal.style.display = "none";
        msg.disabled = false;
        logoutButton.style.display = "inline-block";
        groupIcon.style.display = "inline-block";
        return true;
      } else {
        const refreshed = await refreshAccessToken();
        if (refreshed) return await checkToken();
        signinModal.style.display = "flex";
        msg.disabled = true;
        logoutButton.style.display = "none";
        groupIcon.style.display = "none";
        return false;
      }
    } catch {
      signinModal.style.display = "flex";
      msg.disabled = true;
      logoutButton.style.display = "none";
      groupIcon.style.display = "none";
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
  }

  await init();

  signinForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    signinError.textContent = "";

    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;
    const username = document.getElementById("username").value.trim();

    if (!validateEmail(email)) {
      signinError.textContent = "Введите корректный email";
      return;
    }

    if (!validatePassword(password)) {
      signinError.textContent = "Пароль должен быть от 8 до 32 символов";
      return;
    }

    if (isSignUpMode && !validateUsername(username)) {
      signinError.textContent = "Имя пользователя должно быть от 4 до 32 символов";
      return;
    }

    if (isSignUpMode) {
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
            signinModal.style.display = "none";
            msg.disabled = false;
            groupIcon.style.display = "inline-block";
            signinForm.reset();
            startWebSocket();
            logoutButton.style.display = "block";
          } else {
            signinError.textContent = "Ошибка входа после регистрации";
          }
        } else {
          signinError.textContent = "Ошибка регистрации. Попробуйте позже";
        }
      } catch {
        signinError.textContent = "Ошибка сети";
      }
    } else {
      try {
        const resp = await fetch("/api/auth/sign-in", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        });

        if (resp.ok) {
          const data = await resp.json();
          localStorage.setItem("accessToken", data.token);
          signinModal.style.display = "none";
          msg.disabled = false;
          groupIcon.style.display = "inline-block";
          signinForm.reset();
          startWebSocket();
          logoutButton.style.display = "block";
        } else if (resp.status === 401) {
          signinError.textContent = "Неверный email или пароль";
        } else {
          signinError.textContent = "Ошибка входа, попробуйте позже";
        }
      } catch {
        signinError.textContent = "Ошибка сети";
      }
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
    isSignUpMode = false;
    updateFormMode();
    signinModal.style.display = "flex";
    msg.disabled = true;
    log.innerHTML = "";
    onlineUsersPanel.classList.remove("visible");
    isPanelVisible = false;
  });

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
        dataFulltime += ". 🖊️ " + fullUpdated;
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
        console.log("Отправка запроса на удаление сообщения", parsed.id);
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

  document.getElementById("form").onsubmit = function (e) {
    e.preventDefault();
    if (!conn) return false;

    const text = msg.value.trim();
    if (!validateMessage(text)) {
      alert("Сообщение должно содержать от 1 до 255 символов");
      return false;
    }

    if (editingMessageId) {
      conn.send(JSON.stringify({ type: "update", messageId: editingMessageId, content: text }));
      editingMessageId = null;
    } else {
      conn.send(JSON.stringify({ type: "message", content: text }));
    }
    msg.value = "";
    return false;
  };

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
    console.log("Updating online users with data:", users);
    onlineUsersList.innerHTML = "";
    if (!Array.isArray(users) || users.length === 0) {
      console.warn("No valid users array provided:", users);
      const li = document.createElement("li");
      li.textContent = "Нет пользователей онлайн";
      onlineUsersList.appendChild(li);
      return;
    }
    console.log("Populating online users list:", users);
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
    console.log("Panel toggled, visible:", isPanelVisible);
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
          console.log("Parsed message:", parsed);

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
            const msgEl = document.getElementById(`msg-${parsed.messageId}`);
            if (msgEl) {
              const currentScrollTop = log.scrollTop;
              const msgHeight = msgEl.offsetHeight;
              const msgRect = msgEl.getBoundingClientRect();
              const logRect = log.getBoundingClientRect();
              const isAboveViewport = msgRect.bottom < logRect.top;

              let prevSibling = msgEl.previousElementSibling;

              msgEl.remove();
              console.log("Сообщение удалено:", parsed.messageId);

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
                  console.log("Удалён пустой разделитель дня");

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
                dataFulltime += ". 🖊️ " + fullUpdated;
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
};