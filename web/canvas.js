const canvas = document.getElementById("canvas");
const ctx = canvas.getContext("2d");
const colorInput = document.getElementById("color");
const sizeInput = document.getElementById("size");
const clearButton = document.getElementById("clear");

let last = null;

function drawStroke({ from, to, color, size }) {
  ctx.strokeStyle = color;
  ctx.lineWidth = size;
  ctx.lineCap = "round";
  ctx.beginPath();
  ctx.moveTo(from.x * canvas.width, from.y * canvas.height);
  ctx.lineTo(to.x * canvas.width, to.y * canvas.height);
  ctx.stroke();
}

function clearCanvas() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
}

function toPoint(event) {
  const rect = canvas.getBoundingClientRect();
  const clamp = (v) => Math.min(1, Math.max(0, v));
  return {
    x: clamp((event.clientX - rect.left) / rect.width),
    y: clamp((event.clientY - rect.top) / rect.height),
  };
}

canvas.addEventListener("pointerdown", (event) => {
  if (!canDraw()) return;
  canvas.setPointerCapture(event.pointerId);
  last = toPoint(event);
});

canvas.addEventListener("pointermove", (event) => {
  if (!last || !canDraw()) return;
  const point = toPoint(event);
  const stroke = { from: last, to: point, color: colorInput.value, size: Number(sizeInput.value) };
  drawStroke(stroke);
  send({ type: "draw", stroke });
  last = point;
});

["pointerup", "pointercancel"].forEach((name) =>
  canvas.addEventListener(name, () => (last = null))
);

clearButton.addEventListener("click", () => send({ type: "clear" }));

handlers.draw = (p) => drawStroke(p.stroke);
handlers.drawing = (p) => {
  clearCanvas();
  p.strokes.forEach(drawStroke);
};
handlers.clear = clearCanvas;
