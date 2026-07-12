const API = {
  register: "/api/register",
  login: "/api/login",
  refresh: "/api/refresh",
  logout: "/api/logout",
  todos: "/api/todos",
};

const STORAGE_KEY = "go_todo_auth";

const authPanel = document.querySelector("#authPanel");
const appPanel = document.querySelector("#appPanel");
const statusMessage = document.querySelector("#statusMessage");

const loginForm = document.querySelector("#loginForm");
const loginUsername = document.querySelector("#loginUsername");
const loginPassword = document.querySelector("#loginPassword");
const loginBtn = document.querySelector("#loginBtn");

const registerForm = document.querySelector("#registerForm");
const registerUsername = document.querySelector("#registerUsername");
const registerPassword = document.querySelector("#registerPassword");
const registerBtn = document.querySelector("#registerBtn");

const currentUsername = document.querySelector("#currentUsername");
const logoutBtn = document.querySelector("#logoutBtn");

const createForm = document.querySelector("#createForm");
const titleInput = document.querySelector("#titleInput");
const descInput = document.querySelector("#descInput");
const createBtn = document.querySelector("#createBtn");

const completedFilter = document.querySelector("#completedFilter");
const pageSizeSelect = document.querySelector("#pageSizeSelect");
const prevPageBtn = document.querySelector("#prevPageBtn");
const nextPageBtn = document.querySelector("#nextPageBtn");
const pageInfo = document.querySelector("#pageInfo");
const todoListEl = document.querySelector("#todoList");

let currentPage = 1;
let pageSize = 10;
let currentCompleted = "";
let refreshPromise = null;
let statusTimer = null;

document.addEventListener("DOMContentLoaded", initialize);

function initialize() {
  bindEvents();

  const auth = getAuth();

  if (isValidStoredAuth(auth)) {
    showApp(auth.user);
    loadTodos();
    return;
  }

  clearAuth();
  showAuth();
}

function bindEvents() {
  loginForm.addEventListener("submit", login);
  registerForm.addEventListener("submit", register);
  logoutBtn.addEventListener("click", logout);

  createForm.addEventListener("submit", createTodo);

  completedFilter.addEventListener("change", () => {
    currentCompleted = completedFilter.value;
    currentPage = 1;
    loadTodos();
  });

  pageSizeSelect.addEventListener("change", () => {
    pageSize = Number(pageSizeSelect.value);
    currentPage = 1;
    loadTodos();
  });

  prevPageBtn.addEventListener("click", () => {
    if (currentPage <= 1) {
      return;
    }

    currentPage--;
    loadTodos();
  });

  nextPageBtn.addEventListener("click", () => {
    currentPage++;
    loadTodos();
  });
}

function getAuth() {
  const raw = localStorage.getItem(STORAGE_KEY);

  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw);
  } catch {
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

function saveAuth(auth) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(auth));
}

function clearAuth() {
  localStorage.removeItem(STORAGE_KEY);
}

function isValidStoredAuth(auth) {
  return Boolean(
    auth &&
    auth.user &&
    auth.access_token &&
    auth.refresh_token &&
    auth.session_id
  );
}

function showAuth() {
  authPanel.classList.remove("hidden");
  appPanel.classList.add("hidden");

  currentUsername.textContent = "";
  todoListEl.innerHTML = "";

  currentPage = 1;
  currentCompleted = "";
  completedFilter.value = "";
}

function showApp(user) {
  authPanel.classList.add("hidden");
  appPanel.classList.remove("hidden");

  currentUsername.textContent = user?.username || "";
}

function showStatus(message, type = "info") {
  clearTimeout(statusTimer);

  statusMessage.textContent = message;
  statusMessage.className = `status-message ${type}`;

  statusTimer = setTimeout(() => {
    statusMessage.classList.add("hidden");
  }, 4000);
}

function setButtonLoading(button, loading, loadingText) {
  if (loading) {
    button.dataset.originalText = button.textContent;
    button.textContent = loadingText;
    button.disabled = true;
    return;
  }

  button.textContent = button.dataset.originalText || button.textContent;
  button.disabled = false;
}

