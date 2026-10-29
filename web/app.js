const roomCode = new URLSearchParams(location.search).get("code");
if (!roomCode) location.replace("/");

const protocol = location.protocol === "https:" ? "wss" : "ws";
const ws = new WebSocket(`${protocol}://${location.host}/ws?room=${encodeURIComponent(roomCode)}`);
const form = document.getElementById("form");
const input = document.getElementById("input");
const messages = document.getElementById("messages");
const nicknameForm = document.getElementById("nickname-form");
const nicknameInput = document.getElementById("nickname");

let myId = 0;
let myNickname = "";

function storedNickname() {
  try {
    return localStorage.getItem("nickname") || "";
  } catch {
    return "";
  }
}

function storeNickname(nickname) {
  try {
    localStorage.setItem("nickname", nickname);
  } catch {}
}

function send(packet) {
  if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify(packet));
}

function addSystem(text, className = "system") {
  const li = document.createElement("li");
  li.className = className;
  li.textContent = text;
  append(li);
}

function addChat({ nickname, message }) {
  const li = document.createElement("li");
  if (nickname === myNickname) li.className = "mine";
  const author = document.createElement("span");
  author.className = "author";
  author.textContent = `${nickname}:`;
  li.append(author, document.createTextNode(message));
  append(li);
}

function append(li) {
  messages.appendChild(li);
  messages.scrollTop = messages.scrollHeight;
}

const handlers = {
  nickname: (p) => {
    myId = p.id;
    myNickname = p.nickname;
    nicknameInput.value = p.nickname;
  },
  history: (p) => p.messages.forEach(addChat),
  message: addChat,
  system: (p) => addSystem(p.message),
  error: (p) => addSystem(p.message, "error"),
};

ws.addEventListener("open", () => {
  addSystem("Connected");
  const saved = storedNickname();
  if (saved) send({ type: "nickname", nickname: saved });
  const avatar = loadAvatar();
  if (avatar) send({ type: "avatar", avatar });
});
ws.addEventListener("close", () => {
  addSystem("Disconnected");
  document.getElementById("closed").hidden = false;
});

document.getElementById("room-code").textContent = roomCode;
document.getElementById("copy-link").addEventListener("click", async (event) => {
  try {
    await navigator.clipboard.writeText(location.href);
    event.target.textContent = "Lien copié !";
  } catch {
    event.target.textContent = location.href;
  }
});
ws.addEventListener("message", (event) => {
  const packet = JSON.parse(event.data);
  const handler = handlers[packet.type];
  if (handler) handler(packet);
  else console.error("unknown packet", packet);
});

nicknameForm.addEventListener("submit", (event) => {
  event.preventDefault();
  const nickname = nicknameInput.value.trim();
  if (!nickname || nickname === myNickname) return;
  storeNickname(nickname);
  send({ type: "nickname", nickname });
});

form.addEventListener("submit", (event) => {
  event.preventDefault();
  if (!input.value.trim()) return;
  send({ type: "message", message: input.value });
  input.value = "";
});
