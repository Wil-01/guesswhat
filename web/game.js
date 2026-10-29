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
const teamsInput = document.getElementById("teams");
const settingsInputs = [themeSelect, roundsInput, durationSelect, teamsInput];
const teamTemplate = document.getElementById("team-template");
const teamResult = document.getElementById("team-result");
const TEAM_NAMES = ["", "Équipe rouge", "Équipe bleue"];
const results = document.getElementById("results");
const ranking = document.getElementById("ranking");
const rankTemplate = document.getElementById("rank-template");

let players = [];
let teamScores = [];
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

function playerItem(player) {
  const li = playerTemplate.content.firstElementChild.cloneNode(true);
  li.classList.toggle("me", player.id === myId);
  if (player.team) li.classList.add(`team-${player.team}`);
  li.querySelector(".avatar-slot").replaceChildren(avatarSVG(player.avatar, 32));
  li.querySelector(".name").textContent = player.nickname + (player.id === myId ? " (you)" : "");
  li.querySelector(".badges").textContent = (player.drawer ? "✏️" : "") + (player.admin ? "👑" : "");
  li.querySelector(".score").textContent = player.score;
  return li;
}

function teamItem(team, me) {
  const li = teamTemplate.content.firstElementChild.cloneNode(true);
  li.classList.add(`team-${team}`);
  li.querySelector(".team-name").textContent = TEAM_NAMES[team];
  li.querySelector(".team-score").textContent = `${teamScores[team - 1] || 0} pts`;
  const button = li.querySelector(".join-team");
  button.hidden = gameState !== "waiting" || !me || me.team === team;
  button.addEventListener("click", () => send({ type: "team", team }));
  return li;
}

function myTeam() {
  const me = players.find((player) => player.id === myId);
  return me ? me.team : 0;
}

function renderPlayers() {
  const me = players.find((player) => player.id === myId);
  if (players.some((player) => player.team)) {
    playersList.replaceChildren(
      ...[1, 2].flatMap((team) => [teamItem(team, me), ...players.filter((player) => player.team === team).map(playerItem)])
    );
  } else {
    playersList.replaceChildren(...players.map(playerItem));
  }
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
  teamScores = p.teams || [];
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
  if (p.team) {
    roundLabel.textContent += ` · ${TEAM_NAMES[p.team]}` + (p.team === myTeam() ? "" : " (vous ne pouvez pas deviner)");
  }
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
      if (player.team) li.classList.add(`team-${player.team}`);
      return li;
    })
  );
  teamResult.hidden = !p.teams;
  if (p.teams) {
    const [red, blue] = p.teams;
    const winner = red === blue ? "Égalité" : `Victoire de l'${TEAM_NAMES[red > blue ? 1 : 2].toLowerCase()}`;
    teamResult.textContent = `${winner} (${red} – ${blue})`;
  }
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
  teamsInput.checked = settings.teams;
}

handlers.settings = (p) => {
  if (p.themes && themeSelect.options.length !== p.themes.length + 1) {
    themeSelect.append(...p.themes.map((theme) => new Option(themeLabel(theme), theme)));
  }
  settings = { theme: p.theme, rounds: p.rounds, duration: p.duration, teams: p.teams };
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
    teams: teamsInput.checked,
  });
}

settingsInputs.forEach((input) => input.addEventListener("change", sendSettings));
startButton.addEventListener("click", () => send({ type: "start" }));
