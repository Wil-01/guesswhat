const ws = new WebSocket(`ws://${location.host}/ws`);
const form = document.getElementById("form");
const input = document.getElementById("input");
const output = document.getElementById("output");

ws.addEventListener("message", (event) => {
  const data = JSON.parse(event.data);
  console.log(data);
  const li = document.createElement("li");
  li.textContent = data.message;
  output.appendChild(li);
});

form.addEventListener("submit", (event) => {
  event.preventDefault();
  if (!input.value || ws.readyState !== WebSocket.OPEN) return;
  ws.send(JSON.stringify({ message: input.value }));
  input.value = "";
});
