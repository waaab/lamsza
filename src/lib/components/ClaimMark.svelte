<script>
    let { claimed = false, showLabel = false } = $props();
</script>

<span
    class="claim-mark"
    class:claim-mark--on={claimed}
    aria-label={claimed ? "Átvéve" : "Gazdátlan"}
>
    <svg
        viewBox="0 0 24 24"
        width="14"
        height="14"
        fill="none"
        stroke="currentColor"
        stroke-width="2.25"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
    >
        <circle cx="12" cy="12" r="9" />
        <path d="M9 12.5 11.2 14.7 15.5 9.8" />
    </svg>
    {#if showLabel}
        <span>{claimed ? "Átvéve" : "Gazdátlan"}</span>
    {/if}
</span>

<style>
    .claim-mark {
        display: inline-flex;
        align-items: center;
        gap: 0.25rem;
        margin-left: 0.45rem;
        font-size: var(--text-sm);
        font-weight: 700;
        line-height: 1;
        color: var(--text-faint);
        opacity: 0.38;
        vertical-align: middle;
    }

    .claim-mark--on {
        color: var(--szekely-blue);
        opacity: 1;
    }
</style>
