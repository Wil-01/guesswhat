const ws = new WebSocket(`ws://${location.host}/ws`);
const form = document.getElementById("form");
const input = document.getElementById("input");
const output = document.getElementById("output");

ws.addEventListener("message", (event) => {
  console.log(event.data);
  const li = document.createElement("li");
  li.textContent = event.data;
  output.appendChild(li);
});

form.addEventListener("submit", (event) => {
  event.preventDefault();
  if (!input.value || ws.readyState !== WebSocket.OPEN) return;
  ws.send(input.value);
  input.value = "";
});
