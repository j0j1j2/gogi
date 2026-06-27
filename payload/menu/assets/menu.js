async function refresh() {
  const state = await gogi.state();
  document.getElementById('status').textContent = 'ready';
  const root = document.getElementById('toggles');
  root.innerHTML = '';
  Object.entries(state.patches || {}).forEach(([id, record]) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.textContent = id;
    button.setAttribute('aria-pressed', record.Enabled ? 'true' : 'false');
    button.addEventListener('click', async () => {
      const next = !record.Enabled;
      await gogi.toggle(id, next);
      await refresh();
    });
    root.appendChild(button);
  });
}

refresh().catch(error => {
  document.getElementById('status').textContent = error.message;
});
