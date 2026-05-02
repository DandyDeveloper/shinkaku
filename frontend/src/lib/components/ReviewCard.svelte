<script>
  /** @type {{ pattern: string, meaning: string, example_jp: string, example_en: string, notes: string, jlpt_level: string }} */
  export let grammarPoint;

  /** Whether to reveal the full card (example + notes) */
  export let revealed = false;
</script>

<div class="card">
  <div class="card-header">
    {#if grammarPoint.jlpt_level}
      <span class="badge jlpt-{grammarPoint.jlpt_level.toLowerCase()}">{grammarPoint.jlpt_level}</span>
    {/if}
    <h2 class="pattern">{grammarPoint.pattern}</h2>
  </div>

  {#if revealed}
    <div class="card-body">
      <p class="meaning">{grammarPoint.meaning}</p>

      {#if grammarPoint.example_jp}
        <div class="example">
          <p class="example-jp">{grammarPoint.example_jp}</p>
          {#if grammarPoint.example_en}
            <p class="example-en">{grammarPoint.example_en}</p>
          {/if}
        </div>
      {/if}

      {#if grammarPoint.notes}
        <p class="notes">📝 {grammarPoint.notes}</p>
      {/if}
    </div>
  {:else}
    <p class="hint">Think about the meaning, then flip...</p>
  {/if}
</div>

<style>
  .card {
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 2rem;
    max-width: 640px;
    margin: 0 auto;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  }
  .card-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1.25rem;
  }
  .pattern {
    font-size: 2rem;
    font-weight: 700;
    margin: 0;
    color: #1a202c;
  }
  .badge {
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    text-transform: uppercase;
    white-space: nowrap;
  }
  .jlpt-n5 { background: #c6f6d5; color: #276749; }
  .jlpt-n4 { background: #bee3f8; color: #2c5282; }
  .jlpt-n3 { background: #feebc8; color: #7b341e; }
  .jlpt-n2 { background: #fed7e2; color: #702459; }
  .jlpt-n1 { background: #e9d8fd; color: #44337a; }
  .meaning {
    font-size: 1.2rem;
    color: #4a5568;
    margin-bottom: 1rem;
  }
  .example {
    background: #f7fafc;
    border-left: 3px solid #667eea;
    padding: 0.75rem 1rem;
    border-radius: 0 8px 8px 0;
    margin-bottom: 1rem;
  }
  .example-jp {
    font-size: 1.1rem;
    margin: 0 0 0.25rem;
    color: #2d3748;
  }
  .example-en {
    font-size: 0.9rem;
    color: #718096;
    margin: 0;
  }
  .notes {
    font-size: 0.875rem;
    color: #718096;
    font-style: italic;
  }
  .hint {
    color: #a0aec0;
    font-style: italic;
    text-align: center;
    padding: 1rem 0;
  }
</style>
