(() => {
  const theme = document.querySelector('#theme');
  const choices = ['system', 'light', 'dark'];
  let choice = document.documentElement.dataset.theme || 'system';
  const update = () => { theme.textContent = `Theme: ${choice}`; };
  update();
  theme.addEventListener('click', () => {
    choice = choices[(choices.indexOf(choice) + 1) % choices.length];
    if (choice === 'system') delete document.documentElement.dataset.theme;
    else document.documentElement.dataset.theme = choice;
    try { localStorage.setItem('nanosite-theme', choice); } catch {}
    update();
  });
  const navigation = document.querySelector('#navigation');
  const narrow = matchMedia('(max-width: 760px)');
  const collapse = () => { navigation.open = !narrow.matches; };
  collapse(); narrow.addEventListener('change', collapse);
  const input = document.querySelector('#search');
  const results = document.querySelector('#search-results');
  if (!input) return;
  // The index script resolves against the current page on every deployment prefix,
  // and also when opening generated HTML through file://.
  const indexScript = [...document.scripts].find(s => s.src.endsWith('/assets/search-index.js'));
  const root = new URL('../', indexScript.src);
  input.addEventListener('input', () => {
    const terms = input.value.trim().toLocaleLowerCase().split(/\s+/).filter(Boolean);
    results.replaceChildren(); results.hidden = !terms.length;
    if (!terms.length) return;
    const hits = (globalThis.NANOSITE_SEARCH || []).filter(p => terms.every(t => `${p.title} ${p.text}`.toLocaleLowerCase().includes(t)))
      .sort((a, b) => Number(terms.every(t => b.title.toLocaleLowerCase().includes(t))) - Number(terms.every(t => a.title.toLocaleLowerCase().includes(t))));
    const note = document.createElement('p'); note.textContent = hits.length ? `${hits.length} result${hits.length === 1 ? '' : 's'}` : 'No matching pages'; results.append(note);
    const list = document.createElement('ul');
    for (const hit of hits.slice(0, 20)) {
      const li = document.createElement('li'); const a = document.createElement('a');
      a.href = new URL(hit.url.split('/').map(encodeURIComponent).join('/'), root).href; a.textContent = hit.title;
      const summary = document.createElement('p'); summary.textContent = hit.text.slice(0, 180);
      li.append(a, summary); list.append(li);
    }
    results.append(list);
  });
  document.addEventListener('keydown', event => {
    const editing = /INPUT|TEXTAREA|SELECT/.test(document.activeElement?.tagName) || document.activeElement?.isContentEditable;
    if ((!editing && event.key === '/') || ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k')) { event.preventDefault(); input.focus(); }
    if (event.key === 'Escape') { results.hidden = true; input.blur(); }
    if (event.key === 'ArrowDown' && document.activeElement === input && !results.hidden) { event.preventDefault(); results.querySelector('a')?.focus(); }
  });
  results.addEventListener('keydown', event => {
    const links = [...results.querySelectorAll('a')]; const i = links.indexOf(document.activeElement);
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') { event.preventDefault(); links[Math.max(0, Math.min(links.length - 1, i + (event.key === 'ArrowDown' ? 1 : -1)))]?.focus(); }
  });
})();