async function request(
  url,
  {
    method = "GET",
    body,
    auth = true,
    retry = true,
  } = {}
) {
  const storedAuth = getAuth();

  const headers = {
    Accept: "application/json",
  };

  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  if (auth && storedAuth?.access_token) {
    headers.Authorization = `Bearer ${storedAuth.access_token}`;
  }

  let response;

  try {
    response = await fetch(url, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new Error("无法连接服务器");
  }

  let result = null;

  try {
    result = await response.json();
  } catch {
    result = null;
  }

  if (response.status === 401 && auth && retry) {
    const refreshed = await refreshTokens();

    if (refreshed) {
      return request(url, {
        method,
        body,
        auth,
        retry: false,
      });
    }

    clearAuth();
    showAuth();

    throw new Error("登录状态已过期，请重新登录");
  }

  if (!response.ok || !result || result.code !== 0) {
    throw new Error(
      result?.message || `请求失败，HTTP 状态码：${response.status}`
    );
  }

  return result.data;
}

async function refreshTokens() {
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = performRefresh();

  try {
    return await refreshPromise;
  } finally {
    refreshPromise = null;
  }
}

async function performRefresh() {
  const auth = getAuth();

  if (!auth?.session_id || !auth?.refresh_token) {
    return false;
  }

  let response;

  try {
    response = await fetch(API.refresh, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        session_id: auth.session_id,
        refresh_token: auth.refresh_token,
      }),
    });
  } catch {
    return false;
  }

  let result = null;

  try {
    result = await response.json();
  } catch {
    return false;
  }

  if (!response.ok || !result || result.code !== 0 || !result.data) {
    return false;
  }

  saveAuth({
    user: auth.user,
    access_token: result.data.access_token,
    refresh_token: result.data.refresh_token,
    session_id: result.data.session_id,
  });

  return true;
}

async function login(event) {
  event.preventDefault();

  const username = loginUsername.value.trim();
  const password = loginPassword.value;

  if (!username || !password) {
    showStatus("请输入用户名和密码", "error");
    return;
  }

  setButtonLoading(loginBtn, true, "登录中...");

  try {
    const data = await request(API.login, {
      method: "POST",
      auth: false,
      body: {
        username,
        password,
      },
    });

    saveAuth({
      user: data.user,
      access_token: data.access_token,
      refresh_token: data.refresh_token,
      session_id: data.session_id,
    });

    loginPassword.value = "";

    currentPage = 1;
    showApp(data.user);
    showStatus("登录成功", "success");

    await loadTodos();
  } catch (error) {
    showStatus(error.message, "error");
  } finally {
    setButtonLoading(loginBtn, false);
  }
}

async function register(event) {
  event.preventDefault();

  const username = registerUsername.value.trim();
  const password = registerPassword.value;

  if (!username || !password) {
    showStatus("请输入用户名和密码", "error");
    return;
  }

  setButtonLoading(registerBtn, true, "注册中...");

  try {
    await request(API.register, {
      method: "POST",
      auth: false,
      body: {
        username,
        password,
      },
    });

    loginUsername.value = username;
    loginPassword.value = "";

    registerUsername.value = "";
    registerPassword.value = "";

    showStatus("注册成功，请登录", "success");
    loginPassword.focus();
  } catch (error) {
    showStatus(error.message, "error");
  } finally {
    setButtonLoading(registerBtn, false);
  }
}

async function logout() {
  setButtonLoading(logoutBtn, true, "退出中...");

  try {
    await request(API.logout, {
      method: "POST",
    });
  } catch (error) {
    console.warn("退出接口调用失败：", error);
  } finally {
    clearAuth();
    showAuth();
    setButtonLoading(logoutBtn, false);
    showStatus("已退出登录", "success");
  }
}

function buildListUrl() {
  const params = new URLSearchParams();

  params.set("page", String(currentPage));
  params.set("page_size", String(pageSize));

  if (currentCompleted !== "") {
    params.set("completed", currentCompleted);
  }

  return `${API.todos}?${params.toString()}`;
}

async function loadTodos() {
  if (!isValidStoredAuth(getAuth())) {
    return;
  }

  try {
    const pageData = await request(buildListUrl());

    if (
      pageData.total_pages > 0 &&
      currentPage > pageData.total_pages
    ) {
      currentPage = pageData.total_pages;
      await loadTodos();
      return;
    }

    if (pageData.total_pages === 0) {
      currentPage = 1;
    }

    renderTodos(pageData.items || []);
    updatePager(pageData);
  } catch (error) {
    showStatus(error.message, "error");
  }
}

