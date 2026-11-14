<script>
  export let isOpen = false;
  export let totalScanned = 0;
  export let totalProducts = 0;
  export let matched = 0;
  export let unmatched = 0;
  export let startedAt = new Date();
  export let onClose = () => {};
  export let onReport = () => {};

  $: duration = new Date() - startedAt;
  $: minutes = Math.floor(duration / 60000);
  $: seconds = Math.floor((duration % 60000) / 1000);
  $: durationText = minutes > 0 ? `${minutes}m ${seconds}s` : `${seconds}s`;
</script>

{#if isOpen}
  <div class="overlay" on:click={onClose}>
    <div class="dialog" on:click|stopPropagation>
      <div class="icon">✓</div>
      <h2>Sesi Selesai</h2>

      <div class="summary">
        <div class="card blue">
          <div class="value">{totalScanned}/{totalProducts}</div>
          <div class="label">Total</div>
        </div>
        <div class="card green">
          <div class="value">{matched}</div>
          <div class="label">Cocok</div>
        </div>
        <div class="card orange">
          <div class="value">{unmatched}</div>
          <div class="label">Beda</div>
        </div>
      </div>

      <div class="duration">
        <span>🕐 Durasi: {durationText}</span>
      </div>

      <div class="actions">
        <button class="btn-ghost" on:click={onClose}>Close</button>
        <button class="btn-primary" on:click={onReport}>Lihat Report</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
  }

  .dialog {
    background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
    border-radius: 20px;
    padding: 32px;
    max-width: 360px;
    text-align: center;
  }

  .icon {
    width: 80px;
    height: 80px;
    background: #10b981;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 48px;
    color: white;
    margin: 0 auto 20px;
  }

  h2 {
    font-size: 22px;
    font-weight: bold;
    color: #1f2937;
    margin: 0 0 24px;
  }

  .summary {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    margin-bottom: 16px;
  }

  .card {
    padding: 12px;
    border-radius: 12px;
    border: 2px solid;
  }

  .card.blue {
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.3);
  }

  .card.green {
    background: rgba(16, 185, 129, 0.15);
    border-color: rgba(16, 185, 129, 0.3);
  }

  .card.orange {
    background: rgba(249, 115, 22, 0.15);
    border-color: rgba(249, 115, 22, 0.3);
  }

  .value {
    font-size: 20px;
    font-weight: bold;
    color: #1f2937;
  }

  .label {
    font-size: 11px;
    color: #6b7280;
    margin-top: 4px;
  }

  .duration {
    background: #e5e7eb;
    padding: 8px 12px;
    border-radius: 8px;
    font-size: 12px;
    color: #4b5563;
    margin-bottom: 24px;
  }

  .actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  button {
    padding: 12px;
    border-radius: 10px;
    font-weight: 600;
    font-size: 14px;
    cursor: pointer;
    border: none;
    transition: all 0.2s;
  }

  .btn-ghost {
    background: white;
    color: #4b5563;
    border: 1px solid #d1d5db;
  }

  .btn-ghost:hover {
    background: #f3f4f6;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover {
    background: #2563eb;
  }
</style>
