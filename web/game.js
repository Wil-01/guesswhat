const playersList = document.getElementById("players");
const playerTemplate = document.getElementById("player-template");
const startButton = document.getElementById("start");

let players = [];
let gameState = "waiting";

function renderPlayers() {
  playersList.replaceChildren(
    ...players.map((player) => {
      const li = playerTemplate.content.firstElementChild.cloneNode(true);
      li.classList.toggle("me", player.id === myId);
      li.querySelector(".name").textContent = player.nickname + (player.id === myId ? " (you)" : "");
      li.querySelector(".badges").textContent = player.admin ? "👑" : "";
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
};

startButton.addEventListener("click", () => send({ type: "start" }));
