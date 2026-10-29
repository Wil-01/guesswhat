(() => {
  const root = document.documentElement;
  let saved = null;
  try {
    saved = localStorage.getItem("color-scheme");
  } catch {}
  const dark = saved ? saved === "dark" : matchMedia("(prefers-color-scheme: dark)").matches;
  root.dataset.theme = dark ? "dark" : "light";

  document.addEventListener("DOMContentLoaded", () => {
    const button = document.getElementById("theme-toggle");
    if (!button) return;
    const render = () => {
      const isDark = root.dataset.theme === "dark";
      button.textContent = isDark ? "☀️" : "🌙";
      button.title = isDark ? "Thème clair" : "Thème sombre";
      button.setAttribute("aria-label", button.title);
    };
    button.addEventListener("click", () => {
      root.dataset.theme = root.dataset.theme === "dark" ? "light" : "dark";
      try {
        localStorage.setItem("color-scheme", root.dataset.theme);
      } catch {}
      render();
    });
    render();
  });
})();
