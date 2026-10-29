const playersList = document.getElementById("players");
const playerTemplate = document.getElementById("player-template");
const startButton = document.getElementById("start");
const wordBanner = document.getElementById("word");
const toolbar = document.getElementById("toolbar");
const timer = document.getElementById("timer");
const roundLabel = document.getElementById("round");
const results = document.getElementById("results");
const ranking = document.getElementById("ranking");
const rankTemplate = document.getElementById("rank-template");

let players = [];
let gameState = "waiting";
let drawerId = 0;
let countdown = null;

function canDraw() {
  return gameState !== "playing" || drawerId === myId;
}

function renderBoard() {
  toolbar.classList.toggle("disabled", !canDraw());
  canvas.style.cursor = canDraw() ? "crosshair" : "not-allowed";
  if (gameState !== "playing") {
    roundLabel.hidden = true;
    wordBanner.hidden = true;
    timer.hidden = true;
    stopCountdown();
  }
}

function startCountdown(remaining) {
  stopCountdown();
  const endsAt = Date.now() + remaining;
  const tick = () => {
    const seconds = Math.max(0, Math.ceil((endsAt - Date.now()) / 1000));
    timer.textContent = `⏱ ${seconds}s`;
    if (seconds === 0) stopCountdown();
  };
  timer.hidden = false;
  tick();
  countdown = setInterval(tick, 250);
}

function stopCountdown() {
  clearInterval(countdown);
  countdown = null;
}

function renderPlayers() {
  playersList.replaceChildren(
    ...players.map((player) => {
      const li = playerTemplate.content.firstElementChild.cloneNode(true);
      li.classList.toggle("me", player.id === myId);
      li.querySelector(".name").textContent = player.nickname + (player.id === myId ? " (you)" : "");
      li.querySelector(".badges").textContent = (player.drawer ? "✏️" : "") + (player.admin ? "👑" : "");
      li.querySelector(".score").textContent = player.score;
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
  roundLabel.hidden = false;
  roundLabel.textContent = `Manche ${p.round}/${p.rounds}`;
  wordBanner.hidden = false;
  wordBanner.textContent = p.word ? `Dessine : ${p.word}` : p.hint;
  startCountdown(p.remaining);
  renderBoard();
};

handlers.roundEnd = (p) => {
  drawerId = 0;
  stopCountdown();
  timer.hidden = true;
  wordBanner.textContent = p.word;
  renderBoard();
  addSystem(`The word was: ${p.word}`);
};

handlers.gameEnd = (p) => {
  ranking.replaceChildren(
    ...p.players.map((player) => {
      const li = rankTemplate.content.firstElementChild.cloneNode(true);
      li.querySelector(".name").textContent = player.nickname;
      li.querySelector(".score").textContent = `${player.score} pts`;
      return li;
    })
  );
  results.showModal();
};

startButton.addEventListener("click", () => send({ type: "start" }));
