<script>
    /** @type {{ type: string, id: number, active?: boolean, ontoggle?: () => void }} */
    let { type, id, active = false, ontoggle = () => {} } = $props();
</script>

<div class="favorite-button">
    <button
        type="button"
        class="btn btn-lg btn-favorite"
        aria-pressed={active}
        aria-label={active ? "Kedvenc eltávolítása" : "Kedvenchez adás"}
        onclick={() => ontoggle()}
    >
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
            ><path
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.27 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.77-3.4 6.86-8.55 11.54L12 21.35z"
            ></path></svg
        >
    </button>
</div>
