const ws = new WebSocket(`ws://${location.host}/ws`);
const form = document.getElementById("form");
const input = document.getElementById("input");
const messages = document.getElementById("messages");

function addMessage(text, className) {
  const li = document.createElement("li");
  li.textContent = text;
  if (className) li.className = className;
  messages.appendChild(li);
  messages.scrollTop = messages.scrollHeight;
}

ws.addEventListener("open", () => addMessage("Connected", "system"));
ws.addEventListener("close", () => addMessage("Disconnected", "system"));
ws.addEventListener("message", (event) => {
  const data = JSON.parse(event.data);
  addMessage(data.message);
});

form.addEventListener("submit", (event) => {
  event.preventDefault();
  if (!input.value.trim() || ws.readyState !== WebSocket.OPEN) return;
  ws.send(JSON.stringify({ message: input.value }));
  input.value = "";
});
