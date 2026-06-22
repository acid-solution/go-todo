const todoListEl = document.querySelector("#todoList");
const titleInput = document.querySelector("#titleInput");
const descInput = document.querySelector("#descInput");
const createBtn = document.querySelector("#createBtn");

const API_BASE = "/api/todos";

document.addEventListener("DOMContentLoaded", loadTodos);
createBtn.addEventListener("click", createTodo);

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

async function loadTodos() {
  try {
    const todos = await request(API_BASE);
    renderTodos(todos || []);
  } catch (err) {
    alert(err.message);
  }
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

  const description = prompt("新的任务描述", todo.description);
  if (description === null) return;

  try {
    await request(`${API_BASE}/${todo.id}`, {
      method: "PUT",
      body: JSON.stringify({
        title: title.trim(),
        description: description.trim(),
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
    item.className = `todo-item ${todo.done ? "done" : ""}`;

    item.innerHTML = `
      <div class="todo-title"></div>
      <div class="todo-desc"></div>
      <button class="done-btn">${todo.done ? "已完成" : "完成"}</button>
      <button class="update-btn">更新</button>
      <button class="delete-btn">删除</button>
    `;

    item.querySelector(".todo-title").textContent = todo.title;
    item.querySelector(".todo-desc").textContent = todo.description || "-";

    item.querySelector(".done-btn").addEventListener("click", () => completeTodo(todo.id));
    item.querySelector(".update-btn").addEventListener("click", () => updateTodo(todo));
    item.querySelector(".delete-btn").addEventListener("click", () => deleteTodo(todo.id));

    todoListEl.appendChild(item);
  }
}
