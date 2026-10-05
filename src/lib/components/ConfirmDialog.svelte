<script>
    /**
     * @type {{
     *   open?: boolean,
     *   title?: string,
     *   message?: string,
     *   yesLabel?: string,
     *   noLabel?: string,
     *   onYes?: () => void,
     *   onNo?: () => void,
     * }}
     */
    let {
        open = false,
        title = "Megerősítés",
        message = "",
        yesLabel = "Igen",
        noLabel = "Mégse",
        onYes = () => {},
        onNo = () => {},
    } = $props();
</script>

<svelte:window
    onkeydown={(e) => {
        if (open && e.key === "Escape") onNo();
    }}
/>

{#if open}
    <div
        class="link-dialog-overlay confirm-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        tabindex="-1"
        onclick={(e) => e.target === e.currentTarget && onNo()}
        onkeydown={(e) => e.key === "Escape" && onNo()}
    >
        <div class="link-dialog confirm-dialog" role="presentation" onclick={(e) => e.stopPropagation()}>
            <h3 id="confirm-dialog-title">{title}</h3>
            <p class="confirm-message">{message}</p>
            <div class="link-dialog-actions">
                <button type="button" class="link-dialog-submit" onclick={onYes}>{yesLabel}</button>
                <button type="button" class="link-dialog-cancel" onclick={onNo}>{noLabel}</button>
            </div>
        </div>
    </div>
{/if}

<style>
    .confirm-overlay {
        z-index: 1100;
    }
    .confirm-dialog {
        width: min(28rem, calc(100vw - 2rem));
        max-width: min(28rem, calc(100vw - 2rem));
        min-width: 0;
        padding: 1.25rem 1.5rem;
    }
    .confirm-message {
        margin: 0;
        white-space: pre-line;
        line-height: 1.45;
    }
</style>
