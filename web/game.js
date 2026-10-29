const playersList = document.getElementById("players");
const playerTemplate = document.getElementById("player-template");
const startButton = document.getElementById("start");
const wordBanner = document.getElementById("word");
const toolbar = document.getElementById("toolbar");
const timer = document.getElementById("timer");
const roundLabel = document.getElementById("round");
const roundTheme = document.getElementById("round-theme");
const themeSelect = document.getElementById("theme");
const roundsInput = document.getElementById("rounds");
const durationSelect = document.getElementById("duration");
const settingsInputs = [themeSelect, roundsInput, durationSelect];
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
    roundTheme.hidden = true;
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
      li.querySelector(".avatar-slot").replaceChildren(avatarSVG(player.avatar, 32));
      li.querySelector(".name").textContent = player.nickname + (player.id === myId ? " (you)" : "");
      li.querySelector(".badges").textContent = (player.drawer ? "✏️" : "") + (player.admin ? "👑" : "");
      li.querySelector(".score").textContent = player.score;
      return li;
    })
  );
  const me = players.find((player) => player.id === myId);
  const canManage = Boolean(me && me.admin && gameState === "waiting");
  startButton.hidden = !canManage;
  settingsInputs.forEach((input) => (input.disabled = !canManage));
}

function themeLabel(theme) {
  return theme.charAt(0).toUpperCase() + theme.slice(1);
}

const avatarEditor = createAvatarEditor(document.getElementById("avatar-editor"), loadAvatar() || randomAvatar(), (avatar) => {
  saveAvatar(avatar);
  send({ type: "avatar", avatar });
});

handlers.players = (p) => {
  players = p.players;
  const me = players.find((player) => player.id === myId);
  if (me && !loadAvatar()) {
    saveAvatar(me.avatar);
    avatarEditor.set(me.avatar);
  }
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
  roundTheme.hidden = false;
  roundTheme.textContent = `Thème : ${themeLabel(p.theme)}`;
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
      li.querySelector(".avatar-slot").replaceChildren(avatarSVG(player.avatar, 28));
      li.querySelector(".name").textContent = player.nickname;
      li.querySelector(".score").textContent = `${player.score} pts`;
      return li;
    })
  );
  results.showModal();
};

let settings = null;

function renderSettings() {
  themeSelect.value = settings.theme;
  roundsInput.value = settings.rounds;
  if (![...durationSelect.options].some((option) => Number(option.value) === settings.duration)) {
    durationSelect.append(new Option(`${settings.duration} s`, settings.duration));
  }
  durationSelect.value = settings.duration;
}

handlers.settings = (p) => {
  if (p.themes && themeSelect.options.length !== p.themes.length + 1) {
    themeSelect.append(...p.themes.map((theme) => new Option(themeLabel(theme), theme)));
  }
  settings = { theme: p.theme, rounds: p.rounds, duration: p.duration };
  renderSettings();
};

function sendSettings() {
  const rounds = Number(roundsInput.value);
  if (!Number.isInteger(rounds) || rounds < 1 || rounds > 10) {
    if (settings) renderSettings();
    return;
  }
  send({
    type: "settings",
    theme: themeSelect.value,
    rounds,
    duration: Number(durationSelect.value),
  });
}

settingsInputs.forEach((input) => input.addEventListener("change", sendSettings));
startButton.addEventListener("click", () => send({ type: "start" }));
