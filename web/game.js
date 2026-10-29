const playersList = document.getElementById("players");
const playerTemplate = document.getElementById("player-template");
const startButton = document.getElementById("start");
const wordBanner = document.getElementById("word");
const toolbar = document.getElementById("toolbar");

let players = [];
let gameState = "waiting";
let drawerId = 0;

function canDraw() {
  return gameState !== "playing" || drawerId === myId;
}

function renderBoard() {
  toolbar.classList.toggle("disabled", !canDraw());
  canvas.style.cursor = canDraw() ? "crosshair" : "not-allowed";
  if (gameState !== "playing") wordBanner.hidden = true;
}

function renderPlayers() {
  playersList.replaceChildren(
    ...players.map((player) => {
      const li = playerTemplate.content.firstElementChild.cloneNode(true);
      li.classList.toggle("me", player.id === myId);
      li.querySelector(".name").textContent = player.nickname + (player.id === myId ? " (you)" : "");
      li.querySelector(".badges").textContent = (player.drawer ? "✏️" : "") + (player.admin ? "👑" : "");
      return li;
    })
  );
  const me = players.find((player) => player.id === myId);
  startButton.hidden = !(me && me.admin && gameState === "waiting");
}

handlers.players = (p) => {
  players = p.players;
  renderPlayers();
};

handlers.game = (p) => {
  gameState = p.state;
  renderPlayers();
  renderBoard();
};

handlers.round = (p) => {
  drawerId = p.drawer;
  wordBanner.hidden = false;
  wordBanner.textContent = p.word ? `Dessine : ${p.word}` : p.hint;
  renderBoard();
};

startButton.addEventListener("click", () => send({ type: "start" }));
