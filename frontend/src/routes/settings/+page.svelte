<script>
  import { onMount } from 'svelte';
  import { getSettings, updateSettings } from '$lib/settings.js';

  let includeConversationFurigana = false;

  onMount(() => {
    const settings = getSettings();
    includeConversationFurigana = settings.includeConversationFurigana;
  });

  function handleToggle(event) {
    includeConversationFurigana = event.currentTarget.checked;
    updateSettings({ includeConversationFurigana });
  }
</script>

<svelte:head><title>Settings — 日本語先生</title></svelte:head>

<main>
  <a href="/" class="back">← Dashboard</a>
  <h2>Settings</h2>

  <section class="settings-card">
    <label class="toggle-row" for="furigana-toggle">
      <div>
        <strong>Conversation furigana</strong>
        <p>When enabled, conversation prompts and follow-up replies include furigana as Kanji(かな).</p>
      </div>
      <input
        id="furigana-toggle"
        type="checkbox"
        checked={includeConversationFurigana}
        on:change={handleToggle}
      />
    </label>
  </section>
</main>

<style>
  main {
    max-width: 760px;
    margin: 0 auto;
    padding: 2rem 1rem;
    font-family: system-ui, -apple-system, sans-serif;
  }
  .back { color: #667eea; text-decoration: none; font-size: 0.9rem; }
  h2 { margin: 0.5rem 0 1.5rem; color: #2d3748; }

  .settings-card {
    border: 1px solid #e2e8f0;
    background: #fff;
    border-radius: 12px;
    padding: 1rem;
  }
  .toggle-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    cursor: pointer;
  }
  .toggle-row strong {
    display: block;
    color: #2d3748;
    margin-bottom: 0.2rem;
  }
  .toggle-row p {
    margin: 0;
    color: #718096;
    font-size: 0.9rem;
    line-height: 1.4;
  }
  input[type='checkbox'] {
    width: 18px;
    height: 18px;
    margin-top: 0.2rem;
  }
</style>