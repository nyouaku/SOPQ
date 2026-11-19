<script>
  export let isOpen = false;
  export let totalScanned = 0;
  export let totalProducts = 0;
  export let matched = 0;
  export let unmatched = 0;
  export let startedAt = new Date();
  export let onClose = () => {};
  export let onReport = () => {};

  $: duration = Date.now() - startedAt.getTime();
  $: minutes = Math.floor(duration / 60000);
  $: seconds = Math.floor((duration % 60000) / 1000);
  $: durationText = minutes > 0 ? `${minutes}m ${seconds}s` : `${seconds}s`;

  // Esc menutup dialog. Dipasang di window, bukan di overlay: overlay tidak
  // pernah menerima fokus sendiri saat dibuka, jadi handler di sana tak pernah jalan.
  /** @param {KeyboardEvent} e */
  function onKeydown(e) {
    if (isOpen && e.key === 'Escape') onClose();
  }
</script>

<svelte:window on:keydown={onKeydown} />


{#if isOpen}
  <div
    class="overlay"
    role="dialog"
    aria-modal="true"
    aria-label="Ringkasan sesi selesai"
    tabindex="-1"
  >
    <!-- Backdrop transparan: klik di luar dialog menutupnya. Elemen <button>
         (bukan div) supaya bisa difokus & diaktifkan dengan keyboard. -->
    <button
      type="button"
      class="overlay__backdrop"
      aria-label="Tutup ringkasan"
      on:click={onClose}
    ></button>
    <div class="dialog bg-base-100" role="presentation">
      <div class="icon bg-success text-success-content">✓</div>
      <h2 class="text-base-content">Sesi Selesai</h2>

      <div class="summary">
        <div class="card bg-info/15 border-info/30">
          <div class="value text-base-content">{totalScanned}/{totalProducts}</div>
          <div class="label text-base-content/60">Total</div>
        </div>
        <div class="card bg-success/15 border-success/30">
          <div class="value text-base-content">{matched}</div>
          <div class="label text-base-content/60">Cocok</div>
        </div>
        <div class="card bg-warning/15 border-warning/30">
          <div class="value text-base-content">{unmatched}</div>
          <div class="label text-base-content/60">Beda</div>
        </div>
      </div>

      <div class="duration bg-base-200 text-base-content">
        <span>🕐 Durasi: {durationText}</span>
      </div>

      <div class="actions">
        <button class="btn-ghost bg-base-100 text-base-content border border-base-300" on:click={onClose}>Close</button>
        <button class="btn-primary bg-primary text-primary-content" on:click={onReport}>Lihat Report</button>
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

  /* Backdrop penuh menutupi overlay; dialog di atasnya. */
  .overlay__backdrop {
    position: absolute;
    inset: 0;
    border: 0;
    padding: 0;
    background: transparent;
    cursor: default;
  }

  .dialog {
    position: relative; /* di atas backdrop */
    z-index: 1;
    border-radius: 20px;
    padding: 32px;
    max-width: 360px;
    text-align: center;
  }

  .icon {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 48px;
    margin: 0 auto 20px;
  }

  h2 {
    font-size: 22px;
    font-weight: bold;
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

  .value {
    font-size: 20px;
    font-weight: bold;
  }

  .label {
    font-size: 11px;
    margin-top: 4px;
  }

  .duration {
    padding: 8px 12px;
    border-radius: 8px;
    font-size: 12px;
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

  .btn-ghost:hover {
    background: var(--fallback-b2, oklch(var(--b2)));
  }

  .btn-primary {
    border: none;
  }
</style>
