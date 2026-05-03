<script>
  import { onMount } from 'svelte';
  import { listVocab, createVocab } from '$lib/api.js';

  const JLPT_LEVELS = ['', 'N5', 'N4', 'N3', 'N2', 'N1'];

  let words = [];
  let filtered = [];
  let filterLevel = '';
  let search = '';
  let loading = true;
  let error = '';

  // New vocab form
  let showForm = false;
  let saving = false;
  let saveError = '';
  let form = emptyForm();

  function emptyForm() {
    return {
      word: '',
      reading: '',
      meaning: '',
      part_of_speech: '',
      jlpt_level: '',
      example_jp: '',
      example_en: '',
      notes: ''
    };
  }

  onMount(load);

  async function load() {
    loading = true;
    try {
      words = await listVocab();
      applyFilter();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function applyFilter() {
    filtered = words.filter(v => {
      const matchLevel = !filterLevel || v.jlpt_level === filterLevel;
      const q = search.toLowerCase();
      const matchSearch = !q ||
        v.word.toLowerCase().includes(q) ||
        v.reading.toLowerCase().includes(q) ||
        v.meaning.toLowerCase().includes(q);
      return matchLevel && matchSearch;
    });
  }

  $: { filterLevel; search; if (words.length) applyFilter(); }

  async function handleCreate() {
    if (!form.word || !form.meaning) { saveError = 'Word and meaning are required.'; return; }
    saving = true;
    saveError = '';
    try {
      await createVocab({ ...form, source: 'manual' });
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

<svelte:head><title>Vocabulary — 日本語先生</title></svelte:head>

<main>
  <a href="/" class="back">← Dashboard</a>
  <div class="page-header">
    <h2>Vocabulary</h2>
    <button class="btn primary" on:click={() => { showForm = !showForm; saveError = ''; }}>
      {showForm ? 'Cancel' : '+ Add'}
    </button>
  </div>

  {#if showForm}
    <form class="vocab-form" on:submit|preventDefault={handleCreate}>
      <div class="grid two">
        <label>
          Word
          <input bind:value={form.word} placeholder="食べる" required />
        </label>
        <label>
          Reading
          <input bind:value={form.reading} placeholder="たべる" />
        </label>
      </div>

      <div class="grid two">
        <label>
          Meaning
          <input bind:value={form.meaning} placeholder="to eat" required />
        </label>
        <label>
          Part of speech
          <input bind:value={form.part_of_speech} placeholder="verb (ichidan)" />
        </label>
      </div>

      <label>
        JLPT level
        <select bind:value={form.jlpt_level}>
          {#each JLPT_LEVELS as lvl}
            <option value={lvl}>{lvl || 'Unspecified'}</option>
          {/each}
        </select>
      </label>

      <label>
        Example (JP)
        <input bind:value={form.example_jp} placeholder="毎朝パンを食べる。" />
      </label>
      <label>
        Example (EN)
        <input bind:value={form.example_en} placeholder="I eat bread every morning." />
      </label>
      <label>
        Notes
        <textarea bind:value={form.notes} rows="3" placeholder="Conjugation notes, collocations, etc."></textarea>
      </label>

      {#if saveError}<div class="alert error">{saveError}</div>{/if}

      <div class="actions">
        <button class="btn" type="button" on:click={() => { showForm = false; saveError = ''; }}>Cancel</button>
        <button class="btn primary" type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</button>
      </div>
    </form>
  {/if}

  <section class="filters">
    <label>
      JLPT
      <select bind:value={filterLevel}>
        {#each JLPT_LEVELS as lvl}
          <option value={lvl}>{lvl || 'All'}</option>
        {/each}
      </select>
    </label>

    <label class="search">
      Search
      <input bind:value={search} placeholder="Search by word, reading, meaning..." />
    </label>

    <a href="/vocab/review" class="btn primary review-link">Start Vocab Review</a>
  </section>

  {#if loading}
    <p class="state-msg">Loading vocabulary…</p>
  {:else if error}
    <div class="alert error">{error}</div>
  {:else if filtered.length === 0}
    <p class="state-msg">No vocabulary words found.</p>
  {:else}
    <div class="list">
      {#each filtered as v}
        <article class="card">
          <div class="head">
            <h3>{v.word}</h3>
            <span class="pill">{v.jlpt_level || '—'}</span>
          </div>
          <p class="reading">{v.reading || 'No reading'}</p>
          <p class="meaning">{v.meaning}</p>
          {#if v.part_of_speech}<p class="meta"><strong>POS:</strong> {v.part_of_speech}</p>{/if}
          {#if v.example_jp}<p class="meta"><strong>JP:</strong> {v.example_jp}</p>{/if}
          {#if v.example_en}<p class="meta"><strong>EN:</strong> {v.example_en}</p>{/if}
          {#if v.notes}<p class="meta"><strong>Notes:</strong> {v.notes}</p>{/if}
        </article>
      {/each}
    </div>
  {/if}
</main>

<style>
  main {
    max-width: 820px;
    margin: 0 auto;
    padding: 2rem 1rem;
    font-family: system-ui, -apple-system, sans-serif;
  }
  .back { color: #667eea; text-decoration: none; font-size: 0.9rem; }
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: 0.75rem 0 1rem;
  }
  h2 { margin: 0; color: #2d3748; }

  .filters {
    display: grid;
    grid-template-columns: 180px 1fr auto;
    gap: 0.75rem;
    align-items: end;
    margin-bottom: 1rem;
  }
  .filters .review-link {
    white-space: nowrap;
    text-decoration: none;
    text-align: center;
  }

  label { display: grid; gap: 0.35rem; font-size: 0.9rem; color: #4a5568; }
  input, select, textarea {
    width: 100%;
    border: 1px solid #d1d9e6;
    border-radius: 8px;
    padding: 0.6rem 0.7rem;
    font-size: 0.95rem;
    background: #fff;
  }

  .vocab-form {
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1rem;
    margin-bottom: 1rem;
    display: grid;
    gap: 0.8rem;
  }
  .grid { display: grid; gap: 0.8rem; }
  .grid.two { grid-template-columns: repeat(2, minmax(0, 1fr)); }

  .actions { display: flex; gap: 0.6rem; justify-content: flex-end; }
  .btn {
    padding: 0.55rem 0.85rem;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    background: #fff;
    cursor: pointer;
  }
  .btn.primary {
    border-color: #667eea;
    background: #667eea;
    color: #fff;
  }

  .list { display: grid; gap: 0.85rem; }
  .card {
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 0.95rem;
  }
  .head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.7rem;
  }
  .head h3 { margin: 0; color: #1a202c; }
  .pill {
    padding: 0.2rem 0.55rem;
    border-radius: 999px;
    background: #ebf4ff;
    color: #3b5bdb;
    font-size: 0.8rem;
    font-weight: 700;
  }
  .reading {
    margin: 0.2rem 0 0.45rem;
    color: #4a5568;
  }
  .meaning {
    margin: 0 0 0.5rem;
    color: #2d3748;
    font-weight: 600;
  }
  .meta {
    margin: 0.2rem 0;
    color: #4a5568;
    font-size: 0.92rem;
  }

  .state-msg { color: #718096; }
  .alert.error {
    background: #fff5f5;
    border: 1px solid #fed7d7;
    color: #c53030;
    border-radius: 10px;
    padding: 0.7rem 0.85rem;
  }

  @media (max-width: 780px) {
    .grid.two { grid-template-columns: 1fr; }
    .filters { grid-template-columns: 1fr; }
  }
</style>