function updatePager(pageData) {
  const totalPages = pageData.total_pages;
  const displayTotalPages = Math.max(totalPages, 1);

  pageInfo.textContent =
    `第 ${currentPage} / ${displayTotalPages} 页，共 ${pageData.total} 条`;

  prevPageBtn.disabled = currentPage <= 1;
  nextPageBtn.disabled =
    totalPages === 0 || currentPage >= totalPages;
}

async function createTodo(event) {
  event.preventDefault();

  const title = titleInput.value.trim();
  const description = descInput.value.trim();

  if (!title) {
    showStatus("请输入任务名称", "error");
    return;
  }

  setButtonLoading(createBtn, true, "创建中...");

  try {
    await request(API.todos, {
      method: "POST",
      body: {
        title,
        description,
      },
    });

    titleInput.value = "";
    descInput.value = "";

    currentPage = 1;

    showStatus("任务创建成功", "success");
    await loadTodos();
  } catch (error) {
    showStatus(error.message, "error");
  } finally {
    setButtonLoading(createBtn, false);
  }
}

async function completeTodo(id) {
  try {
    await request(`${API.todos}/${id}/done`, {
      method: "PATCH",
    });

    showStatus("任务已完成", "success");
    await loadTodos();
  } catch (error) {
    showStatus(error.message, "error");
  }
}

async function updateTodo(todo) {
  const title = prompt("新的任务名称", todo.title);

  if (title === null) {
    return;
  }

  const description = prompt(
    "新的任务描述",
    todo.description || ""
  );

  if (description === null) {
    return;
  }

  const trimmedTitle = title.trim();
  const trimmedDescription = description.trim();

  if (!trimmedTitle) {
    showStatus("任务名称不能为空", "error");
    return;
  }

  try {
    await request(`${API.todos}/${todo.id}`, {
      method: "PUT",
      body: {
        title: trimmedTitle,
        description: trimmedDescription,
      },
    });

    showStatus("任务更新成功", "success");
    await loadTodos();
  } catch (error) {
    showStatus(error.message, "error");
  }
}

async function deleteTodo(id) {
  if (!confirm("确定删除这个任务吗？")) {
    return;
  }

  try {
    await request(`${API.todos}/${id}`, {
      method: "DELETE",
    });

    showStatus("任务删除成功", "success");
    await loadTodos();
  } catch (error) {
    showStatus(error.message, "error");
  }
}

function renderTodos(todos) {
  todoListEl.innerHTML = "";

  if (todos.length === 0) {
    const empty = document.createElement("div");
    empty.className = "empty";
    empty.textContent = "暂无任务";
    todoListEl.appendChild(empty);
    return;
  }

  for (const todo of todos) {
    const item = document.createElement("article");
    item.className = `todo-item ${todo.completed ? "done" : ""}`;

    const content = document.createElement("div");
    content.className = "todo-content";

    const title = document.createElement("div");
    title.className = "todo-title";
    title.textContent = todo.title;

    const description = document.createElement("div");
    description.className = "todo-desc";
    description.textContent = todo.description || "-";

    content.appendChild(title);
    content.appendChild(description);

    const actions = document.createElement("div");
    actions.className = "todo-actions";

    const doneBtn = document.createElement("button");
    doneBtn.type = "button";
    doneBtn.className = "done-btn";
    doneBtn.textContent = todo.completed ? "已完成" : "完成";

    if (todo.completed) {
      doneBtn.disabled = true;
    } else {
      doneBtn.addEventListener("click", () => {
        completeTodo(todo.id);
      });
    }

    const updateBtn = document.createElement("button");
    updateBtn.type = "button";
    updateBtn.className = "update-btn";
    updateBtn.textContent = "更新";
    updateBtn.addEventListener("click", () => {
      updateTodo(todo);
    });

    const deleteBtn = document.createElement("button");
    deleteBtn.type = "button";
    deleteBtn.className = "delete-btn";
    deleteBtn.textContent = "删除";
    deleteBtn.addEventListener("click", () => {
      deleteTodo(todo.id);
    });

    actions.appendChild(doneBtn);
    actions.appendChild(updateBtn);
    actions.appendChild(deleteBtn);

    item.appendChild(content);
    item.appendChild(actions);

    todoListEl.appendChild(item);
  }
}