<script>
  export let message = 'Data tersimpan';
  export let scannedCount = 0;
  export let totalProducts = 0;

  let visible = true;
  let opacity = 0;

  $: if (visible) {
    opacity = 1;
    setTimeout(() => {
      visible = false;
    }, 2000);
  }
</script>

<div class="notification" style="opacity: {opacity}">
  <div class="content">
    <div class="icon">✓</div>
    <div class="text">
      <div class="title">{message}</div>
      <div class="subtitle">Scan {scannedCount}/{totalProducts}</div>
    </div>
  </div>
</div>

<style>
  .notification {
    position: fixed;
    top: 24px;
    left: 24px;
    background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
    border-radius: 12px;
    padding: 16px;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
    z-index: 9999;
    transition: opacity 0.3s ease-out;
    max-width: 320px;
  }

  .content {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }

  .icon {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.3);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    flex-shrink: 0;
  }

  .text {
    flex: 1;
  }

  .title {
    color: white;
    font-weight: 600;
    font-size: 13px;
  }

  .subtitle {
    color: rgba(255, 255, 255, 0.85);
    font-size: 11px;
    margin-top: 2px;
  }
</style>
