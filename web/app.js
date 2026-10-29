const ws = new WebSocket(`ws://${location.host}/ws`);
const form = document.getElementById("form");
const input = document.getElementById("input");
const messages = document.getElementById("messages");
const nicknameForm = document.getElementById("nickname-form");
const nicknameInput = document.getElementById("nickname");

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

function addSystem(text) {
  const li = document.createElement("li");
  li.className = "system";
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
    myNickname = p.nickname;
    nicknameInput.value = p.nickname;
  },
  history: (p) => p.messages.forEach(addChat),
  message: addChat,
  system: (p) => addSystem(p.message),
};

ws.addEventListener("open", () => {
  addSystem("Connected");
  const saved = storedNickname();
  if (saved) send({ type: "nickname", nickname: saved });
});
ws.addEventListener("close", () => addSystem("Disconnected"));
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
