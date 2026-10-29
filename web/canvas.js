const canvas = document.getElementById("canvas");
const ctx = canvas.getContext("2d", { willReadFrequently: true });
const colorInput = document.getElementById("color");
const sizeInput = document.getElementById("size");
const palette = document.getElementById("palette");
const toolButtons = document.querySelectorAll(".tool");

const BACKGROUND = "#ffffff";
const COLORS = [
  "#000000", "#7f7f7f", "#c62828", "#ef6c00", "#fdd835", "#43a047", "#00897b",
  "#ffffff", "#d0d0d0", "#1e88e5", "#3949ab", "#8e24aa", "#d81b60", "#6d4c41",
];
const SPRAY_DOTS = 12;
const SPRAY_INTERVAL = 40;
const FILL_TOLERANCE = 48;

let tool = "brush";
let last = null;
let pointer = null;
let sprayTimer = null;
let beginAction = false;

function drawStroke(stroke) {
  switch (stroke.kind) {
    case "line":
      drawLine(stroke);
      break;
    case "spray":
      drawSpray(stroke);
      break;
    case "fill":
      floodFill(stroke);
      break;
  }
}

function drawLine({ from, to, color, size }) {
  ctx.strokeStyle = color;
  ctx.lineWidth = size;
  ctx.lineCap = "round";
  ctx.beginPath();
  ctx.moveTo(from.x * canvas.width, from.y * canvas.height);
  ctx.lineTo(to.x * canvas.width, to.y * canvas.height);
  ctx.stroke();
}

function drawSpray({ points, color }) {
  ctx.fillStyle = color;
  for (const p of points) {
    ctx.fillRect(p.x * canvas.width - 1, p.y * canvas.height - 1, 2, 2);
  }
}

function hexToRgb(hex) {
  const n = parseInt(hex.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

function floodFill({ from, color }) {
  const { width, height } = canvas;
  const x = Math.min(width - 1, Math.floor(from.x * width));
  const y = Math.min(height - 1, Math.floor(from.y * height));
  const image = ctx.getImageData(0, 0, width, height);
  const data = image.data;
  const start = (y * width + x) * 4;
  const target = [data[start], data[start + 1], data[start + 2]];
  const fill = hexToRgb(color);
  if (target.every((v, i) => v === fill[i])) return;

  const matches = (i) =>
    Math.abs(data[i] - target[0]) <= FILL_TOLERANCE &&
    Math.abs(data[i + 1] - target[1]) <= FILL_TOLERANCE &&
    Math.abs(data[i + 2] - target[2]) <= FILL_TOLERANCE;
  const visited = new Uint8Array(width * height);
  const stack = [y * width + x];
  while (stack.length > 0) {
    const p = stack.pop();
    if (visited[p]) continue;
    visited[p] = 1;
    const i = p * 4;
    if (!matches(i)) continue;
    data[i] = fill[0];
    data[i + 1] = fill[1];
    data[i + 2] = fill[2];
    data[i + 3] = 255;
    const px = p % width;
    if (px > 0) stack.push(p - 1);
    if (px < width - 1) stack.push(p + 1);
    if (p >= width) stack.push(p - width);
    if (p < width * (height - 1)) stack.push(p + width);
  }
  ctx.putImageData(image, 0, 0);
}

function clearCanvas() {
  ctx.fillStyle = BACKGROUND;
  ctx.fillRect(0, 0, canvas.width, canvas.height);
}

function emit(stroke) {
  if (beginAction) {
    stroke.begin = true;
    beginAction = false;
  }
  drawStroke(stroke);
  send({ type: "draw", stroke });
}

function currentColor() {
  return tool === "eraser" ? BACKGROUND : colorInput.value;
}

function currentSize() {
  return Number(sizeInput.value);
}

function toPoint(event) {
  const rect = canvas.getBoundingClientRect();
  const clamp = (v) => Math.min(1, Math.max(0, v));
  return {
    x: clamp((event.clientX - rect.left) / rect.width),
    y: clamp((event.clientY - rect.top) / rect.height),
  };
}

function sprayTick() {
  const radius = currentSize() * 2;
  const clamp = (v) => Math.min(1, Math.max(0, v));
  const points = Array.from({ length: SPRAY_DOTS }, () => {
    const angle = Math.random() * 2 * Math.PI;
    const dist = radius * Math.sqrt(Math.random());
    return {
      x: clamp(pointer.x + (Math.cos(angle) * dist) / canvas.width),
      y: clamp(pointer.y + (Math.sin(angle) * dist) / canvas.height),
    };
  });
  emit({ kind: "spray", points, color: currentColor(), size: currentSize() });
}

function stopDrawing() {
  last = null;
  pointer = null;
  clearInterval(sprayTimer);
  sprayTimer = null;
}

canvas.addEventListener("pointerdown", (event) => {
  if (!canDraw()) return;
  canvas.setPointerCapture(event.pointerId);
  const point = toPoint(event);
  beginAction = true;
  switch (tool) {
    case "fill":
      emit({ kind: "fill", from: point, color: currentColor() });
      break;
    case "spray":
      pointer = point;
      sprayTick();
      sprayTimer = setInterval(sprayTick, SPRAY_INTERVAL);
      break;
    default:
      last = point;
      emit({ kind: "line", from: point, to: point, color: currentColor(), size: currentSize() });
  }
});

canvas.addEventListener("pointermove", (event) => {
  if (!canDraw()) return stopDrawing();
  const point = toPoint(event);
  if (pointer) pointer = point;
  if (!last) return;
  emit({ kind: "line", from: last, to: point, color: currentColor(), size: currentSize() });
  last = point;
});

["pointerup", "pointercancel"].forEach((name) => canvas.addEventListener(name, stopDrawing));

function selectTool(name) {
  tool = name;
  toolButtons.forEach((button) => button.classList.toggle("active", button.dataset.tool === name));
}

function selectColor(color) {
  colorInput.value = color;
  palette.querySelectorAll(".swatch").forEach((swatch) =>
    swatch.classList.toggle("active", swatch.dataset.color === color)
  );
  if (tool === "eraser") selectTool("brush");
}

toolButtons.forEach((button) => button.addEventListener("click", () => selectTool(button.dataset.tool)));

palette.append(
  ...COLORS.map((color) => {
    const swatch = document.createElement("button");
    swatch.type = "button";
    swatch.className = "swatch";
    swatch.dataset.color = color;
    swatch.style.background = color;
    swatch.title = color;
    swatch.addEventListener("click", () => selectColor(color));
    return swatch;
  })
);
colorInput.addEventListener("input", () => selectColor(colorInput.value));
selectColor(colorInput.value);

function undo() {
  if (canDraw()) send({ type: "undo" });
}

document.getElementById("undo").addEventListener("click", undo);
document.getElementById("clear").addEventListener("click", () => send({ type: "clear" }));
document.addEventListener("keydown", (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key === "z" && !(event.target instanceof HTMLInputElement)) {
    event.preventDefault();
    undo();
  }
});

clearCanvas();

handlers.draw = (p) => drawStroke(p.stroke);
handlers.drawing = (p) => {
  clearCanvas();
  p.strokes.forEach(drawStroke);
};
handlers.clear = clearCanvas;
