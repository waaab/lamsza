<script>
    let { nearby = [], related = [], history = [], currentLocationSlug = "" } = $props();
    let here = $derived(String(currentLocationSlug ?? "").trim());
    let showNearby = $derived(Array.isArray(nearby) && nearby.length > 0);
    let showRelated = $derived(Array.isArray(related) && related.length > 0);
    let showHistory = $derived(Array.isArray(history) && history.length > 0);

    /** @param {Record<string, unknown>} row */
    function placeLabel(row) {
        const place = String(row?.location ?? "").trim();
        if (!place) return "";
        const slug = String(row?.location_slug ?? "").trim();
        if (here && slug && slug === here) return "";
        return place;
    }
</script>

{#snippet linkList(rows)}
    <ul class="entry-related__list">
        {#each rows as row (row.slug || row.id)}
            <li>
                <a href="/bejegyzes/{row.slug}">{row.name}</a>
                {#if placeLabel(row)}
                    <span class="entry-related__place">{placeLabel(row)}</span>
                {/if}
            </li>
        {/each}
    </ul>
{/snippet}

{#if showNearby || showHistory || showRelated}
    <div class="entry-related">
        {#if showNearby}
            <section class="entry-related__col" aria-labelledby="entry-nearby-title">
                <h2 id="entry-nearby-title" class="entry-related__title">A közelben...</h2>
                {@render linkList(nearby)}
            </section>
        {/if}
        {#if showHistory}
            <section class="entry-related__col" aria-labelledby="entry-history-title">
                <h2 id="entry-history-title" class="entry-related__title">Böngészési előzmények</h2>
                {@render linkList(history)}
            </section>
        {/if}
        {#if showRelated}
            <section class="entry-related__col" aria-labelledby="entry-related-title">
                <h2 id="entry-related-title" class="entry-related__title">Kapcsolódó bejegyzések</h2>
                {@render linkList(related)}
            </section>
        {/if}
    </div>
{/if}

<style>
    .entry-related {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 1.5rem 2rem;
        padding-top: 0.5rem;
        border-top: 1px solid var(--border-color);
    }
    @media (max-width: 991px) {
        .entry-related {
            grid-template-columns: 1fr;
        }
    }
    .entry-related__title {
        margin: 0 0 0.65rem;
        font-size: 1.05rem;
    }
    .entry-related__list {
        list-style: none;
        margin: 0;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }
    .entry-related__list a {
        color: var(--szekely-blue, #1d4ed8);
        font-weight: 600;
        text-decoration: none;
    }
    .entry-related__list a:hover {
        text-decoration: underline;
    }
    .entry-related__place {
        margin-left: 0.35rem;
        color: var(--text-muted);
        font-weight: 400;
    }
</style>
