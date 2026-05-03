<script>
  import { onMount } from 'svelte';
  import { listGrammar, createGrammar } from '$lib/api.js';

  const JLPT_LEVELS = ['', 'N5', 'N4', 'N3', 'N2', 'N1'];

  let points = [];
  let filtered = [];
  let filterLevel = '';
  let search = '';
  let loading = true;
  let error = '';

  // New grammar form
  let showForm = false;
  let saving = false;
  let saveError = '';
  let form = emptyForm();

  function emptyForm() {
    return { pattern: '', meaning: '', jlpt_level: '', example_jp: '', example_en: '', notes: '' };
  }

  onMount(load);

  async function load() {
    loading = true;
    try {
      points = await listGrammar();
      applyFilter();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function applyFilter() {
    filtered = points.filter(p => {
      const matchLevel = !filterLevel || p.jlpt_level === filterLevel;
      const q = search.toLowerCase();
      const matchSearch = !q ||
        p.pattern.toLowerCase().includes(q) ||
        p.meaning.toLowerCase().includes(q);
      return matchLevel && matchSearch;
    });
  }

  $: { filterLevel; search; if (points.length) applyFilter(); }

  async function handleCreate() {
    if (!form.pattern || !form.meaning) { saveError = 'Pattern and meaning required.'; return; }
    saving = true;
    saveError = '';
    try {
      await createGrammar({ ...form, source: 'manual' });
      form = emptyForm();
      showForm = false;
      await load();
    } catch (e) {
      saveError = e.message;
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head><title>Grammar — 日本語先生</title></svelte:head>

<main>
  <a href="/" class="back">← Dashboard</a>
  <div class="page-header">
    <h2>Grammar Points</h2>
    <button class="btn primary" on:click={() => { showForm = !showForm; saveError = ''; }}>
      {showForm ? 'Cancel' : '+ Add'}
    </button>
  </div>

  {#if showForm}
    <form class="grammar-form" on:submit|preventDefault={handleCreate}>
      <div class="form-row">
        <label>Pattern *
          <input bind:value={form.pattern} placeholder="〜てもいい" required />
        </label>
        <label>JLPT Level
          <select bind:value={form.jlpt_level}>
            {#each JLPT_LEVELS as l}
              <option value={l}>{l || '—'}</option>
            {/each}
          </select>
        </label>
      </div>
      <label>Meaning *
        <input bind:value={form.meaning} placeholder="It is okay to…" required />
      </label>
      <div class="form-row">
        <label>Example (Japanese)
          <input bind:value={form.example_jp} placeholder="ここで食べてもいいですか。" />
        </label>
        <label>Example (English)
          <input bind:value={form.example_en} placeholder="May I eat here?" />
        </label>
      </div>
      <label>Notes
        <input bind:value={form.notes} placeholder="Optional cautions or nuances" />
      </label>
      {#if saveError}<p class="form-error">{saveError}</p>{/if}
      <button class="btn primary" type="submit" disabled={saving}>
        {saving ? 'Saving…' : 'Save Grammar Point'}
      </button>
    </form>
  {/if}

  <div class="filters">
    <input class="search-input" bind:value={search} placeholder="Search patterns, meanings…" />
    <div class="level-tabs">
      {#each JLPT_LEVELS as l}
        <button
          class="level-tab"
          class:active={filterLevel === l}
          on:click={() => filterLevel = l}
        >{l || 'All'}</button>
      {/each}
    </div>
  </div>

  {#if loading}
    <p class="state-msg">Loading…</p>
  {:else if error}
    <div class="alert error">{error}</div>
  {:else if filtered.length === 0}
    <p class="state-msg">No grammar points found.</p>
  {:else}
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Pattern</th>
            <th>Meaning</th>
            <th>JLPT</th>
            <th>Example</th>
            <th>Practice</th>
          </tr>
        </thead>
        <tbody>
          {#each filtered as gp}
            <tr>
              <td class="pattern-cell">{gp.pattern}</td>
              <td>{gp.meaning}</td>
              <td><span class="badge">{gp.jlpt_level || '—'}</span></td>
              <td class="example-cell">
                {#if gp.example_jp}
                  <span>{gp.example_jp}</span>
                  {#if gp.example_en}<br/><small>{gp.example_en}</small>{/if}
                {:else}
                  <span class="muted">—</span>
                {/if}
              </td>
              <td class="actions-cell">
                <a class="table-link" href={`/challenge?grammar=${gp.id}&mode=conversation`}>Conversation</a>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    <p class="count">{filtered.length} of {points.length} grammar point{points.length !== 1 ? 's' : ''}</p>
  {/if}
</main>

<style>
  main {
    max-width: 900px;
    margin: 0 auto;
    padding: 2rem 1rem;
    font-family: system-ui, -apple-system, sans-serif;
  }
  .back { color: #667eea; text-decoration: none; font-size: 0.9rem; }
  .page-header { display: flex; justify-content: space-between; align-items: center; margin: 0.5rem 0 1.5rem; }
  h2 { margin: 0; color: #2d3748; }

  .grammar-form {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
  .grammar-form label { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.85rem; color: #4a5568; font-weight: 600; }
  .grammar-form input, .grammar-form select {
    padding: 0.5rem 0.75rem;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    font-size: 0.95rem;
    font-family: inherit;
    font-weight: 400;
  }
  .grammar-form input:focus, .grammar-form select:focus { outline: none; border-color: #667eea; }
  .form-error { color: #c53030; font-size: 0.85rem; }

  .filters { display: flex; gap: 0.75rem; align-items: center; margin-bottom: 1rem; flex-wrap: wrap; }
  .search-input {
    flex: 1;
    min-width: 200px;
    padding: 0.5rem 0.75rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 0.9rem;
  }
  .search-input:focus { outline: none; border-color: #667eea; }
  .level-tabs { display: flex; gap: 0.25rem; }
  .level-tab {
    padding: 0.4rem 0.7rem;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    background: white;
    font-size: 0.8rem;
    cursor: pointer;
    font-weight: 600;
    color: #718096;
  }
  .level-tab.active { background: #667eea; color: white; border-color: #667eea; }

  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
  th { text-align: left; padding: 0.6rem 0.75rem; background: #f7fafc; color: #718096; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.04em; border-bottom: 1px solid #e2e8f0; }
  td { padding: 0.65rem 0.75rem; border-bottom: 1px solid #f0f4f8; vertical-align: top; }
  tr:hover td { background: #f7fafc; }
  .pattern-cell { font-weight: 700; color: #2d3748; white-space: nowrap; }
  .example-cell small { color: #a0aec0; }
  .actions-cell { white-space: nowrap; }
  .table-link {
    display: inline-flex;
    align-items: center;
    padding: 0.35rem 0.6rem;
    border-radius: 999px;
    text-decoration: none;
    font-size: 0.8rem;
    font-weight: 700;
    background: #edf2ff;
    color: #4055c8;
  }
  .table-link:hover { background: #e0e7ff; }
  .badge { background: #ebf8ff; color: #2b6cb0; font-size: 0.75rem; font-weight: 700; padding: 0.15rem 0.4rem; border-radius: 4px; }
  .muted { color: #cbd5e0; }
  .count { font-size: 0.8rem; color: #a0aec0; text-align: right; margin-top: 0.5rem; }

  .btn { padding: 0.6rem 1.1rem; border-radius: 8px; font-size: 0.9rem; font-weight: 600; cursor: pointer; border: 1px solid #e2e8f0; background: #f7fafc; color: #2d3748; }
  .btn.primary { background: #667eea; color: white; border-color: #667eea; }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .state-msg { text-align: center; color: #a0aec0; padding: 3rem 0; }
  .alert { border-radius: 8px; padding: 1rem; }
  .alert.error { background: #fff5f5; border: 1px solid #fc8181; color: #c53030; }
</style>
