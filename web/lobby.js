const nicknameInput = document.getElementById("nickname");
const privateInput = document.getElementById("private");
const themeSelect = document.getElementById("theme");
const roundsInput = document.getElementById("rounds");
const durationSelect = document.getElementById("duration");
const createButton = document.getElementById("create");
const createError = document.getElementById("create-error");
const joinForm = document.getElementById("join-form");
const codeInput = document.getElementById("code");
const roomsList = document.getElementById("rooms");
const noRooms = document.getElementById("no-rooms");
const roomTemplate = document.getElementById("room-template");

const MAX_PLAYERS = 10;

try {
  nicknameInput.value = localStorage.getItem("nickname") || "";
} catch {}

nicknameInput.addEventListener("change", () => {
  try {
    localStorage.setItem("nickname", nicknameInput.value.trim());
  } catch {}
});

const initialAvatar = loadAvatar() || randomAvatar();
saveAvatar(initialAvatar);
createAvatarEditor(document.getElementById("avatar-editor"), initialAvatar, saveAvatar);

function goToRoom(code) {
  location.href = `/room.html?code=${encodeURIComponent(code.toUpperCase())}`;
}

createButton.addEventListener("click", async () => {
  createError.hidden = true;
  try {
    const res = await fetch("/api/rooms", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        private: privateInput.checked,
        theme: themeSelect.value,
        rounds: Number(roundsInput.value),
        duration: Number(durationSelect.value),
      }),
    });
    if (!res.ok) throw new Error(await res.text());
    const { code } = await res.json();
    goToRoom(code);
  } catch (err) {
    createError.textContent = `Impossible de créer la room : ${err.message}`;
    createError.hidden = false;
  }
});

joinForm.addEventListener("submit", (event) => {
  event.preventDefault();
  const code = codeInput.value.trim();
  if (code) goToRoom(code);
});

async function refreshRooms() {
  try {
    const res = await fetch("/api/rooms");
    if (!res.ok) throw new Error(res.statusText);
    renderRooms(await res.json());
  } catch (err) {
    console.error("cannot list rooms:", err);
  }
}

function renderRooms(rooms) {
  noRooms.hidden = rooms.length > 0;
  roomsList.replaceChildren(
    ...rooms.map((room) => {
      const li = roomTemplate.content.firstElementChild.cloneNode(true);
      const link = li.querySelector(".room-link");
      link.href = `/room.html?code=${encodeURIComponent(room.code)}`;
      link.querySelector(".code").textContent = room.code;
      const state = room.state === "playing" ? "en jeu" : "en attente";
      const theme = room.theme ? themeLabel(room.theme) : "Tous les thèmes";
      link.querySelector(".meta").textContent = `${theme} · ${room.players}/${MAX_PLAYERS} joueurs · ${state}`;
      return li;
    })
  );
}

function themeLabel(theme) {
  return theme.charAt(0).toUpperCase() + theme.slice(1);
}

async function loadThemes() {
  try {
    const res = await fetch("/api/themes");
    if (!res.ok) throw new Error(res.statusText);
    const themes = await res.json();
    themeSelect.append(...themes.map((theme) => new Option(themeLabel(theme), theme)));
  } catch (err) {
    console.error("cannot list themes:", err);
  }
}

document.getElementById("refresh").addEventListener("click", refreshRooms);
loadThemes();
refreshRooms();
setInterval(refreshRooms, 3000);
