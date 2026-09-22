<script>
    let { nearby = [], related = [], currentLocationSlug = "" } = $props();
    let here = $derived(String(currentLocationSlug ?? "").trim());
    let showNearby = $derived(Array.isArray(nearby) && nearby.length > 0);
    let showRelated = $derived(Array.isArray(related) && related.length > 0);
</script>

{#if showNearby || showRelated}
    <div class="entry-related">
        {#if showNearby}
            <section class="entry-related__col" aria-labelledby="entry-nearby-title">
                <h2 id="entry-nearby-title" class="entry-related__title">Közeli helyek</h2>
                <ul class="entry-related__list">
                    {#each nearby as row (row.slug || row.id)}
                        <li>
                            <a href="/bejegyzes/{row.slug}">{row.name}</a>
                            {#if here && row.location_slug && row.location_slug !== here && row.location}
                                <span class="entry-related__place">{row.location}</span>
                            {/if}
                        </li>
                    {/each}
                </ul>
            </section>
        {/if}
        {#if showRelated}
            <section class="entry-related__col" aria-labelledby="entry-related-title">
                <h2 id="entry-related-title" class="entry-related__title">Kapcsolódó bejegyzések</h2>
                <ul class="entry-related__list">
                    {#each related as row (row.slug || row.id)}
                        <li>
                            <a href="/bejegyzes/{row.slug}">{row.name}</a>
                        </li>
                    {/each}
                </ul>
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
