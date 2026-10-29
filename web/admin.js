const loginForm = document.getElementById("login");
const passwordInput = document.getElementById("password");
const panel = document.getElementById("panel");
const statusLine = document.getElementById("status");
const createForm = document.getElementById("create-theme");
const themeNameInput = document.getElementById("theme-name");
const themeWordsInput = document.getElementById("theme-words");
const filterInput = document.getElementById("filter");
const themesContainer = document.getElementById("themes");
const themeTemplate = document.getElementById("theme-template");
const chipTemplate = document.getElementById("chip-template");

let password = "";
let themes = {};

try {
  password = sessionStorage.getItem("adminPassword") || "";
} catch {}

function showStatus(text, isError = false) {
  statusLine.textContent = text;
  statusLine.classList.toggle("error", isError);
  statusLine.hidden = false;
}

function setPassword(value) {
  password = value;
  try {
    if (value) sessionStorage.setItem("adminPassword", value);
    else sessionStorage.removeItem("adminPassword");
  } catch {}
}

function logout(message) {
  setPassword("");
  panel.hidden = true;
  loginForm.hidden = false;
  if (message) showStatus(message, true);
}

async function api(method, path, body) {
  const options = { method, headers: { Authorization: `Bearer ${password}` } };
  if (body !== undefined) {
    options.headers["Content-Type"] = "application/json";
    options.body = JSON.stringify(body);
  }
  const res = await fetch(path, options);
  if (res.status === 401) {
    logout("Mot de passe incorrect.");
    throw new Error("unauthorized");
  }
  if (!res.ok) throw new Error((await res.text()).trim() || res.statusText);
  return res.status === 204 ? null : res.json();
}

async function run(action, success) {
  try {
    await action();
    if (success) showStatus(success);
    await load();
  } catch (err) {
    if (err.message !== "unauthorized") showStatus(err.message, true);
  }
}

async function load() {
  themes = await api("GET", "/api/admin/themes");
  loginForm.hidden = true;
  panel.hidden = false;
  render();
}

function render() {
  const filter = filterInput.value.trim().toLowerCase();
  themesContainer.replaceChildren(
    ...Object.keys(themes)
      .sort()
      .map((name) => renderTheme(name, themes[name], filter))
  );
}

function renderTheme(name, words, filter) {
  const section = themeTemplate.content.firstElementChild.cloneNode(true);
  section.querySelector(".theme-name").textContent = name;
  section.querySelector(".count").textContent = `(${words.length} mots)`;
  section.querySelector(".delete-theme").addEventListener("click", () => {
    if (!confirm(`Supprimer le thème « ${name} » et ses ${words.length} mots ?`)) return;
    run(() => api("DELETE", `/api/admin/themes/${encodeURIComponent(name)}`), `Thème « ${name} » supprimé.`);
  });

  const visible = words.filter((word) => word.toLowerCase().includes(filter));
  section.querySelector(".chips").replaceChildren(
    ...visible.map((word) => {
      const chip = chipTemplate.content.firstElementChild.cloneNode(true);
      chip.querySelector(".word").textContent = word;
      chip.querySelector(".remove").addEventListener("click", () =>
        run(
          () => api("DELETE", `/api/admin/themes/${encodeURIComponent(name)}/words/${encodeURIComponent(word)}`),
          `« ${word} » supprimé.`
        )
      );
      return chip;
    })
  );
  if (filter && visible.length === 0) section.hidden = true;

  const form = section.querySelector(".add-word");
  form.addEventListener("submit", (event) => {
    event.preventDefault();
    const word = form.elements[0].value;
    run(() => api("POST", `/api/admin/themes/${encodeURIComponent(name)}/words`, { word }), `« ${word.trim()} » ajouté à ${name}.`);
  });
  return section;
}

loginForm.addEventListener("submit", (event) => {
  event.preventDefault();
  setPassword(passwordInput.value);
  statusLine.hidden = true;
  run(() => Promise.resolve());
});

createForm.addEventListener("submit", (event) => {
  event.preventDefault();
  const name = themeNameInput.value.trim();
  const words = themeWordsInput.value.split(",").filter((word) => word.trim());
  run(async () => {
    await api("POST", "/api/admin/themes", { name, words });
    createForm.reset();
  }, `Thème « ${name} » créé.`);
});

filterInput.addEventListener("input", render);
document.getElementById("logout").addEventListener("click", () => logout());

if (password) run(() => Promise.resolve());
