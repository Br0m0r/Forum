// Function to set theme and logo according to localStorage or default
function applyThemeFromStorage() {
  const darkMode = localStorage.getItem('darkMode');
  if (darkMode === 'enabled') {
    document.documentElement.classList.add('dark-mode');
    setLogo('dark');
  } else {
    document.documentElement.classList.remove('dark-mode');
    setLogo('light');
  }
}

// Function to set the logo image depending on mode
function setLogo(mode) {
  const logo = document.getElementById('logo');
  if (!logo) return;
  logo.src = mode === 'dark' ? '/static/img/logo_dark.png' : '/static/img/logo_white.png';
}

// On page load, apply theme and logo
window.addEventListener('DOMContentLoaded', applyThemeFromStorage);

// On toggle, update theme, localStorage, and logo
document.getElementById('theme').onclick = function() {
  const isDark = document.documentElement.classList.toggle('dark-mode');
  localStorage.setItem('darkMode', isDark ? 'enabled' : 'disabled');
  setLogo(isDark ? 'dark' : 'light');
};