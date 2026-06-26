const todoListEl = document.querySelector("#todoList");
const titleInput = document.querySelector("#titleInput");
const descInput = document.querySelector("#descInput");
const createBtn = document.querySelector("#createBtn");

const completedFilter = document.querySelector("#completedFilter");
const pageSizeSelect = document.querySelector("#pageSizeSelect");
const prevPageBtn = document.querySelector("#prevPageBtn");
const nextPageBtn = document.querySelector("#nextPageBtn");
const pageInfo = document.querySelector("#pageInfo");

const API_BASE = "/api/todos";

let currentPage = 1;
let pageSize = 10;
let currentCompleted = "";

document.addEventListener("DOMContentLoaded", loadTodos);

createBtn.addEventListener("click", createTodo);

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
  if (currentPage <= 1) return;

  currentPage--;
  loadTodos();
});

nextPageBtn.addEventListener("click", () => {
  currentPage++;
  loadTodos();
});

async function request(url, options = {}) {
  const resp = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
    },
    ...options,
  });

  const result = await resp.json();

  if (!resp.ok || result.code !== 0) {
    throw new Error(result.message || "请求失败");
  }

  return result.data;
}

function buildListUrl() {
  const params = new URLSearchParams();

  params.set("page", currentPage);
  params.set("page_size", pageSize);

  if (currentCompleted !== "") {
    params.set("completed", currentCompleted);
  }

  return `${API_BASE}?${params.toString()}`;
}

async function loadTodos() {
  try {
    const todos = await request(buildListUrl());
    renderTodos(todos || []);
    updatePager(todos || []);
  } catch (err) {
    alert(err.message);
  }
}

function updatePager(todos) {
  pageInfo.textContent = `第 ${currentPage} 页`;

  prevPageBtn.disabled = currentPage <= 1;

  // 后端暂时没有返回 total，所以这里只能粗略判断：
  // 如果当前页数量少于 pageSize，就认为后面没有下一页。
  nextPageBtn.disabled = todos.length < pageSize;
}

async function createTodo() {
  const title = titleInput.value.trim();
  const description = descInput.value.trim();

  if (!title) {
    alert("请输入任务名称");
    return;
  }

  try {
    await request(API_BASE, {
      method: "POST",
      body: JSON.stringify({
        title,
        description,
      }),
    });

    titleInput.value = "";
    descInput.value = "";

    // 新任务按 created_at 倒序排，创建后回到第一页更合理
    currentPage = 1;
    await loadTodos();
  } catch (err) {
    alert(err.message);
  }
}

async function completeTodo(id) {
  try {
    await request(`${API_BASE}/${id}/done`, {
      method: "PATCH",
    });

    await loadTodos();
  } catch (err) {
    alert(err.message);
  }
}

async function updateTodo(todo) {
  const title = prompt("新的任务名称", todo.title);
  if (title === null) return;

  const description = prompt("新的任务描述", todo.description || "");
  if (description === null) return;

  const trimmedTitle = title.trim();
  const trimmedDescription = description.trim();

  if (!trimmedTitle) {
    alert("任务名称不能为空");
    return;
  }

  try {
    await request(`${API_BASE}/${todo.id}`, {
      method: "PUT",
      body: JSON.stringify({
        title: trimmedTitle,
        description: trimmedDescription,
      }),
    });

    await loadTodos();
  } catch (err) {
    alert(err.message);
  }
}

async function deleteTodo(id) {
  if (!confirm("确定删除这个任务吗？")) {
    return;
  }

  try {
    await request(`${API_BASE}/${id}`, {
      method: "DELETE",
    });

    await loadTodos();
  } catch (err) {
    alert(err.message);
  }
}

function renderTodos(todos) {
  todoListEl.innerHTML = "";

  if (todos.length === 0) {
    todoListEl.innerHTML = `<div class="empty">暂无任务</div>`;
    return;
  }

  for (const todo of todos) {
    const item = document.createElement("div");
    item.className = `todo-item ${todo.completed ? "done" : ""}`;

    item.innerHTML = `
      <div class="todo-title"></div>
      <div class="todo-desc"></div>
      <button class="done-btn">${todo.completed ? "已完成" : "完成"}</button>
      <button class="update-btn">更新</button>
      <button class="delete-btn">删除</button>
    `;

    item.querySelector(".todo-title").textContent = todo.title;
    item.querySelector(".todo-desc").textContent = todo.description || "-";

    const doneBtn = item.querySelector(".done-btn");

    if (todo.completed) {
      doneBtn.disabled = true;
    } else {
      doneBtn.addEventListener("click", () => completeTodo(todo.id));
    }

    item.querySelector(".update-btn").addEventListener("click", () => updateTodo(todo));
    item.querySelector(".delete-btn").addEventListener("click", () => deleteTodo(todo.id));

    todoListEl.appendChild(item);
  }
}