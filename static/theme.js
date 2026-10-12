(() => {
  try {
    const choice = localStorage.getItem('nanosite-theme');
    if (choice === 'light' || choice === 'dark') document.documentElement.dataset.theme = choice;
  } catch { /* Rendering still follows system colors if storage is unavailable. */ }
})();
