const SVG_NS = "http://www.w3.org/2000/svg";

const line = (d, color = "#222") =>
  `<path d="${d}" stroke="${color}" stroke-width="4" fill="none" stroke-linecap="round"/>`;
const dot = (x, y, r = 6) => `<circle cx="${x}" cy="${y}" r="${r}" fill="#222"/>`;

const AVATAR_COLORS = [
  "#ffd54f", "#ff8a65", "#e57373", "#f06292", "#ba68c8", "#9575cd",
  "#64b5f6", "#4dd0e1", "#4db6ac", "#81c784", "#aed581", "#a1887f",
];

const AVATAR_EYES = [
  dot(35, 42) + dot(65, 42),
  `<circle cx="35" cy="42" r="11" fill="#fff" stroke="#222" stroke-width="3"/>` +
    `<circle cx="65" cy="42" r="11" fill="#fff" stroke="#222" stroke-width="3"/>` +
    dot(37, 44, 5) + dot(67, 44, 5),
  line("M27 45 Q35 35 43 45 M57 45 Q65 35 73 45"),
  line("M27 43 H43 M57 43 H73"),
  dot(35, 44) + dot(65, 44) + line("M25 31 L43 38 M75 31 L57 38"),
  dot(35, 42) + line("M57 44 Q65 36 73 44"),
];

const AVATAR_MOUTHS = [
  line("M32 62 Q50 78 68 62"),
  `<path d="M30 60 Q50 88 70 60 Z" fill="#222"/>`,
  line("M36 67 H64"),
  dot(50, 67, 7),
  line("M32 62 Q50 76 68 62") + `<path d="M43 68 Q50 84 57 68 Z" fill="#e57373"/>`,
  line("M34 72 Q50 58 66 72"),
];

const AVATAR_PARTS = [
  { key: "color", label: "Couleur", count: AVATAR_COLORS.length },
  { key: "eyes", label: "Yeux", count: AVATAR_EYES.length },
  { key: "mouth", label: "Bouche", count: AVATAR_MOUTHS.length },
];

function validAvatar(avatar) {
  return (
    avatar !== null &&
    typeof avatar === "object" &&
    AVATAR_PARTS.every(({ key, count }) => Number.isInteger(avatar[key]) && avatar[key] >= 0 && avatar[key] < count)
  );
}

function randomAvatar() {
  return Object.fromEntries(AVATAR_PARTS.map(({ key, count }) => [key, Math.floor(Math.random() * count)]));
}

function avatarSVG(avatar, size = 40) {
  const svg = document.createElementNS(SVG_NS, "svg");
  svg.setAttribute("viewBox", "0 0 100 100");
  svg.setAttribute("width", size);
  svg.setAttribute("height", size);
  svg.setAttribute("class", "avatar");
  svg.setAttribute("aria-hidden", "true");
  if (!validAvatar(avatar)) return svg;
  svg.innerHTML =
    `<circle cx="50" cy="50" r="46" fill="${AVATAR_COLORS[avatar.color]}" stroke="#333" stroke-width="3"/>` +
    AVATAR_EYES[avatar.eyes] +
    AVATAR_MOUTHS[avatar.mouth];
  return svg;
}

function loadAvatar() {
  try {
    const avatar = JSON.parse(localStorage.getItem("avatar"));
    return validAvatar(avatar) ? avatar : null;
  } catch {
    return null;
  }
}

function saveAvatar(avatar) {
  try {
    localStorage.setItem("avatar", JSON.stringify(avatar));
  } catch {}
}

function createAvatarEditor(container, avatar, onChange) {
  let current = { ...avatar };
  const preview = document.createElement("div");
  preview.className = "avatar-preview";
  const controls = document.createElement("div");
  controls.className = "avatar-controls";

  const render = () => preview.replaceChildren(avatarSVG(current, 72));
  const update = (next) => {
    current = next;
    render();
    onChange({ ...current });
  };

  for (const { key, label, count } of AVATAR_PARTS) {
    const row = document.createElement("div");
    row.className = "avatar-row";
    const name = document.createElement("span");
    name.textContent = label;
    const button = (text, step) => {
      const b = document.createElement("button");
      b.type = "button";
      b.textContent = text;
      b.setAttribute("aria-label", `${label} ${step > 0 ? "suivant" : "précédent"}`);
      b.addEventListener("click", () => update({ ...current, [key]: (current[key] + step + count) % count }));
      return b;
    };
    row.append(button("◀", -1), name, button("▶", 1));
    controls.append(row);
  }

  const dice = document.createElement("button");
  dice.type = "button";
  dice.textContent = "🎲 Au hasard";
  dice.addEventListener("click", () => update(randomAvatar()));
  controls.append(dice);

  container.replaceChildren(preview, controls);
  render();
  return {
    set(next) {
      current = { ...next };
      render();
    },
  };
}
