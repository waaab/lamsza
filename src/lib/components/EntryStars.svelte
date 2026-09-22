<script>
    const STAR_SLOTS = [1, 2, 3, 4, 5];

    /** @type {{ rating?: number, size?: number }} */
    let { rating = 0, size = 16 } = $props();

    let clamped = $derived(
        Number.isFinite(Number(rating))
            ? Math.min(5, Math.max(0, Number(rating)))
            : 0,
    );
</script>

<span class="entry-stars" aria-hidden="true">
    {#each STAR_SLOTS as slot (slot)}
        {@const filled = clamped >= slot}
        {@const half = !filled && clamped >= slot - 0.5}
        <svg
            class={[
                "entry-stars__star",
                filled && "entry-stars__star--filled",
                half && "entry-stars__star--half",
            ]}
            viewBox="0 0 24 24"
            width={size}
            height={size}
        >
            <path
                d="M12 3.2 14.6 9l6.4.6-4.8 4.1 1.5 6.2L12 16.7 6.3 19.9 7.8 13.7 3 9.6 9.4 9z"
            />
        </svg>
    {/each}
</span>

<style>
    .entry-stars {
        display: inline-flex;
        align-items: center;
        gap: 0.08rem;
    }
    .entry-stars__star {
        fill: none;
        stroke: var(--text-faint);
        stroke-width: 1.6;
        flex-shrink: 0;
    }
    .entry-stars__star--filled,
    .entry-stars__star--half {
        stroke: var(--warning-orange);
    }
    .entry-stars__star--filled {
        fill: var(--warning-orange);
    }
    .entry-stars__star--half {
        fill: color-mix(in srgb, var(--warning-orange) 45%, transparent);
    }
</style>
